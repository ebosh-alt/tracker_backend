package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"
)

type RefreshInput struct {
	RefreshToken string
	IP           *string
	UserAgent    *string
}

type RefreshOutput struct {
	Token TokenPair
}

type RefreshTokenParser interface {
	ParseDetailed(tokenStr string, expectedType string) (*TokenMeta, error)
}

type TokenMeta struct {
	UserID    int64
	TokenType string
	JTI       string
	ExpiresAt time.Time
}

type RefreshUseCase struct {
	parser   RefreshTokenParser
	sessions SessionRepository
	tokens   TokenIssuer
	hasher   TokenHasher
}

func NewRefreshUseCase(
	parser RefreshTokenParser,
	sessions SessionRepository,
	tokens TokenIssuer,
	hasher TokenHasher,
) *RefreshUseCase {
	if parser == nil {
		panic(fmt.Errorf("%w: refresh parser dependency is required", shared.ErrInvalidInput))
	}
	if sessions == nil {
		panic(fmt.Errorf("%w: session repository dependency is required", shared.ErrInvalidInput))
	}
	if tokens == nil {
		panic(fmt.Errorf("%w: token issuer dependency is required", shared.ErrInvalidInput))
	}
	if hasher == nil {
		panic(fmt.Errorf("%w: token hasher dependency is required", shared.ErrInvalidInput))
	}

	return &RefreshUseCase{
		parser:   parser,
		sessions: sessions,
		tokens:   tokens,
		hasher:   hasher,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, in *RefreshInput) (*RefreshOutput, error) {
	if in == nil {
		return nil, fmt.Errorf("%w: input is required", shared.ErrInvalidInput)
	}
	raw := strings.TrimSpace(in.RefreshToken)
	if raw == "" {
		return nil, fmt.Errorf("%w: refreshToken is required", shared.ErrInvalidInput)
	}
	parsed, err := uc.parser.ParseDetailed(raw, "refresh")
	if err != nil {
		return nil, shared.ErrUnauthorized
	}
	if parsed == nil {
		return nil, shared.ErrUnauthorized
	}
	session, err := uc.sessions.GetByRefreshJTI(ctx, parsed.JTI)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) || errors.Is(err, shared.ErrUnauthorized) {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}

	if session == nil {
		return nil, shared.ErrUnauthorized
	}
	// guard
	if parsed.UserID != session.UserID {
		return nil, shared.ErrUnauthorized
	}
	ok, err := uc.hasher.Equal(raw, session.Refresh.TokenHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, shared.ErrUnauthorized
	}

	err = session.CanRotate(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	issued, err := uc.tokens.IssueDetailed(session.UserID)
	if err != nil {
		return nil, err
	}

	accessHash, err := uc.hasher.Hash(issued.Access.Value)
	if err != nil {
		return nil, err
	}

	refreshHash, err := uc.hasher.Hash(issued.Refresh.Value)
	if err != nil {
		return nil, err
	}

	next := &identity.Session{
		UserID: session.UserID,
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
		Meta: identity.Meta{
			IP:        in.IP,
			UserAgent: in.UserAgent,
		},
	}
	_, err = uc.sessions.Rotate(ctx, session.ID, next)
	if err != nil {
		return nil, err
	}
	return &RefreshOutput{
		Token: TokenPair{
			AccessToken:  issued.Access.Value,
			RefreshToken: issued.Refresh.Value,
		},
	}, nil
}
