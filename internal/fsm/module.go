package fsm

import (
	"fmt"

	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"fsm",
		logger.WithNamedLogger("fsm"),
		fx.Provide(func(factory cachefx.Factory) (cache.Cache, error) {
			storage, err := factory.New("fsm")
			if err != nil {
				return nil, fmt.Errorf("create cache: %w", err)
			}

			return storage, nil
		}, fx.Private),
		fx.Provide(NewService),
	)
}
