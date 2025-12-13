package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/discussions"
	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Questions struct {
	base

	giveawaysSvc   *giveaways.Service
	discussionsSvc *discussions.Service
}

func NewQuestions(
	bot *bot.Bot,
	giveawaysSvc *giveaways.Service,
	discussionsSvc *discussions.Service,
	logger *zap.Logger,
) Task {
	return &Questions{
		base: base{
			bot:    bot,
			logger: logger,
		},

		giveawaysSvc:   giveawaysSvc,
		discussionsSvc: discussionsSvc,
	}
}

func (t *Questions) Name() string {
	return "Questions"
}

func (t *Questions) Run(ctx context.Context) error {
	discuss, err := t.discussionsSvc.Generate(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate questions: %w", err)
	}

	indexed := lo.KeyBy(
		discuss,
		func(item discussions.Discussion) int64 {
			return item.GiveawayID
		},
	)

	givs, err := t.giveawaysSvc.ListByIDs(ctx, lo.Keys(indexed))
	if err != nil {
		return fmt.Errorf("failed to list giveaways: %w", err)
	}

	errs := make([]error, 0, len(givs))
	for _, v := range givs {
		d := indexed[v.ID]

		errs = append(errs, t.send(ctx, d, v))
	}

	return errors.Join(errs...)
}

func (t *Questions) send(ctx context.Context, d discussions.Discussion, ga giveaways.Giveaway) error {
	res, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    ga.Group.TelegramID,
		Text:      bot.EscapeMarkdown(d.Text),
		ParseMode: models.ParseModeMarkdown,
		ReplyParameters: &models.ReplyParameters{
			MessageID:                int(ga.TelegramMessageID),
			AllowSendingWithoutReply: false,
		},
	})
	if err != nil {
		t.logger.Error("failed to send message", zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}

	t.logger.Info("message sent", zap.Int("message_id", res.ID))

	if setErr := t.discussionsSvc.SetTelegramID(ctx, d.ID, int64(res.ID)); setErr != nil {
		t.logger.Error("failed to save discussion telegram ID", zap.Error(setErr))
		return fmt.Errorf("failed to save discussion telegram ID: %w", setErr)
	}

	return nil
}
