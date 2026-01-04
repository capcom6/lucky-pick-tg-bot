package giveaways

import (
	"fmt"
	"strconv"

	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
)

type Settings struct {
	LLMDescription bool
}

func NewSettings(settings map[string]string) (Settings, error) {
	s := DefaultSettings()

	if d := settings["giveaways.llm_description"]; d != "" {
		description, err := strconv.ParseBool(d)
		if err != nil {
			return s, fmt.Errorf("failed to parse giveaways.llm_description setting: %w", err)
		}
		s.LLMDescription = description
	}

	return s, nil
}

func DefaultSettings() Settings {
	return Settings{
		LLMDescription: false,
	}
}

func SettingDefinitions() []settings.SettingDefinition {
	//nolint:exhaustruct //default values
	return []settings.SettingDefinition{
		{
			Key:          "giveaways.llm_description",
			Category:     "🎯 Giveaways",
			Label:        "Use LLM for Descriptions",
			Description:  "Generate giveaway descriptions using AI",
			Type:         settings.Boolean,
			DefaultValue: "false",
		},
	}
}
