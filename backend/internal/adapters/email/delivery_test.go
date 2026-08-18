package email

import (
	"context"
	"testing"
	"time"

	"rigmark/internal/modules/identity/ports"
)

func TestDevelopmentSinkRecordsIntentWithoutClaimingDelivery(t *testing.T) {
	sink := NewDevelopmentSink()
	expires := time.Now().UTC().Add(time.Hour)
	receipt, err := sink.SendPasswordReset(context.Background(), ports.PasswordResetMessage{
		Recipient: "person@example.invalid", Token: "one-time-secret", ExpiresAt: expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Accepted {
		t.Fatal("development sink falsely claimed delivery")
	}
	messages := sink.Messages("person@example.invalid")
	if len(messages) != 1 || messages[0].Token != "one-time-secret" || messages[0].Kind != "password_reset" {
		t.Fatalf("development messages = %#v", messages)
	}
}

func TestDisabledAdapterNeverClaimsDelivery(t *testing.T) {
	receipt, err := (Disabled{}).SendVerification(context.Background(), ports.VerificationMessage{})
	if err != nil || receipt.Accepted {
		t.Fatalf("disabled receipt=%#v error=%v", receipt, err)
	}
}
