package user

import (
	"context"
	"errors"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/extractors"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type contextKey string

const (
	userKey contextKey = "user"
)

var ErrContextKeyNotFound = errors.New("context key not found")

func NewMiddleware(usersSvc *users.Service, logger *zap.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			tgUser := extractors.User(update)

			if tgUser == nil {
				next(ctx, b, update)
				return
			}

			user, err := usersSvc.RegisterUser(ctx, ToDomain(tgUser))
			if err != nil {
				logger.Error("register user", zap.Error(err))
				_, _ = b.SendMessage(
					ctx,
					&bot.SendMessageParams{
						ChatID: extractors.ChatID(update),
						Text:   "❌ Failed to register user. Please try again.",
					},
				)
				return
			}

			ctx = context.WithValue(ctx, userKey, user)

			next(ctx, b, update)
		}
	}
}

func FromContext(ctx context.Context) (*users.User, error) {
	user, ok := ctx.Value(userKey).(*users.User)
	if !ok {
		return nil, ErrContextKeyNotFound
	}

	return user, nil
}
