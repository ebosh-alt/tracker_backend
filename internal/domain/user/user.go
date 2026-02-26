package user

import (
	"fmt"
	"strings"
	"time"

	"tracker/internal/domain/shared"
)

const (
	// DefaultStepsGoal значение по умолчанию для дневной цели шагов.
	DefaultStepsGoal = 10000
	minStepsGoal     = 1000
	maxStepsGoal     = 30000
	stepsGoalStep    = 500
)

// StepsGoal хранит дневную цель шагов и гарантирует доменные ограничения.
type StepsGoal int

// NewStepsGoal валидирует и создает цель по шагам.
func NewStepsGoal(v int) (StepsGoal, error) {
	if v < minStepsGoal || v > maxStepsGoal || v%stepsGoalStep != 0 {
		return 0, fmt.Errorf(
			"%w: steps goal should be in [%d..%d] with step %d",
			shared.ErrInvalidInput,
			minStepsGoal,
			maxStepsGoal,
			stepsGoalStep,
		)
	}
	return StepsGoal(v), nil
}

// Int возвращает числовое значение цели.
func (g StepsGoal) Int() int {
	return int(g)
}

// User доменная сущность пользователя mini app.
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	Timezone   string
	StepsGoal  StepsGoal
	Streak     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewUser создает пользователя с базовой валидацией обязательных полей.
func NewUser(
	telegramID int64,
	username string,
	firstName string,
	lastName string,
	timezone string,
	now time.Time,
) (User, error) {
	if telegramID <= 0 {
		return User{}, fmt.Errorf("%w: telegram id should be positive", shared.ErrInvalidInput)
	}
	if err := validateTimezone(timezone); err != nil {
		return User{}, err
	}

	goal, err := NewStepsGoal(DefaultStepsGoal)
	if err != nil {
		return User{}, err
	}

	return User{
		TelegramID: telegramID,
		Username:   strings.TrimSpace(username),
		FirstName:  strings.TrimSpace(firstName),
		LastName:   strings.TrimSpace(lastName),
		Timezone:   timezone,
		StepsGoal:  goal,
		Streak:     0,
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC(),
	}, nil
}

// UpdateSettings обновляет часовой пояс и/или цель шагов (partial update).
func (u *User) UpdateSettings(patch SettingsPatch, now time.Time) error {
	if patch.Timezone != nil {
		if err := validateTimezone(*patch.Timezone); err != nil {
			return err
		}
		u.Timezone = *patch.Timezone
	}

	if patch.StepsGoal != nil {
		goal, err := NewStepsGoal(*patch.StepsGoal)
		if err != nil {
			return err
		}
		u.StepsGoal = goal
	}

	if patch.Streak != nil {
		if *patch.Streak < 0 {
			return fmt.Errorf("%w: streak should be >= 0", shared.ErrInvalidInput)
		}
		u.Streak = *patch.Streak
	}

	u.UpdatedAt = now.UTC()
	return nil
}

func (u *User) ApplyTelegramProfile(p TelegramProfile, now time.Time) error {
	if u.ID != p.TelegramID {
		return fmt.Errorf("%w: telegram id should be the same", shared.ErrInvalidInput)
	}
	u.Username = strings.TrimSpace(p.Username)
	u.FirstName = strings.TrimSpace(p.FirstName)
	u.LastName = strings.TrimSpace(p.Username)
	u.UpdatedAt = now.UTC()
	return nil
}

func validateTimezone(tz string) error {
	v := strings.TrimSpace(tz)
	if v == "" {
		return fmt.Errorf("%w: timezone should not be empty", shared.ErrInvalidInput)
	}
	if _, err := time.LoadLocation(v); err != nil {
		return fmt.Errorf("%w: invalid timezone %q", shared.ErrInvalidInput, tz)
	}
	return nil
}

// SettingsPatch описывает частичное изменение пользовательских настроек.
type SettingsPatch struct {
	Timezone  *string
	StepsGoal *int
	Streak    *int
}

type TelegramProfile struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	Timezone   string
}

func NewTelegramProfile(telegramID int64, username string, firstName string, lastName string, timezone string) (*TelegramProfile, error) {
	if telegramID <= 0 || strings.TrimSpace(username) == "" || strings.TrimSpace(firstName) == "" {
		return nil, fmt.Errorf("%w: invalid telegram profile data", shared.ErrInvalidInput)
	}
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = "UTC"
	}

	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, fmt.Errorf("%w: invalid timezone %q", shared.ErrInvalidInput, timezone)
	}
	return &TelegramProfile{
		TelegramID: telegramID,
		Username:   strings.TrimSpace(username),
		FirstName:  strings.TrimSpace(firstName),
		LastName:   strings.TrimSpace(lastName),
		Timezone:   timezone,
	}, nil
}

type UpsertFromTelegramResult struct {
	User    *User
	Created bool
}

func (p TelegramProfile) NewUserFromTelegram(timezone string, now time.Time) (*User, error) {
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		tz = "UTC"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	if p.TelegramID <= 0 {
		return nil, fmt.Errorf("telegram id must be positive")
	}
	return &User{
		TelegramID: p.TelegramID,
		Username:   strings.TrimSpace(p.Username),
		FirstName:  strings.TrimSpace(p.FirstName),
		LastName:   strings.TrimSpace(p.LastName),
		Timezone:   tz,
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC(),
	}, nil
}
