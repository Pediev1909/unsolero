package auth

import (
	"bytes"
	"errors"
	"testing"
)

func TestPasswordHasherHashesAndVerifies(t *testing.T) {
	hasher, err := NewPasswordHasher(PasswordParams{
		MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error = %v", err)
	}

	first, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	second, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() second error = %v", err)
	}
	if first == second {
		t.Fatal("Hash() reused a password salt")
	}

	valid, err := hasher.Verify(first, "correct horse battery staple")
	if err != nil || !valid {
		t.Fatalf("Verify() = %v, %v; want true, nil", valid, err)
	}
	valid, err = hasher.Verify(first, "incorrect password")
	if err != nil || valid {
		t.Fatalf("Verify() wrong password = %v, %v; want false, nil", valid, err)
	}
}

func TestPasswordHasherRejectsUntrustedParameters(t *testing.T) {
	hasher, err := NewPasswordHasher(PasswordParams{
		MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error = %v", err)
	}
	_, err = hasher.Verify(
		"$argon2id$v=19$m=999999999,t=1,p=1$c2FsdHNhbHRzYWx0c2FsdA$YWJjZGVmZ2hpamtsbW5vcA",
		"password",
	)
	if !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("Verify() error = %v, want ErrInvalidPasswordHash", err)
	}
}

func TestPasswordHasherIdentifiesCostChanges(t *testing.T) {
	current, err := NewPasswordHasher(PasswordParams{
		MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() error = %v", err)
	}
	legacy, err := NewPasswordHasher(PasswordParams{
		MemoryKiB: 8 * 1024, Iterations: 2, Parallelism: 1, KeyLength: 32,
	})
	if err != nil {
		t.Fatalf("NewPasswordHasher() legacy error = %v", err)
	}
	hash, err := legacy.Hash("a sufficiently long password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !current.NeedsRehash(hash) {
		t.Fatal("NeedsRehash() = false for different parameters")
	}
}

func TestSessionTokenManagerGeneratesOpaqueHashedTokens(t *testing.T) {
	manager := NewSessionTokenManager()
	raw, storedHash, err := manager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if raw == "" || bytes.Contains(storedHash, []byte(raw)) {
		t.Fatal("Generate() did not separate the raw token from its stored hash")
	}
	resolvedHash, err := manager.Hash(raw)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !bytes.Equal(storedHash, resolvedHash) {
		t.Fatal("Hash() did not reproduce the stored token hash")
	}
	if _, err := manager.Hash("not-a-valid-token"); !errors.Is(err, ErrInvalidSessionToken) {
		t.Fatalf("Hash() error = %v, want ErrInvalidSessionToken", err)
	}
}
