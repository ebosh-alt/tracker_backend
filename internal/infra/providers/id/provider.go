package id

import (
	"crypto/rand"
	"encoding/hex"
)

const randomBytesLength = 16

// Provider генерирует инфраструктурные идентификаторы.
type Provider struct{}

// NewID возвращает случайный hex-id длиной 32 символа.
func (Provider) NewID() (string, error) {
	buf := make([]byte, randomBytesLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
