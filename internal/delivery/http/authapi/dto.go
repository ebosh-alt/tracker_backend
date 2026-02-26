package authapi

import (
	"time"

	appauth "tracker/internal/application/auth"
)

type tokenPairDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

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

type telegramAuthResponseDTO struct {
	User  userDTO      `json:"user"`
	Token tokenPairDTO `json:"token"`
}

func mapTelegramAuthResponse(out appauth.TelegramAuthOutput) telegramAuthResponseDTO {
	return telegramAuthResponseDTO{
		User: userDTO{
			ID:        out.User.ID,
			TgID:      out.User.TelegramID,
			Username:  out.User.Username,
			FirstName: out.User.FirstName,
			LastName:  out.User.LastName,
			Timezone:  out.User.Timezone,
			StepsGoal: out.User.StepsGoal.Int(),
			Streak:    out.User.Streak,
			CreatedAt: out.User.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: out.User.UpdatedAt.UTC().Format(time.RFC3339),
		},
		Token: tokenPairDTO{
			AccessToken:  out.Token.AccessToken,
			RefreshToken: out.Token.RefreshToken,
		},
	}
}

type refreshResponseDTO struct {
	Token tokenPairDTO `json:"token"`
}

func mapRefreshResponse(out appauth.RefreshOutput) refreshResponseDTO {
	return refreshResponseDTO{
		Token: tokenPairDTO{
			AccessToken:  out.Token.AccessToken,
			RefreshToken: out.Token.RefreshToken,
		},
	}
}
