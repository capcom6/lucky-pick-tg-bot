package settings

import (
	"sort"
	"sync"
)

// SettingRegistry manages the registration and retrieval of setting definitions.
type SettingRegistry struct {
	settings map[string]SettingDefinition
	mu       sync.RWMutex
}

// NewSettingRegistry creates a new empty setting registry.
func NewSettingRegistry() *SettingRegistry {
	return &SettingRegistry{
		settings: make(map[string]SettingDefinition),
		mu:       sync.RWMutex{},
	}
}

// RegisterSetting adds a setting definition to the registry.
func (r *SettingRegistry) RegisterSetting(def SettingDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.settings[def.Key] = def
}

// GetSetting retrieves a setting definition by its key
// Returns the setting definition and a boolean indicating if it exists.
func (r *SettingRegistry) GetSetting(key string) (SettingDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	setting, exists := r.settings[key]
	return setting, exists
}

func (r *SettingRegistry) listSettings(filter func(SettingDefinition) bool) []SettingDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []SettingDefinition
	for _, setting := range r.settings {
		if filter(setting) {
			result = append(result, setting)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// ListSettingsByCategory returns all setting definitions for a specific category.
func (r *SettingRegistry) ListSettingsByCategory(category string) []SettingDefinition {
	return r.listSettings(func(s SettingDefinition) bool {
		return s.Category == category
	})
}

// ListAllSettings returns all registered setting definitions.
func (r *SettingRegistry) ListAllSettings() []SettingDefinition {
	return r.listSettings(func(_ SettingDefinition) bool { return true })
}

// ListCategories returns all unique categories that have registered settings.
func (r *SettingRegistry) ListCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make(map[string]bool)
	for _, setting := range r.settings {
		categories[setting.Category] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	sort.Strings(result)
	return result
}
