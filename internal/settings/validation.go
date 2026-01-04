package settings

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ValidateSetting validates a setting value based on its definition
func ValidateSetting(key string, value interface{}, registry *SettingRegistry) error {
	setting, exists := registry.GetSetting(key)
	if !exists {
		return fmt.Errorf("unknown setting: %s", key)
	}

	return validateValue(value, setting)
}

// validateValue routes validation to the appropriate type-specific validator
func validateValue(value interface{}, setting SettingDefinition) error {
	switch setting.Type {
	case Text:
		return validateText(value, setting.Validation)
	case Number:
		return validateNumber(value, setting.Validation)
	case Boolean:
		return validateBoolean(value)
	case Duration:
		return validateDuration(value, setting.Validation)
	default:
		return fmt.Errorf("unknown setting type: %s", setting.Type)
	}
}

// validateText validates text-based setting values
func validateText(value interface{}, validation *SettingValidation) error {
	if value == nil {
		return nil // Let Required check handle nil values
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value for text setting, got %T", value)
	}

	if validation != nil {
		if validation.Required && str == "" {
			return fmt.Errorf("value is required")
		}

		if validation.MinLength != nil && len(str) < *validation.MinLength {
			return fmt.Errorf("value must be at least %d characters long", *validation.MinLength)
		}

		if validation.MaxLength != nil && len(str) > *validation.MaxLength {
			return fmt.Errorf("value must be at most %d characters long", *validation.MaxLength)
		}

		if validation.Pattern != nil {
			matched, err := regexp.MatchString(*validation.Pattern, str)
			if err != nil {
				return fmt.Errorf("invalid validation pattern: %v", err)
			}
			if !matched {
				return fmt.Errorf("value does not match required pattern")
			}
		}
	}

	return nil
}

// validateNumber validates numeric setting values
func validateNumber(value interface{}, validation *SettingValidation) error {
	if value == nil {
		return nil // Let Required check handle nil values
	}

	var num float64
	var isInt bool

	switch v := value.(type) {
	case int:
		num = float64(v)
		isInt = true
	case int32:
		num = float64(v)
		isInt = true
	case int64:
		num = float64(v)
		isInt = true
	case float32:
		num = float64(v)
	case float64:
		num = v
	case string:
		var err error
		num, err = parseNumber(v)
		if err != nil {
			return fmt.Errorf("invalid number format: %v", err)
		}
	default:
		return fmt.Errorf("expected numeric value, got %T", value)
	}

	if validation != nil {
		if validation.Required && ((isInt && num == 0) || (!isInt && num == 0.0)) {
			return fmt.Errorf("value is required")
		}

		if validation.MinValue != nil && num < *validation.MinValue {
			return fmt.Errorf("value must be at least %v", *validation.MinValue)
		}

		if validation.MaxValue != nil && num > *validation.MaxValue {
			return fmt.Errorf("value must be at most %v", *validation.MaxValue)
		}
	}

	return nil
}

// parseNumber parses a string to float64, supporting integers and decimals
func parseNumber(s string) (float64, error) {
	// Try to parse as float64 directly
	return strconv.ParseFloat(s, 64)
}

// validateBoolean validates boolean setting values
func validateBoolean(value interface{}) error {
	if value == nil {
		return nil // Booleans are never required since false is a valid value
	}

	switch v := value.(type) {
	case bool:
		return nil
	case string:
		_, err := strconv.ParseBool(v)
		return err
	default:
		return fmt.Errorf("expected boolean value, got %T", value)
	}
}

// validateDuration validates duration setting values
func validateDuration(value interface{}, validation *SettingValidation) error {
	if value == nil {
		return nil // Let Required check handle nil values
	}

	var duration time.Duration

	switch v := value.(type) {
	case time.Duration:
		duration = v
	case string:
		// Parse HH:MM:SS format
		durationValue, err := ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid duration format: %v", err)
		}
		duration = durationValue.Duration
	case DurationValue:
		duration = v.Duration
	default:
		return fmt.Errorf("expected duration value, got %T", value)
	}

	if validation != nil {
		if validation.Required && duration == 0 {
			return fmt.Errorf("value is required")
		}

		if validation.MinValue != nil && duration < time.Second*time.Duration(*validation.MinValue) {
			return fmt.Errorf("duration must be at least %v", time.Second*time.Duration(*validation.MinValue))
		}

		if validation.MaxValue != nil && duration > time.Second*time.Duration(*validation.MaxValue) {
			return fmt.Errorf("duration must be at most %v", time.Second*time.Duration(*validation.MaxValue))
		}
	}

	return nil
}

// ConvertValue converts a string value to the appropriate type based on the setting definition
func ConvertValue(value string, setting SettingDefinition) (interface{}, error) {
	switch setting.Type {
	case Text:
		return value, nil
	case Number:
		return parseNumber(value)
	case Boolean:
		return strconv.ParseBool(value)
	case Duration:
		return ParseDuration(value)
	default:
		return nil, fmt.Errorf("unknown setting type: %s", setting.Type)
	}
}
