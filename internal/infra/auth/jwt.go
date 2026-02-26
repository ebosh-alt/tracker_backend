package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

// TokenPair access-token короткоживущий, refresh-token долгоживущий.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Token struct {
	Value     string
	UserID    int64
	TokenType string
	JTI       string
	ExpiresAt time.Time
}

type Issues struct {
	AccessToken  Token
	RefreshToken Token
	IssuedAt     time.Time
}

// Manager отвечает за issue/parse JWT.
type Manager struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func (m *Manager) ParseDetailed(tokenStr string, expectedType string) (*Token, error) {
	parsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		if expectedType == "refresh" {
			return m.RefreshSecret, nil
		}
		return m.AccessSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	typ, ok := claims["typ"].(string)
	if !ok || typ != expectedType {
		return nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}
	return &Token{
		UserID:    int64(sub),
		TokenType: typ,
		ExpiresAt: time.Unix(int64(exp), 0).UTC(),
		JTI:       jti,
	}, nil
}

// Parse валидирует JWT и возвращает user id.
func (m *Manager) Parse(tokenStr, expectedType string) (int64, error) {
	parsed, err := m.ParseDetailed(tokenStr, expectedType)
	if err != nil {
		return 0, err
	}
	return parsed.UserID, nil
}

func (m *Manager) IssueDetailed(userID int64) (*Issues, error) {
	now := time.Now().UTC()
	iat := now.Unix()
	accessToken := Token{
		UserID:    userID,
		TokenType: "access",
		JTI:       uuid.NewString(),
		ExpiresAt: now.Add(m.AccessTTL),
	}
	accessClaims := jwt.MapClaims{
		"sub": accessToken.UserID,
		"jti": accessToken.JTI,
		"iat": iat,
		"exp": accessToken.ExpiresAt.Unix(),
		"typ": accessToken.TokenType,
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.AccessSecret)
	if err != nil {
		return nil, err
	}
	accessToken.Value = access

	refreshToken := Token{
		UserID:    userID,
		TokenType: "refresh",
		JTI:       uuid.NewString(),
		ExpiresAt: now.Add(m.RefreshTTL),
	}
	refreshClaims := jwt.MapClaims{
		"sub": refreshToken.UserID,
		"jti": refreshToken.JTI,
		"iat": iat,
		"exp": refreshToken.ExpiresAt.Unix(),
		"typ": refreshToken.TokenType,
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(m.RefreshSecret)
	if err != nil {
		return nil, err
	}
	refreshToken.Value = refresh

	return &Issues{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IssuedAt:     now,
	}, nil
}

// Issue выпускает пару токенов для пользователя.
func (m *Manager) Issue(userID int64) (TokenPair, error) {
	issues, err := m.IssueDetailed(userID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: issues.AccessToken.Value, RefreshToken: issues.RefreshToken.Value}, nil
}
