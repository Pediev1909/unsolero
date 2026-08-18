package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

const (
	maximumImportPages    = 100
	maximumImportRecords  = 10_000
	defaultImportAttempts = 3
)

type ProviderSelector interface {
	Select(string) ports.ProviderAdapter
}

type ImportService struct {
	repository ports.ImportRepository
	providers  ProviderSelector
	now        func() time.Time
}

func NewImportService(repository ports.ImportRepository, providers ProviderSelector) *ImportService {
	return &ImportService{repository: repository, providers: providers, now: time.Now}
}

func (service *ImportService) CreateConfiguration(ctx context.Context, actor identity.UserID, input domain.ProviderConfigurationInput) (domain.ProviderConfiguration, error) {
	if actor == "" || input.NormalizeAndValidate() != nil {
		return domain.ProviderConfiguration{}, ErrInvalidAttribution
	}
	return service.repository.CreateProviderConfiguration(ctx, actor, input)
}

func (service *ImportService) ListConfigurations(ctx context.Context) ([]domain.ProviderConfiguration, error) {
	return service.repository.ListProviderConfigurations(ctx)
}

func (service *ImportService) SetLifecycle(ctx context.Context, actor identity.UserID, id domain.ProviderConfigurationID, status domain.ProviderLifecycle) (domain.ProviderConfiguration, error) {
	if actor == "" || id == "" {
		return domain.ProviderConfiguration{}, ErrInvalidAttribution
	}
	configuration, err := service.repository.GetProviderConfiguration(ctx, id)
	if err != nil {
		return domain.ProviderConfiguration{}, err
	}
	verified := false
	switch status {
	case domain.ProviderDisabled, domain.ProviderSuspended:
	case domain.ProviderConfigured, domain.ProviderActive, domain.ProviderDegraded:
		if configuration.CredentialReference == nil {
			return domain.ProviderConfiguration{}, ports.ErrProviderDisabled
		}
		if err := service.providers.Select(configuration.AdapterKey).ValidateConfiguration(ctx, configuration); err != nil {
			return domain.ProviderConfiguration{}, err
		}
		verified = true
	default:
		return domain.ProviderConfiguration{}, ErrInvalidAttribution
	}
	return service.repository.SetProviderLifecycle(ctx, actor, id, status, verified)
}

func (service *ImportService) TriggerManual(ctx context.Context, actor identity.UserID, id domain.ProviderConfigurationID, key string) (domain.ImportRun, error) {
	key = strings.TrimSpace(key)
	if actor == "" || id == "" || len(key) < 8 || len(key) > 200 {
		return domain.ImportRun{}, ErrInvalidAttribution
	}
	configuration, err := service.repository.GetProviderConfiguration(ctx, id)
	if err != nil {
		return domain.ImportRun{}, err
	}
	if configuration.LifecycleStatus != domain.ProviderActive && configuration.LifecycleStatus != domain.ProviderDegraded {
		return domain.ImportRun{}, ports.ErrProviderDisabled
	}
	if err := service.providers.Select(configuration.AdapterKey).ValidateConfiguration(ctx, configuration); err != nil {
		return domain.ImportRun{}, err
	}
	return service.repository.QueueImport(ctx, &actor, id, domain.ImportManual, key, defaultImportAttempts)
}

func (service *ImportService) Retry(ctx context.Context, actor identity.UserID, id domain.ImportRunID, key string) (domain.ImportRun, error) {
	if actor == "" || id == "" || len(strings.TrimSpace(key)) < 8 || len(key) > 200 {
		return domain.ImportRun{}, ErrInvalidAttribution
	}
	return service.repository.RetryImport(ctx, actor, id, strings.TrimSpace(key))
}

func (service *ImportService) ListImports(ctx context.Context, page, pageSize int) ([]domain.ImportRun, int64, error) {
	offset, limit, err := importPagination(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return service.repository.ListImports(ctx, limit, offset)
}

func (service *ImportService) ListFailures(ctx context.Context, runID domain.ImportRunID, page, pageSize int) ([]domain.ImportFailure, int64, error) {
	offset, limit, err := importPagination(page, pageSize)
	if err != nil || runID == "" {
		return nil, 0, ErrInvalidAttribution
	}
	return service.repository.ListImportFailures(ctx, runID, limit, offset)
}

func (service *ImportService) QueueScheduled(ctx context.Context, limit int) (int, error) {
	if limit < 1 || limit > 100 {
		return 0, ErrInvalidAttribution
	}
	return service.repository.QueueDueImports(ctx, service.now().UTC(), limit)
}

func (service *ImportService) RecoverStalled(ctx context.Context, lease time.Duration, limit int) (int, error) {
	if lease < time.Minute || limit < 1 || limit > 100 {
		return 0, ErrInvalidAttribution
	}
	now := service.now().UTC()
	return service.repository.RecoverStalledImports(ctx, now.Add(-lease), now, limit)
}

func (service *ImportService) ProcessNext(ctx context.Context) (bool, error) {
	run, err := service.repository.ClaimNextImport(ctx, service.now().UTC())
	if errors.Is(err, ports.ErrImportNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	provider := service.providers.Select(run.ProviderConfiguration.AdapterKey)
	if err := provider.ValidateConfiguration(ctx, run.ProviderConfiguration); err != nil {
		return true, service.recordImportFailure(ctx, run, err)
	}

	cursor := run.CursorBefore
	total := 0
	for page := 0; page < maximumImportPages; page++ {
		batch, fetchErr := provider.FetchOffers(ctx, run.ProviderConfiguration, cursor)
		if fetchErr != nil {
			return true, service.recordImportFailure(ctx, run, fetchErr)
		}
		if len(batch.Offers)+total > maximumImportRecords {
			return true, service.recordImportFailure(ctx, run, errors.New("provider import exceeded record limit"))
		}
		validated, failures := service.validateBatch(run.ProviderConfiguration, batch.Offers)
		result, applyErr := service.repository.ApplyImport(ctx, run, validated, failures, batch, service.now().UTC())
		if applyErr != nil {
			return true, service.recordImportFailure(ctx, run, applyErr)
		}
		total += len(batch.Offers)
		run.RecordsReceived += result.Received
		run.RecordsApplied += result.Applied
		run.RecordsRejected += result.Rejected
		run.OffersDeactivated += result.OffersDeactivated
		run.CursorAfter = result.NextCursor
		if batch.Complete {
			return true, service.repository.CompleteImport(ctx, run.ID, domain.ImportApplyResult{
				Received: run.RecordsReceived, Applied: run.RecordsApplied, Rejected: run.RecordsRejected,
				OffersDeactivated: run.OffersDeactivated, NextCursor: result.NextCursor, Complete: true,
			}, service.now().UTC())
		}
		if batch.NextCursor == nil || (cursor != nil && *batch.NextCursor == *cursor) {
			return true, service.recordImportFailure(ctx, run, errors.New("provider cursor did not advance"))
		}
		cursor = batch.NextCursor
	}
	return true, service.recordImportFailure(ctx, run, errors.New("provider import exceeded page limit"))
}

func (service *ImportService) AnonymizeExpiredClicks(ctx context.Context, limit int) (int64, error) {
	if limit < 1 || limit > 10_000 {
		return 0, ErrInvalidAttribution
	}
	return service.repository.AnonymizeExpiredClicks(ctx, service.now().UTC(), limit)
}

func (service *ImportService) validateBatch(configuration domain.ProviderConfiguration, records []domain.ProviderOffer) ([]domain.ValidatedOffer, []domain.ImportRecordFailure) {
	validated := make([]domain.ValidatedOffer, 0, len(records))
	failures := make([]domain.ImportRecordFailure, 0)
	seen := make(map[string]bool, len(records))
	for _, record := range records {
		externalID := strings.TrimSpace(record.ExternalOfferID)
		if externalID != "" && seen[externalID] {
			fingerprint := safeRecordFingerprint(record)
			failures = append(failures, domain.ImportRecordFailure{ExternalRecordID: pointer(externalID),
				Code: "record.duplicate", Message: "Duplicate external offer ID in provider page.", RecordFingerprint: &fingerprint})
			continue
		}
		if externalID != "" {
			seen[externalID] = true
		}
		offer, err := domain.ValidateProviderOffer(record, service.now().UTC(), time.Duration(configuration.FreshnessTTLMinutes)*time.Minute)
		if err != nil {
			fingerprint := safeRecordFingerprint(record)
			failure := domain.ImportRecordFailure{Code: "record.invalid", Message: boundedError(err), RecordFingerprint: &fingerprint}
			if externalID != "" && len(externalID) <= 200 {
				failure.ExternalRecordID = pointer(externalID)
			}
			failures = append(failures, failure)
			continue
		}
		validated = append(validated, offer)
	}
	return validated, failures
}

func (service *ImportService) recordImportFailure(ctx context.Context, run domain.ImportRun, err error) error {
	code := "provider.failed"
	if errors.Is(err, ports.ErrProviderDisabled) {
		code = "provider.disabled"
	} else if errors.Is(err, ports.ErrProviderUnavailable) {
		code = "provider.unavailable"
	}
	if repositoryErr := service.repository.FailImport(ctx, run, code, boundedError(err), service.now().UTC()); repositoryErr != nil {
		return fmt.Errorf("record import failure after %v: %w", err, repositoryErr)
	}
	return nil
}

func importPagination(page, pageSize int) (int, int, error) {
	if page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		return 0, 0, ErrInvalidAttribution
	}
	return (page - 1) * pageSize, pageSize, nil
}

func safeRecordFingerprint(record domain.ProviderOffer) string {
	value := fmt.Sprintf("%s\x1f%s\x1f%s\x1f%d\x1f%s\x1f%s", record.ExternalOfferID,
		record.ProductID, record.MerchantSKU, record.PriceMinor, record.Currency, record.Availability)
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func boundedError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func pointer(value string) *string { return &value }
