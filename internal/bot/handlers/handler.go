package handlers

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/extractors"
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

func (h *BaseHandler) sendMessage(ctx context.Context, params *bot.SendMessageParams) {
	if params.ChatID == 0 {
		h.logger.Error("failed to send message: missing chat ID", zap.Any("params", params))
		return
	}

	if _, err := h.bot.SendMessage(ctx, params); err != nil {
		h.logger.Error(
			"failed to send message",
			zap.Any("params", params),
			zap.Error(err),
		)
	}
}

func (h *BaseHandler) sendReply(ctx context.Context, update *models.Update, params *bot.SendMessageParams) {
	fromID := extractors.From(update)

	params.ChatID = fromID
	if params.ReplyParameters != nil && update.Message != nil {
		params.ReplyParameters.MessageID = update.Message.ID
	}

	h.sendMessage(ctx, params)
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
	h.withContext(update).Error("handling error", zap.Error(err))
	h.sendReply(
		ctx,
		update,
		&bot.SendMessageParams{Text: "К сожалению, возникла ошибка. Обратитесь к администратору."},
	)
}
