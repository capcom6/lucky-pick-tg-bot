package giveaways

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"giveaways",
		logger.WithNamedLogger("giveaways"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewLLM, fx.Private),
		fx.Provide(NewService),
		fx.Invoke(func(settingsSvc *settings.Service) {
			for _, v := range SettingDefinitions() {
				settingsSvc.RegisterDefinition(v)
			}
		}),
	)
}
