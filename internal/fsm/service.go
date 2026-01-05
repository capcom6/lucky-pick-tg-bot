package fsm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-core-fx/cachefx/cache"
	"go.uber.org/zap"
)

type Service struct {
	storage *cache.Typed[*State]

	logger *zap.Logger
}

func NewService(storage cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		storage: cache.NewTyped[*State](storage),
		logger:  logger,
	}
}

func (s *Service) Get(ctx context.Context, userID int64) (*State, error) {
	item, err := s.storage.Get(ctx, strconv.FormatInt(userID, 10))
	if errors.Is(err, cache.ErrKeyNotFound) {
		return &State{Name: "", Data: map[string]string{}}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}

	return item, nil
}

func (s *Service) Set(ctx context.Context, userID int64, state *State) error {
	if err := s.storage.Set(ctx, strconv.FormatInt(userID, 10), state); err != nil {
		return fmt.Errorf("set state: %w", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, userID int64) error {
	if err := s.storage.Delete(ctx, strconv.FormatInt(userID, 10)); err != nil {
		if errors.Is(err, cache.ErrKeyNotFound) {
			return nil
		}
		return fmt.Errorf("delete state: %w", err)
	}

	return nil
}
