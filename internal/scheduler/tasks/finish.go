package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Finish struct {
	base

	giveawaysSvc *giveaways.Service
	usersSvc     *users.Service
}

func NewFinish(
	bot *gotelegrambotfx.Bot,
	giveawaysSvc *giveaways.Service,
	usersSvc *users.Service,
	logger *zap.Logger,
) Task {
	return &Finish{
		base: base{
			bot:    bot,
			logger: logger,
		},

		giveawaysSvc: giveawaysSvc,
		usersSvc:     usersSvc,
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
	if winner.Giveaway.IsAnonymous && winner.Participant != nil {
		if err := f.notifyWinnerDirect(ctx, winner); err != nil {
			if isBotForbiddenError(err) {
				return errors.Join(err, f.handleForbiddenError(ctx, winner))
			}

			return err
		}
	}

	params := &bot.SendMessageParams{
		ChatID: winner.Giveaway.Group.TelegramID,
		Text:   f.formatText(winner),
		ReplyParameters: &models.ReplyParameters{
			MessageID:                int(winner.Giveaway.TelegramMessageID),
			ChatID:                   winner.Giveaway.Group.TelegramID,
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

	if winner.Giveaway.IsAnonymous {
		return bot.EscapeMarkdown(
			"🏆 Победитель выбран.\n\n📩 Мы уже отправили победителю личное сообщение с инструкциями.",
		)
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

func (f *Finish) notifyWinnerDirect(ctx context.Context, winner giveaways.Winner) error {
	if winner.Participant == nil {
		return nil
	}

	_, err := f.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: winner.Participant.UserTelegramID,
		Text:   "🎉 Поздравляем! Вы стали победителем розыгрыша.\n\nСвяжитесь с администратором для получения приза.",
	})
	if err != nil {
		return fmt.Errorf("failed to send winner direct message: %w", err)
	}

	return nil
}

func (f *Finish) handleForbiddenError(ctx context.Context, winner giveaways.Winner) error {
	var err error

	if rerollErr := f.giveawaysSvc.ReopenForReroll(ctx, winner.Giveaway.ID); rerollErr != nil {
		f.logger.Error("failed to reopen giveaway for reroll",
			zap.Int64("giveaway_id", winner.Giveaway.ID),
			zap.Error(rerollErr),
		)

		err = errors.Join(err, fmt.Errorf("failed to reopen giveaway for reroll: %w", rerollErr))
	}

	if setErr := f.usersSvc.SetActive(ctx, winner.Participant.UserID, false); setErr != nil {
		f.logger.Error("failed to mark user as inactive",
			zap.Int64("user_id", winner.Participant.UserID),
			zap.Error(setErr),
		)

		err = errors.Join(err, fmt.Errorf("failed to mark user as inactive: %w", setErr))
	}

	return err
}

func isBotForbiddenError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, bot.ErrorForbidden)
}
