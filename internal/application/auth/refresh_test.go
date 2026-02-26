package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"
)

type refreshParserStub struct {
	out *TokenMeta
	err error
}

func (s refreshParserStub) ParseDetailed(tokenStr string, expectedType string) (*TokenMeta, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.out, nil
}

type refreshSessionRepoStub struct {
	getOut     *identity.Session
	getErr     error
	rotateOut  *identity.Session
	rotateErr  error
	lastOldID  int64
	lastRotate *identity.Session
}

func (s *refreshSessionRepoStub) Create(ctx context.Context, in *identity.Session) (*identity.Session, error) {
	return in, nil
}

func (s *refreshSessionRepoStub) GetByRefreshJTI(ctx context.Context, refreshJTI string) (*identity.Session, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.getOut, nil
}

func (s *refreshSessionRepoStub) Rotate(ctx context.Context, oldSessionID int64, next *identity.Session) (*identity.Session, error) {
	s.lastOldID = oldSessionID
	s.lastRotate = next
	if s.rotateErr != nil {
		return nil, s.rotateErr
	}
	if s.rotateOut != nil {
		return s.rotateOut, nil
	}
	return next, nil
}

type refreshTokenIssuerStub struct {
	out IssuedPair
	err error
}

func (s refreshTokenIssuerStub) Issue(userID int64) (TokenPair, error) {
	return TokenPair{}, nil
}

func (s refreshTokenIssuerStub) IssueDetailed(userID int64) (IssuedPair, error) {
	if s.err != nil {
		return IssuedPair{}, s.err
	}
	return s.out, nil
}

type refreshHasherStub struct {
	hashes   map[string]string
	equalOK  bool
	equalErr error
	hashErr  error
}

func (s refreshHasherStub) Hash(raw string) (string, error) {
	if s.hashErr != nil {
		return "", s.hashErr
	}
	if v, ok := s.hashes[raw]; ok {
		return v, nil
	}
	return "hash:" + raw, nil
}

func (s refreshHasherStub) Equal(raw, hashed string) (bool, error) {
	if s.equalErr != nil {
		return false, s.equalErr
	}
	return s.equalOK, nil
}

func TestRefreshUseCase_Success(t *testing.T) {
	now := time.Now().UTC()
	repo := &refreshSessionRepoStub{
		getOut: &identity.Session{
			ID:     10,
			UserID: 42,
			Refresh: identity.Claim{
				JTI:       "old-r-jti",
				TokenHash: "refresh_old_hash",
				ExpiresAt: now.Add(time.Hour),
			},
			Access: identity.Claim{
				JTI:       "old-a-jti",
				TokenHash: "access_old_hash",
				ExpiresAt: now.Add(10 * time.Minute),
			},
		},
	}

	uc := NewRefreshUseCase(
		refreshParserStub{out: &TokenMeta{UserID: 42, JTI: "old-r-jti", TokenType: "refresh", ExpiresAt: now.Add(time.Hour)}},
		repo,
		refreshTokenIssuerStub{out: IssuedPair{
			Access: IssuedToken{
				Value:     "access.new.raw",
				JTI:       "new-a-jti",
				ExpiresAt: now.Add(15 * time.Minute),
			},
			Refresh: IssuedToken{
				Value:     "refresh.new.raw",
				JTI:       "new-r-jti",
				ExpiresAt: now.Add(24 * time.Hour),
			},
			IssuedAt: now,
		}},
		refreshHasherStub{
			equalOK: true,
			hashes: map[string]string{
				"access.new.raw":  "access_new_hash",
				"refresh.new.raw": "refresh_new_hash",
			},
		},
	)

	out, err := uc.Execute(context.Background(), &RefreshInput{RefreshToken: "refresh.old.raw"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out == nil {
		t.Fatal("expected output")
	}
	if out.Token.AccessToken != "access.new.raw" || out.Token.RefreshToken != "refresh.new.raw" {
		t.Fatalf("unexpected token pair: %+v", out.Token)
	}
	if repo.lastOldID != 10 {
		t.Fatalf("expected rotate old session id=10, got %d", repo.lastOldID)
	}
	if repo.lastRotate == nil {
		t.Fatal("expected rotate payload")
	}
	if repo.lastRotate.Refresh.JTI != "new-r-jti" {
		t.Fatalf("expected new refresh jti, got %q", repo.lastRotate.Refresh.JTI)
	}
}

func TestRefreshUseCase_ParserErrorUnauthorized(t *testing.T) {
	uc := NewRefreshUseCase(
		refreshParserStub{err: errors.New("bad token")},
		&refreshSessionRepoStub{},
		refreshTokenIssuerStub{},
		refreshHasherStub{},
	)

	_, err := uc.Execute(context.Background(), &RefreshInput{RefreshToken: "bad"})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestRefreshUseCase_SessionInfraErrorPropagates(t *testing.T) {
	infraErr := errors.New("db down")
	uc := NewRefreshUseCase(
		refreshParserStub{out: &TokenMeta{UserID: 42, JTI: "jti", TokenType: "refresh", ExpiresAt: time.Now().UTC().Add(time.Hour)}},
		&refreshSessionRepoStub{getErr: infraErr},
		refreshTokenIssuerStub{},
		refreshHasherStub{},
	)

	_, err := uc.Execute(context.Background(), &RefreshInput{RefreshToken: "raw"})
	if !errors.Is(err, infraErr) {
		t.Fatalf("expected infra error, got %v", err)
	}
}

func TestRefreshUseCase_HashMismatchUnauthorized(t *testing.T) {
	now := time.Now().UTC()
	uc := NewRefreshUseCase(
		refreshParserStub{out: &TokenMeta{UserID: 42, JTI: "jti", TokenType: "refresh", ExpiresAt: now.Add(time.Hour)}},
		&refreshSessionRepoStub{
			getOut: &identity.Session{
				ID:     1,
				UserID: 42,
				Refresh: identity.Claim{
					JTI:       "jti",
					TokenHash: "stored_hash",
					ExpiresAt: now.Add(time.Hour),
				},
				Access: identity.Claim{
					JTI:       "a-jti",
					TokenHash: "a_hash",
					ExpiresAt: now.Add(10 * time.Minute),
				},
			},
		},
		refreshTokenIssuerStub{},
		refreshHasherStub{equalOK: false},
	)

	_, err := uc.Execute(context.Background(), &RefreshInput{RefreshToken: "raw"})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestRefreshUseCase_RotateErrorPropagates(t *testing.T) {
	now := time.Now().UTC()
	rotateErr := errors.New("rotate failed")
	uc := NewRefreshUseCase(
		refreshParserStub{out: &TokenMeta{UserID: 42, JTI: "jti", TokenType: "refresh", ExpiresAt: now.Add(time.Hour)}},
		&refreshSessionRepoStub{
			getOut: &identity.Session{
				ID:     1,
				UserID: 42,
				Refresh: identity.Claim{
					JTI:       "jti",
					TokenHash: "stored_hash",
					ExpiresAt: now.Add(time.Hour),
				},
				Access: identity.Claim{
					JTI:       "a-jti",
					TokenHash: "a_hash",
					ExpiresAt: now.Add(10 * time.Minute),
				},
			},
			rotateErr: rotateErr,
		},
		refreshTokenIssuerStub{out: IssuedPair{
			Access: IssuedToken{
				Value:     "access.new.raw",
				JTI:       "new-a-jti",
				ExpiresAt: now.Add(15 * time.Minute),
			},
			Refresh: IssuedToken{
				Value:     "refresh.new.raw",
				JTI:       "new-r-jti",
				ExpiresAt: now.Add(24 * time.Hour),
			},
			IssuedAt: now,
		}},
		refreshHasherStub{
			equalOK: true,
			hashes: map[string]string{
				"access.new.raw":  "access_new_hash",
				"refresh.new.raw": "refresh_new_hash",
			},
		},
	)

	_, err := uc.Execute(context.Background(), &RefreshInput{RefreshToken: "raw"})
	if !errors.Is(err, rotateErr) {
		t.Fatalf("expected rotate error, got %v", err)
	}
}
