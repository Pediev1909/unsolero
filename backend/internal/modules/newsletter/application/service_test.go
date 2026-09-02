package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"rigmark/internal/modules/newsletter/domain"
	"rigmark/internal/modules/newsletter/ports"
)

type repositoryStub struct {
	pending      []domain.PendingSubscription
	created      bool
	upsertErr    error
	confirmed    [][]byte
	confirmErr   error
	unsubscribed [][]byte
	purged       int64
}

func (stub *repositoryStub) UpsertPending(_ context.Context, pending domain.PendingSubscription) (bool, error) {
	stub.pending = append(stub.pending, pending)
	return stub.created, stub.upsertErr
}

func (stub *repositoryStub) Confirm(_ context.Context, hash []byte, _ time.Time) error {
	stub.confirmed = append(stub.confirmed, hash)
	return stub.confirmErr
}

func (stub *repositoryStub) Unsubscribe(_ context.Context, hash []byte, _ time.Time) error {
	stub.unsubscribed = append(stub.unsubscribed, hash)
	return nil
}

func (stub *repositoryStub) PurgeExpiredPending(context.Context, time.Time) (int64, error) {
	return stub.purged, nil
}

type deliveryStub struct {
	messages []ports.ConfirmationMessage
	err      error
}

func (stub *deliveryStub) SendNewsletterConfirmation(_ context.Context, message ports.ConfirmationMessage) error {
	stub.messages = append(stub.messages, message)
	return stub.err
}

// tokenStub hands out numbered tokens so a test can tell the confirmation
// token from the unsubscribe token that is generated right after it.
type tokenStub struct {
	generated int
}

func (stub *tokenStub) Generate() (string, []byte, error) {
	stub.generated++
	return fmt.Sprintf("raw-token-%d", stub.generated), []byte(fmt.Sprintf("hash-%d", stub.generated)), nil
}

func (stub *tokenStub) Hash(raw string) ([]byte, error) {
	if !strings.HasPrefix(raw, "raw-token-") {
		return nil, errors.New("malformed token")
	}
	return []byte("hash-" + strings.TrimPrefix(raw, "raw-token-")), nil
}

type clockStub struct{ now time.Time }

func (clock clockStub) Now() time.Time { return clock.now }

var testNow = time.Date(2026, 9, 2, 9, 0, 0, 0, time.UTC)

func newTestService(t *testing.T, repository *repositoryStub, delivery *deliveryStub) *Service {
	t.Helper()
	service, err := newService(repository, &tokenStub{}, delivery, clockStub{now: testNow}, Config{ConfirmationTTL: DefaultConfirmationTTL})
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}
	return service
}

func TestSubscribeRecordsPendingRowAndMailsOnlyTheConfirmationToken(t *testing.T) {
	repository := &repositoryStub{created: true}
	delivery := &deliveryStub{}
	service := newTestService(t, repository, delivery)

	receipt, err := service.Subscribe(context.Background(), " Reader@Example.com ", "article:mailchimp-alternatives")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if !receipt.Recorded {
		t.Error("Subscribe() receipt was not recorded")
	}
	if len(repository.pending) != 1 {
		t.Fatalf("pending rows = %d, want 1", len(repository.pending))
	}
	pending := repository.pending[0]
	if pending.Email != "reader@example.com" {
		t.Errorf("stored email = %q, want normalized address", pending.Email)
	}
	if pending.Source != "article:mailchimp-alternatives" || pending.ConsentTextVersion != domain.ConsentTextVersion {
		t.Errorf("stored source/consent = %q/%q", pending.Source, pending.ConsentTextVersion)
	}
	if string(pending.ConfirmTokenHash) != "hash-1" || string(pending.UnsubscribeTokenHash) != "hash-2" {
		t.Errorf("stored hashes = %q/%q, want distinct confirm and unsubscribe hashes", pending.ConfirmTokenHash, pending.UnsubscribeTokenHash)
	}
	if want := testNow.Add(48 * time.Hour); !pending.ConfirmExpiresAt.Equal(want) || !pending.RequestedAt.Equal(testNow) {
		t.Errorf("expiry/requested = %v/%v, want %v/%v", pending.ConfirmExpiresAt, pending.RequestedAt, want, testNow)
	}
	if len(delivery.messages) != 1 {
		t.Fatalf("delivered messages = %d, want 1", len(delivery.messages))
	}
	message := delivery.messages[0]
	if message.Recipient != "reader@example.com" || message.Token != "raw-token-1" || !message.ExpiresAt.Equal(pending.ConfirmExpiresAt) {
		t.Errorf("confirmation message = %+v", message)
	}
}

func TestSubscribeStaysSilentForConfirmedAddress(t *testing.T) {
	repository := &repositoryStub{created: false}
	delivery := &deliveryStub{}
	service := newTestService(t, repository, delivery)

	receipt, err := service.Subscribe(context.Background(), "reader@example.com", "footer")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if !receipt.Recorded {
		t.Error("confirmed address must receive the same neutral receipt")
	}
	if len(delivery.messages) != 0 {
		t.Error("confirmed address must not be mailed again")
	}
}

func TestSubscribeRejectsInvalidInputBeforeTouchingStorage(t *testing.T) {
	cases := map[string]struct {
		email, source string
		field         string
	}{
		"missing at sign":       {"reader.example.com", "footer", "email"},
		"display name":          {"Reader <reader@example.com>", "footer", "email"},
		"too long":              {strings.Repeat("a", 250) + "@example.com", "footer", "email"},
		"uppercase source":      {"reader@example.com", "Footer", "source"},
		"source with spaces":    {"reader@example.com", "article mailchimp", "source"},
		"source leading symbol": {"reader@example.com", ":footer", "source"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			repository := &repositoryStub{created: true}
			delivery := &deliveryStub{}
			service := newTestService(t, repository, delivery)

			_, err := service.Subscribe(context.Background(), testCase.email, testCase.source)
			var validation ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("Subscribe() error = %v, want ValidationError", err)
			}
			if _, ok := validation.Fields[testCase.field]; !ok {
				t.Errorf("validation fields = %v, want %q", validation.Fields, testCase.field)
			}
			if len(repository.pending) != 0 || len(delivery.messages) != 0 {
				t.Error("invalid input must not reach storage or delivery")
			}
		})
	}
}

func TestSubscribeReportsDeliveryFailureWithRecordedReceipt(t *testing.T) {
	repository := &repositoryStub{created: true}
	delivery := &deliveryStub{err: errors.New("smtp down")}
	service := newTestService(t, repository, delivery)

	receipt, err := service.Subscribe(context.Background(), "reader@example.com", "footer")
	if err == nil {
		t.Fatal("Subscribe() must surface the delivery failure to the caller")
	}
	if !receipt.Recorded {
		t.Error("the row was written, so the receipt must say recorded")
	}
}

func TestSubscribeReportsStorageFailureWithoutRecordedReceipt(t *testing.T) {
	repository := &repositoryStub{upsertErr: errors.New("database unavailable")}
	delivery := &deliveryStub{}
	service := newTestService(t, repository, delivery)

	receipt, err := service.Subscribe(context.Background(), "reader@example.com", "footer")
	if err == nil || receipt.Recorded {
		t.Errorf("Subscribe() = %+v, %v; want unrecorded receipt with error", receipt, err)
	}
	if len(delivery.messages) != 0 {
		t.Error("nothing may be mailed when the row was not written")
	}
}

func TestConfirmConsumesTokenHash(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &deliveryStub{})

	if err := service.Confirm(context.Background(), "  raw-token-1 "); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if len(repository.confirmed) != 1 || string(repository.confirmed[0]) != "hash-1" {
		t.Errorf("confirmed hashes = %q, want the token's hash", repository.confirmed)
	}
}

func TestConfirmTranslatesUnknownAndMalformedTokens(t *testing.T) {
	repository := &repositoryStub{confirmErr: ports.ErrNotFound}
	service := newTestService(t, repository, &deliveryStub{})

	if err := service.Confirm(context.Background(), "raw-token-9"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("unknown token error = %v, want ErrInvalidToken", err)
	}
	if err := service.Confirm(context.Background(), "not-a-token"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("malformed token error = %v, want ErrInvalidToken", err)
	}
	if len(repository.confirmed) != 1 {
		t.Error("a malformed token must be rejected before the repository is asked")
	}
}

func TestConfirmWrapsUnexpectedRepositoryErrors(t *testing.T) {
	repository := &repositoryStub{confirmErr: errors.New("connection reset")}
	service := newTestService(t, repository, &deliveryStub{})

	err := service.Confirm(context.Background(), "raw-token-1")
	if err == nil || errors.Is(err, ErrInvalidToken) {
		t.Errorf("Confirm() error = %v, want a wrapped infrastructure error", err)
	}
}

func TestUnsubscribeIsIdempotent(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &deliveryStub{})

	for range 2 {
		if err := service.Unsubscribe(context.Background(), "raw-token-2"); err != nil {
			t.Fatalf("Unsubscribe() error = %v", err)
		}
	}
	if len(repository.unsubscribed) != 2 || string(repository.unsubscribed[1]) != "hash-2" {
		t.Errorf("unsubscribed hashes = %q", repository.unsubscribed)
	}
	if err := service.Unsubscribe(context.Background(), "garbage"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("malformed unsubscribe token error = %v, want ErrInvalidToken", err)
	}
}

func TestNewServiceRejectsMissingDependenciesAndUnsafeLifetimes(t *testing.T) {
	if _, err := newService(nil, &tokenStub{}, &deliveryStub{}, clockStub{}, Config{ConfirmationTTL: time.Hour}); err == nil {
		t.Error("nil repository must be rejected")
	}
	if _, err := newService(&repositoryStub{}, &tokenStub{}, &deliveryStub{}, clockStub{}, Config{ConfirmationTTL: 30 * time.Minute}); err == nil {
		t.Error("a confirmation lifetime under an hour must be rejected")
	}
	if _, err := newService(&repositoryStub{}, &tokenStub{}, &deliveryStub{}, clockStub{}, Config{ConfirmationTTL: 8 * 24 * time.Hour}); err == nil {
		t.Error("a confirmation lifetime over a week must be rejected")
	}
}
