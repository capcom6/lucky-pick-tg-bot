package handlers

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Groups struct {
	BaseHandler

	groupsSvc *groups.Service
}

func NewGroups(bot *bot.Bot, groupsSvc *groups.Service, logger *zap.Logger) Handler {
	return &Groups{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

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

	switch update.MyChatMember.NewChatMember.Type {
	case models.ChatMemberTypeOwner, models.ChatMemberTypeAdministrator:
		if err := h.groupsSvc.CreateOrUpdate(ctx, update.MyChatMember.Chat.ID, update.MyChatMember.Chat.Title); err != nil {
			h.logger.Error("failed to create or update group", zap.Error(err))
		}
	case models.ChatMemberTypeMember,
		models.ChatMemberTypeRestricted,
		models.ChatMemberTypeLeft,
		models.ChatMemberTypeBanned:
		if err := h.groupsSvc.Disable(ctx, update.MyChatMember.Chat.ID); err != nil {
			h.logger.Error("failed to disable group", zap.Error(err))
		}
	default:
		h.logger.Warn("unknown chat member type", zap.String("type", string(update.MyChatMember.NewChatMember.Type)))
	}
}
