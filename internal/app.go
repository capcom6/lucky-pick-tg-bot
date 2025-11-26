package internal

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot"
	"github.com/capcom6/lucky-pick-tg-bot/internal/config"
	"github.com/capcom6/lucky-pick-tg-bot/internal/db"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/server"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-core-fx/bunfx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/goosefx"
	"github.com/go-core-fx/logger"
	"github.com/go-core-fx/sqlfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Run() {
	fx.New(
		logger.Module(),
		logger.WithFxDefaultLogger(),
		config.Module(),
		sqlfx.Module(),
		goosefx.Module(),
		bunfx.Module(),
		fiberfx.Module(),
		gotelegrambotfx.Module(),
		//
		db.Module(),
		server.Module(),
		bot.Module(),
		//
		users.Module(),
		groups.Module(),
		//
		fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("app started")
					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("app stopped")
					return nil
				},
			})
		}),
	).Run()
}
