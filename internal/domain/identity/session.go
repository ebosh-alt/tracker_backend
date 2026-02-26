package identity

import (
	"fmt"
	"time"
	"tracker/internal/domain/shared"
)

type Meta struct {
	RotatedFromID *int64
	RevokedAt     *time.Time
	RevokeReason  *string
	IP            *string
	UserAgent     *string
}

type Claim struct {
	JTI       string
	TokenHash string
	ExpiresAt time.Time
}

type Session struct {
	ID     int64
	UserID int64

	Refresh Claim
	Access  Claim
	Meta    Meta

	IssuedAt  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s Session) Validate() error {
	if s.UserID <= 0 {
		return fmt.Errorf("%w: invalid user ID", shared.ErrInvalidInput)
	}
	if s.Access.JTI == "" || s.Refresh.JTI == "" {
		return fmt.Errorf("%w: JTI cannot be empty", shared.ErrInvalidInput)
	}
	if s.Access.TokenHash == "" || s.Refresh.TokenHash == "" {
		return fmt.Errorf("%w: token hash cannot be empty", shared.ErrInvalidInput)
	}
	if s.Refresh.ExpiresAt.Before(s.Access.ExpiresAt) {
		return fmt.Errorf("%w: refresh expires before access", shared.ErrInvalidInput)
	}
	return nil
}

func (s Session) IsActive(now time.Time) bool {
	return s.Meta.RevokedAt == nil && s.Refresh.ExpiresAt.After(now.UTC())
}

func (s Session) CanRotate(now time.Time) error {
	if s.Meta.RevokedAt != nil {
		return fmt.Errorf("%w: session revoked", shared.ErrUnauthorized)
	}
	if !s.Refresh.ExpiresAt.After(now.UTC()) {
		return fmt.Errorf("%w: refresh token expired", shared.ErrUnauthorized)
	}
	return nil
}
