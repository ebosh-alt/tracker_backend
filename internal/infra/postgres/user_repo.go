package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

const (
	userGetByIDSQL = `SELECT id, tg_id, username, first_name, last_name, timezone, steps_goal, streak, created_at, updated_at
FROM users
WHERE id = $1`
	upsertUserFromTelegramSQL = `INSERT INTO users (
  tg_id,
  username,
  first_name,
  last_name,
  timezone,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (tg_id) DO UPDATE
SET
  username   = EXCLUDED.username,
  first_name = EXCLUDED.first_name,
  last_name  = EXCLUDED.last_name,
  timezone   = COALESCE(NULLIF(EXCLUDED.timezone, ''), users.timezone),
  updated_at = NOW()
RETURNING
  id,
  tg_id,
  username,
  first_name,
  last_name,
  timezone,
  steps_goal,
  streak,
  created_at,
  updated_at,
  (xmax = 0) AS created;

`
)

// UserRepo читает пользователя для application-слоя.
type UserRepo struct {
	db *pgxpool.Pool
}

// NewUserReader создает reader пользователя.
func NewUserReader(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// GetByID возвращает профиль пользователя.
func (ur *UserRepo) GetByID(ctx context.Context, userID int64) (*domainUser.User, error) {
	var (
		id        int64
		tgID      int64
		username  sql.NullString
		firstName sql.NullString
		lastName  sql.NullString
		timezone  string
		stepsGoal int
		streak    int
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	exec := executorFromContext(ctx, ur.db)
	err := exec.QueryRow(
		ctx,
		userGetByIDSQL,
		userID,
	).Scan(&id, &tgID, &username, &firstName, &lastName, &timezone, &stepsGoal, &streak, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	goal, _ := domainUser.NewStepsGoal(stepsGoal)
	user := domainUser.User{
		ID:         id,
		TelegramID: tgID,
		Username:   username.String,
		FirstName:  firstName.String,
		LastName:   lastName.String,
		Timezone:   timezone,
		StepsGoal:  goal,
		Streak:     streak,
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.Time.UTC()
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time.UTC()
	}
	return &user, nil
}

func (ur *UserRepo) UpsertFromTelegram(ctx context.Context, tgUser *domainUser.TelegramProfile) (*domainUser.UpsertFromTelegramResult, error) {
	var (
		id        int64
		tgID      int64
		username  sql.NullString
		firstName sql.NullString
		lastName  sql.NullString
		timezone  string
		stepsGoal int
		streak    int
		createdAt time.Time
		updatedAt time.Time
		created   bool
	)

	exec := executorFromContext(ctx, ur.db)
	err := exec.QueryRow(
		ctx,
		upsertUserFromTelegramSQL,
		tgUser.TelegramID,
		tgUser.Username,
		tgUser.FirstName,
		tgUser.LastName,
		tgUser.Timezone,
	).Scan(&id, &tgID, &username, &firstName, &lastName, &timezone, &stepsGoal, &streak, &createdAt, &updatedAt, &created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	goal, err := domainUser.NewStepsGoal(stepsGoal)
	if err != nil {
		return nil, err
	}

	user := domainUser.User{
		ID:         id,
		TelegramID: tgID,
		Username:   username.String,
		FirstName:  firstName.String,
		LastName:   lastName.String,
		Timezone:   timezone,
		StepsGoal:  goal,
		Streak:     streak,
		CreatedAt:  createdAt.UTC(),
		UpdatedAt:  updatedAt.UTC(),
	}

	return &domainUser.UpsertFromTelegramResult{
		User:    &user,
		Created: created,
	}, nil
}
