package groups

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/actions"
)

// Service provides group management operations.
type Service struct {
	groups *Repository

	actionsSvc *actions.Service
}

// NewService creates a new instance of the service.
func NewService(groups *Repository, actionsSvc *actions.Service) *Service {
	return &Service{
		groups: groups,

		actionsSvc: actionsSvc,
	}
}

// CreateOrUpdate creates a new group record or updates an existing one.
func (s *Service) CreateOrUpdate(ctx context.Context, group Group, admin Admin) error {
	if err := s.groups.CreateOrUpdate(ctx, &group, []Admin{admin}); err != nil {
		return fmt.Errorf("failed to create or update group: %w", err)
	}

	// Log the action
	s.actionsSvc.LogAction(
		ctx,
		"group.enabled",
		0,
		0,
		fmt.Sprintf("Enable group %q with telegram ID %d", group.Title, group.TelegramID),
	)

	return nil
}

// Disable implements Service.
func (s *Service) Disable(ctx context.Context, telegramID int64) error {
	if err := s.groups.UpdateStatus(ctx, telegramID, false); err != nil {
		return fmt.Errorf("failed to disable group: %w", err)
	}

	// Log the action
	s.actionsSvc.LogAction(ctx, "group.disabled", 0, 0, fmt.Sprintf("Disable group with telegram ID %d", telegramID))

	return nil
}
