package gotelegrambotfx

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx/extractors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Bot struct {
	*bot.Bot
}

func New(config Config, options []bot.Option, logger *zap.Logger) (*Bot, error) {
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

	return &Bot{b}, nil
}

func (b *Bot) SendReply(
	ctx context.Context,
	update *models.Update,
	params *bot.SendMessageParams,
) (*models.Message, error) {
	fromID := extractors.From(update)

	p := *params
	p.ChatID = fromID
	if p.ReplyParameters != nil {
		rp := *p.ReplyParameters
		switch {
		case update != nil && update.Message != nil:
			rp.MessageID = update.Message.ID
		case update != nil && update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil:
			rp.MessageID = update.CallbackQuery.Message.Message.ID
		}
		p.ReplyParameters = &rp
	}

	return b.SendMessage(ctx, &p) //nolint:wrapcheck // pass upstream
}
