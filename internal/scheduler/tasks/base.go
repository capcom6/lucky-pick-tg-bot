package tasks

import (
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type base struct {
	bot    *bot.Bot
	logger *zap.Logger
}
