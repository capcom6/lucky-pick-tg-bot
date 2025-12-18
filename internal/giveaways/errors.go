package giveaways

import "errors"

var (
	ErrLLMFailed             = errors.New("llm failed")
	ErrNotEnoughParticipants = errors.New("not enough participants")
	ErrNotFound              = errors.New("giveaway not found")
)
