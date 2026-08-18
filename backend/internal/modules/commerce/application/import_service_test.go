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

type providerSelectorStub struct{ adapter ports.ProviderAdapter }

func (stub providerSelectorStub) Select(string) ports.ProviderAdapter { return stub.adapter }

type providerStub struct {
	batches []domain.ProviderBatch
	err     error
	cursors []*string
}

func (stub *providerStub) Key() string { return "fixture" }
func (stub *providerStub) ValidateConfiguration(context.Context, domain.ProviderConfiguration) error {
	return stub.err
}
func (stub *providerStub) FetchOffers(_ context.Context, _ domain.ProviderConfiguration, cursor *string) (domain.ProviderBatch, error) {
	stub.cursors = append(stub.cursors, cursor)
	if stub.err != nil {
		return domain.ProviderBatch{}, stub.err
	}
	batch := stub.batches[0]
	stub.batches = stub.batches[1:]
	return batch, nil
}

type importRepositoryStub struct {
	run       domain.ImportRun
	claimed   bool
	applied   [][]domain.ValidatedOffer
	failures  [][]domain.ImportRecordFailure
	completed *domain.ImportApplyResult
	failed    string
}

func (stub *importRepositoryStub) CreateProviderConfiguration(context.Context, identity.UserID, domain.ProviderConfigurationInput) (domain.ProviderConfiguration, error) {
	return domain.ProviderConfiguration{}, nil
}
func (stub *importRepositoryStub) ListProviderConfigurations(context.Context) ([]domain.ProviderConfiguration, error) {
	return nil, nil
}
func (stub *importRepositoryStub) GetProviderConfiguration(context.Context, domain.ProviderConfigurationID) (domain.ProviderConfiguration, error) {
	return stub.run.ProviderConfiguration, nil
}
func (stub *importRepositoryStub) SetProviderLifecycle(context.Context, identity.UserID, domain.ProviderConfigurationID, domain.ProviderLifecycle, bool) (domain.ProviderConfiguration, error) {
	return domain.ProviderConfiguration{}, nil
}
func (stub *importRepositoryStub) QueueImport(context.Context, *identity.UserID, domain.ProviderConfigurationID, domain.ImportTrigger, string, int16) (domain.ImportRun, error) {
	return stub.run, nil
}
func (stub *importRepositoryStub) QueueDueImports(context.Context, time.Time, int) (int, error) {
	return 0, nil
}
func (stub *importRepositoryStub) RecoverStalledImports(context.Context, time.Time, time.Time, int) (int, error) {
	return 0, nil
}
func (stub *importRepositoryStub) ClaimNextImport(context.Context, time.Time) (domain.ImportRun, error) {
	if stub.claimed {
		return domain.ImportRun{}, ports.ErrImportNotFound
	}
	stub.claimed = true
	return stub.run, nil
}
func (stub *importRepositoryStub) ApplyImport(_ context.Context, _ domain.ImportRun, offers []domain.ValidatedOffer, failures []domain.ImportRecordFailure, batch domain.ProviderBatch, _ time.Time) (domain.ImportApplyResult, error) {
	stub.applied = append(stub.applied, offers)
	stub.failures = append(stub.failures, failures)
	return domain.ImportApplyResult{Received: len(offers) + len(failures), Applied: len(offers),
		Rejected: len(failures), NextCursor: batch.NextCursor, Complete: batch.Complete}, nil
}
func (stub *importRepositoryStub) CompleteImport(_ context.Context, _ domain.ImportRunID, result domain.ImportApplyResult, _ time.Time) error {
	stub.completed = &result
	return nil
}
func (stub *importRepositoryStub) FailImport(_ context.Context, _ domain.ImportRun, code, _ string, _ time.Time) error {
	stub.failed = code
	return nil
}
func (stub *importRepositoryStub) RetryImport(context.Context, identity.UserID, domain.ImportRunID, string) (domain.ImportRun, error) {
	return stub.run, nil
}
func (stub *importRepositoryStub) ListImports(context.Context, int, int) ([]domain.ImportRun, int64, error) {
	return nil, 0, nil
}
func (stub *importRepositoryStub) ListImportFailures(context.Context, domain.ImportRunID, int, int) ([]domain.ImportFailure, int64, error) {
	return nil, 0, nil
}
func (stub *importRepositoryStub) AnonymizeExpiredClicks(context.Context, time.Time, int) (int64, error) {
	return 0, nil
}

func validImportRun() domain.ImportRun {
	return domain.ImportRun{ID: "f82be2ea-e25c-473e-8466-1a9adab854af", AttemptCount: 1, MaxAttempts: 3,
		ProviderConfiguration: domain.ProviderConfiguration{ID: "80cb0af6-1c46-486c-87fd-32a838ad4f71",
			MerchantID: "f3e605ce-f657-419f-b627-477373051085", AdapterKey: "fixture",
			ProviderKey: "fixture", LifecycleStatus: domain.ProviderActive, FreshnessTTLMinutes: 4320}}
}

func validProviderOffer(id string) domain.ProviderOffer {
	return domain.ProviderOffer{ExternalOfferID: id, ProductID: "56c11ce4-d2b3-4d3b-994c-a04afe3b9b16",
		MerchantSKU: "sku-" + id, ProductURL: "https://merchant.example/products/" + id,
		PriceMinor: 1000, Currency: "USD", Availability: "in_stock", Condition: "new"}
}

func TestImportProcessesCursorsAndCompletes(t *testing.T) {
	next := "page-2"
	provider := &providerStub{batches: []domain.ProviderBatch{
		{Offers: []domain.ProviderOffer{validProviderOffer("one")}, NextCursor: &next},
		{Offers: []domain.ProviderOffer{validProviderOffer("two")}, Complete: true},
	}}
	repository := &importRepositoryStub{run: validImportRun()}
	service := NewImportService(repository, providerSelectorStub{adapter: provider})
	service.now = func() time.Time { return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC) }
	didWork, err := service.ProcessNext(context.Background())
	if err != nil || !didWork {
		t.Fatalf("ProcessNext() = %v, %v", didWork, err)
	}
	if len(provider.cursors) != 2 || provider.cursors[0] != nil || provider.cursors[1] == nil || *provider.cursors[1] != next {
		t.Fatalf("unexpected cursors: %#v", provider.cursors)
	}
	if repository.completed == nil || repository.completed.Applied != 2 || repository.completed.Rejected != 0 {
		t.Fatalf("completion = %#v", repository.completed)
	}
}

func TestImportRejectsDuplicateAndMalformedRecords(t *testing.T) {
	bad := validProviderOffer("bad")
	bad.ProductURL = "https://127.0.0.1/private"
	provider := &providerStub{batches: []domain.ProviderBatch{{Offers: []domain.ProviderOffer{
		validProviderOffer("same"), validProviderOffer("same"), bad,
	}, Complete: true}}}
	repository := &importRepositoryStub{run: validImportRun()}
	service := NewImportService(repository, providerSelectorStub{adapter: provider})
	service.now = func() time.Time { return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC) }
	if _, err := service.ProcessNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(repository.applied) != 1 || len(repository.applied[0]) != 1 || len(repository.failures[0]) != 2 {
		t.Fatalf("applied=%d failures=%d", len(repository.applied[0]), len(repository.failures[0]))
	}
	if repository.completed == nil || repository.completed.Rejected != 2 {
		t.Fatalf("completion = %#v", repository.completed)
	}
}

func TestDisabledProviderFailsClosedAndRecordsFailure(t *testing.T) {
	provider := &providerStub{err: ports.ErrProviderDisabled}
	repository := &importRepositoryStub{run: validImportRun()}
	service := NewImportService(repository, providerSelectorStub{adapter: provider})
	didWork, err := service.ProcessNext(context.Background())
	if err != nil || !didWork || repository.failed != "provider.disabled" {
		t.Fatalf("ProcessNext()=%v,%v failed=%q", didWork, err, repository.failed)
	}
}

func TestSetLifecycleRequiresProviderVerification(t *testing.T) {
	run := validImportRun()
	reference := "secret/fixture"
	run.ProviderConfiguration.CredentialReference = &reference
	repository := &importRepositoryStub{run: run}
	service := NewImportService(repository, providerSelectorStub{adapter: &providerStub{err: ports.ErrProviderUnavailable}})
	_, err := service.SetLifecycle(context.Background(), "565a3e84-2c44-433a-a316-a898d5a18bdc",
		run.ProviderConfiguration.ID, domain.ProviderActive)
	if !errors.Is(err, ports.ErrProviderUnavailable) {
		t.Fatalf("SetLifecycle error=%v", err)
	}
}
