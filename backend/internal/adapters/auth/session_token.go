package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const sessionTokenBytes = 32

var ErrInvalidSessionToken = errors.New("invalid session token")

type SessionTokenManager struct {
	random io.Reader
}

func NewSessionTokenManager() *SessionTokenManager {
	return &SessionTokenManager{random: rand.Reader}
}

func (manager *SessionTokenManager) Generate() (string, []byte, error) {
	value := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(manager.random, value); err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(value)
	return raw, hashToken(raw), nil
}

func (manager *SessionTokenManager) Hash(raw string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(raw)
	if err != nil || len(decoded) != sessionTokenBytes {
		return nil, ErrInvalidSessionToken
	}
	return hashToken(raw), nil
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
