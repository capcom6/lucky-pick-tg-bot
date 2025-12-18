package discussions

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"discussions",
		logger.WithNamedLogger("discussions"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewLLM, fx.Private),
		fx.Provide(NewService),
	)
}
