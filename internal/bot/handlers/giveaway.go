package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/keyboards"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/state"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	// Giveaway state constants.
	giveawayStatePrefix           = "giveaway:"
	giveawayStateWaitGroup        = giveawayStatePrefix + "wait_group"
	giveawayStateWaitPhoto        = giveawayStatePrefix + "wait_photo"
	giveawayStateWaitPublishDate  = giveawayStatePrefix + "wait_publish_date"
	giveawayStateWaitConfirmation = giveawayStatePrefix + "wait_confirmation"

	// Command constants.
	giveawayCommand = "/giveaway"
	cancelCommand   = "/cancel"

	giveawayCallbackGroup   = "giveaway:group:"
	giveawayCallbackConfirm = "giveaway:confirm"
	giveawayCallbackCancel  = "giveaway:cancel"

	// Giveaway data constants.
	giveawayDataGroupID             = "groupID"
	giveawayDataPhotoID             = "photoID"
	giveawayDataDescription         = "description"
	giveawayDataOriginalDescription = "original_description"
	giveawayDataPublishDate         = "publishDate"
	giveawayDataApplicationEndDate  = "applicationEndDate"
	giveawayDataResultsDate         = "resultsDate"
)

// GiveawayScheduler handles giveaway scheduling flow.
type GiveawayScheduler struct {
	BaseHandler

	fsmService *fsm.Service

	usersSvc     *users.Service
	groupsSvc    *groups.Service
	giveawaysSvc *giveaways.Service
}

func NewGiveawayScheduler(
	bot *bot.Bot,
	fsmService *fsm.Service,
	usersSvc *users.Service,
	groupsSvc *groups.Service,
	giveawaysSvc *giveaways.Service,
	logger *zap.Logger,
) Handler {
	return &GiveawayScheduler{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

		fsmService: fsmService,

		usersSvc:     usersSvc,
		groupsSvc:    groupsSvc,
		giveawaysSvc: giveawaysSvc,
	}
}

func (g *GiveawayScheduler) Register(b *bot.Bot) {
	// Register command handler
	isEmptyState := state.NewStateFilter("", g.fsmService, g.logger)
	hasPrefixState := state.NewStatePrefixFilter(giveawayStatePrefix, g.fsmService, g.logger)
	commandFilter := func(command string) bot.MatchFunc {
		return func(update *models.Update) bool {
			if update.Message == nil {
				return false
			}
			return update.Message.Text == command
		}
	}
	combinator := func(filters ...bot.MatchFunc) bot.MatchFunc {
		return func(update *models.Update) bool {
			for _, f := range filters {
				if !f(update) {
					return false
				}
			}
			return true
		}
	}

	b.RegisterHandlerMatchFunc(
		combinator(
			isEmptyState,
			commandFilter(giveawayCommand),
		),
		g.handleGiveawayCommand,
	)

	b.RegisterHandlerMatchFunc(
		combinator(
			hasPrefixState,
			commandFilter(cancelCommand),
		),
		g.handleCancelCommand,
	)

	b.RegisterHandlerMatchFunc(
		combinator(
			state.NewStateFilter(giveawayStateWaitGroup, g.fsmService, g.logger),
			func(update *models.Update) bool {
				return update.CallbackQuery != nil &&
					strings.HasPrefix(update.CallbackQuery.Data, giveawayCallbackGroup)
			},
		),
		g.handleGiveawayGroup,
	)

	b.RegisterHandlerMatchFunc(
		combinator(
			state.NewStateFilter(giveawayStateWaitPhoto, g.fsmService, g.logger),
			func(update *models.Update) bool {
				return update.Message != nil
			},
		),
		g.handlePhotoAndDescription,
	)

	b.RegisterHandlerMatchFunc(
		combinator(
			state.NewStateFilter(giveawayStateWaitPublishDate, g.fsmService, g.logger),
			func(update *models.Update) bool {
				return update.Message != nil
			},
		),
		g.handlePublishDate,
	)

	// Add callback query handlers after the cancel command registration
	b.RegisterHandlerMatchFunc(
		combinator(
			hasPrefixState,
			func(update *models.Update) bool {
				return update.CallbackQuery != nil &&
					update.CallbackQuery.Data == giveawayCallbackConfirm
			},
		),
		g.handleConfirmation,
	)

	b.RegisterHandlerMatchFunc(
		combinator(
			hasPrefixState,
			func(update *models.Update) bool {
				return update.CallbackQuery != nil &&
					update.CallbackQuery.Data == giveawayCallbackCancel
			},
		),
		g.handleCancelCommand,
	)
}

func (g *GiveawayScheduler) handleGiveawayCommand(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		g.withContext(update).Error("invalid update: missing message or sender")
		return
	}

	// Ensure this is not in a group chat
	if update.Message.Chat.Type != models.ChatTypePrivate {
		g.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ Giveaway scheduling is only available in private chats."},
		)
		return
	}

	logger := g.withContext(update)

	// Register or get user
	user, err := g.usersSvc.RegisterUser(ctx, UserToDomain(update.Message.From))
	if err != nil {
		logger.Error("failed to register user", zap.Error(err))
		g.sendReply(ctx, update, &bot.SendMessageParams{Text: "❌ Failed to process user. Please try again."})
		return
	}

	// Check if user is admin of any groups
	adminGroups, err := g.groupsSvc.GetUserAdminGroups(ctx, user.ID)
	if err != nil {
		logger.Error("failed to get user admin groups", zap.Error(err))
		g.sendReply(ctx, update, &bot.SendMessageParams{Text: "❌ Failed to verify admin status. Please try again."})
		return
	}

	if len(adminGroups) == 0 {
		g.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ You must be an admin of a group to schedule giveaways."},
		)
		return
	}

	state, err := state.GetState(ctx)
	if err != nil {
		g.handleError(ctx, update, fmt.Errorf("failed to get state: %w", err))
		return
	}

	// If user is admin of exactly one group, auto-select it
	if len(adminGroups) == 1 {
		state.SetName(giveawayStateWaitPhoto)
		state.AddData(giveawayDataGroupID, strconv.FormatInt(adminGroups[0].ID, 10))

		g.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "📸 Please send a photo with description caption for the giveaway."},
		)
		return
	}

	// Show group selection keyboard
	state.SetName(giveawayStateWaitGroup)
	g.showGroupSelectionKeyboard(ctx, update.Message.Chat.ID, adminGroups)
}

func (g *GiveawayScheduler) showGroupSelectionKeyboard(
	ctx context.Context,
	chatID int64,
	groups []groups.GroupWithSettings,
) {
	markup := keyboards.GroupSelectionKeyboard(
		giveawayCallbackGroup,
		groups,
	)

	g.sendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "👥 Select a group for the giveaway:",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: markup,
		},
	)
}

func (g *GiveawayScheduler) handleGiveawayGroup(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	logger := g.withContext(update)
	state, err := state.GetState(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	// Parse group ID from callback data
	groupIDStr := strings.TrimPrefix(update.CallbackQuery.Data, giveawayCallbackGroup)
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		logger.Error("failed to parse group ID", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	state.SetName(giveawayStateWaitPhoto)
	state.AddData(giveawayDataGroupID, strconv.FormatInt(groupID, 10))

	g.sendReply(
		ctx,
		update,
		&bot.SendMessageParams{Text: "📸 Please send a photo with description caption for the giveaway."},
	)
}

func (g *GiveawayScheduler) handlePhotoAndDescription(ctx context.Context, _ *bot.Bot, update *models.Update) {
	logger := g.withContext(update)

	state, err := state.GetState(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	if len(update.Message.Photo) == 0 || update.Message.Caption == "" {
		g.sendReply(ctx, update, &bot.SendMessageParams{Text: "❌ Please send a photo with description caption."})
		return
	}

	photo := lo.MaxBy(
		update.Message.Photo,
		func(a, b models.PhotoSize) bool {
			return a.FileSize > b.FileSize
		},
	)

	state.SetName(giveawayStateWaitPublishDate)
	state.AddData(giveawayDataPhotoID, photo.FileID)
	state.AddData(giveawayDataOriginalDescription, update.Message.Caption)

	// Request start time
	g.sendReply(
		ctx,
		update,
		&bot.SendMessageParams{
			Text: "⏰ Please specify the start time in format: YYYY-MM-DD HH:MM (e.g., 2023-12-25 14:30)",
		},
	)
}

func (g *GiveawayScheduler) handlePublishDate(ctx context.Context, _ *bot.Bot, update *models.Update) {
	logger := g.withContext(update)

	state, err := state.GetState(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	if update.Message == nil || update.Message.Text == "" {
		g.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{
				Text: "⏰ Please specify the start time in format: YYYY-MM-DD HH:MM (e.g., 2023-12-25 14:30)",
			},
		)
		return
	}

	startTime, err := parseDateTime(update.Message.Text)
	if err != nil {
		g.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{
				Text: "⏰ Please specify the start time in format: YYYY-MM-DD HH:MM (e.g., 2023-12-25 14:30)",
			},
		)
		return
	}

	const duration = 24 * time.Hour
	const resultsDuration = 26 * time.Hour

	state.SetName(giveawayStateWaitConfirmation)
	state.AddData(giveawayDataPublishDate, formatDateTime(startTime))
	state.AddData(giveawayDataApplicationEndDate, formatDateTime(startTime.Add(duration)))
	state.AddData(giveawayDataResultsDate, formatDateTime(startTime.Add(resultsDuration)))

	g.showPreviewAndConfirmation(ctx, update.Message.Chat.ID, state)
}

func (g *GiveawayScheduler) showPreviewAndConfirmation(ctx context.Context, chatID int64, state *fsm.State) {
	group, settings, err := g.loadGroupAndSettings(ctx)
	if err != nil {
		g.sendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Failed to load group and settings. Please try again.",
		})
		return
	}

	state.AddData(giveawayDataDescription, state.GetData(giveawayDataOriginalDescription))
	if settings.LLMDescription {
		photo, downErr := g.downloadPhoto(ctx, state.GetData(giveawayDataPhotoID))
		if downErr != nil {
			g.logger.Error("failed to download photo", zap.Error(downErr))
			g.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Failed to download photo. Please try again.",
			})
			return
		}

		publishDate, downErr := parseDateTime(state.GetData(giveawayDataPublishDate))
		if downErr != nil {
			g.logger.Error("failed to parse publish date", zap.Error(downErr))
			g.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Failed to parse publish date. Please try again.",
			})
			return
		}

		description, downErr := g.giveawaysSvc.GenerateDescription(
			ctx,
			state.GetData(giveawayDataOriginalDescription),
			publishDate,
			photo,
		)
		if downErr != nil {
			g.logger.Error("failed to generate description", zap.Error(downErr))
			g.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Failed to generate description. Please try again.",
			})
			return
		}

		state.AddData(giveawayDataDescription, description)
	}

	previewText := fmt.Sprintf(`🎯 *Preview*

📱 Group: %s
📝 Description: %s
⏰ Start time: %s
📝 Application end: %s
🎉 Results: %s`,
		bot.EscapeMarkdown(group.Title),
		bot.EscapeMarkdown(state.GetData(giveawayDataDescription)),
		bot.EscapeMarkdown(state.GetData(giveawayDataPublishDate)),
		bot.EscapeMarkdown(state.GetData(giveawayDataApplicationEndDate)),
		bot.EscapeMarkdown(state.GetData(giveawayDataResultsDate)),
	)

	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "✅ Confirm",
					CallbackData: giveawayCallbackConfirm,
				},
				{
					Text:         "❌ Cancel",
					CallbackData: giveawayCallbackCancel,
				},
			},
		},
	}

	_, err = g.bot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      chatID,
		Photo:       &models.InputFileString{Data: state.GetData(giveawayDataPhotoID)},
		Caption:     previewText,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: markup,
	})
	if err != nil {
		g.logger.Error("failed to send message with keyboard", zap.Error(err))
	}
}

func (g *GiveawayScheduler) handleConfirmation(ctx context.Context, _ *bot.Bot, update *models.Update) {
	logger := g.withContext(update)

	user, err := g.usersSvc.RegisterUser(ctx, UserToDomain(&update.CallbackQuery.From))
	if err != nil {
		logger.Error("failed to register user", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	state, err := state.GetState(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	groupID, err := strconv.ParseInt(state.GetData(giveawayDataGroupID), 10, 64)
	if err != nil {
		logger.Error("failed to parse group ID", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	if ok, adminErr := g.groupsSvc.IsAdmin(ctx, groupID, user.ID); adminErr != nil {
		g.handleError(ctx, update, fmt.Errorf("failed to check if user is group admin: %w", adminErr))
		return
	} else if !ok {
		logger.Error("user is not group admin", zap.Int64("group_id", groupID), zap.Int64("user_id", user.ID))
		g.sendReply(ctx, update, &bot.SendMessageParams{Text: "❌ You are not group admin."})
		return
	}

	publishDate, err := parseDateTime(state.GetData(giveawayDataPublishDate))
	if err != nil {
		logger.Error("failed to parse publish date", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	applicationEndDate, err := parseDateTime(state.GetData(giveawayDataApplicationEndDate))
	if err != nil {
		logger.Error("failed to parse application end date", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}
	resultsDate, err := parseDateTime(state.GetData(giveawayDataResultsDate))
	if err != nil {
		logger.Error("failed to parse results date", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	if createErr := g.giveawaysSvc.Create(ctx, giveaways.GiveawayPrepared{
		GiveawayDraft: giveaways.GiveawayDraft{
			GroupID:            groupID,
			AdminUserID:        user.ID,
			PhotoFileID:        state.GetData(giveawayDataPhotoID),
			Description:        state.GetData(giveawayDataDescription),
			PublishDate:        publishDate,
			ApplicationEndDate: applicationEndDate,
			ResultsDate:        resultsDate,
			IsAnonymous:        false,
		},
		OriginalDescription: state.GetData(giveawayDataOriginalDescription),
	}); createErr != nil {
		logger.Error("failed to create giveaway", zap.Error(createErr))
		g.handleError(ctx, update, createErr)
		return
	}

	state.Clear()
	g.sendReply(ctx, update, &bot.SendMessageParams{Text: "✅ Giveaway scheduled successfully!"})
}

func (g *GiveawayScheduler) handleCancelCommand(ctx context.Context, _ *bot.Bot, update *models.Update) {
	logger := g.withContext(update)

	state, err := state.GetState(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		g.handleError(ctx, update, err)
		return
	}

	state.Clear()

	g.sendReply(ctx, update, &bot.SendMessageParams{Text: "🔄 Operation cancelled."})
}

func (g *GiveawayScheduler) loadGroupAndSettings(ctx context.Context) (*groups.Group, *giveaways.Settings, error) {
	state, err := state.GetState(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get state: %w", err)
	}

	groupID, err := strconv.ParseInt(state.GetData(giveawayDataGroupID), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse group ID: %w", err)
	}

	group, err := g.groupsSvc.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get group: %w", err)
	}

	settings, err := giveaways.NewSettings(group.Settings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return &group.Group, &settings, nil
}

func (g *GiveawayScheduler) downloadPhoto(ctx context.Context, fileID string) ([]byte, error) {
	const maxPhotoSize = 10 * 1024 * 1024 // 10MB limit
	const downloadTimeout = 30 * time.Second

	f, err := g.bot.GetFile(ctx, &bot.GetFileParams{
		FileID: fileID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	link := g.bot.FileDownloadLink(f)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %s", resp.Status) //nolint:err113 //generic error is enough
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPhotoSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func parseDateTime(timeStr string) (time.Time, error) {
	// Parse time in YYYY-MM-DD HH:MM format
	const layout = "2006-01-02 15:04"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}
	return t, nil
}

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
