package callback

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

func NewMiddleware(logger *zap.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			next(ctx, b, update)

			if update.CallbackQuery == nil {
				return
			}

			if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
			}); err != nil {
				logger.Error("failed to answer callback query", zap.Error(err))
			}

			if update.CallbackQuery.Message.Message != nil {
				if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
					MessageID: update.CallbackQuery.Message.Message.ID,
				}); err != nil {
					logger.Error("failed to delete message", zap.Error(err))
				}
			}
		}
	}
}
