package settings

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SettingType represents the type of a setting value
type SettingType string

const (
	Text     SettingType = "text"
	Number   SettingType = "number"
	Boolean  SettingType = "boolean"
	Duration SettingType = "duration"
)

// SettingDefinition defines a group setting with all its metadata
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
	DefaultValue interface{} `json:"default_value"`
	// Validation contains rules for validating setting values
	Validation *SettingValidation `json:"validation,omitempty"`
	// Options defines available choices for dropdown/enum settings
	Options []SettingOption `json:"options,omitempty"`
}

// SettingValidation defines validation rules for setting values
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

// SettingOption defines a choice for dropdown/enum settings
type SettingOption struct {
	// Label is the human-readable option name
	Label string `json:"label"`
	// Value is the actual value stored for this option
	Value interface{} `json:"value"`
}

// DurationValue represents a time duration with parsing and formatting utilities
type DurationValue struct {
	time.Duration
}

// String returns the duration in HH:MM:SS format
func (d DurationValue) String() string {
	if d.Duration == 0 {
		return "00:00:00"
	}

	hours := int(d.Duration.Hours())
	minutes := int(d.Duration.Minutes()) % 60
	seconds := int(d.Duration.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// ParseDuration parses a duration from HH:MM:SS format
func ParseDuration(s string) (DurationValue, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return DurationValue{}, fmt.Errorf("invalid duration format, expected HH:MM:SS")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return DurationValue{}, fmt.Errorf("invalid hours: %v", err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return DurationValue{}, fmt.Errorf("invalid minutes: %v", err)
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return DurationValue{}, fmt.Errorf("invalid seconds: %v", err)
	}

	duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

	return DurationValue{duration}, nil
}

func Ptr[T any](value T) *T {
	return &value
}
