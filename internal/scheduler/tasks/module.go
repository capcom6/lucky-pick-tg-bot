package tasks

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"tasks",
		logger.WithNamedLogger("tasks"),
		fx.Provide(fx.Annotate(NewPublish, fx.ResultTags(`group:"tasks"`))),
		fx.Provide(fx.Annotate(NewClose, fx.ResultTags(`group:"tasks"`))),
		fx.Provide(fx.Annotate(NewFinish, fx.ResultTags(`group:"tasks"`))),
	)
}
