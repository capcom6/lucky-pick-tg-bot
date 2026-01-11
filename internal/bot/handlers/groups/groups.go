package groups

import (
	"context"
	"strconv"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/keyboards"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx/extractors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

const (
	// Group command constants.
	groupsCommand = "/groups"

	// Group callback constants.
	groupsSelectionCallback = "groups:select:"
)

// Handler handles group management operations.
type Handler struct {
	handler.BaseHandler

	usersSvc  *users.Service
	groupsSvc *groups.Service
}

func NewHandler(
	bot *gotelegrambotfx.Bot,
	usersSvc *users.Service,
	groupsSvc *groups.Service,
	logger *zap.Logger,
) handler.Handler {
	return &Handler{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},

		usersSvc:  usersSvc,
		groupsSvc: groupsSvc,
	}
}

func (h *Handler) Register(b *gotelegrambotfx.Bot) {
	// Register chat member handler (existing functionality)
	b.RegisterHandlerMatchFunc(
		h.filterChatMember,
		h.handleChatMember,
	)

	// Register command handler
	b.RegisterHandlerMatchFunc(
		h.filterGroupsCommand,
		adaptor.New(h.handleGroupsCommand),
	)

	// Register callback handlers
	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				strings.HasPrefix(update.CallbackQuery.Data, groupsSelectionCallback)
		},
		h.handleGroupSelection,
	)

	b.RegisterHandlerMatchFunc(
		func(update *models.Update) bool {
			return update.CallbackQuery != nil &&
				update.CallbackQuery.Data == "groups:back"
		},
		adaptor.New(h.handleGroupsCommand),
	)
}

func (h *Handler) filterChatMember(update *models.Update) bool {
	return update.MyChatMember != nil
}

func (h *Handler) filterGroupsCommand(update *models.Update) bool {
	if update.Message == nil || update.Message.Text == "" {
		return false
	}
	return update.Message.Text == groupsCommand && update.Message.Chat.Type == models.ChatTypePrivate
}

func (h *Handler) handleChatMember(ctx context.Context, _ *bot.Bot, update *models.Update) {
	h.Logger.Debug("my chat member update", zap.Any("update", update))

	if update.MyChatMember == nil {
		h.Logger.Error("invalid update: missing MyChatMember")
		return
	}

	user, err := h.usersSvc.RegisterUser(ctx, users.UserIn{
		TelegramUserID: update.MyChatMember.From.ID,
		Username:       update.MyChatMember.From.Username,
		FirstName:      update.MyChatMember.From.FirstName,
		LastName:       update.MyChatMember.From.LastName,
	})
	if err != nil {
		h.Logger.Error("failed to register user", zap.Error(err))
		return
	}

	switch update.MyChatMember.NewChatMember.Type {
	case models.ChatMemberTypeOwner, models.ChatMemberTypeAdministrator:
		if createErr := h.groupsSvc.CreateOrUpdate(
			ctx,
			groups.GroupDraft{TelegramID: update.MyChatMember.Chat.ID, Title: update.MyChatMember.Chat.Title},
			groups.Admin{UserID: user.ID},
		); createErr != nil {
			h.Logger.Error("failed to create or update group", zap.Error(createErr))
		}
	case models.ChatMemberTypeMember,
		models.ChatMemberTypeRestricted,
		models.ChatMemberTypeLeft,
		models.ChatMemberTypeBanned:
		if disableErr := h.groupsSvc.Disable(ctx, update.MyChatMember.Chat.ID); disableErr != nil {
			h.Logger.Error("failed to disable group", zap.Error(disableErr))
		}
	default:
		h.Logger.Warn("unknown chat member type", zap.String("type", string(update.MyChatMember.NewChatMember.Type)))
	}
}

func (h *Handler) handleGroupsCommand(ctx *adaptor.Context, update *models.Update) {
	logger := h.WithContext(update)

	// Register or get user
	user, err := ctx.User()
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		h.SendReply(
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
		h.SendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ Failed to verify admin status. Please try again."},
		)
		return
	}

	if len(adminGroups) == 0 {
		h.SendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ You must be an admin of a group to manage group settings."},
		)
		return
	}

	// If user is admin of exactly one group, show its settings directly
	if len(adminGroups) == 1 {
		h.showGroupMenu(ctx, update, adminGroups[0].ID)
		return
	}

	// Show group selection keyboard
	chatID := extractors.ChatID(update)
	if chatID == 0 {
		logger.Error("failed to extract chat ID")
		return
	}
	h.showGroupSelectionKeyboard(ctx, chatID, adminGroups)
}

func (h *Handler) showGroupSelectionKeyboard(ctx context.Context, chatID int64, groups []groups.GroupWithSettings) {
	markup := keyboards.GroupSelectionKeyboard(
		groupsSelectionCallback,
		groups,
	)

	h.SendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "👥 Select a group to manage settings:",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: markup,
		},
	)
}

func (h *Handler) handleGroupSelection(ctx context.Context, _ *bot.Bot, update *models.Update) {
	logger := h.WithContext(update)

	if update.CallbackQuery == nil {
		return
	}

	groupIDStr := strings.TrimPrefix(update.CallbackQuery.Data, groupsSelectionCallback)
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		logger.Error("failed to parse group ID", zap.Error(err))
		return
	}

	h.showGroupMenu(ctx, update, groupID)
}

func (h *Handler) showGroupMenu(ctx context.Context, update *models.Update, groupID int64) {
	h.SendReply(
		ctx,
		update,
		&bot.SendMessageParams{
			Text:        "👥 Group settings",
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: managementKeyboard(groupID),
		},
	)
}
