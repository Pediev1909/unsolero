package email

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"rigmark/internal/modules/identity/ports"
	newsletterports "rigmark/internal/modules/newsletter/ports"
)

var ErrDeliveryDisabled = errors.New("email delivery is disabled")

type Disabled struct{}

func (Disabled) SendVerification(context.Context, ports.VerificationMessage) (ports.DeliveryReceipt, error) {
	return ports.DeliveryReceipt{Accepted: false, Reference: "disabled"}, nil
}
func (Disabled) SendPasswordReset(context.Context, ports.PasswordResetMessage) (ports.DeliveryReceipt, error) {
	return ports.DeliveryReceipt{Accepted: false, Reference: "disabled"}, nil
}
func (Disabled) SendSecurityNotification(context.Context, ports.SecurityNotification) (ports.DeliveryReceipt, error) {
	return ports.DeliveryReceipt{Accepted: false, Reference: "disabled"}, nil
}
func (Disabled) SendNewsletterConfirmation(context.Context, newsletterports.ConfirmationMessage) error {
	return nil
}

// DevelopmentSink records explicit delivery intents in process memory. It is
// not a mail transport and never reports delivery. Production configuration
// rejects this adapter before the server starts.
type DevelopmentSink struct {
	mu       sync.Mutex
	nextID   uint64
	messages []ports.DevelopmentMessage
}

func NewDevelopmentSink() *DevelopmentSink { return &DevelopmentSink{} }

func (sink *DevelopmentSink) SendVerification(_ context.Context, message ports.VerificationMessage) (ports.DeliveryReceipt, error) {
	return sink.record("email_verification", message.Recipient, message.Token, message.ExpiresAt)
}

func (sink *DevelopmentSink) SendPasswordReset(_ context.Context, message ports.PasswordResetMessage) (ports.DeliveryReceipt, error) {
	return sink.record("password_reset", message.Recipient, message.Token, message.ExpiresAt)
}

func (sink *DevelopmentSink) SendSecurityNotification(_ context.Context, message ports.SecurityNotification) (ports.DeliveryReceipt, error) {
	return sink.record("security_notification", message.Recipient, "", time.Time{})
}

func (sink *DevelopmentSink) SendNewsletterConfirmation(_ context.Context, message newsletterports.ConfirmationMessage) error {
	_, err := sink.record("newsletter_confirmation", message.Recipient, message.Token, message.ExpiresAt)
	return err
}

func (sink *DevelopmentSink) Messages(recipient string) []ports.DevelopmentMessage {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	result := make([]ports.DevelopmentMessage, 0, len(sink.messages))
	for _, message := range sink.messages {
		if recipient == "" || message.Recipient == recipient {
			result = append(result, message)
		}
	}
	return result
}

func (sink *DevelopmentSink) record(kind, recipient, token string, expires time.Time) (ports.DeliveryReceipt, error) {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	sink.nextID++
	id := fmt.Sprintf("dev-intent-%d", sink.nextID)
	sink.messages = append(sink.messages, ports.DevelopmentMessage{ID: id, Kind: kind, Recipient: recipient,
		Token: token, ExpiresAt: expires, CreatedAt: time.Now().UTC()})
	return ports.DeliveryReceipt{Accepted: false, Reference: id}, nil
}
