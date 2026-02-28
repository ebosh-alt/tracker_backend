package meapi

import (
	"time"

	domainUser "tracker/internal/domain/user"
)

type userDTO struct {
	ID        int64  `json:"id"`
	TgID      int64  `json:"tgId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
	Timezone  string `json:"timezone"`
	StepsGoal int    `json:"stepsGoal"`
	Streak    int    `json:"streak"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type updateSettingsResponseDTO struct {
	User userDTO `json:"user"`
}

func mapUser(user domainUser.User) userDTO {
	return userDTO{
		ID:        user.ID,
		TgID:      user.TelegramID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Timezone:  user.Timezone,
		StepsGoal: user.StepsGoal.Int(),
		Streak:    user.Streak,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapUpdateSettingsResponse(user domainUser.User) updateSettingsResponseDTO {
	return updateSettingsResponseDTO{User: mapUser(user)}
}
