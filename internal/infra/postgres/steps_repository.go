package postgres

import (
	"context"
	"errors"
	"time"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StepsRepository struct {
	db *pgxpool.Pool
}

const (
	getByDateRangeSQL = `SELECT
  user_id,
  date,
  count,
  source,
  created_at,
  updated_at
FROM steps
WHERE user_id = $1
  AND date >= $2
  AND date <= $3
ORDER BY date ASC;`

	upsertByDateSQL = `INSERT INTO steps (
  user_id,
  date,
  count,
  source,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (user_id, date) DO UPDATE
SET
  count = EXCLUDED.count,
  source = EXCLUDED.source,
  updated_at = NOW()
RETURNING
  user_id,
  date,
  count,
  source,
  created_at,
  updated_at;`

	addDeltaSQL = `INSERT INTO steps (
  user_id,
  date,
  count,
  source,
  created_at,
  updated_at
)
VALUES ($1, $2, GREATEST($3, 0), $4, NOW(), NOW())
ON CONFLICT (user_id, date) DO UPDATE
SET
  count = steps.count + $3,
  source = EXCLUDED.source,
  updated_at = NOW()
WHERE steps.count + $3 >= 0
RETURNING
  user_id,
  date,
  count,
  source,
  created_at,
  updated_at;`
	deleteByDateSQL = `DELETE FROM steps
WHERE user_id = $1
  AND date = $2;`
)

func NewStepsRepository(db *pgxpool.Pool) *StepsRepository {
	return &StepsRepository{
		db: db,
	}
}

func (s *StepsRepository) GetByDateRange(ctx context.Context, userID int64, from shared.LocalDate, to shared.LocalDate) ([]domainSteps.DailyEntry, error) {
	if userID <= 0 {
		return nil, shared.ErrInvalidInput
	}

	if err := from.Validate(); err != nil {
		return nil, shared.ErrInvalidInput
	}

	if err := to.Validate(); err != nil {
		return nil, shared.ErrInvalidInput
	}

	if from > to {
		return nil, shared.ErrInvalidInput
	}

	exec := executorFromContext(ctx, s.db)
	rows, err := exec.Query(ctx, getByDateRangeSQL,
		userID,
		from.String(),
		to.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyEntry := make([]domainSteps.DailyEntry, 0)

	for rows.Next() {
		var (
			UserID    int64
			Date      time.Time
			Count     int
			Source    string
			CreatedAt time.Time
			UpdatedAt time.Time
		)

		if err := rows.Scan(
			&UserID,
			&Date,
			&Count,
			&Source,
			&CreatedAt,
			&UpdatedAt); err != nil {
			return nil, err
		}

		source, err := domainSteps.ParseSource(Source)
		if err != nil {
			return nil, shared.ErrInvalidInput
		}
		date, err := shared.ParseLocalDate(Date.Format("2006-01-02"))

		if err != nil {
			return nil, shared.ErrInvalidInput
		}
		daily, err := domainSteps.NewDailyEntry(UserID, date, Count, source, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		daily.CreatedAt = CreatedAt.UTC()
		daily.UpdatedAt = UpdatedAt.UTC()
		dailyEntry = append(dailyEntry, *daily)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return dailyEntry, nil
}

func (s *StepsRepository) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	var out domainSteps.DailyEntry
	if entry.UserID <= 0 {
		return out, shared.ErrInvalidInput
	}
	if err := entry.Date.Validate(); err != nil {
		return out, shared.ErrInvalidInput
	}
	if entry.Count < 0 {
		return out, shared.ErrInvalidInput
	}
	if _, err := domainSteps.ParseSource(string(entry.Source)); err != nil {
		return out, shared.ErrInvalidInput
	}

	var (
		userID    int64
		dateDB    time.Time
		count     int
		sourceDB  string
		createdAt time.Time
		updatedAt time.Time
	)
	exec := executorFromContext(ctx, s.db)

	err := exec.QueryRow(
		ctx,
		upsertByDateSQL,
		entry.UserID,
		entry.Date.String(),
		entry.Count,
		string(entry.Source),
	).Scan(&userID, &dateDB, &count, &sourceDB, &createdAt, &updatedAt)
	if err != nil {
		return out, err
	}

	source, err := domainSteps.ParseSource(sourceDB)
	if err != nil {
		return out, shared.ErrInvalidInput
	}
	date, err := shared.ParseLocalDate(dateDB.Format("2006-01-02"))
	if err != nil {
		return out, shared.ErrInvalidInput
	}

	daily, err := domainSteps.NewDailyEntry(userID, date, count, source, time.Now().UTC())
	if err != nil {
		return out, err
	}
	daily.CreatedAt = createdAt.UTC()
	daily.UpdatedAt = updatedAt.UTC()

	return *daily, nil
}

func (s *StepsRepository) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	var out domainSteps.DailyEntry

	if userID <= 0 {
		return out, shared.ErrInvalidInput
	}
	if err := date.Validate(); err != nil {
		return out, shared.ErrInvalidInput
	}
	if delta == 0 {
		return out, shared.ErrInvalidInput
	}
	if _, err := domainSteps.ParseSource(string(source)); err != nil {
		return out, shared.ErrInvalidInput
	}

	exec := executorFromContext(ctx, s.db)

	var (
		userIDDB  int64
		dateDB    time.Time
		countDB   int
		sourceDB  string
		createdAt time.Time
		updatedAt time.Time
	)

	err := exec.QueryRow(
		ctx,
		addDeltaSQL,
		userID,
		date.String(),
		delta,
		string(source),
	).Scan(&userIDDB, &dateDB, &countDB, &sourceDB, &createdAt, &updatedAt)
	if err != nil {
		// negative resulting count -> no row from UPDATE ... WHERE
		if errors.Is(err, pgx.ErrNoRows) {
			return out, shared.ErrInvalidInput
		}
		return out, err
	}

	src, err := domainSteps.ParseSource(sourceDB)
	if err != nil {
		return out, shared.ErrInvalidInput
	}
	localDate, err := shared.ParseLocalDate(dateDB.Format("2006-01-02"))
	if err != nil {
		return out, shared.ErrInvalidInput
	}

	daily, err := domainSteps.NewDailyEntry(userIDDB, localDate, countDB, src, time.Now().UTC())
	if err != nil {
		return out, err
	}
	daily.CreatedAt = createdAt.UTC()
	daily.UpdatedAt = updatedAt.UTC()

	return *daily, nil
}

func (s *StepsRepository) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	if userID <= 0 {
		return shared.ErrInvalidInput
	}
	if err := date.Validate(); err != nil {
		return shared.ErrInvalidInput
	}

	exec := executorFromContext(ctx, s.db)
	tag, err := exec.Exec(ctx, deleteByDateSQL, userID, date.String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

var _ domainSteps.Repository = (*StepsRepository)(nil)
