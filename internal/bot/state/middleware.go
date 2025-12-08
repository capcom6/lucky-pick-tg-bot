package state

import (
	"context"
	"errors"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/extractors"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type contextKey string

const stateKey contextKey = "state"

var ErrContextKeyNotFound = errors.New("context key not found")

func NewMiddleware(svc *fsm.Service, logger *zap.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			userID := extractors.UserID(update)

			if userID == 0 {
				next(ctx, bot, update)
				return
			}

			state, err := svc.Get(ctx, userID)
			if err != nil {
				logger.Error("get state", zap.Error(err))
				state = new(fsm.State)
			}

			ctx = context.WithValue(ctx, stateKey, state)

			next(ctx, bot, update)

			if setErr := svc.Set(ctx, userID, state); setErr != nil {
				logger.Error("set state", zap.Error(setErr))
			}
		}
	}
}

func GetState(ctx context.Context) (*fsm.State, error) {
	if v, ok := ctx.Value(stateKey).(*fsm.State); ok {
		return v, nil
	}

	return nil, ErrContextKeyNotFound
}
