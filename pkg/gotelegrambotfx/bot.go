package gotelegrambotfx

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

func New(config Config, options []bot.Option, logger *zap.Logger) (*bot.Bot, error) {
	opts := []bot.Option{
		bot.WithErrorsHandler(func(err error) {
			logger.Error("something went wrong", zap.Error(err))
		}),
		bot.WithDebugHandler(func(format string, args ...any) {
			logger.Debug(fmt.Sprintf(format, args...), zap.String("format", format), zap.Any("args", args))
		}),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			logger.Debug("update is not handled", zap.Any("update", update))
		}),
	}

	opts = append(opts, options...)

	b, err := bot.New(config.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return b, nil
}
