package settings

import (
	"fmt"
	"strconv"
)

// parseNumber parses a string to float64, supporting integers and decimals.
func parseNumber(s string) (float64, error) {
	// Try to parse as float64 directly
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %w", err)
	}

	return f, nil
}

// parseBoolean parses a string to bool.
func parseBoolean(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("invalid boolean format: %w", err)
	}
	return b, nil
}
