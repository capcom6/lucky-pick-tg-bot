package gotelegrambotfx

import (
	"context"

	"github.com/go-core-fx/logger"
	"github.com/go-telegram/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"gotelegrambotfx",
		logger.WithNamedLogger("gotelegrambotfx"),
		fx.Provide(New),
		fx.Invoke(func(b *bot.Bot, log *zap.Logger, lc fx.Lifecycle) {
			ctx, cancel := context.WithCancel(context.Background())
			closeChan := make(chan struct{})
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					go func() {
						b.Start(ctx)
						close(closeChan)
					}()
					log.Info("bot starting")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					cancel()
					select {
					case <-closeChan:
					case <-ctx.Done():
						log.Warn("bot stop timeout exceeded, forcing shutdown")
						return ctx.Err()
					}
					log.Info("bot stopped")
					return nil
				},
			})
		}),
	)
}
