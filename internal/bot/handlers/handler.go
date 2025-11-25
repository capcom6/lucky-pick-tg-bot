package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Handler interface {
	Register(bot *bot.Bot)
}

type BaseHandler struct {
	bot *bot.Bot

	logger *zap.Logger
}

func (h *BaseHandler) handleError(ctx context.Context, update *models.Update, err error) {
	_, sendErr := h.bot.SendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "К сожалению, возникла ошибка. Обратитесь к администратору.",
		},
	)

	if sendErr != nil {
		h.logger.Error("Failed to send error message", zap.Error(sendErr), zap.NamedError("source", err))
	}
}
