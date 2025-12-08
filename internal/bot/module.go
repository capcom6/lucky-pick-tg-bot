package bot

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/callback"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/state"
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"bot",
		logger.WithNamedLogger("bot"),
		fx.Provide(fx.Annotate(callback.NewMiddleware, fx.ResultTags(`group:"middlewares"`))),
		fx.Provide(fx.Annotate(state.NewMiddleware, fx.ResultTags(`group:"middlewares"`))),
		fx.Provide(
			fx.Annotate(
				func(mws []bot.Middleware) []bot.Option {
					return []bot.Option{
						bot.WithAllowedUpdates(bot.AllowedUpdates{
							"message",
							"callback_query",
							"my_chat_member",
						}),
						bot.WithMiddlewares(mws...),
					}
				},
				fx.ParamTags(`group:"middlewares"`),
			),
		),
		handlers.Module(),
	)
}
