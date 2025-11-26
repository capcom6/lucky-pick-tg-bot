package handlers

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Start struct {
	BaseHandler

	usersSvc *users.Service
}

func NewStart(bot *bot.Bot, usersSvc *users.Service, logger *zap.Logger) Handler {
	return &Start{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

		usersSvc: usersSvc,
	}
}

func (s *Start) Register(b *bot.Bot) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"start",
		bot.MatchTypeCommandStartOnly,
		s.handleStart,
	)
}

func (s *Start) handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		s.withContext(update).Error("invalid update: missing message or sender")
		return
	}

	logger := s.withContext(update)

	user, err := s.usersSvc.RegisterUser(
		ctx,
		UserToDomain(update.Message.From),
	)
	if err != nil {
		s.handleError(ctx, update, err)
		return
	}

	displayName := user.Username
	if displayName == "" {
		displayName = user.FirstName
	}
	if displayName == "" {
		displayName = "пользователь"
	}

	if _, replyErr := b.SendMessage(
		ctx,
		&bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Привет, " + displayName + "!\n\nДобро пожаловать в Lucky Pick Bot!\n\nТеперь ты сможешь получать уведомления о выигрыше здесь.",
		},
	); replyErr != nil {
		logger.Error("failed to send reply message", zap.Error(replyErr))
	}
}
