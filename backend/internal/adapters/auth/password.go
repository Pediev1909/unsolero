package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Version   = 19
	maxMemoryKiB    = 256 * 1024
	maxIterations   = 10
	maxParallelism  = 8
	defaultSaltSize = 16
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

type PasswordParams struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

type PasswordHasher struct {
	params PasswordParams
	random io.Reader
}

func NewPasswordHasher(params PasswordParams) (*PasswordHasher, error) {
	if err := validateParams(params); err != nil {
		return nil, err
	}
	return &PasswordHasher{params: params, random: rand.Reader}, nil
}

func (hasher *PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, defaultSaltSize)
	if _, err := io.ReadFull(hasher.random, salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey(
		[]byte(password),
		salt,
		hasher.params.Iterations,
		hasher.params.MemoryKiB,
		hasher.params.Parallelism,
		hasher.params.KeyLength,
	)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Version,
		hasher.params.MemoryKiB,
		hasher.params.Iterations,
		hasher.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (hasher *PasswordHasher) Verify(encodedHash, password string) (bool, error) {
	params, salt, expected, err := parseHash(encodedHash)
	if err != nil {
		return false, err
	}
	actual := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.MemoryKiB,
		params.Parallelism,
		uint32(len(expected)),
	)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func (hasher *PasswordHasher) NeedsRehash(encodedHash string) bool {
	params, _, _, err := parseHash(encodedHash)
	return err != nil || params != hasher.params
}

func parseHash(encodedHash string) (PasswordParams, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	if parts[2] != fmt.Sprintf("v=%d", argon2Version) {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	var params PasswordParams
	count, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&params.MemoryKiB,
		&params.Iterations,
		&params.Parallelism,
	)
	if err != nil || count != 3 || parts[3] != fmt.Sprintf(
		"m=%d,t=%d,p=%d",
		params.MemoryKiB,
		params.Iterations,
		params.Parallelism,
	) {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 16 {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) < 16 || len(key) > 64 {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	params.KeyLength = uint32(len(key))
	if err := validateParams(params); err != nil {
		return PasswordParams{}, nil, nil, ErrInvalidPasswordHash
	}
	return params, salt, key, nil
}

func validateParams(params PasswordParams) error {
	if params.MemoryKiB < 8*1024 || params.MemoryKiB > maxMemoryKiB {
		return errors.New("password hash memory must be between 8192 and 262144 KiB")
	}
	if params.Iterations < 1 || params.Iterations > maxIterations {
		return errors.New("password hash iterations must be between 1 and 10")
	}
	if params.Parallelism < 1 || params.Parallelism > maxParallelism {
		return errors.New("password hash parallelism must be between 1 and 8")
	}
	if params.KeyLength < 16 || params.KeyLength > 64 {
		return errors.New("password hash key length must be between 16 and 64 bytes")
	}
	return nil
}
