package handlers

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Start struct {
	handler.BaseHandler
}

func NewStart(bot *gotelegrambotfx.Bot, logger *zap.Logger) handler.Handler {
	return &Start{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},
	}
}

func (s *Start) Register(b *gotelegrambotfx.Bot) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"start",
		bot.MatchTypeCommandStartOnly,
		adaptor.New(s.handleStart),
	)
}

func (s *Start) handleStart(ctx *adaptor.Context, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		s.WithContext(update).Error("invalid update: missing message or sender")
		return
	}

	logger := s.WithContext(update)

	user, err := ctx.User()
	if err != nil {
		s.HandleError(ctx, update, err)
		return
	}

	displayName := user.Username
	if displayName == "" {
		displayName = user.FirstName
	}
	if displayName == "" {
		displayName = "пользователь"
	}

	if _, replyErr := s.Bot.SendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Привет, " + displayName + "!\n\nДобро пожаловать в Lucky Pick Bot!\n\nТеперь ты сможешь получать уведомления о выигрыше здесь.",
		},
	); replyErr != nil {
		logger.Error("failed to send reply message", zap.Error(replyErr))
	}
}
