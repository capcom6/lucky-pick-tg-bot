package tasks

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type Close struct {
	base

	giveawaysSvc *giveaways.Service
}

func NewClose(bot *bot.Bot, giveawaysSvc *giveaways.Service, logger *zap.Logger) Task {
	return &Close{
		base: base{
			bot:    bot,
			logger: logger,
		},

		giveawaysSvc: giveawaysSvc,
	}
}

func (c *Close) Name() string {
	return "Close"
}

func (c *Close) Run(ctx context.Context) error {
	active, err := c.giveawaysSvc.ListApplicationFinished(ctx)
	if err != nil {
		return fmt.Errorf("failed to list giveaways: %w", err)
	}

	for _, giveaway := range active {
		if pubErr := c.close(ctx, &giveaway); pubErr != nil {
			c.logger.Error("failed to close giveaway",
				zap.Int64("giveaway_id", giveaway.ID),
				zap.Error(pubErr),
			)
		}
	}

	return nil
}

func (c *Close) close(ctx context.Context, giveaway *giveaways.Giveaway) error {
	if _, err := c.bot.UnpinChatMessage(
		ctx,
		&bot.UnpinChatMessageParams{
			ChatID:    giveaway.Group.TelegramID,
			MessageID: int(giveaway.TelegramMessageID),
		}); err != nil {
		c.logger.Error("failed to unpin message",
			zap.Int64("message_id", giveaway.TelegramMessageID),
			zap.Error(err),
		)
	}

	if err := c.giveawaysSvc.Close(ctx, giveaway.ID); err != nil {
		return fmt.Errorf("failed to close giveaway: %w", err)
	}

	return nil
}
