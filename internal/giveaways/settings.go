package giveaways

import (
	"fmt"
	"strconv"
)

type Settings struct {
	LLMDescription bool
}

func DefaultSettings() Settings {
	return Settings{
		LLMDescription: false,
	}
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
