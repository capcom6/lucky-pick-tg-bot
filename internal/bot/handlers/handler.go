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

func (h *BaseHandler) withContext(update *models.Update) *zap.Logger {
	logger := h.logger

	if update == nil {
		return h.logger
	}

	switch {
	case update.CallbackQuery != nil:
		logger = logger.With(
			zap.String("update_type", "callback_query"),
			zap.Int64("from_id", update.CallbackQuery.From.ID),
			zap.String("callback_data", update.CallbackQuery.Data),
		)
	case update.Message != nil:
		logger = logger.With(
			zap.String("update_type", "message"),
			zap.Int64("chat_id", update.Message.Chat.ID),
			zap.String("text", update.Message.Text),
		)
	case update.EditedMessage != nil:
		logger = logger.With(
			zap.String("update_type", "edited_message"),
			zap.Int64("chat_id", update.EditedMessage.Chat.ID),
			zap.Int64("message_id", int64(update.EditedMessage.ID)),
			zap.String("text", update.EditedMessage.Text),
		)
	case update.EditedChannelPost != nil:
		logger = logger.With(
			zap.String("update_type", "edited_channel_post"),
			zap.Int64("chat_id", update.EditedChannelPost.Chat.ID),
			zap.Int64("message_id", int64(update.EditedChannelPost.ID)),
			zap.String("text", update.EditedChannelPost.Text),
		)
	case update.InlineQuery != nil:
		logger = logger.With(
			zap.String("update_type", "inline_query"),
			zap.String("query", update.InlineQuery.Query),
		)
	case update.ChosenInlineResult != nil:
		logger = logger.With(
			zap.String("update_type", "chosen_inline_result"),
			zap.Int64("from_id", update.ChosenInlineResult.From.ID),
			zap.String("query", update.ChosenInlineResult.Query),
			zap.String("result_id", update.ChosenInlineResult.ResultID),
		)
	}
	return logger
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
		h.withContext(update).Error("failed to send error message", zap.Error(sendErr), zap.NamedError("source", err))
	}
}
