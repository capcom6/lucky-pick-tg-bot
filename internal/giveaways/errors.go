package giveaways

import "errors"

var (
	ErrNotFound              = errors.New("giveaway not found")
	ErrNotEnoughParticipants = errors.New("not enough participants")
)
