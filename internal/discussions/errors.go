package discussions

import "errors"

var (
	ErrLLMFailed = errors.New("llm failed")
	ErrNotFound  = errors.New("discussion not found")
)
