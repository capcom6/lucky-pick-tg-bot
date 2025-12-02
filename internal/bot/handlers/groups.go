package handlers

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Groups struct {
	BaseHandler

	usersSvc  *users.Service
	groupsSvc *groups.Service
}

func NewGroups(bot *bot.Bot, usersSvc *users.Service, groupsSvc *groups.Service, logger *zap.Logger) Handler {
	return &Groups{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

		usersSvc:  usersSvc,
		groupsSvc: groupsSvc,
	}
}

func (h *Groups) Register(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(
		h.filter,
		h.handle,
	)
}

func (h *Groups) filter(update *models.Update) bool {
	return update.MyChatMember != nil
}

func (h *Groups) handle(ctx context.Context, _ *bot.Bot, update *models.Update) {
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
			groups.Group{TelegramID: update.MyChatMember.Chat.ID, Title: update.MyChatMember.Chat.Title},
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
