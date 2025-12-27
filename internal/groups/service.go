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
func (s *Service) CreateOrUpdate(ctx context.Context, group GroupDraft, admin Admin) error {
	if err := s.groups.CreateOrUpdate(ctx, &group, []Admin{admin}); err != nil {
		return fmt.Errorf("failed to create or update group: %w", err)
	}

	// Log the action
	s.actionsSvc.LogAction(
		ctx,
		"group.enabled",
		admin.UserID,
		0,
		fmt.Sprintf("Enable group %q with telegram ID %d", group.Title, group.TelegramID),
	)

	return nil
}

func (s *Service) SelectByIDs(ctx context.Context, ids []int64) ([]GroupWithSettings, error) {
	return s.groups.SelectByIDs(ctx, ids)
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

// GetUserAdminGroups returns groups where the user is an admin.
func (s *Service) GetUserAdminGroups(ctx context.Context, userID int64) ([]GroupWithSettings, error) {
	return s.groups.GetByUser(ctx, userID)
}

// IsAdmin returns true if the user is an admin of the group.
func (s *Service) IsAdmin(ctx context.Context, groupID int64, userID int64) (bool, error) {
	return s.groups.IsAdmin(ctx, groupID, userID)
}

// GetByID returns a group by its ID.
func (s *Service) GetByID(ctx context.Context, groupID int64) (*Group, error) {
	return s.groups.GetByID(ctx, groupID)
}

// UpdateSettings updates the settings of a group.
func (s *Service) UpdateSettings(ctx context.Context, groupID int64, settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	return s.groups.UpdateSettings(ctx, groupID, settings)
}
