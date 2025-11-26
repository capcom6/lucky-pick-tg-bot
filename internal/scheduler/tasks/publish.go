package tasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Publish struct {
	base

	giveawaysSvc *giveaways.Service
}

func NewPublish(bot *bot.Bot, giveawaysSvc *giveaways.Service, logger *zap.Logger) Task {
	return &Publish{
		base: base{
			bot:    bot,
			logger: logger,
		},

		giveawaysSvc: giveawaysSvc,
	}
}

func (p *Publish) Name() string {
	return "Publish"
}

func (p *Publish) Run(ctx context.Context) error {
	scheduled, err := p.giveawaysSvc.ListScheduled(ctx)
	if err != nil {
		return fmt.Errorf("failed to list scheduled giveaways: %w", err)
	}

	for _, giveaway := range scheduled {
		if pubErr := p.publish(ctx, &giveaway); pubErr != nil {
			p.logger.Error("failed to publish giveaway",
				zap.Int64("giveaway_id", giveaway.ID),
				zap.Error(pubErr),
			)
		}
	}

	return nil
}

func (p *Publish) publish(ctx context.Context, giveaway *giveaways.Giveaway) error {
	// Кнопки
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Хочу!", CallbackData: "participate:" + strconv.FormatInt(giveaway.ID, 10)},
				// {Text: "❌ Отменить", CallbackData: "cancel"},
			},
		},
	}

	caption := fmt.Sprintf(`%s

*Завершение*: %s
*Итоги*: %s`,
		bot.EscapeMarkdown(giveaway.Description),
		bot.EscapeMarkdown(giveaway.ApplicationEndDate.Format("02.01.2006 15:04")),
		bot.EscapeMarkdown(giveaway.ResultsDate.Format("02.01.2006 15:04")),
	)

	params := &bot.SendPhotoParams{
		ChatID: giveaway.TelegramGroupID,
		Photo: &models.InputFileString{
			Data: giveaway.PhotoFileID,
		},
		Caption:     caption,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: markup,
	}
	message, err := p.bot.SendPhoto(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}

	if updErr := p.giveawaysSvc.Published(ctx, giveaway.ID, int64(message.ID)); updErr != nil {
		return fmt.Errorf("failed to update giveaway: %w", updErr)
	}

	if _, pinErr := p.bot.PinChatMessage(ctx, &bot.PinChatMessageParams{
		ChatID:              giveaway.TelegramGroupID,
		MessageID:           message.ID,
		DisableNotification: false,
	}); pinErr != nil {
		p.logger.Error("failed to pin message",
			zap.Int("message_id", message.ID),
			zap.Error(pinErr),
		)
	}

	return nil
}
