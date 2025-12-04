package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

const (
	// Callback data prefixes.
	giveawayInitPrefix    = "giveaway:init:"
	giveawayGroupPrefix   = "giveaway:group:"
	giveawayConfirmPrefix = "giveaway:confirm:"
	giveawayCancelPrefix  = "giveaway:cancel"

	// Command constants.
	giveawayCommand = "/giveaway"
	cancelCommand   = "/cancel"
)

// SchedulingState holds the temporary state for a giveaway scheduling session.
type SchedulingState struct {
	AdminUserID     int64
	SelectedGroupID int64
	PhotoFileID     string
	Description     string
	StartTime       time.Time
	ApplicationEnd  time.Time
	ResultsTime     time.Time
	TimeoutTimer    *time.Timer
	LastUpdateTime  time.Time
}

// GiveawayScheduler handles giveaway scheduling flow.
type GiveawayScheduler struct {
	BaseHandler

	usersSvc     *users.Service
	groupsSvc    *groups.Service
	giveawaysSvc *giveaways.Service

	// In-memory storage for scheduling sessions
	schedulingStates map[int64]*SchedulingState
}

func NewGiveawayScheduler(
	bot *bot.Bot,
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

		usersSvc:     usersSvc,
		groupsSvc:    groupsSvc,
		giveawaysSvc: giveawaysSvc,

		schedulingStates: make(map[int64]*SchedulingState),
	}
}

func (g *GiveawayScheduler) Register(b *bot.Bot) {
	// Register command handler
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		giveawayCommand,
		bot.MatchTypeCommandStartOnly,
		g.handleGiveawayCommand,
	)

	// Register callback handlers
	b.RegisterHandlerMatchFunc(g.filterGiveawayInit, g.handleGiveawayInit)
	b.RegisterHandlerMatchFunc(g.filterGiveawayGroup, g.handleGiveawayGroup)
	b.RegisterHandlerMatchFunc(g.filterGiveawayConfirm, g.handleGiveawayConfirm)
	b.RegisterHandlerMatchFunc(g.filterGiveawayCancel, g.handleGiveawayCancel)

	// Register cancel command handler
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		cancelCommand,
		bot.MatchTypeCommand,
		g.handleCancelCommand,
	)

	// Register message handlers for the flow
	b.RegisterHandlerMatchFunc(g.filterSchedulingMessage, g.handleSchedulingMessage)
}

func (g *GiveawayScheduler) filterGiveawayInit(update *models.Update) bool {
	return update.CallbackQuery != nil && strings.HasPrefix(update.CallbackQuery.Data, giveawayInitPrefix)
}

func (g *GiveawayScheduler) filterGiveawayGroup(update *models.Update) bool {
	return update.CallbackQuery != nil && strings.HasPrefix(update.CallbackQuery.Data, giveawayGroupPrefix)
}

func (g *GiveawayScheduler) filterGiveawayConfirm(update *models.Update) bool {
	return update.CallbackQuery != nil && strings.HasPrefix(update.CallbackQuery.Data, giveawayConfirmPrefix)
}

func (g *GiveawayScheduler) filterGiveawayCancel(update *models.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data == giveawayCancelPrefix
}

func (g *GiveawayScheduler) filterSchedulingMessage(update *models.Update) bool {
	userID := update.Message.From.ID
	_, exists := g.schedulingStates[userID]
	return exists && update.Message != nil
}

func (g *GiveawayScheduler) handleGiveawayCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		g.withContext(update).Error("invalid update: missing message or sender")
		return
	}

	// Ensure this is not in a group chat
	if update.Message.Chat.Type != models.ChatTypePrivate {
		g.sendMessage(ctx, update.Message.Chat.ID, "❌ Giveaway scheduling is only available in private chats.")
		return
	}

	logger := g.withContext(update)

	// Register or get user
	user, err := g.usersSvc.RegisterUser(ctx, UserToDomain(update.Message.From))
	if err != nil {
		logger.Error("failed to register user", zap.Error(err))
		g.sendMessage(ctx, update.Message.Chat.ID, "❌ Failed to process user. Please try again.")
		return
	}

	// Check if user is admin of any groups
	adminGroups, err := g.groupsSvc.GetUserAdminGroups(ctx, user.ID)
	if err != nil {
		logger.Error("failed to get user admin groups", zap.Error(err))
		g.sendMessage(ctx, update.Message.Chat.ID, "❌ Failed to verify admin status. Please try again.")
		return
	}

	if len(adminGroups) == 0 {
		g.sendMessage(ctx, update.Message.Chat.ID, "❌ You must be an admin of a group to schedule giveaways.")
		return
	}

	// Initialize scheduling state
	g.initializeSchedulingState(user.ID)

	// If user is admin of exactly one group, auto-select it
	if len(adminGroups) == 1 {
		g.schedulingStates[user.ID].SelectedGroupID = adminGroups[0].ID
		g.sendMessage(ctx, update.Message.Chat.ID, "📸 Please send a photo with description caption for the giveaway.")
		return
	}

	// Show group selection keyboard
	g.showGroupSelectionKeyboard(ctx, update.Message.Chat.ID, adminGroups)
}

func (g *GiveawayScheduler) handleGiveawayInit(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	if _, exists := g.schedulingStates[userID]; !exists {
		g.answerCallbackQuery(ctx, update, "❌ Session expired. Please start again with /giveaway")
		return
	}

	// Answer callback query
	g.answerCallbackQuery(ctx, update, "📸 Please send a photo with description caption.")
}

func (g *GiveawayScheduler) handleGiveawayGroup(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	if _, exists := g.schedulingStates[userID]; !exists {
		g.answerCallbackQuery(ctx, update, "❌ Session expired. Please start again with /giveaway")
		return
	}

	// Parse group ID from callback data
	groupIDStr := strings.TrimPrefix(update.CallbackQuery.Data, giveawayGroupPrefix)
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		g.answerCallbackQuery(ctx, update, "❌ Invalid group selection.")
		return
	}

	g.schedulingStates[userID].SelectedGroupID = groupID
	g.answerCallbackQuery(ctx, update, "📸 Please send a photo with description caption.")
}

func (g *GiveawayScheduler) handleGiveawayConfirm(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	state, exists := g.schedulingStates[userID]
	if !exists {
		g.answerCallbackQuery(ctx, update, "❌ Session expired. Please start again with /giveaway")
		return
	}

	// Create giveaway in database
	giveaway := giveaways.GiveawayDraft{
		GroupID:            state.SelectedGroupID,
		AdminUserID:        state.AdminUserID,
		PhotoFileID:        state.PhotoFileID,
		Description:        state.Description,
		PublishDate:        state.StartTime,
		ApplicationEndDate: state.ApplicationEnd,
		ResultsDate:        state.ResultsTime,
		IsAnonymous:        false,
	}

	if _, err := g.giveawaysSvc.Create(ctx, giveaway); err != nil {
		g.answerCallbackQuery(ctx, update, "❌ Failed to schedule giveaway. Please try again.")
		return
	}

	// Clean up state
	g.cleanupSchedulingState(userID)

	g.answerCallbackQuery(
		ctx,
		update,
		"🎉 Giveaway scheduled! Results will post automatically at "+state.ResultsTime.Format("2006-01-02 15:04"),
	)
}

func (g *GiveawayScheduler) handleGiveawayCancel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	userID := update.CallbackQuery.From.ID
	g.cleanupSchedulingState(userID)

	g.answerCallbackQuery(ctx, update, "🔄 Operation cancelled.")
}

func (g *GiveawayScheduler) handleCancelCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	if _, exists := g.schedulingStates[userID]; exists {
		g.cleanupSchedulingState(userID)
		g.sendMessage(ctx, update.Message.Chat.ID, "🔄 Operation cancelled.")
	}
}

func (g *GiveawayScheduler) handleSchedulingMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	state, exists := g.schedulingStates[userID]
	if !exists {
		return
	}

	// Reset timeout timer
	if state.TimeoutTimer != nil {
		state.TimeoutTimer.Stop()
	}
	state.LastUpdateTime = time.Now()
	state.TimeoutTimer = time.AfterFunc(10*time.Minute, func() {
		g.cleanupSchedulingState(userID)
	})

	// Handle different steps based on state
	if state.PhotoFileID == "" {
		// Expecting photo with caption
		if update.Message.Photo == nil || update.Message.Caption == "" {
			g.sendMessage(ctx, update.Message.Chat.ID, "📸 Please send a photo with description caption.")
			return
		}

		// Store photo and description
		state.PhotoFileID = update.Message.Photo[0].FileID
		state.Description = update.Message.Caption

		// Request start time
		g.sendMessage(
			ctx,
			update.Message.Chat.ID,
			"⏰ Please specify the start time in format: YYYY-MM-DD HH:MM (e.g., 2023-12-25 14:30)",
		)
		return
	}

	if state.StartTime.IsZero() {
		// Expecting start time
		startTime, err := g.parseStartTime(update.Message.Text)
		if err != nil {
			g.sendMessage(
				ctx,
				update.Message.Chat.ID,
				"⏰ Invalid format. Use: YYYY-MM-DD HH:MM (e.g., 2023-12-25 14:30)",
			)
			return
		}

		// Validate time is at least 5 minutes from now
		now := time.Now()
		if startTime.Before(now.Add(5 * time.Minute)) {
			g.sendMessage(ctx, update.Message.Chat.ID, "⏰ Start time must be at least 5 minutes from now.")
			return
		}

		// Calculate other times
		state.StartTime = startTime
		state.ApplicationEnd = startTime.Add(24 * time.Hour)
		state.ResultsTime = startTime.Add(26 * time.Hour)

		// Show preview and confirmation
		g.showPreviewAndConfirmation(ctx, update.Message.Chat.ID, state)
		return
	}
}

func (g *GiveawayScheduler) initializeSchedulingState(userID int64) {
	g.cleanupSchedulingState(userID)

	g.schedulingStates[userID] = &SchedulingState{
		AdminUserID:     userID,
		SelectedGroupID: 0,
		PhotoFileID:     "",
		Description:     "",
		StartTime:       time.Time{},
		ApplicationEnd:  time.Time{},
		ResultsTime:     time.Time{},
		TimeoutTimer:    nil,
		LastUpdateTime:  time.Now(),
	}
}

func (g *GiveawayScheduler) cleanupSchedulingState(userID int64) {
	if state, exists := g.schedulingStates[userID]; exists {
		if state.TimeoutTimer != nil {
			state.TimeoutTimer.Stop()
		}
		delete(g.schedulingStates, userID)
	}
}

func (g *GiveawayScheduler) showGroupSelectionKeyboard(ctx context.Context, chatID int64, groups []groups.GroupModel) {
	buttons := make([][]models.InlineKeyboardButton, 0, len(groups))
	for _, group := range groups {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         group.Title,
				CallbackData: giveawayGroupPrefix + strconv.FormatInt(group.ID, 10),
			},
		})
	}

	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	g.sendMessageWithKeyboard(ctx, chatID, "👥 Select a group for the giveaway:", markup)
}

func (g *GiveawayScheduler) showPreviewAndConfirmation(ctx context.Context, chatID int64, state *SchedulingState) {
	// Get group info
	group, err := g.groupsSvc.GetByID(ctx, state.SelectedGroupID)
	if err != nil {
		g.sendMessage(ctx, chatID, "❌ Failed to load group information. Please try again.")
		return
	}

	previewText := fmt.Sprintf(`🎯 *Preview*

📱 Group: %s
📸 Photo: %s
📝 Description: %s
⏰ Start time (UTC): %s
📝 Application end: %s
🎉 Results: %s`,
		group.Title,
		"Photo uploaded",
		state.Description,
		state.StartTime.Format("2006-01-02 15:04"),
		state.ApplicationEnd.Format("2006-01-02 15:04"),
		state.ResultsTime.Format("2006-01-02 15:04"),
	)

	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "✅ Confirm",
					CallbackData: giveawayConfirmPrefix + "yes",
				},
				{
					Text:         "❌ Cancel",
					CallbackData: giveawayCancelPrefix,
				},
			},
		},
	}

	g.sendMessageWithKeyboard(ctx, chatID, previewText, markup)
}

func (g *GiveawayScheduler) parseStartTime(timeStr string) (time.Time, error) {
	// Parse time in YYYY-MM-DD HH:MM format
	layout := "2006-01-02 15:04"
	return time.Parse(layout, timeStr)
}

func (g *GiveawayScheduler) sendMessageWithKeyboard(
	ctx context.Context,
	chatID int64,
	text string,
	keyboard *models.InlineKeyboardMarkup,
) {
	_, err := g.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeMarkdown,
	})
	if err != nil {
		g.logger.Error("failed to send message with keyboard", zap.Error(err))
	}
}

func (g *GiveawayScheduler) answerCallbackQuery(ctx context.Context, update *models.Update, text string) {
	_, err := g.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            text,
	})
	if err != nil {
		g.logger.Error("failed to answer callback query", zap.Error(err))
	}
}
