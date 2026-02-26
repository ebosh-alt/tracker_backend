package telegram

import (
	"fmt"

	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"
)

// InitDataVerifierAdapter адаптирует Telegram client к доменному порту verify.
type InitDataVerifierAdapter struct {
	client *Client
}

// NewInitDataVerifierAdapter создает verifier-адаптер.
func NewInitDataVerifierAdapter(client *Client) *InitDataVerifierAdapter {
	return &InitDataVerifierAdapter{client: client}
}

// VerifyInitData валидирует initData и возвращает профиль пользователя.
func (a *InitDataVerifierAdapter) VerifyInitData(initData string) (*identity.VerifiedTelegram, error) {
	payload, err := a.client.VerifyInitData(initData)
	if err != nil {
		return nil, err
	}

	userPayload, ok := ExtractUser(payload)
	if !ok {
		return nil, fmt.Errorf("%w: user payload is missing", shared.ErrInvalidInput)
	}

	telegramID, ok := parseInt64Any(userPayload["id"])
	if !ok || telegramID <= 0 {
		return nil, fmt.Errorf("%w: telegram id is invalid", shared.ErrInvalidInput)
	}

	firstName, ok := parseStringAny(userPayload["first_name"])
	if !ok || firstName == "" {
		return nil, fmt.Errorf("%w: first name is required", shared.ErrInvalidInput)
	}

	username, _ := parseStringAny(userPayload["username"])
	lastName, hasLastName := parseStringAny(userPayload["last_name"])

	var lastNamePtr *string
	if hasLastName && lastName != "" {
		lastNamePtr = &lastName
	}

	return &identity.VerifiedTelegram{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastNamePtr,
	}, nil
}

func parseInt64Any(v any) (int64, bool) {
	switch value := v.(type) {
	case float64:
		return int64(value), true
	case int64:
		return value, true
	case int:
		return int64(value), true
	default:
		return 0, false
	}
}

func parseStringAny(v any) (string, bool) {
	value, ok := v.(string)
	if !ok {
		return "", false
	}
	return value, true
}
