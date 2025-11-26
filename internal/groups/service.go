package groups

import (
	"context"
	"fmt"
)

// Service provides group management operations.
type Service struct {
	groups *Repository
}

// NewService creates a new instance of the service.
func NewService(groups *Repository) *Service {
	return &Service{groups: groups}
}

// CreateOrUpdate creates a new group record or updates an existing one.
func (s *Service) CreateOrUpdate(ctx context.Context, telegramID int64, title string) error {
	if err := s.groups.CreateOrUpdate(ctx, &Group{
		TelegramID: telegramID,
		Title:      title,
	}); err != nil {
		return fmt.Errorf("failed to create or update group: %w", err)
	}

	return nil
}

// Disable implements Service.
func (s *Service) Disable(ctx context.Context, telegramID int64) error {
	if err := s.groups.UpdateStatus(ctx, telegramID, false); err != nil {
		return fmt.Errorf("failed to disable group: %w", err)
	}

	return nil
}
