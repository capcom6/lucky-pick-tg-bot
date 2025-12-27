package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/keyboards"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/state"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	//
	groupsStatePrefix = "groups:"

	//
	groupsDataGroupID = groupsStatePrefix + "group_id"

	// Group command constants.
	groupsCommand = "/groups"

	// Group callback constants.
	groupsCallbackShowSettings = "groups:show_settings:"
	groupsCallbackEditSettings = "groups:edit_settings:"
)

// Groups handles group management operations.
type Groups struct {
	BaseHandler

	fsmService *fsm.Service
	usersSvc   *users.Service
	groupsSvc  *groups.Service
}

func NewGroups(
	bot *bot.Bot,
	fsmService *fsm.Service,
	usersSvc *users.Service,
	groupsSvc *groups.Service,
	logger *zap.Logger,
) Handler {
	return &Groups{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

		fsmService: fsmService,
		usersSvc:   usersSvc,
		groupsSvc:  groupsSvc,
	}
}

func (h *Groups) Register(b *bot.Bot) {
	// Register chat member handler (existing functionality)
	b.RegisterHandlerMatchFunc(
		h.filterChatMember,
		h.handleChatMember,
	)

	// Register command handler
	b.RegisterHandlerMatchFunc(
		h.filterGroupsCommand,
		h.handleGroupsCommand,
	)

	// Register callback handlers
	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, groupsCallbackShowSettings)
		},
		h.handleShowGroupSettings,
	)
}

func (h *Groups) filterChatMember(update *models.Update) bool {
	return update.MyChatMember != nil
}

func (h *Groups) filterGroupsCommand(update *models.Update) bool {
	if update.Message == nil || update.Message.Text == "" {
		return false
	}
	return update.Message.Text == groupsCommand && update.Message.Chat.Type == models.ChatTypePrivate
}

// func (h *Groups) filterLLMCallback(update *models.Update) bool {
// 	if update.CallbackQuery == nil {
// 		return false
// 	}
// 	return strings.HasPrefix(update.CallbackQuery.Data, groupsCallbackShowLLMSettings) ||
// 		strings.HasPrefix(update.CallbackQuery.Data, groupsCallbackToggleLLM) ||
// 		strings.HasPrefix(update.CallbackQuery.Data, groupsCallbackChangeLLMModel) ||
// 		strings.HasPrefix(update.CallbackQuery.Data, groupsCallbackSelectLLMModel)
// }

func (h *Groups) handleChatMember(ctx context.Context, _ *bot.Bot, update *models.Update) {
	h.logger.Debug("my chat member update", zap.Any("update", update))

	if update.MyChatMember == nil {
		h.logger.Error("invalid update: missing MyChatMember")
		return
	}

	user, err := h.usersSvc.RegisterUser(ctx, users.UserIn{
		TelegramUserID: update.MyChatMember.From.ID,
		Username:       update.MyChatMember.From.Username,
		FirstName:      update.MyChatMember.From.FirstName,
		LastName:       update.MyChatMember.From.LastName,
	})
	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err))
		return
	}

	switch update.MyChatMember.NewChatMember.Type {
	case models.ChatMemberTypeOwner, models.ChatMemberTypeAdministrator:
		if createErr := h.groupsSvc.CreateOrUpdate(
			ctx,
			groups.GroupDraft{TelegramID: update.MyChatMember.Chat.ID, Title: update.MyChatMember.Chat.Title},
			groups.Admin{UserID: user.ID},
		); createErr != nil {
			h.logger.Error("failed to create or update group", zap.Error(createErr))
		}
	case models.ChatMemberTypeMember,
		models.ChatMemberTypeRestricted,
		models.ChatMemberTypeLeft,
		models.ChatMemberTypeBanned:
		if disableErr := h.groupsSvc.Disable(ctx, update.MyChatMember.Chat.ID); disableErr != nil {
			h.logger.Error("failed to disable group", zap.Error(disableErr))
		}
	default:
		h.logger.Warn("unknown chat member type", zap.String("type", string(update.MyChatMember.NewChatMember.Type)))
	}
}

func (h *Groups) handleGroupsCommand(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		h.withContext(update).Error("invalid update: missing message or sender")
		return
	}

	logger := h.withContext(update)

	// Register or get user
	user, err := h.usersSvc.RegisterUser(ctx, UserToDomain(update.Message.From))
	if err != nil {
		logger.Error("failed to register user", zap.Error(err))
		h.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ Failed to process user. Please try again."},
		)
		return
	}

	// Check if user is admin of any groups
	adminGroups, err := h.groupsSvc.GetUserAdminGroups(ctx, user.ID)
	if err != nil {
		logger.Error("failed to get user admin groups", zap.Error(err))
		h.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ Failed to verify admin status. Please try again."},
		)
		return
	}

	if len(adminGroups) == 0 {
		h.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ You must be an admin of a group to manage group settings."},
		)
		return
	}

	// If user is admin of exactly one group, show its settings directly
	if len(adminGroups) == 1 {
		h.showGroupSettings(ctx, update, adminGroups[0].ID)
		return
	}

	// Show group selection keyboard
	h.showGroupSelectionKeyboard(ctx, update.Message.Chat.ID, adminGroups)
}

func (h *Groups) showGroupSelectionKeyboard(ctx context.Context, chatID int64, groups []groups.GroupWithSettings) {
	markup := keyboards.GroupSelectionKeyboard(
		groupsCallbackShowSettings,
		groups,
	)

	h.sendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "👥 Select a group to manage settings:",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: markup,
		},
	)
}

func (h *Groups) handleShowGroupSettings(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		h.withContext(update).Error("invalid update: missing callback query")
		return
	}

	data := update.CallbackQuery.Data
	groupIDstr := strings.TrimPrefix(data, groupsCallbackShowSettings)
	groupID, err := strconv.ParseInt(groupIDstr, 10, 64)
	if err != nil {
		h.withContext(update).Error("failed to parse group ID", zap.Error(err))
		h.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{
				Text:      "❌ Failed to parse group ID. Please try again.",
				ParseMode: models.ParseModeMarkdown,
			},
		)
		return
	}

	h.showGroupSettings(ctx, update, groupID)
}

func (h *Groups) showGroupSettings(ctx context.Context, update *models.Update, groupID int64) {
	// Get current settings
	if !h.isUserGroupAdmin(ctx, groupID, &update.CallbackQuery.From) {
		h.sendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ You are not an admin of this group."},
		)
		return
	}

	group, err := h.groupsSvc.GetByID(ctx, groupID)
	if err != nil {
		h.withContext(update).Error("failed to get group", zap.Error(err))
		h.handleError(ctx, update, err)
		return
	}

	settings := group.Settings

	messageText := strings.Builder{}

	messageText.WriteString(fmt.Sprintf(`⚙️ *Group Settings*

📱 Group: %s
`,
		bot.EscapeMarkdown(group.Title),
	))

	for k, v := range settings {
		messageText.WriteString(fmt.Sprintf("\n🔧 %s: %s", bot.EscapeMarkdown(k), bot.EscapeMarkdown(v)))
	}

	// Get buttons
	const columns = 2
	buttons := lo.Map(
		lo.ChunkEntries(settings, columns),
		func(chunk map[string]string, _ int) []models.InlineKeyboardButton {
			return lo.MapToSlice(chunk, func(key, value string) models.InlineKeyboardButton {
				return models.InlineKeyboardButton{
					Text:         key + ": " + value,
					CallbackData: groupsCallbackEditSettings + key,
				}
			})
		},
	)

	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	state, err := state.GetState(ctx)
	if err != nil {
		h.handleError(ctx, update, fmt.Errorf("failed to get state: %w", err))
		return
	}

	state.AddData(groupsDataGroupID, strconv.FormatInt(group.ID, 10))

	h.sendReply(
		ctx,
		update,
		&bot.SendMessageParams{
			Text:        messageText.String(),
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: markup,
		},
	)
}

func (h *Groups) isUserGroupAdmin(ctx context.Context, groupID int64, user *models.User) bool {
	// Register user to get internal user ID
	u, err := h.usersSvc.RegisterUser(ctx, UserToDomain(user))
	if err != nil {
		h.logger.Error("failed to register user for admin check", zap.Error(err))
		return false
	}

	// Check if user is admin of the group
	isAdmin, err := h.groupsSvc.IsAdmin(ctx, groupID, u.ID)
	if err != nil {
		h.logger.Error("failed to check admin status", zap.Error(err))
		return false
	}

	return isAdmin
}
