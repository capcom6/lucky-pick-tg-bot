package settings

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
)

// Service provides business logic for group settings management.
type Service struct {
	groupsSvc *groups.Service
	registry  *SettingRegistry
}

// NewService creates a new instance of the settings service.
func NewService(groupsSvc *groups.Service, registry *SettingRegistry) *Service {
	return &Service{
		groupsSvc: groupsSvc,
		registry:  registry,
	}
}

func (s *Service) RegisterDefinition(def SettingDefinition) {
	s.registry.RegisterSetting(def)
}

// GetSetting retrieves a setting value for a group with proper type conversion.
// Returns the converted value and whether the setting exists.
func (s *Service) GetSetting(ctx context.Context, groupID int64, key string) (string, bool, error) {
	// Get setting definition to know the type
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return "", false, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	// Get raw string value from repository
	rawValue, err := s.groupsSvc.GetSetting(ctx, groupID, key)
	if err != nil {
		return "", false, fmt.Errorf("failed to get setting %s: %w", key, err)
	}

	// If no value is set, return default value
	if rawValue == "" {
		return def.DefaultValue, false, nil
	}

	return rawValue, true, nil
}

// GetAllSettings retrieves all settings for a group with proper type conversion.
// Returns a map of setting keys to converted values, plus a map of which settings have custom values.
func (s *Service) GetAllSettings(ctx context.Context, groupID int64) (map[string]string, map[string]bool, error) {
	// Get all raw settings from repository
	rawSettings, err := s.groupsSvc.GetAllSettings(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get all settings: %w", err)
	}

	result := make(map[string]string)
	hasCustomValue := make(map[string]bool)

	// Convert each setting
	for key, rawValue := range rawSettings {
		def, exists := s.registry.GetSetting(key)
		if !exists {
			// Skip unknown settings for backward compatibility
			continue
		}

		if rawValue == "" {
			result[key] = def.DefaultValue
			hasCustomValue[key] = false
		} else {
			result[key] = rawValue
			hasCustomValue[key] = true
		}
	}

	// Add default values for settings not explicitly set
	for _, def := range s.registry.ListAllSettings() {
		if _, exists := result[def.Key]; !exists {
			result[def.Key] = def.DefaultValue
			hasCustomValue[def.Key] = false
		}
	}

	return result, hasCustomValue, nil
}

// UpdateSetting updates a single setting for a group with validation.
func (s *Service) UpdateSetting(ctx context.Context, groupID int64, key string, value string) error {
	// Validate the setting exists
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	// Validate the value
	if err := def.Validate(value); err != nil {
		return fmt.Errorf("invalid setting value: %w", err)
	}

	// Update in repository
	if err := s.groupsSvc.UpdateSetting(ctx, groupID, key, value); err != nil {
		return fmt.Errorf("failed to update setting %s: %w", key, err)
	}

	// Log the action (if actions service is available)
	// Note: This would require injecting the actions service, keeping it simple for now

	return nil
}

// ValidateSetting validates a setting value without updating.
func (s *Service) ValidateSetting(key string, value string) error {
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	return def.Validate(value)
}

// GetSettingDefinition returns the definition of a setting.
func (s *Service) GetSettingDefinition(key string) (SettingDefinition, bool) {
	return s.registry.GetSetting(key)
}

// ListSettingsByCategory returns all settings in a category.
func (s *Service) ListSettingsByCategory(category string) []SettingDefinition {
	return s.registry.ListSettingsByCategory(category)
}

// ListCategories returns all available setting categories.
func (s *Service) ListCategories() []string {
	return s.registry.ListCategories()
}
