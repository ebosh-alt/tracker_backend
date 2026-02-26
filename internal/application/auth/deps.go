package auth

import (
	"context"
	"tracker/internal/domain/identity"
)

type SessionRepository interface {
	Create(ctx context.Context, s *identity.Session) (*identity.Session, error)
	GetByRefreshJTI(ctx context.Context, refreshJTI string) (*identity.Session, error)
	Rotate(ctx context.Context, oldSessionID int64, next *identity.Session) (*identity.Session, error)
}

type TokenHasher interface {
	Hash(raw string) (string, error)
	Equal(raw, hashed string) (bool, error)
}
