package openrouterfx

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"openrouterfx",
		logger.WithNamedLogger("openrouterfx"),
		fx.Provide(New),
	)
}
