package handlers

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers/cancel"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers/settings"
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"handlers",
		logger.WithNamedLogger("handlers"),
		fx.Provide(fx.Annotate(NewStart, fx.ResultTags(`group:"handlers"`))),
		fx.Provide(fx.Annotate(NewParticipant, fx.ResultTags(`group:"handlers"`))),
		fx.Provide(fx.Annotate(groups.NewHandler, fx.ResultTags(`group:"handlers"`))),
		fx.Provide(fx.Annotate(NewGiveawayScheduler, fx.ResultTags(`group:"handlers"`))),
		fx.Provide(fx.Annotate(settings.NewHandler, fx.ResultTags(`group:"handlers"`))),
		fx.Provide(fx.Annotate(cancel.NewHandler, fx.ResultTags(`group:"handlers"`))),
		fx.Invoke(fx.Annotate(
			func(handlers []handler.Handler, b *bot.Bot) {
				for _, handler := range handlers {
					handler.Register(b)
				}
			},
			fx.ParamTags(`group:"handlers"`),
		)),
	)
}
