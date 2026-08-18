package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

type conversionProviderStub struct {
	verified    domain.VerifiedWebhook
	validateErr error
	verifyErr   error
	fetchErr    error
}

func (stub conversionProviderStub) Key() string { return "fixture" }
func (stub conversionProviderStub) ValidateConversionConfiguration(context.Context, domain.ProviderConfiguration) error {
	return stub.validateErr
}
func (stub conversionProviderStub) VerifyWebhook(context.Context, domain.ProviderConfiguration, domain.WebhookRequest) (domain.VerifiedWebhook, error) {
	return stub.verified, stub.verifyErr
}
func (stub conversionProviderStub) FetchConversions(context.Context, domain.ProviderConfiguration, *string) (domain.ConversionBatch, error) {
	return domain.ConversionBatch{}, stub.fetchErr
}

type conversionSelectorStub struct {
	provider ports.ConversionProviderAdapter
}

func (stub conversionSelectorStub) Select(string) ports.ConversionProviderAdapter {
	return stub.provider
}

type conversionRepositoryStub struct {
	configuration domain.ProviderConfiguration
	recordErr     error
	recordedState string
	delivery      domain.WebhookDelivery
	applied       []domain.VerifiedConversionEvent
}

func (stub *conversionRepositoryStub) GetProviderConfiguration(context.Context, domain.ProviderConfigurationID) (domain.ProviderConfiguration, error) {
	return stub.configuration, nil
}
func (stub *conversionRepositoryStub) SetConversionProviderEnabled(context.Context, identity.UserID, domain.ProviderConfigurationID, bool, time.Time) (domain.ProviderConfiguration, error) {
	return stub.configuration, nil
}
func (stub *conversionRepositoryStub) RecordWebhookDelivery(_ context.Context, _ domain.ProviderConfigurationID, _, _, state string, _ *time.Time, _ *string, _ time.Time) (domain.WebhookDelivery, error) {
	stub.recordedState = state
	if stub.delivery.ID == "" {
		stub.delivery = domain.WebhookDelivery{ID: "delivery", VerificationState: state}
	}
	return stub.delivery, stub.recordErr
}
func (stub *conversionRepositoryStub) ApplyWebhookEvents(_ context.Context, _ domain.WebhookDeliveryID, _ domain.ProviderConfiguration, events []domain.VerifiedConversionEvent, _ time.Time) (int, error) {
	stub.applied = events
	return len(events), nil
}
func (stub *conversionRepositoryStub) ResolveConversionAttribution(context.Context, domain.ProviderConfiguration, *string, time.Time, time.Duration) (domain.ConversionAttribution, error) {
	return domain.ConversionAttribution{Status: "unattributed"}, nil
}
func (stub *conversionRepositoryStub) QueueConversionImport(context.Context, *identity.UserID, domain.ProviderConfigurationID, domain.ImportTrigger, string, int16) (domain.ConversionImportRun, error) {
	return domain.ConversionImportRun{}, nil
}
func (stub *conversionRepositoryStub) QueueDueConversionImports(context.Context, time.Time, int) (int, error) {
	return 0, nil
}
func (stub *conversionRepositoryStub) RecoverStalledConversionImports(context.Context, time.Time, time.Time, int) (int, error) {
	return 0, nil
}
func (stub *conversionRepositoryStub) ClaimNextConversionImport(context.Context, time.Time) (domain.ConversionImportRun, error) {
	return domain.ConversionImportRun{}, ports.ErrImportNotFound
}
func (stub *conversionRepositoryStub) ApplyConversionImport(context.Context, domain.ConversionImportRun, []domain.VerifiedConversionEvent, []domain.ImportRecordFailure, domain.ConversionBatch, time.Time) (int, error) {
	return 0, nil
}
func (stub *conversionRepositoryStub) CompleteConversionImport(context.Context, domain.ConversionImportRun, domain.ConversionBatch, int, int, time.Time) error {
	return nil
}
func (stub *conversionRepositoryStub) FailConversionImport(context.Context, domain.ConversionImportRun, string, string, time.Time) error {
	return nil
}
func (stub *conversionRepositoryStub) RetryConversionImport(context.Context, identity.UserID, domain.ConversionImportRunID, string) (domain.ConversionImportRun, error) {
	return domain.ConversionImportRun{}, nil
}
func (stub *conversionRepositoryStub) ListConversions(context.Context, domain.ConversionFilter) ([]domain.Conversion, int64, error) {
	return nil, 0, nil
}
func (stub *conversionRepositoryStub) ListConversionImports(context.Context, int, int) ([]domain.ConversionImportRun, int64, error) {
	return nil, 0, nil
}
func (stub *conversionRepositoryStub) ListReconciliations(context.Context, int, int) ([]domain.ReconciliationRun, int64, error) {
	return nil, 0, nil
}
func (stub *conversionRepositoryStub) ReconcileConversionImport(context.Context, *identity.UserID, domain.ConversionImportRunID, string, time.Time) (domain.ReconciliationRun, error) {
	return domain.ReconciliationRun{}, nil
}
func (stub *conversionRepositoryStub) MonetizationReport(context.Context, time.Time, time.Time) (domain.MonetizationReport, error) {
	return domain.MonetizationReport{}, nil
}

func activeConversionConfiguration() domain.ProviderConfiguration {
	credential := "secret/test"
	return domain.ProviderConfiguration{ID: "80cb0af6-1c46-486c-87fd-32a838ad4f71",
		MerchantID: "f3e605ce-f657-419f-b627-477373051085", ProviderKey: "fixture",
		AdapterKey: "fixture", CredentialReference: &credential, LifecycleStatus: domain.ProviderActive,
		ConversionEnabled: true}
}

func validConversionWebhook(now time.Time) domain.VerifiedWebhook {
	return domain.VerifiedWebhook{SignatureTimestamp: now, Events: []domain.ProviderConversionEvent{{
		ProviderEventID: "evt-1", EventType: domain.EventConversionCreated,
		ExternalConversionID: "conversion-1", OrderStatus: domain.OrderPending,
		EventTimestamp: now.Add(-time.Minute),
	}}}
}

func TestForgedConversionWebhookIsRejectedAndAudited(t *testing.T) {
	now := time.Now().UTC()
	repository := &conversionRepositoryStub{configuration: activeConversionConfiguration()}
	service := NewConversionService(repository, conversionSelectorStub{provider: conversionProviderStub{verifyErr: errors.New("bad signature")}})
	service.now = func() time.Time { return now }
	_, err := service.IngestWebhook(context.Background(), repository.configuration.ID,
		domain.WebhookRequest{Body: []byte(`{"event":"forged"}`), ReceivedAt: now})
	if !errors.Is(err, ports.ErrWebhookRejected) || repository.recordedState != "rejected" || len(repository.applied) != 0 {
		t.Fatalf("err=%v state=%q applied=%d", err, repository.recordedState, len(repository.applied))
	}
}

func TestDuplicateConversionWebhookIsIdempotent(t *testing.T) {
	now := time.Now().UTC()
	repository := &conversionRepositoryStub{configuration: activeConversionConfiguration(),
		delivery: domain.WebhookDelivery{ID: "delivery", VerificationState: "verified", Processed: true}}
	service := NewConversionService(repository, conversionSelectorStub{provider: conversionProviderStub{verified: validConversionWebhook(now)}})
	service.now = func() time.Time { return now }
	result, err := service.IngestWebhook(context.Background(), repository.configuration.ID,
		domain.WebhookRequest{Body: []byte(`{"event":"same"}`), ReceivedAt: now})
	if err != nil || !result.Duplicate || len(repository.applied) != 0 {
		t.Fatalf("result=%#v err=%v applied=%d", result, err, len(repository.applied))
	}
}

func TestUnprocessedDuplicateWebhookResumesAfterTimeout(t *testing.T) {
	now := time.Now().UTC()
	repository := &conversionRepositoryStub{configuration: activeConversionConfiguration(),
		delivery: domain.WebhookDelivery{ID: "delivery", VerificationState: "verified"}}
	service := NewConversionService(repository, conversionSelectorStub{provider: conversionProviderStub{verified: validConversionWebhook(now)}})
	service.now = func() time.Time { return now }
	result, err := service.IngestWebhook(context.Background(), repository.configuration.ID,
		domain.WebhookRequest{Body: []byte(`{"event":"same"}`), ReceivedAt: now})
	if err != nil || result.Duplicate || result.Accepted != 1 || len(repository.applied) != 1 {
		t.Fatalf("result=%#v err=%v applied=%d", result, err, len(repository.applied))
	}
}

func TestDisabledConversionProviderCannotIngest(t *testing.T) {
	now := time.Now().UTC()
	configuration := activeConversionConfiguration()
	configuration.ConversionEnabled = false
	repository := &conversionRepositoryStub{configuration: configuration}
	service := NewConversionService(repository, conversionSelectorStub{provider: conversionProviderStub{verified: validConversionWebhook(now)}})
	_, err := service.IngestWebhook(context.Background(), configuration.ID,
		domain.WebhookRequest{Body: []byte(`{}`), ReceivedAt: now})
	if !errors.Is(err, ports.ErrProviderDisabled) || repository.recordedState != "" {
		t.Fatalf("err=%v state=%q", err, repository.recordedState)
	}
}

func TestVerifiedWebhookNeverAcceptsStaleSignature(t *testing.T) {
	now := time.Now().UTC()
	verified := validConversionWebhook(now)
	verified.SignatureTimestamp = now.Add(-11 * time.Minute)
	repository := &conversionRepositoryStub{configuration: activeConversionConfiguration()}
	service := NewConversionService(repository, conversionSelectorStub{provider: conversionProviderStub{verified: verified}})
	_, err := service.IngestWebhook(context.Background(), repository.configuration.ID,
		domain.WebhookRequest{Body: []byte(`{}`), ReceivedAt: now})
	if !errors.Is(err, ports.ErrWebhookRejected) || repository.recordedState != "rejected" {
		t.Fatalf("err=%v state=%q", err, repository.recordedState)
	}
}
