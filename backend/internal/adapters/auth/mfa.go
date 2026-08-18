package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // TOTP is standardized with HMAC-SHA1 by RFC 6238.
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type AESGCMSecretBox struct {
	aead cipher.AEAD
}

func NewAESGCMSecretBox(key []byte) (*AESGCMSecretBox, error) {
	if len(key) != 32 {
		return nil, errors.New("MFA encryption key must be exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create MFA cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create MFA AEAD: %w", err)
	}
	return &AESGCMSecretBox{aead: aead}, nil
}

func (box *AESGCMSecretBox) Seal(plaintext []byte) ([]byte, []byte, int16, error) {
	nonce := make([]byte, box.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, 0, fmt.Errorf("generate MFA nonce: %w", err)
	}
	return box.aead.Seal(nil, nonce, plaintext, []byte("unsolero:mfa:v1")), nonce, 1, nil
}

func (box *AESGCMSecretBox) Open(ciphertext, nonce []byte, version int16) ([]byte, error) {
	if version != 1 || len(nonce) != box.aead.NonceSize() {
		return nil, errors.New("unsupported MFA ciphertext")
	}
	plaintext, err := box.aead.Open(nil, nonce, ciphertext, []byte("unsolero:mfa:v1"))
	if err != nil {
		return nil, errors.New("MFA ciphertext authentication failed")
	}
	return plaintext, nil
}

type TOTP struct{}

func (TOTP) Verify(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	for drift := int64(-1); drift <= 1; drift++ {
		if hmac.Equal([]byte(totpCode(secret, now.Unix()/30+drift)), []byte(code)) {
			return true
		}
	}
	return false
}

func totpCode(secret string, counter int64) string {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, uint64(counter))
	mac := hmac.New(sha1.New, decoded)
	_, _ = mac.Write(value)
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	number := (uint32(digest[offset])&0x7f)<<24 | uint32(digest[offset+1])<<16 |
		uint32(digest[offset+2])<<8 | uint32(digest[offset+3])
	return fmt.Sprintf("%06d", number%1_000_000)
}
