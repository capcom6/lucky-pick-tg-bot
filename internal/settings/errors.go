package settings

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrValidationFailed = errors.New("validation failed")
)
