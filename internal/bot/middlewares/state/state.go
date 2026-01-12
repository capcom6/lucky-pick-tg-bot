package state

import (
	"context"
	"errors"

	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx/extractors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type contextKey string

const stateKey contextKey = "state"

var ErrContextKeyNotFound = errors.New("context key not found")

func NewMiddleware(svc *fsm.Service, logger *zap.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			userID := extractors.UserID(update)

			if userID == 0 {
				next(ctx, b, update)
				return
			}

			state, err := svc.Get(ctx, userID)
			if err != nil {
				logger.Error("get state", zap.Error(err))
				if _, sendErr := (&gotelegrambotfx.Bot{Bot: b}).SendReply(ctx, update, &bot.SendMessageParams{
					Text: "❌ Failed to get state. Please try again.",
				}); sendErr != nil {
					logger.Error("failed to send error message", zap.Error(sendErr))
				}
				return
			}

			ctx = context.WithValue(ctx, stateKey, state)

			next(ctx, b, update)

			if setErr := svc.Set(ctx, userID, state); setErr != nil {
				logger.Error("set state", zap.Error(setErr))
			}
		}
	}
}

func FromContext(ctx context.Context) (*fsm.State, error) {
	if v, ok := ctx.Value(stateKey).(*fsm.State); ok {
		return v, nil
	}

	return nil, ErrContextKeyNotFound
}
