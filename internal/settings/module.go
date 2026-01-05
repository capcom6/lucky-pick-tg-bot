package settings

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"settings",
		logger.WithNamedLogger("settings"),
		fx.Provide(NewSettingRegistry, fx.Private),
		fx.Provide(NewService),
	)
}
