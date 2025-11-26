package groups

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"groups",
		logger.WithNamedLogger("groups"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewService),
	)
}
