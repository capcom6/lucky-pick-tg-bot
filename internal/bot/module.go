package bot

import (
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"bot",
		logger.WithNamedLogger("bot"),
		fx.Provide(func() []bot.Option {
			return []bot.Option{}
		}),
		// fx.Provide(New),
	)
}
