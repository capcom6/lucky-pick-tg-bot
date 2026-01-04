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
			settingsSvc.RegisterDefinition(settings.SettingDefinition{
				Key:          "giveaways.llm_description",
				Category:     "giveaways",
				Label:        "Use LLM for Descriptions",
				Description:  "Generate giveaway descriptions using AI",
				Type:         settings.Boolean,
				DefaultValue: false,
				Validation: &settings.SettingValidation{
					Required: false,
				},
			})
		}),
	)
}
