package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

const defaultTimezone = "UTC"

// TokenPair пара токенов авторизации.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// TelegramAuthInput входные данные для Telegram auth.
type TelegramAuthInput struct {
	InitData string
	Timezone *string
}

// TelegramAuthOutput результат Telegram auth.
type TelegramAuthOutput struct {
	User    domainUser.User
	Token   TokenPair
	Created bool
}

type IssuedToken struct {
	Value     string
	JTI       string
	ExpiresAt time.Time
}

type IssuedPair struct {
	Access   IssuedToken
	Refresh  IssuedToken
	IssuedAt time.Time
}

// TelegramVerifier проверяет подпись и payload initData.
type TelegramVerifier interface {
	VerifyInitData(initData string) (*identity.VerifiedTelegram, error)
}

// UserRepository сохраняет/читает пользователя при auth.
type UserRepository interface {
	UpsertFromTelegram(ctx context.Context, cmd *domainUser.TelegramProfile) (*domainUser.UpsertFromTelegramResult, error)
}

// TokenIssuer выпускает пару access/refresh токенов.
type TokenIssuer interface {
	Issue(userID int64) (TokenPair, error)
	IssueDetailed(userID int64) (IssuedPair, error)
}

// TelegramAuthUseCase выполняет сценарий авторизации через Telegram.
type TelegramAuthUseCase struct {
	verifier TelegramVerifier
	users    UserRepository
	sessions SessionRepository
	tokens   TokenIssuer
	hasher   TokenHasher
}

// NewTelegramAuthUseCase создает use case Telegram auth.
func NewTelegramAuthUseCase(verifier TelegramVerifier, users UserRepository, tokens TokenIssuer, hasher TokenHasher, sessions SessionRepository) *TelegramAuthUseCase {
	if verifier == nil {
		panic(fmt.Errorf("%w: telegram verifier dependency is required", shared.ErrInvalidInput))
	}
	if users == nil {
		panic(fmt.Errorf("%w: user repository dependency is required", shared.ErrInvalidInput))
	}
	if tokens == nil {
		panic(fmt.Errorf("%w: token issuer dependency is required", shared.ErrInvalidInput))
	}
	if sessions == nil {
		panic(fmt.Errorf("%w: sessions repostirory dependency is required", shared.ErrInvalidInput))
	}
	if hasher == nil {
		panic(fmt.Errorf("%w: hasher dependency is required", shared.ErrInvalidInput))
	}

	return &TelegramAuthUseCase{
		verifier: verifier,
		users:    users,
		tokens:   tokens,
		sessions: sessions,
		hasher:   hasher,
	}
}

// Execute выполняет auth flow: validate -> verify -> upsert -> issue tokens.
func (uc *TelegramAuthUseCase) Execute(ctx context.Context, in TelegramAuthInput) (TelegramAuthOutput, error) {
	initData := strings.TrimSpace(in.InitData)
	if initData == "" {
		return TelegramAuthOutput{}, fmt.Errorf("%w: initData is required", shared.ErrInvalidInput)
	}

	verified, err := uc.verifier.VerifyInitData(initData)
	if err != nil {
		return TelegramAuthOutput{}, fmt.Errorf("%w: telegram initData verification failed", shared.ErrUnauthorized)
	}
	if verified == nil || verified.TelegramID <= 0 {
		return TelegramAuthOutput{}, fmt.Errorf("%w: invalid telegram profile", shared.ErrInvalidInput)
	}

	profile, err := domainUser.NewTelegramProfile(
		verified.TelegramID,
		verified.Username,
		verified.FirstName,
		resolveLastName(verified.LastName),
		resolveTimezone(in.Timezone),
	)
	if err != nil {
		return TelegramAuthOutput{}, fmt.Errorf("%w: invalid telegram profile data: %v", shared.ErrInvalidInput, err)
	}

	upsertResult, err := uc.users.UpsertFromTelegram(ctx, profile)
	if err != nil {
		return TelegramAuthOutput{}, err
	}
	if upsertResult == nil || upsertResult.User == nil {
		return TelegramAuthOutput{}, fmt.Errorf("%w: upsert user returned empty result", shared.ErrConflict)
	}

	issued, err := uc.tokens.IssueDetailed(upsertResult.User.ID)
	if err != nil {
		return TelegramAuthOutput{}, err
	}
	accessHash, err := uc.hasher.Hash(issued.Access.Value)
	if err != nil {
		return TelegramAuthOutput{}, err
	}

	refreshHash, err := uc.hasher.Hash(issued.Refresh.Value)
	if err != nil {
		return TelegramAuthOutput{}, err
	}

	_, err = uc.sessions.Create(ctx, &identity.Session{
		UserID: upsertResult.User.ID,
		Access: identity.Claim{
			JTI:       issued.Access.JTI,
			TokenHash: accessHash,
			ExpiresAt: issued.Access.ExpiresAt,
		},
		Refresh: identity.Claim{
			JTI:       issued.Refresh.JTI,
			TokenHash: refreshHash,
			ExpiresAt: issued.Refresh.ExpiresAt,
		},
		IssuedAt: issued.IssuedAt,
	})
	if err != nil {
		return TelegramAuthOutput{}, err
	}
	return TelegramAuthOutput{
		User: *upsertResult.User,
		Token: TokenPair{
			AccessToken:  issued.Access.Value,
			RefreshToken: issued.Refresh.Value,
		},
		Created: upsertResult.Created,
	}, nil
}

func resolveTimezone(timezone *string) string {
	if timezone == nil {
		return defaultTimezone
	}
	v := strings.TrimSpace(*timezone)
	if v == "" {
		return defaultTimezone
	}
	return v
}

func resolveLastName(lastName *string) string {
	if lastName == nil {
		return ""
	}
	return strings.TrimSpace(*lastName)
}
