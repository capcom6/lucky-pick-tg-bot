package users

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	users  *Repository
	logger *zap.Logger
}

func NewService(users *Repository, logger *zap.Logger) *Service {
	return &Service{
		users:  users,
		logger: logger,
	}
}

// RegisterUser creates or updates a user from Telegram user data.
func (s *Service) RegisterUser(ctx context.Context, user UserIn) (*User, error) {
	model := NewUserModel(
		user.TelegramUserID,
		user.Username,
		user.FirstName,
		user.LastName,
	)

	// Create or update user in database
	if err := s.users.CreateOrUpdate(ctx, model); err != nil {
		s.logger.Error("Failed to create or update user",
			zap.Int64("telegram_user_id", user.TelegramUserID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	s.logger.Info("User registered successfully",
		zap.Int64("id", model.ID),
		zap.Int64("telegram_user_id", model.TelegramUserID),
		zap.String("username", model.Username),
		zap.String("first_name", model.FirstName))

	return &User{
		UserIn: UserIn{
			TelegramUserID: model.TelegramUserID,
			Username:       model.Username,
			FirstName:      model.FirstName,
			LastName:       model.LastName,
		},
		ID:           model.ID,
		RegisteredAt: model.RegisteredAt,
		IsActive:     model.IsActive,
	}, nil
}

// // GetUser retrieves a user by their Telegram ID
// func (s *Service) GetUser(ctx context.Context, telegramID int64) (*UserModel, error) {
// 	user, err := s.users.GetByTelegramID(ctx, telegramID)
// 	if err != nil {
// 		s.logger.Warn("User not found",
// 			zap.Int64("telegram_user_id", telegramID),
// 			zap.Error(err))
// 		return nil, fmt.Errorf("user not found: %w", err)
// 	}

// 	return user, nil
// }

// // IsUserRegistered checks if a user is registered and active
// func (s *Service) IsUserRegistered(ctx context.Context, telegramID int64) (bool, error) {
// 	isRegistered, err := s.users.IsUserRegistered(ctx, telegramID)
// 	if err != nil {
// 		s.logger.Error("Failed to check user registration",
// 			zap.Int64("telegram_user_id", telegramID),
// 			zap.Error(err))
// 		return false, fmt.Errorf("failed to check user registration: %w", err)
// 	}

// 	return isRegistered, nil
// }

// // RegisterOrGetUser registers a new user or retrieves existing one
// func (s *Service) RegisterOrGetUser(ctx context.Context, telegramUser *models.User) (*UserModel, error) {
// 	if telegramUser == nil {
// 		return nil, fmt.Errorf("telegram user is nil")
// 	}

// 	// First try to get existing user
// 	existingUser, err := s.users.GetByTelegramID(ctx, telegramUser.ID)
// 	if err == nil {
// 		// User exists, return it
// 		return existingUser, nil
// 	}

// 	// User doesn't exist, register them
// 	return s.RegisterUser(ctx, telegramUser)
// }
