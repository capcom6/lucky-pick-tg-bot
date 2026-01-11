package tasks

import (
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"go.uber.org/zap"
)

type base struct {
	bot    *gotelegrambotfx.Bot
	logger *zap.Logger
}
