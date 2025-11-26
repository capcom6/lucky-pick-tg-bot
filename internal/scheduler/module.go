package scheduler

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/scheduler/tasks"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"scheduler",
		logger.WithNamedLogger("scheduler"),
		tasks.Module(),
		fx.Provide(fx.Annotate(
			NewService,
			fx.ParamTags(`group:"tasks"`),
		)),
		fx.Invoke(func(s *Service, lc fx.Lifecycle) {
			ctx, cancel := context.WithCancel(context.Background())
			waitChan := make(chan struct{})

			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					go func() {
						s.Run(ctx)
						close(waitChan)
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					cancel()
					select {
					case <-waitChan:
					case <-ctx.Done():
					}
					return nil
				},
			})
		}),
	)
}
