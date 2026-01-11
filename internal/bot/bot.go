package bot

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

func RegisterCommands(ctx context.Context, b *gotelegrambotfx.Bot, logger *zap.Logger) {
	// Register bot commands
	commands := []models.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "giveaway", Description: "Create a new giveaway"},
		{Command: "cancel", Description: "Cancel current operation"},
		{Command: "groups", Description: "List your groups"},
	}
	_, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commands})
	if err != nil {
		logger.Error("failed to set bot commands", zap.Error(err))
	}
}
