package scheduler

import (
	"context"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/scheduler/tasks"
	"go.uber.org/zap"
)

type Service struct {
	tasks []tasks.Task

	logger *zap.Logger
}

func NewService(tasks []tasks.Task, logger *zap.Logger) *Service {
	return &Service{
		tasks: tasks,

		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Debug("running tasks", zap.Int("count", len(s.tasks)))
			for _, task := range s.tasks {
				start := time.Now()
				if err := task.Run(ctx); err != nil {
					s.logger.Error("failed to run task",
						zap.String("task", task.Name()),
						zap.Error(err),
					)
				}
				duration := time.Since(start)
				s.logger.Info("task finished",
					zap.String("task", task.Name()),
					zap.Duration("duration", duration),
				)
			}

		case <-ctx.Done():
			return
		}
	}
}
