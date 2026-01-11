package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/filter"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/state"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

// Settings handles the group settings editing flow.
type Settings struct {
	handler.BaseHandler

	fsmService  *fsm.Service
	groupsSvc   *groups.Service
	settingsSvc *settings.Service
}

func NewHandler(
	bot *gotelegrambotfx.Bot,
	fsmService *fsm.Service,
	groupsSvc *groups.Service,
	settingsSvc *settings.Service,
	logger *zap.Logger,
) handler.Handler {
	return &Settings{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},

		fsmService:  fsmService,
		groupsSvc:   groupsSvc,
		settingsSvc: settingsSvc,
	}
}

func (s *Settings) Register(b *gotelegrambotfx.Bot) {
	// Register state-specific handlers
	s.registerStateHandlers(b)

	// Register callback handlers for settings flow
	s.registerCallbackHandlers(b)
}

func (s *Settings) registerStateHandlers(b *gotelegrambotfx.Bot) {
	// Category selection handler
	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, callbackCategoryPrefix)
		},
		adaptor.New(s.handleCategorySelect),
	)

	// Setting list handler
	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, callbackSettingPrefix)
		},
		adaptor.New(s.handleSettingSelect),
	)

	// Input handlers for different types
	b.RegisterHandlerMatchFunc(
		filter.And(
			func(update *models.Update) bool { return update.Message != nil },
			state.NewStatePrefixFilter(settingInputPrefix, s.fsmService, s.Logger),
		),
		adaptor.New(s.handleTextInput),
	)

	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, callbackInputBooleanPrefix)
		},
		adaptor.New(s.handleBooleanInput),
	)
}

func (s *Settings) registerCallbackHandlers(b *gotelegrambotfx.Bot) {
	// Settings list
	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, callbackGroupPrefix)
		},
		adaptor.New(s.handleSettingsList),
	)
}

// Permission check

func (s *Settings) checkAdminPermission(ctx *adaptor.Context, groupID int64) bool {
	user, err := ctx.User()
	if err != nil {
		s.Logger.Error("failed to get user from context", zap.Error(err))
		return false
	}

	isAdmin, err := s.groupsSvc.IsAdmin(ctx, groupID, user.ID)
	if err != nil {
		s.Logger.Error("failed to check if user is group admin", zap.Error(err))
		return false
	}

	return isAdmin
}

// Handler implementations

func (s *Settings) handleSettingsList(ctx *adaptor.Context, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	// Extract group ID from callback data
	callbackData := update.CallbackQuery.Data
	groupIDStr := strings.TrimPrefix(callbackData, callbackGroupPrefix)
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		s.HandleError(ctx, update, err)
		return
	}

	s.showCategoriesList(ctx, update, groupID)
}

func (s *Settings) showCategoriesList(ctx *adaptor.Context, update *models.Update, groupID int64) {
	logger := s.WithContext(update)

	// Check admin permission
	if !s.checkAdminPermission(ctx, groupID) {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ You must be an admin of this group to edit settings.",
		})
		return
	}

	// Get group info
	group, err := s.groupsSvc.GetByID(ctx, groupID)
	if err != nil {
		logger.Error("failed to get group", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	// Get available categories
	categories := s.settingsSvc.ListCategories()
	if len(categories) == 0 {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "⚙️ No settings categories available.",
		})
		return
	}

	// Build keyboard with categories
	keyboard := categoriesKeyboard(categories)

	// Update state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	state.SetName(categoriesList)
	state.SetGroupID(groupID)

	// Send message
	s.SendReply(ctx, update, &bot.SendMessageParams{
		Text: fmt.Sprintf(`⚙️ *Settings Categories*

📱 Group: %s

Select a category to edit:`, bot.EscapeMarkdown(group.Title)),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
}

func (s *Settings) handleCategorySelect(ctx *adaptor.Context, update *models.Update) {
	logger := s.WithContext(update)

	if update.CallbackQuery == nil {
		return
	}

	// Extract category from callback data
	category := strings.TrimPrefix(update.CallbackQuery.Data, callbackCategoryPrefix)
	if category == "" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Invalid category selection.",
		})
		return
	}

	// Get current state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	groupID := state.GroupID()
	if groupID == 0 {
		logger.Error("missing group ID in state")
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Missing group context. Please start from the groups menu.",
		})
		return
	}

	s.showSettingsList(ctx, update, groupID, category)
}

func (s *Settings) showSettingsList(ctx *adaptor.Context, update *models.Update, groupID int64, category string) {
	logger := s.WithContext(update)

	// Check admin permission
	if !s.checkAdminPermission(ctx, groupID) {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ You must be an admin of this group to edit settings.",
		})
		return
	}

	// Get settings for category
	settingsList := s.settingsSvc.ListSettingsByCategory(category)
	if len(settingsList) == 0 {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: fmt.Sprintf("❌ No settings found in category: %s", category),
		})
		return
	}

	// Get current values
	currentValues, _, err := s.settingsSvc.GetAllSettings(ctx, groupID)
	if err != nil {
		logger.Error("failed to get current settings", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	// Build keyboard with settings
	keyboard := settingsKeyboard(groupID, settingsList, currentValues)

	// Update state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	state.SetName(settingList)
	state.SetCategory(category)

	// Send message
	s.SendReply(ctx, update, &bot.SendMessageParams{
		Text: fmt.Sprintf(`⚙️ *%s Settings*

Select a setting to edit:`, bot.EscapeMarkdown(category)),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
}

func (s *Settings) handleSettingSelect(ctx *adaptor.Context, update *models.Update) {
	logger := s.WithContext(update)

	if update.CallbackQuery == nil {
		return
	}

	settingKey := strings.TrimPrefix(update.CallbackQuery.Data, callbackSettingPrefix)
	if settingKey == "" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Invalid setting selection.",
		})
		return
	}

	// Get current state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	groupID := state.GroupID()
	if groupID == 0 {
		logger.Error("missing group ID in state")
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Missing group context. Please start from the groups menu.",
		})
		return
	}

	s.showSettingEdit(ctx, update, groupID, settingKey)
}

func (s *Settings) showSettingEdit(
	ctx *adaptor.Context,
	update *models.Update,
	groupID int64,
	settingKey string,
) {
	logger := s.WithContext(update)

	// Check admin permission
	if !s.checkAdminPermission(ctx, groupID) {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ You must be an admin of this group to edit settings.",
		})
		return
	}

	// Get setting definition
	setting, exists := s.settingsSvc.GetSettingDefinition(settingKey)
	if !exists {
		logger.Error("setting definition not found", zap.String("key", settingKey))
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Setting not found.",
		})
		return
	}

	// Get current value
	currentValue, _, err := s.settingsSvc.GetSetting(ctx, groupID, settingKey)
	if err != nil {
		logger.Error("failed to get current setting value", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	// Build prompt message
	promptMessage := fmt.Sprintf("✏️ *Edit Setting: %s*\n\n%s\nCurrent value: `%s`\n\nEnter the new value:",
		bot.EscapeMarkdown(setting.Label),
		bot.EscapeMarkdown(setting.Description),
		bot.EscapeMarkdown(currentValue),
	)

	// Update state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	state.SetName(settingInputPrefix + string(setting.Type))
	state.SetSetting(settingKey)

	// Send message
	s.SendReply(ctx, update, &bot.SendMessageParams{
		Text:        promptMessage,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: editKeyboard(setting, currentValue),
	})
}

func (s *Settings) handleTextInput(ctx *adaptor.Context, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Please enter a value.",
		})
		return
	}

	s.processSettingInput(ctx, update, update.Message.Text)
}

func (s *Settings) handleBooleanInput(ctx *adaptor.Context, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	// Extract boolean value from callback data
	boolValueStr := strings.TrimPrefix(update.CallbackQuery.Data, callbackInputBooleanPrefix)
	if boolValueStr != "true" && boolValueStr != "false" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Invalid boolean value in callback data.",
		})
		return
	}

	s.processSettingInput(ctx, update, boolValueStr)
}

func (s *Settings) processSettingInput(ctx *adaptor.Context, update *models.Update, inputValue string) {
	logger := s.WithContext(update)

	// Get current state
	state, err := s.state(ctx)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		s.HandleError(ctx, update, err)
		return
	}

	groupID := state.GroupID()
	if groupID == 0 {
		logger.Error("missing group ID in state")
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Missing group context. Please start from the groups menu.",
		})
		return
	}

	category := state.Category()
	if category == "" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Missing category context. Please go back to the categories menu.",
		})
		return
	}

	settingKey := state.Setting()
	if settingKey == "" {
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Missing setting key. Please select a setting to edit.",
		})
		return
	}

	// Validate input
	if validErr := s.settingsSvc.ValidateSetting(settingKey, inputValue); validErr != nil {
		logger.Warn("invalid setting value", zap.Error(validErr))
		s.SendReply(ctx, update, &bot.SendMessageParams{
			Text: fmt.Sprintf("❌ Invalid input: %s", validErr.Error()),
		})
		return
	}

	// Save setting
	if updErr := s.settingsSvc.UpdateSetting(ctx, groupID, settingKey, inputValue); updErr != nil {
		logger.Error("failed to save setting", zap.Error(updErr))
		s.HandleError(ctx, update, updErr)
		return
	}

	// Show confirmation and updated setting
	s.SendReply(ctx, update, &bot.SendMessageParams{
		Text: "✅ Setting saved successfully!",
	})

	// Show updated setting
	s.showSettingsList(ctx, update, groupID, category)
}

func (s *Settings) state(ctx *adaptor.Context) (*internalState, error) {
	state, err := ctx.State()
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	return newInternalState(state), nil
}
