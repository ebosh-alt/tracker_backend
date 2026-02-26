package postgres

import (
	"context"
	"errors"
	"strings"
	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sessionCreateSQL = `INSERT INTO auth_sessions (
  user_id,
  access_jti,
  refresh_jti,
  access_token_hash,
  refresh_token_hash,
  issued_at,
  access_expires_at,
  refresh_expires_at,
  rotated_from_id,
  revoked_at,
  revoke_reason,
  ip,
  user_agent,
  created_at,
  updated_at
)
VALUES (
  $1,  -- user_id
  $2,  -- access_jti
  $3,  -- refresh_jti
  $4,  -- access_token_hash
  $5,  -- refresh_token_hash
  $6,  -- issued_at
  $7,  -- access_expires_at
  $8,  -- refresh_expires_at
  $9,  -- rotated_from_id
  $10, -- revoked_at
  $11, -- revoke_reason
  $12, -- ip
  $13, -- user_agent
  NOW(),
  NOW()
)
RETURNING
  id,
  user_id,
  access_jti,
  refresh_jti,
  access_token_hash,
  refresh_token_hash,
  issued_at,
  access_expires_at,
  refresh_expires_at,
  rotated_from_id,
  revoked_at,
  revoke_reason,
  ip,
  user_agent,
  created_at,
  updated_at;`
	sessionGetByRefreshJTISQL = `SELECT
  id,
  user_id,
  access_jti,
  refresh_jti,
  access_token_hash,
  refresh_token_hash,
  issued_at,
  access_expires_at,
  refresh_expires_at,
  rotated_from_id,
  revoked_at,
  revoke_reason,
  ip,
  user_agent,
  created_at,
  updated_at
FROM auth_sessions
WHERE refresh_jti = $1
ORDER BY id DESC LIMIT 1;`
	sessionRotateSQL = `WITH revoked AS (
  UPDATE auth_sessions
  SET
    revoked_at = NOW(),
    revoke_reason = COALESCE($2, 'rotated'),
    updated_at = NOW()
  WHERE id = $1
    AND revoked_at IS NULL
    AND refresh_expires_at > NOW()
  RETURNING id, user_id
)
INSERT INTO auth_sessions (
  user_id,
  access_jti,
  refresh_jti,
  access_token_hash,
  refresh_token_hash,
  issued_at,
  access_expires_at,
  refresh_expires_at,
  rotated_from_id,
  ip,
  user_agent,
  created_at,
  updated_at
)
SELECT
  r.user_id,
  $3,   -- access_jti
  $4,   -- refresh_jti
  $5,   -- access_token_hash
  $6,   -- refresh_token_hash
  $7,   -- issued_at
  $8,   -- access_expires_at
  $9,   -- refresh_expires_at
  r.id, -- rotated_from_id
  $10,  -- ip
  $11,  -- user_agent
  NOW(),
  NOW()
FROM revoked r
RETURNING
  id,
  user_id,
  access_jti,
  refresh_jti,
  access_token_hash,
  refresh_token_hash,
  issued_at,
  access_expires_at,
  refresh_expires_at,
  rotated_from_id,
  revoked_at,
  revoke_reason,
  ip,
  user_agent,
  created_at,
  updated_at;
`
	sessionRevokeAllByUserSQL = `UPDATE auth_sessions
SET
  revoked_at = NOW(),
  revoke_reason = COALESCE($2, 'logout_all'),
  updated_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;`
)

type AuthSessionRepo struct {
	db *pgxpool.Pool
}

func NewAuthSessionRepository(db *pgxpool.Pool) *AuthSessionRepo {
	return &AuthSessionRepo{db: db}
}

func (a AuthSessionRepo) GetByRefreshJTI(ctx context.Context, refreshJTI string) (*identity.Session, error) {
	if strings.TrimSpace(refreshJTI) == "" {
		return nil, shared.ErrInvalidInput
	}
	exec := executorFromContext(ctx, a.db)
	row := exec.QueryRow(ctx, sessionGetByRefreshJTISQL, refreshJTI)
	session, err := scanAuthSessionRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}
	return session, nil
}

func (a AuthSessionRepo) Create(ctx context.Context, s *identity.Session) (*identity.Session, error) {
	if s == nil {
		return nil, shared.ErrInvalidInput
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	exec := executorFromContext(ctx, a.db)
	row := exec.QueryRow(ctx, sessionCreateSQL,
		s.UserID,
		s.Access.JTI,
		s.Refresh.JTI,
		s.Access.TokenHash,
		s.Refresh.TokenHash,
		s.IssuedAt,
		s.Access.ExpiresAt,
		s.Refresh.ExpiresAt,
		s.Meta.RotatedFromID,
		s.Meta.RevokedAt,
		s.Meta.RevokeReason,
		s.Meta.IP,
		s.Meta.UserAgent,
	)
	session, err := scanAuthSessionRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return session, nil
}

func (a AuthSessionRepo) Rotate(ctx context.Context, oldSessionID int64, next *identity.Session) (*identity.Session, error) {
	if next == nil {
		return nil, shared.ErrInvalidInput
	}
	if err := next.Validate(); err != nil {
		return nil, err
	}
	if oldSessionID <= 0 {
		return nil, shared.ErrInvalidInput
	}

	exec := executorFromContext(ctx, a.db)
	row := exec.QueryRow(
		ctx,
		sessionRotateSQL,
		oldSessionID,
		"rotated",
		next.Access.JTI,
		next.Refresh.JTI,
		next.Access.TokenHash,
		next.Refresh.TokenHash,
		next.IssuedAt,
		next.Access.ExpiresAt,
		next.Refresh.ExpiresAt,
		next.Meta.IP,
		next.Meta.UserAgent,
	)
	session, err := scanAuthSessionRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}
	return session, nil
}

func (a AuthSessionRepo) RevokeAllByUser(ctx context.Context, userID int64, reason string) error {
	if userID <= 0 {
		return shared.ErrInvalidInput
	}
	if strings.TrimSpace(reason) == "" {
		reason = "logout_all"
	}
	exec := executorFromContext(ctx, a.db)
	_, err := exec.Exec(ctx, sessionRevokeAllByUserSQL, userID, reason)
	return err
}

func scanAuthSessionRow(row pgx.Row) (*identity.Session, error) {
	var s identity.Session
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.Access.JTI,
		&s.Refresh.JTI,
		&s.Access.TokenHash,
		&s.Refresh.TokenHash,
		&s.IssuedAt,
		&s.Access.ExpiresAt,
		&s.Refresh.ExpiresAt,
		&s.Meta.RotatedFromID,
		&s.Meta.RevokedAt,
		&s.Meta.RevokeReason,
		&s.Meta.IP,
		&s.Meta.UserAgent,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

var _ identity.SessionReader = (*AuthSessionRepo)(nil)
var _ identity.SessionWriter = (*AuthSessionRepo)(nil)
