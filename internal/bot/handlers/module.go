package handlers

import (
	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"handlers",
		logger.WithNamedLogger("handlers"),
		fx.Provide(fx.Annotate(NewStart, fx.ResultTags(`group:"handlers"`))),
		fx.Invoke(fx.Annotate(
			func(b *bot.Bot, handlers []Handler) {
				for _, handler := range handlers {
					handler.Register(b)
				}
			},
			fx.ParamTags(``, `group:"handlers"`),
		)),
	)
}
