package auth

import (
	"bytes"
	"testing"
	"time"
)

func TestAESGCMSecretBoxRoundTripAndTamperDetection(t *testing.T) {
	box, err := NewAESGCMSecretBox(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, nonce, version, err := box.Seal([]byte("JBSWY3DPEHPK3PXP"))
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := box.Open(ciphertext, nonce, version)
	if err != nil || string(plaintext) != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("Open() = %q,%v", plaintext, err)
	}
	ciphertext[0] ^= 1
	if _, err := box.Open(ciphertext, nonce, version); err == nil {
		t.Fatal("tampered ciphertext was accepted")
	}
}

func TestTOTPUsesBoundedClockDrift(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	instant := time.Unix(59, 0).UTC()
	code := totpCode(secret, instant.Unix()/30)
	if !((TOTP{}).Verify(secret, code, instant)) {
		t.Fatal("current TOTP was rejected")
	}
	if (TOTP{}).Verify(secret, "000000", instant) {
		t.Fatal("invalid TOTP was accepted")
	}
}
