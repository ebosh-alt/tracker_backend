package identity

import (
	"context"
)

type VerifiedTelegram struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   *string
}

type TelegramVerifier interface {
	VerifyInitData(initData string) (*VerifiedTelegram, error)
}

type SessionReader interface {
	GetByRefreshJTI(ctx context.Context, refreshJTI string) (*Session, error)
}

type SessionWriter interface {
	Create(ctx context.Context, s *Session) (*Session, error)
	Rotate(ctx context.Context, oldSessionID int64, next *Session) (*Session, error)
	RevokeAllByUser(ctx context.Context, userID int64, reason string) error
}
