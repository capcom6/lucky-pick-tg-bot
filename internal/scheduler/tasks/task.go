package tasks

import "context"

type Task interface {
	Name() string
	Run(ctx context.Context) error
}
