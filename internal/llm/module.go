package llm

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"llm",
		logger.WithNamedLogger("llm"),
		fx.Provide(NewService),
	)
}
