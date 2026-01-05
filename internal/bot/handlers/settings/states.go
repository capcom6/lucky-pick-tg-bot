package settings

// Settings-related FSM states for group settings editing flow.

import (
	"strconv"

	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
)

const (
	// categoriesList shows the main settings menu with categories.
	categoriesList = "settings:categories"

	// settingList shows settings within a selected category.
	settingList = "settings:setting_list"

	// settingInputPrefix is a prefix for input states.
	settingInputPrefix = "settings:input:"
)

// State data keys for settings flow.
const (
	// Group context.
	settingsDataGroupID = "settings:group_id"

	// Current context.
	settingsDataCategory = "settings:category"
	settingsDataSetting  = "settings:setting"
)

type internalState struct {
	*fsm.State
}

func newInternalState(s *fsm.State) *internalState {
	return &internalState{
		State: s,
	}
}

func (s *internalState) SetGroupID(groupID int64) {
	s.AddData(settingsDataGroupID, strconv.FormatInt(groupID, 10))
}

func (s *internalState) GroupID() int64 {
	groupIDStr := s.GetData(settingsDataGroupID)
	if groupIDStr == "" {
		return 0
	}
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		return 0
	}
	return groupID
}

func (s *internalState) SetCategory(category string) {
	s.AddData(settingsDataCategory, category)
}

func (s *internalState) Category() string {
	return s.GetData(settingsDataCategory)
}

func (s *internalState) SetSetting(setting string) {
	s.AddData(settingsDataSetting, setting)
}

func (s *internalState) Setting() string {
	return s.GetData(settingsDataSetting)
}
