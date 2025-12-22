package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	giveawayDataGroupID            = "groupID"
	giveawayDataPhotoID            = "photoID"
	giveawayDataDescription        = "description"
	giveawayDataPublishDate        = "publishDate"
	giveawayDataApplicationEndDate = "applicationEndDate"
	giveawayDataResultsDate        = "resultsDate"
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

func (g *GiveawayScheduler) showGroupSelectionKeyboard(ctx context.Context, chatID int64, groups []groups.Group) {
	buttons := make([][]models.InlineKeyboardButton, 0, len(groups))
	for _, group := range groups {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         group.Title,
				CallbackData: giveawayCallbackGroup + strconv.FormatInt(group.ID, 10),
			},
		})
	}

	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

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
	state.AddData(giveawayDataDescription, update.Message.Caption)

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
	groupID, err := strconv.ParseInt(state.GetData(giveawayDataGroupID), 10, 64)
	if err != nil {
		g.sendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Failed to parse group information. Please try again.",
		})
		return
	}

	// Get group info
	group, err := g.groupsSvc.GetByID(ctx, groupID)
	if err != nil {
		g.logger.Error("failed to load group information", zap.Error(err), zap.Int64("group_id", groupID))
		g.sendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Failed to load group information. Please try again.",
		})
		return
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

	description := state.GetData(giveawayDataDescription)

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

	if createErr := g.giveawaysSvc.Create(ctx, giveaways.GiveawayDraft{
		GroupID:            groupID,
		AdminUserID:        user.ID,
		PhotoFileID:        state.GetData(giveawayDataPhotoID),
		Description:        description,
		PublishDate:        publishDate,
		ApplicationEndDate: applicationEndDate,
		ResultsDate:        resultsDate,
		IsAnonymous:        false,
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
