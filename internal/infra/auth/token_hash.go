package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

type TokenHasher interface {
	Hash(raw string) (string, error)
	Equal(raw, hashed string) (bool, error)
}

type SHA256TokenHasher struct{}

func (h SHA256TokenHasher) Hash(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("token is empty")
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}

func (h SHA256TokenHasher) Equal(raw, hashed string) (bool, error) {
	if hashed == "" {
		return false, errors.New("hashed token is empty")
	}
	computed, err := h.Hash(raw)
	if err != nil {
		return false, err
	}
	if len(computed) != len(hashed) {
		return false, nil
	}
	ok := subtle.ConstantTimeCompare([]byte(computed), []byte(hashed)) == 1
	return ok, nil
}
