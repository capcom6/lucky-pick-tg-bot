package state

import (
	"context"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/extractors"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

func NewStateFilter(target string, fsm *fsm.Service, logger *zap.Logger) bot.MatchFunc {
	return func(update *models.Update) bool {
		userID := extractors.UserID(update)

		if userID == 0 {
			return false
		}

		state, err := fsm.Get(context.Background(), userID)
		if err != nil {
			logger.Error("get state", zap.Error(err))
			return false
		}

		return state.Name == target
	}
}

func NewStatePrefixFilter(prefix string, fsm *fsm.Service, logger *zap.Logger) bot.MatchFunc {
	return func(update *models.Update) bool {
		userID := extractors.UserID(update)

		if userID == 0 {
			return false
		}

		state, err := fsm.Get(context.Background(), userID)
		if err != nil {
			logger.Error("get state", zap.Error(err))
			return false
		}

		return strings.HasPrefix(state.Name, prefix)
	}
}
