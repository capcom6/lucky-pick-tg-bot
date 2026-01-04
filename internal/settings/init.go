package settings

// // InitSettingsRegistry creates and initializes a new setting registry with all predefined settings
// func InitSettingsRegistry() *SettingRegistry {
// 	registry := NewSettingRegistry()

// 	// Register discussions settings
// 	registry.RegisterSetting(SettingDefinition{
// 		Key:          "discussions.delay",
// 		Category:     "discussions",
// 		Label:        "Discussion Delay",
// 		Description:  "Minimum time between discussion messages from the same user",
// 		Type:         Duration,
// 		DefaultValue: DurationValue{Duration: 30 * time.Minute},
// 		Validation: &SettingValidation{
// 			MinValue: floatPtr(1),    // Minimum 1 second
// 			MaxValue: floatPtr(3600), // Maximum 1 hour
// 			Required: false,
// 		},
// 	})

// 	// Register giveaways settings
// 	registry.RegisterSetting(SettingDefinition{
// 		Key:          "giveaways.llm_description",
// 		Category:     "giveaways",
// 		Label:        "Use LLM for Descriptions",
// 		Description:  "Generate giveaway descriptions using AI when not provided",
// 		Type:         Boolean,
// 		DefaultValue: true,
// 		Validation: &SettingValidation{
// 			Required: false,
// 		},
// 	})

// 	return registry
// }

// // floatPtr returns a pointer to a float64 value (helper function)
// func floatPtr(f float64) *float64 {
// 	return &f
// }

// // intPtr returns a pointer to an int value (helper function)
// func intPtr(i int) *int {
// 	return &i
// }

// // RegisterAllSettings registers all predefined settings in the given registry
// func RegisterAllSettings(registry *SettingRegistry) {
// 	// If no registry provided, create a new one
// 	if registry == nil {
// 		registry = NewSettingRegistry()
// 	}

// 	// Register all predefined settings
// 	settings := []SettingDefinition{
// 		{
// 			Key:          "discussions.delay",
// 			Category:     "discussions",
// 			Label:        "Discussion Delay",
// 			Description:  "Minimum time between discussion messages from the same user",
// 			Type:         Duration,
// 			DefaultValue: DurationValue{Duration: 30 * time.Minute},
// 			Validation: &SettingValidation{
// 				MinValue: floatPtr(1),    // Minimum 1 second
// 				MaxValue: floatPtr(3600), // Maximum 1 hour
// 				Required: false,
// 			},
// 		},
// 		{
// 			Key:          "giveaways.llm_description",
// 			Category:     "giveaways",
// 			Label:        "Use LLM for Descriptions",
// 			Description:  "Generate giveaway descriptions using AI when not provided",
// 			Type:         Boolean,
// 			DefaultValue: true,
// 			Validation: &SettingValidation{
// 				Required: false,
// 			},
// 		},
// 	}

// 	for _, setting := range settings {
// 		registry.RegisterSetting(setting)
// 	}

// 	// Note: This function can be extended in the future to register additional settings
// 	// as the system grows and new settings are added
// }

// // GetDefaultValue returns the default value for a setting key, converted to string
// func (r *SettingRegistry) GetDefaultValueAsString(key string) string {
// 	setting, exists := r.GetSetting(key)
// 	if !exists {
// 		return ""
// 	}

// 	return formatValue(setting.DefaultValue, setting.Type)
// }

// // formatValue formats a value based on its type for display/storage
// func formatValue(value interface{}, settingType SettingType) string {
// 	if value == nil {
// 		return ""
// 	}

// 	switch settingType {
// 	case Text:
// 		if str, ok := value.(string); ok {
// 			return str
// 		}
// 	case Number:
// 		switch v := value.(type) {
// 		case int:
// 			return strconv.Itoa(v)
// 		case int32:
// 			return strconv.Itoa(int(v))
// 		case int64:
// 			return strconv.FormatInt(v, 10)
// 		case float32:
// 			return strconv.FormatFloat(float64(v), 'f', -1, 32)
// 		case float64:
// 			return strconv.FormatFloat(v, 'f', -1, 64)
// 		case string:
// 			return v
// 		}
// 	case Boolean:
// 		if boolVal, ok := value.(bool); ok {
// 			if boolVal {
// 				return "true"
// 			}
// 			return "false"
// 		}
// 	case Duration:
// 		if duration, ok := value.(time.Duration); ok {
// 			return DurationValue{Duration: duration}.String()
// 		}
// 		if durationValue, ok := value.(DurationValue); ok {
// 			return durationValue.String()
// 		}
// 	}

// 	// Fallback: convert to string
// 	return ""
// }
