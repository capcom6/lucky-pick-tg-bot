package settings

// SettingRegistry manages the registration and retrieval of setting definitions
type SettingRegistry struct {
	settings map[string]SettingDefinition
}

// NewSettingRegistry creates a new empty setting registry
func NewSettingRegistry() *SettingRegistry {
	return &SettingRegistry{
		settings: make(map[string]SettingDefinition),
	}
}

// RegisterSetting adds a setting definition to the registry
func (r *SettingRegistry) RegisterSetting(def SettingDefinition) {
	r.settings[def.Key] = def
}

// GetSetting retrieves a setting definition by its key
// Returns the setting definition and a boolean indicating if it exists
func (r *SettingRegistry) GetSetting(key string) (SettingDefinition, bool) {
	setting, exists := r.settings[key]
	return setting, exists
}

// ListSettingsByCategory returns all setting definitions for a specific category
func (r *SettingRegistry) ListSettingsByCategory(category string) []SettingDefinition {
	var result []SettingDefinition
	for _, setting := range r.settings {
		if setting.Category == category {
			result = append(result, setting)
		}
	}
	return result
}

// ListAllSettings returns all registered setting definitions
func (r *SettingRegistry) ListAllSettings() []SettingDefinition {
	var result []SettingDefinition
	for _, setting := range r.settings {
		result = append(result, setting)
	}
	return result
}

// ListCategories returns all unique categories that have registered settings
func (r *SettingRegistry) ListCategories() []string {
	categories := make(map[string]bool)
	for _, setting := range r.settings {
		categories[setting.Category] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// HasSetting checks if a setting with the given key exists in the registry
func (r *SettingRegistry) HasSetting(key string) bool {
	_, exists := r.settings[key]
	return exists
}

// RemoveSetting removes a setting definition from the registry
// Returns true if the setting was found and removed, false otherwise
func (r *SettingRegistry) RemoveSetting(key string) bool {
	if _, exists := r.settings[key]; exists {
		delete(r.settings, key)
		return true
	}
	return false
}
