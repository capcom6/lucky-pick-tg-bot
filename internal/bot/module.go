package bot

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/callback"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/state"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/user"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"bot",
		logger.WithNamedLogger("bot"),
		fx.Provide(
			func(usersSvc *users.Service, stateSvc *fsm.Service, log *zap.Logger) []bot.Option {
				return []bot.Option{
					bot.WithAllowedUpdates(bot.AllowedUpdates{
						"message",
						"callback_query",
						"my_chat_member",
					}),
					bot.WithMiddlewares(
						user.NewMiddleware(usersSvc, log),
						state.NewMiddleware(stateSvc, log),
						callback.NewMiddleware(log),
					),
				}
			},
		),
		handlers.Module(),
		fx.Invoke(func(lc fx.Lifecycle, b *gotelegrambotfx.Bot, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					RegisterCommands(ctx, b, log)

					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)
}
