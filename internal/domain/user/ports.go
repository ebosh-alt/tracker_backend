package user

import "context"

// Repository определяет операции хранения пользователя.
// Интерфейс объявлен в домене, реализации остаются в инфраструктуре.
type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (User, error)
	Save(ctx context.Context, user User) (User, error)
	UpdateSettings(ctx context.Context, userID int64, patch SettingsPatch) (User, error)
	UpsertFromTelegram(ctx context.Context, cmd *TelegramProfile) (*UpsertFromTelegramResult, error)
	AuthRefresh(ctx context.Context, userID int64) error
}
