package bot

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers"
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"bot",
		logger.WithNamedLogger("bot"),
		fx.Provide(func() []bot.Option {
			return []bot.Option{
				bot.WithAllowedUpdates(bot.AllowedUpdates{
					"message",
					"callback_query",
					"my_chat_member",
				}),
			}
		}),
		handlers.Module(),
	)
}
