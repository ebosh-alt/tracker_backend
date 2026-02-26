package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"tracker/internal/domain/identity"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

type verifierStub struct {
	profile *identity.VerifiedTelegram
	err     error
}

func (s verifierStub) VerifyInitData(initData string) (*identity.VerifiedTelegram, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.profile, nil
}

type repoStub struct {
	result *domainUser.UpsertFromTelegramResult
	err    error
	last   *domainUser.TelegramProfile
}

func (s *repoStub) UpsertFromTelegram(ctx context.Context, cmd *domainUser.TelegramProfile) (*domainUser.UpsertFromTelegramResult, error) {
	s.last = cmd
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

type tokenStub struct {
	pair         TokenPair
	detailedPair IssuedPair
	err          error
	lastID       int64
	detailedCall bool
}

func (s *tokenStub) Issue(userID int64) (TokenPair, error) {
	s.lastID = userID
	if s.err != nil {
		return TokenPair{}, s.err
	}
	if s.pair.AccessToken == "" && s.pair.RefreshToken == "" {
		return TokenPair{
			AccessToken:  s.detailedPair.Access.Value,
			RefreshToken: s.detailedPair.Refresh.Value,
		}, nil
	}
	return s.pair, nil
}

func (s *tokenStub) IssueDetailed(userID int64) (IssuedPair, error) {
	s.lastID = userID
	s.detailedCall = true
	if s.err != nil {
		return IssuedPair{}, s.err
	}
	return s.detailedPair, nil
}

type hasherStub struct {
	hashes   map[string]string
	errRaw   string
	errValue error
}

func (s *hasherStub) Hash(raw string) (string, error) {
	if s.errRaw != "" && raw == s.errRaw {
		if s.errValue != nil {
			return "", s.errValue
		}
		return "", errors.New("hash error")
	}
	if hashed, ok := s.hashes[raw]; ok {
		return hashed, nil
	}
	return "hash:" + raw, nil
}

func (s *hasherStub) Equal(raw, hashed string) (bool, error) {
	computed, err := s.Hash(raw)
	if err != nil {
		return false, err
	}
	return computed == hashed, nil
}

type sessionRepoStub struct {
	created    *identity.Session
	err        error
	calls      int
	getResult  *identity.Session
	getErr     error
	rotateErr  error
	lastRotate *identity.Session
}

func (s *sessionRepoStub) Create(ctx context.Context, in *identity.Session) (*identity.Session, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	copySession := *in
	s.created = &copySession
	out := copySession
	out.ID = 1
	return &out, nil
}

func (s *sessionRepoStub) GetByRefreshJTI(ctx context.Context, refreshJTI string) (*identity.Session, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.getResult != nil {
		return s.getResult, nil
	}
	return nil, shared.ErrNotFound
}

func (s *sessionRepoStub) Rotate(ctx context.Context, oldSessionID int64, next *identity.Session) (*identity.Session, error) {
	if s.rotateErr != nil {
		return nil, s.rotateErr
	}
	if next == nil {
		return nil, shared.ErrInvalidInput
	}
	copySession := *next
	copySession.ID = oldSessionID + 1
	s.lastRotate = &copySession
	return &copySession, nil
}

func TestTelegramAuth_EmptyInitData(t *testing.T) {
	uc := NewTelegramAuthUseCase(verifierStub{}, &repoStub{}, &tokenStub{}, &hasherStub{}, &sessionRepoStub{})

	_, err := uc.Execute(context.Background(), TelegramAuthInput{InitData: "   "})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTelegramAuth_VerifyError(t *testing.T) {
	uc := NewTelegramAuthUseCase(verifierStub{err: errors.New("bad signature")}, &repoStub{}, &tokenStub{}, &hasherStub{}, &sessionRepoStub{})

	_, err := uc.Execute(context.Background(), TelegramAuthInput{InitData: "valid"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestTelegramAuth_Success_DefaultTimezoneAndSessionPersist(t *testing.T) {
	lastName := "Doe"
	ip := "10.20.30.40"
	ua := "telegram-test-agent"
	now := time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC)
	createdUser := domainUser.User{
		ID:         42,
		TelegramID: 1001,
		Username:   "john",
		FirstName:  "John",
		LastName:   "Doe",
		Timezone:   "UTC",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	repo := &repoStub{
		result: &domainUser.UpsertFromTelegramResult{
			User:    &createdUser,
			Created: true,
		},
	}
	tokens := &tokenStub{
		detailedPair: IssuedPair{
			Access: IssuedToken{
				Value:     "access.raw.jwt",
				JTI:       "access-jti-1",
				ExpiresAt: now.Add(15 * time.Minute),
			},
			Refresh: IssuedToken{
				Value:     "refresh.raw.jwt",
				JTI:       "refresh-jti-1",
				ExpiresAt: now.Add(24 * time.Hour),
			},
			IssuedAt: now,
		},
	}
	hasher := &hasherStub{
		hashes: map[string]string{
			"access.raw.jwt":  "access_hash",
			"refresh.raw.jwt": "refresh_hash",
		},
	}
	sessions := &sessionRepoStub{}

	uc := NewTelegramAuthUseCase(verifierStub{profile: &identity.VerifiedTelegram{
		TelegramID: 1001,
		Username:   "john",
		FirstName:  "John",
		LastName:   &lastName,
	}}, repo, tokens, hasher, sessions)

	out, err := uc.Execute(context.Background(), TelegramAuthInput{
		InitData:  "valid",
		IP:        &ip,
		UserAgent: &ua,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.last == nil {
		t.Fatal("expected upsert to be called")
	}
	if repo.last.Timezone != "UTC" {
		t.Fatalf("expected default timezone UTC, got %q", repo.last.Timezone)
	}
	if tokens.lastID != createdUser.ID {
		t.Fatalf("expected token issue for user %d, got %d", createdUser.ID, tokens.lastID)
	}
	if out.User.ID != createdUser.ID {
		t.Fatalf("expected user id %d, got %d", createdUser.ID, out.User.ID)
	}
	if !tokens.detailedCall {
		t.Fatal("expected IssueDetailed to be called")
	}
	if out.Token.AccessToken != "access.raw.jwt" || out.Token.RefreshToken != "refresh.raw.jwt" {
		t.Fatalf("unexpected token pair: %+v", out.Token)
	}
	if !out.Created {
		t.Fatal("expected created=true")
	}
	if sessions.calls != 1 {
		t.Fatalf("expected sessions.Create to be called once, got %d", sessions.calls)
	}
	if sessions.created == nil {
		t.Fatal("expected created session payload")
	}
	if sessions.created.UserID != createdUser.ID {
		t.Fatalf("expected session user id %d, got %d", createdUser.ID, sessions.created.UserID)
	}
	if sessions.created.Access.JTI != "access-jti-1" || sessions.created.Refresh.JTI != "refresh-jti-1" {
		t.Fatalf("unexpected session JTI: access=%q refresh=%q", sessions.created.Access.JTI, sessions.created.Refresh.JTI)
	}
	if sessions.created.Access.TokenHash != "access_hash" || sessions.created.Refresh.TokenHash != "refresh_hash" {
		t.Fatalf("unexpected session token hashes: access=%q refresh=%q", sessions.created.Access.TokenHash, sessions.created.Refresh.TokenHash)
	}
	if sessions.created.Meta.IP == nil || *sessions.created.Meta.IP != ip {
		t.Fatalf("expected session ip %q, got %#v", ip, sessions.created.Meta.IP)
	}
	if sessions.created.Meta.UserAgent == nil || *sessions.created.Meta.UserAgent != ua {
		t.Fatalf("expected session userAgent %q, got %#v", ua, sessions.created.Meta.UserAgent)
	}
}

func TestTelegramAuth_HashAccessError(t *testing.T) {
	lastName := "Doe"
	createdUser := domainUser.User{ID: 42, TelegramID: 1001, Username: "john", FirstName: "John", LastName: "Doe", Timezone: "UTC"}

	repo := &repoStub{
		result: &domainUser.UpsertFromTelegramResult{User: &createdUser, Created: false},
	}
	tokens := &tokenStub{
		detailedPair: IssuedPair{
			Access:   IssuedToken{Value: "access.raw.jwt", JTI: "a-jti", ExpiresAt: time.Now().UTC().Add(time.Minute)},
			Refresh:  IssuedToken{Value: "refresh.raw.jwt", JTI: "r-jti", ExpiresAt: time.Now().UTC().Add(time.Hour)},
			IssuedAt: time.Now().UTC(),
		},
	}
	hashErr := errors.New("hash failed")
	hasher := &hasherStub{errRaw: "access.raw.jwt", errValue: hashErr}
	sessions := &sessionRepoStub{}

	uc := NewTelegramAuthUseCase(verifierStub{profile: &identity.VerifiedTelegram{
		TelegramID: 1001,
		Username:   "john",
		FirstName:  "John",
		LastName:   &lastName,
	}}, repo, tokens, hasher, sessions)

	_, err := uc.Execute(context.Background(), TelegramAuthInput{InitData: "valid"})
	if !errors.Is(err, hashErr) {
		t.Fatalf("expected hash error, got %v", err)
	}
	if sessions.calls != 0 {
		t.Fatalf("expected sessions.Create not to be called, got %d", sessions.calls)
	}
}

func TestTelegramAuth_SessionCreateError(t *testing.T) {
	lastName := "Doe"
	createdUser := domainUser.User{ID: 42, TelegramID: 1001, Username: "john", FirstName: "John", LastName: "Doe", Timezone: "UTC"}

	repo := &repoStub{
		result: &domainUser.UpsertFromTelegramResult{User: &createdUser, Created: false},
	}
	tokens := &tokenStub{
		detailedPair: IssuedPair{
			Access:   IssuedToken{Value: "access.raw.jwt", JTI: "a-jti", ExpiresAt: time.Now().UTC().Add(time.Minute)},
			Refresh:  IssuedToken{Value: "refresh.raw.jwt", JTI: "r-jti", ExpiresAt: time.Now().UTC().Add(time.Hour)},
			IssuedAt: time.Now().UTC(),
		},
	}
	hasher := &hasherStub{
		hashes: map[string]string{
			"access.raw.jwt":  "access_hash",
			"refresh.raw.jwt": "refresh_hash",
		},
	}
	createErr := errors.New("create session failed")
	sessions := &sessionRepoStub{err: createErr}

	uc := NewTelegramAuthUseCase(verifierStub{profile: &identity.VerifiedTelegram{
		TelegramID: 1001,
		Username:   "john",
		FirstName:  "John",
		LastName:   &lastName,
	}}, repo, tokens, hasher, sessions)

	_, err := uc.Execute(context.Background(), TelegramAuthInput{InitData: "valid"})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create session error, got %v", err)
	}
	if sessions.calls != 1 {
		t.Fatalf("expected sessions.Create to be called once, got %d", sessions.calls)
	}
}
