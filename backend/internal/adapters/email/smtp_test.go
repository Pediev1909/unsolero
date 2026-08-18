package email

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"rigmark/internal/modules/identity/ports"
)

type captureSender struct {
	message OutboundMessage
	err     error
}

func (sender *captureSender) Send(_ context.Context, message OutboundMessage) error {
	sender.message = message
	return sender.err
}

func TestSMTPDeliveryBuildsFragmentTokenLinksWithoutInventingDelivery(t *testing.T) {
	sender := &captureSender{}
	delivery, err := NewSMTPDeliveryWithSender(sender, "https://unsolero.example")
	if err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("a", 43)
	receipt, err := delivery.SendVerification(context.Background(), ports.VerificationMessage{
		Recipient: "Person <person@example.com>", Token: token, ExpiresAt: time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC),
	})
	if err != nil || !receipt.Accepted || receipt.Reference != "smtp_accepted" {
		t.Fatalf("receipt=%#v error=%v", receipt, err)
	}
	if !strings.Contains(sender.message.TextBody, "/verify-email#"+token) || strings.Contains(sender.message.TextBody, "?token=") {
		t.Fatalf("unsafe verification body = %q", sender.message.TextBody)
	}
}

func TestSMTPDeliveryRejectsHeaderInjectionAndUnknownSecurityEvents(t *testing.T) {
	delivery, err := NewSMTPDeliveryWithSender(&captureSender{}, "https://unsolero.example")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := delivery.SendPasswordReset(context.Background(), ports.PasswordResetMessage{
		Recipient: "person@example.com\r\nBcc: attacker@example.com", Token: strings.Repeat("a", 43), ExpiresAt: time.Now().Add(time.Hour),
	}); err == nil {
		t.Fatal("header injection recipient accepted")
	}
	if _, err := delivery.SendSecurityNotification(context.Background(), ports.SecurityNotification{
		Recipient: "person@example.com", EventType: "arbitrary_message", OccurredAt: time.Now(),
	}); err == nil {
		t.Fatal("unknown security notification accepted")
	}
}

func TestSMTPDeliveryReportsTransportFailureWithoutAcceptance(t *testing.T) {
	sender := &captureSender{err: errors.New("synthetic SMTP outage")}
	delivery, err := NewSMTPDeliveryWithSender(sender, "https://unsolero.example")
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := delivery.SendSecurityNotification(context.Background(), ports.SecurityNotification{
		Recipient: "person@example.com", EventType: "mfa_enabled", OccurredAt: time.Now(),
	})
	if err == nil || receipt.Accepted {
		t.Fatalf("receipt=%#v error=%v", receipt, err)
	}
}

func TestSMTPConfigurationRequiresBoundedTransportSettings(t *testing.T) {
	if _, err := NewSMTPDelivery(SMTPConfig{Address: "smtp.example:587", SenderAddress: "security@example.com", PublicSiteURL: "https://unsolero.example"}); err == nil {
		t.Fatal("SMTP delivery accepted zero timeout")
	}
}
