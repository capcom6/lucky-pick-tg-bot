package settings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// SettingType represents the type of a setting value.
type SettingType string

const (
	Text     SettingType = "text"
	Number   SettingType = "number"
	Boolean  SettingType = "boolean"
	Duration SettingType = "duration"
)

// SettingDefinition defines a group setting with all its metadata.
type SettingDefinition struct {
	// Key is the unique identifier for the setting (e.g., "discussions.delay")
	Key string `json:"key"`
	// Category is the logical group this setting belongs to (e.g., "discussions")
	Category string `json:"category"`
	// Label is the human-readable name displayed to users
	Label string `json:"label"`
	// Description explains what this setting does
	Description string `json:"description"`
	// Type determines how the setting value is validated and displayed
	Type SettingType `json:"type"`
	// DefaultValue is the value used when no custom value is set
	DefaultValue string `json:"default_value"`
	// Validation contains rules for validating setting values
	Validation *SettingValidation `json:"validation,omitempty"`
	// Options defines available choices for dropdown/enum settings
	Options []SettingOption `json:"options,omitempty"`
}

func (s SettingDefinition) Format(currentValue string) string {
	if s.Type == Boolean {
		switch currentValue {
		case "true":
			return "✅ True"
		case "false":
			return "❌ False"
		}
	}

	return currentValue
}

func (s SettingDefinition) Validate(value string) error {
	v, err := s.parseValue(value)
	if err != nil {
		return fmt.Errorf("invalid setting value: %w", err)
	}

	return s.Validation.validateAny(v)
}

func (s SettingDefinition) parseValue(value string) (any, error) {
	switch s.Type {
	case Text:
		return value, nil
	case Number:
		return parseNumber(value)
	case Boolean:
		return parseBoolean(value)
	case Duration:
		return ParseDuration(value)
	default:
		return nil, fmt.Errorf("%w: unknown setting type: %s", ErrValidationFailed, s.Type)
	}
}

// SettingValidation defines validation rules for setting values.
type SettingValidation struct {
	// MinValue is the minimum allowed value for numeric settings
	MinValue *float64 `json:"min_value,omitempty"`
	// MaxValue is the maximum allowed value for numeric settings
	MaxValue *float64 `json:"max_value,omitempty"`
	// MinLength is the minimum length for text settings
	MinLength *int `json:"min_length,omitempty"`
	// MaxLength is the maximum length for text settings
	MaxLength *int `json:"max_length,omitempty"`
	// Pattern is a regex pattern for text validation
	Pattern *string `json:"pattern,omitempty"`
	// Required indicates if this setting must have a value
	Required bool `json:"required"`
}

func (s *SettingValidation) validateAny(value any) error {
	if s == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return s.validateText(v)
	case float64:
		return s.validateNumber(v)
	case bool:
		return s.validateBoolean(v)
	case time.Duration:
		return s.validateDuration(v)
	case DurationValue:
		return s.validateDuration(v.Duration)
	default:
		return fmt.Errorf("%w: unknown setting value type: %T", ErrValidationFailed, value)
	}
}

// validateText validates a text setting value.
func (s *SettingValidation) validateText(value string) error {
	if s == nil {
		return nil
	}

	value = strings.TrimSpace(value)

	if !s.Required && value == "" {
		return nil
	}

	if s.Required && value == "" {
		return fmt.Errorf("%w: value is required", ErrValidationFailed)
	}

	if s.MinLength != nil && utf8.RuneCountInString(value) < *s.MinLength {
		return fmt.Errorf("%w: value must be at least %d characters long", ErrValidationFailed, *s.MinLength)
	}

	if s.MaxLength != nil && utf8.RuneCountInString(value) > *s.MaxLength {
		return fmt.Errorf("%w: value must be at most %d characters long", ErrValidationFailed, *s.MaxLength)
	}

	if s.Pattern != nil {
		matched, err := regexp.MatchString(*s.Pattern, value)
		if err != nil {
			return fmt.Errorf("%w: invalid validation pattern: %w", ErrValidationFailed, err)
		}
		if !matched {
			return fmt.Errorf("%w: value does not match required pattern", ErrValidationFailed)
		}
	}

	return nil
}

// validateNumber validates a numeric setting value.
func (s *SettingValidation) validateNumber(value float64) error {
	if s == nil {
		return nil
	}

	if !s.Required && value == 0 {
		return nil
	}

	if s.Required && value == 0 {
		return fmt.Errorf("%w: value is required", ErrValidationFailed)
	}

	if s.MinValue != nil && value < *s.MinValue {
		return fmt.Errorf("%w: value must be at least %f", ErrValidationFailed, *s.MinValue)
	}

	if s.MaxValue != nil && value > *s.MaxValue {
		return fmt.Errorf("%w: value must be at most %f", ErrValidationFailed, *s.MaxValue)
	}

	return nil
}

// validateDuration validates a duration setting value.
func (s *SettingValidation) validateDuration(value time.Duration) error {
	if s == nil {
		return nil
	}

	if !s.Required && value == 0 {
		return nil
	}

	if s.Required && value == 0 {
		return fmt.Errorf("%w: value is required", ErrValidationFailed)
	}

	if s.MinValue != nil && value < time.Second*time.Duration(*s.MinValue) {
		return fmt.Errorf(
			"%w: duration must be at least %v",
			ErrValidationFailed,
			DurationValue{time.Second * time.Duration(*s.MinValue)},
		)
	}

	if s.MaxValue != nil && value > time.Second*time.Duration(*s.MaxValue) {
		return fmt.Errorf(
			"%w: duration must be at most %v",
			ErrValidationFailed,
			DurationValue{time.Second * time.Duration(*s.MaxValue)},
		)
	}

	return nil
}

func (s *SettingValidation) validateBoolean(_ bool) error {
	return nil
}

// SettingOption defines a choice for dropdown/enum settings.
type SettingOption struct {
	// Label is the human-readable option name
	Label string `json:"label"`
	// Value is the actual value stored for this option
	Value any `json:"value"`
}

// DurationValue represents a time duration with parsing and formatting utilities.
type DurationValue struct {
	time.Duration
}

// String returns the duration in HH:MM:SS format.
func (d DurationValue) String() string {
	if d.Duration == 0 {
		return "00:00:00"
	}

	const (
		minutesPerHour   = 60
		secondsPerMinute = 60
	)

	hours := int(d.Duration.Hours())
	minutes := int(d.Duration.Minutes()) % minutesPerHour
	seconds := int(d.Duration.Seconds()) % secondsPerMinute

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// ParseDuration parses a duration from HH:MM:SS format.
func ParseDuration(s string) (DurationValue, error) {
	const partsCount = 3

	parts := strings.Split(s, ":")
	if len(parts) != partsCount {
		return DurationValue{}, fmt.Errorf("%w: invalid duration format, expected HH:MM:SS", ErrValidationFailed)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return DurationValue{}, fmt.Errorf("%w: invalid hours: %w", ErrValidationFailed, err)
	}
	if hours < 0 {
		return DurationValue{}, fmt.Errorf("%w: hours must be non-negative", ErrValidationFailed)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return DurationValue{}, fmt.Errorf("%w: invalid minutes: %w", ErrValidationFailed, err)
	}
	if minutes < 0 || minutes > 59 {
		return DurationValue{}, fmt.Errorf("%w: minutes must be between 0 and 59", ErrValidationFailed)
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return DurationValue{}, fmt.Errorf("%w: invalid seconds: %w", ErrValidationFailed, err)
	}
	if seconds < 0 || seconds > 59 {
		return DurationValue{}, fmt.Errorf("%w: seconds must be between 0 and 59", ErrValidationFailed)
	}

	duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

	return DurationValue{duration}, nil
}

func Ptr[T any](value T) *T {
	return &value
}
