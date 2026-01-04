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
func (s *Service) GetSetting(ctx context.Context, groupID int64, key string) (interface{}, bool, error) {
	// Get setting definition to know the type
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return nil, false, fmt.Errorf("unknown setting key: %s", key)
	}

	// Get raw string value from repository
	rawValue, err := s.groupsSvc.GetSetting(ctx, groupID, key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get setting %s: %w", key, err)
	}

	// If no value is set, return default value
	if rawValue == "" {
		return def.DefaultValue, false, nil
	}

	// Convert to proper type
	converted, err := ConvertValue(rawValue, def)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert setting %s value: %w", key, err)
	}

	return converted, true, nil
}

// GetAllSettings retrieves all settings for a group with proper type conversion.
// Returns a map of setting keys to converted values, plus a map of which settings have custom values.
func (s *Service) GetAllSettings(ctx context.Context, groupID int64) (map[string]interface{}, map[string]bool, error) {
	// Get all raw settings from repository
	rawSettings, err := s.groupsSvc.GetAllSettings(ctx, groupID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get all settings: %w", err)
	}

	result := make(map[string]interface{})
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
			converted, err := ConvertValue(rawValue, def)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert setting %s: %w", key, err)
			}
			result[key] = converted
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
func (s *Service) UpdateSetting(ctx context.Context, groupID int64, userID int64, key string, value interface{}) error {
	// Validate the setting exists
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return fmt.Errorf("unknown setting key: %s", key)
	}

	// Validate the value
	if err := validateValue(value, def); err != nil {
		return fmt.Errorf("invalid setting value: %w", err)
	}

	// Convert value to string for storage
	var stringValue string
	if value != nil {
		switch v := value.(type) {
		case string:
			stringValue = v
		case bool:
			if v {
				stringValue = "true"
			} else {
				stringValue = "false"
			}
		case int, int32, int64:
			stringValue = fmt.Sprintf("%d", v)
		case float32, float64:
			stringValue = fmt.Sprintf("%g", v)
		case DurationValue:
			stringValue = v.String()
		default:
			return fmt.Errorf("unsupported value type: %T", v)
		}
	}

	// Update in repository
	if err := s.groupsSvc.UpdateSetting(ctx, groupID, key, stringValue); err != nil {
		return fmt.Errorf("failed to update setting %s: %w", key, err)
	}

	// Log the action (if actions service is available)
	// Note: This would require injecting the actions service, keeping it simple for now

	return nil
}

// UpdateSettings updates multiple settings for a group at once with validation.
func (s *Service) UpdateSettings(
	ctx context.Context,
	groupID int64,
	userID int64,
	settings map[string]interface{},
) error {
	// Validate all settings first
	validatedSettings := make(map[string]string)

	for key, value := range settings {
		// Validate setting exists
		def, exists := s.registry.GetSetting(key)
		if !exists {
			return fmt.Errorf("unknown setting key: %s", key)
		}

		// Validate value
		if err := validateValue(value, def); err != nil {
			return fmt.Errorf("invalid setting %s value: %w", key, err)
		}

		// Convert to string
		var stringValue string
		if value != nil {
			switch v := value.(type) {
			case string:
				stringValue = v
			case bool:
				if v {
					stringValue = "true"
				} else {
					stringValue = "false"
				}
			case int, int32, int64:
				stringValue = fmt.Sprintf("%d", v)
			case float32, float64:
				stringValue = fmt.Sprintf("%g", v)
			case DurationValue:
				stringValue = v.String()
			default:
				return fmt.Errorf("unsupported value type for %s: %T", key, v)
			}
		}

		validatedSettings[key] = stringValue
	}

	// Update all settings in repository
	if err := s.groupsSvc.UpdateSettings(ctx, groupID, validatedSettings); err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// ValidateSetting validates a setting value without updating.
func (s *Service) ValidateSetting(key string, value interface{}) error {
	def, exists := s.registry.GetSetting(key)
	if !exists {
		return fmt.Errorf("unknown setting key: %s", key)
	}

	return validateValue(value, def)
}

// ValidateSettingValue validates a setting value (alias for ValidateSetting).
func (s *Service) ValidateSettingValue(key string, value interface{}) error {
	return s.ValidateSetting(key, value)
}

// SetSetting sets a setting value for a group (alias for UpdateSetting without userID).
func (s *Service) SetSetting(ctx context.Context, groupID int64, key string, value interface{}) error {
	return s.UpdateSetting(ctx, groupID, 0, key, value)
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
