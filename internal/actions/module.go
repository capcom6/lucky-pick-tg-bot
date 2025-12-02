package actions

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"actions",
		logger.WithNamedLogger("actions"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewService),
	)
}
