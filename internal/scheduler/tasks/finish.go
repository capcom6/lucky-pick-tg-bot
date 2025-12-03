package tasks

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Finish struct {
	base

	giveawaysSvc *giveaways.Service
}

func NewFinish(bot *bot.Bot, giveawaysSvc *giveaways.Service, logger *zap.Logger) Task {
	return &Finish{
		base: base{
			bot:    bot,
			logger: logger,
		},

		giveawaysSvc: giveawaysSvc,
	}
}

func (f *Finish) Name() string {
	return "Finish"
}

func (f *Finish) Run(ctx context.Context) error {
	winners, err := f.giveawaysSvc.ListWinners(ctx)
	if err != nil {
		return fmt.Errorf("failed to list winners: %w", err)
	}

	for _, winner := range winners {
		if pubErr := f.notify(ctx, winner); pubErr != nil {
			f.logger.Error(
				"failed to notify winner",
				zap.Int64("giveaway_id", winner.Giveaway.ID),
				zap.Error(pubErr),
			)
		}
	}

	return nil
}

func (f *Finish) notify(ctx context.Context, winner giveaways.Winner) error {
	params := &bot.SendMessageParams{
		ChatID: winner.Giveaway.TelegramGroupID,
		Text:   f.formatText(winner),
		ReplyParameters: &models.ReplyParameters{
			MessageID:                int(winner.Giveaway.TelegramMessageID),
			ChatID:                   winner.Giveaway.TelegramGroupID,
			AllowSendingWithoutReply: false,
		},
		ParseMode: models.ParseModeMarkdown,
	}
	_, err := f.bot.SendMessage(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (f *Finish) formatText(winner giveaways.Winner) string {
	if winner.Participant == nil {
		return bot.EscapeMarkdown("🏆 Победитель: не выбран\n\nК сожалению, участников оказалось недостаточно.")
	}

	var username string
	switch {
	case winner.Participant.UserUsername != "":
		username = "@" + bot.EscapeMarkdown(winner.Participant.UserUsername)
	case winner.Participant.UserFirstName != "":
		username = fmt.Sprintf(
			"[%s](tg://user?id=%d)",
			bot.EscapeMarkdown(winner.Participant.UserFirstName),
			winner.Participant.UserTelegramID,
		)
	default:
		username = fmt.Sprintf(
			"[%d](tg://user?id=%d)",
			winner.Participant.UserTelegramID,
			winner.Participant.UserTelegramID,
		)
	}

	return fmt.Sprintf(
		bot.EscapeMarkdown("🏆 Победитель: %s\n\n🎉Поздравляем!\nСвяжитесь с администратором для получения приза."),
		username,
	)
}
