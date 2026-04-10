package handlers

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Start struct {
	handler.BaseHandler

	usersSvc *users.Service
}

func NewStart(bot *gotelegrambotfx.Bot, usersSvc *users.Service, logger *zap.Logger) handler.Handler {
	return &Start{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},

		usersSvc: usersSvc,
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

	user, err := ctx.User()
	if err != nil {
		s.HandleError(ctx, update, err)
		return
	}

	if actErr := s.usersSvc.SetActive(ctx, user.ID, true); actErr != nil {
		s.Logger.Error("set user active", zap.Error(actErr))
	}

	displayName := user.Username
	if displayName == "" {
		displayName = user.FirstName
	}
	if displayName == "" {
		displayName = "пользователь"
	}

	s.SendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Привет, " + displayName + "!\n\nДобро пожаловать в Lucky Pick Bot!\n\nТеперь ты сможешь получать уведомления о выигрыше здесь.",
		},
	)
}
