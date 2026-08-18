package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

const (
	maximumConversionEvents = 10_000
	maximumConversionPages  = 100
	maximumWebhookEvents    = 1_000
	conversionAttempts      = 3
	attributionWindow       = 30 * 24 * time.Hour
	webhookClockSkew        = 10 * time.Minute
)

type ConversionProviderSelector interface {
	Select(string) ports.ConversionProviderAdapter
}

type WebhookResult struct {
	Accepted  int  `json:"accepted"`
	Duplicate bool `json:"duplicate"`
}

type ConversionService struct {
	repository ports.ConversionRepository
	providers  ConversionProviderSelector
	now        func() time.Time
}

func NewConversionService(repository ports.ConversionRepository, providers ConversionProviderSelector) *ConversionService {
	return &ConversionService{repository: repository, providers: providers, now: time.Now}
}

func (service *ConversionService) IngestWebhook(ctx context.Context, configurationID domain.ProviderConfigurationID, request domain.WebhookRequest) (WebhookResult, error) {
	configuration, err := service.repository.GetProviderConfiguration(ctx, configurationID)
	if err != nil {
		return WebhookResult{}, err
	}
	if !configuration.ConversionEnabled || configuration.LifecycleStatus != domain.ProviderActive && configuration.LifecycleStatus != domain.ProviderDegraded {
		return WebhookResult{}, ports.ErrProviderDisabled
	}
	adapter := service.providers.Select(configuration.AdapterKey)
	if err = adapter.ValidateConversionConfiguration(ctx, configuration); err != nil {
		return WebhookResult{}, err
	}
	verified, err := adapter.VerifyWebhook(ctx, configuration, request)
	if err != nil {
		code := "webhook.verification_failed"
		fingerprint := domain.RequestFingerprint(configurationID, request.Body, request.ReceivedAt)
		_, _ = service.repository.RecordWebhookDelivery(ctx, configurationID, fingerprint,
			domain.BodyFingerprint(request.Body), "rejected", nil, &code, request.ReceivedAt)
		return WebhookResult{}, ports.ErrWebhookRejected
	}
	if verified.SignatureTimestamp.IsZero() || verified.SignatureTimestamp.Before(request.ReceivedAt.Add(-webhookClockSkew)) ||
		verified.SignatureTimestamp.After(request.ReceivedAt.Add(webhookClockSkew)) {
		code := "webhook.timestamp_invalid"
		fingerprint := domain.RequestFingerprint(configurationID, request.Body, verified.SignatureTimestamp)
		_, _ = service.repository.RecordWebhookDelivery(ctx, configurationID, fingerprint,
			domain.BodyFingerprint(request.Body), "rejected", &verified.SignatureTimestamp, &code, request.ReceivedAt)
		return WebhookResult{}, ports.ErrWebhookRejected
	}
	if len(verified.Events) > maximumWebhookEvents {
		return WebhookResult{}, ErrInvalidAttribution
	}
	requestFingerprint := domain.RequestFingerprint(configurationID, request.Body, verified.SignatureTimestamp)
	delivery, err := service.repository.RecordWebhookDelivery(ctx, configurationID, requestFingerprint,
		domain.BodyFingerprint(request.Body), "verified", &verified.SignatureTimestamp, nil, request.ReceivedAt)
	if err != nil {
		return WebhookResult{}, err
	}
	if delivery.VerificationState != "verified" {
		return WebhookResult{}, ports.ErrWebhookRejected
	}
	if delivery.Processed {
		return WebhookResult{Duplicate: true}, nil
	}
	events, err := service.prepareEvents(ctx, configuration, verified.Events, request.ReceivedAt, &delivery.ID, nil)
	if err != nil {
		return WebhookResult{}, err
	}
	applied, err := service.repository.ApplyWebhookEvents(ctx, delivery.ID, configuration, events, service.now().UTC())
	return WebhookResult{Accepted: applied}, err
}

func (service *ConversionService) TriggerManualImport(ctx context.Context, actor identity.UserID, configurationID domain.ProviderConfigurationID, key string) (domain.ConversionImportRun, error) {
	key = strings.TrimSpace(key)
	if actor == "" || len(key) < 8 || len(key) > 200 {
		return domain.ConversionImportRun{}, ErrInvalidAttribution
	}
	configuration, err := service.repository.GetProviderConfiguration(ctx, configurationID)
	if err != nil {
		return domain.ConversionImportRun{}, err
	}
	if !configuration.ConversionEnabled || configuration.LifecycleStatus != domain.ProviderActive && configuration.LifecycleStatus != domain.ProviderDegraded {
		return domain.ConversionImportRun{}, ports.ErrProviderDisabled
	}
	if err = service.providers.Select(configuration.AdapterKey).ValidateConversionConfiguration(ctx, configuration); err != nil {
		return domain.ConversionImportRun{}, err
	}
	return service.repository.QueueConversionImport(ctx, &actor, configurationID, domain.ImportManual, key, conversionAttempts)
}

func (service *ConversionService) SetProviderEnabled(ctx context.Context, actor identity.UserID, configurationID domain.ProviderConfigurationID, enabled bool) (domain.ProviderConfiguration, error) {
	if actor == "" || configurationID == "" {
		return domain.ProviderConfiguration{}, ErrInvalidAttribution
	}
	configuration, err := service.repository.GetProviderConfiguration(ctx, configurationID)
	if err != nil {
		return domain.ProviderConfiguration{}, err
	}
	if enabled {
		if configuration.CredentialReference == nil ||
			(configuration.LifecycleStatus != domain.ProviderActive && configuration.LifecycleStatus != domain.ProviderDegraded) {
			return domain.ProviderConfiguration{}, ports.ErrProviderDisabled
		}
		if err = service.providers.Select(configuration.AdapterKey).ValidateConversionConfiguration(ctx, configuration); err != nil {
			return domain.ProviderConfiguration{}, err
		}
	}
	return service.repository.SetConversionProviderEnabled(ctx, actor, configurationID, enabled, service.now().UTC())
}

func (service *ConversionService) RetryImport(ctx context.Context, actor identity.UserID, id domain.ConversionImportRunID, key string) (domain.ConversionImportRun, error) {
	if actor == "" || id == "" || len(strings.TrimSpace(key)) < 8 || len(key) > 200 {
		return domain.ConversionImportRun{}, ErrInvalidAttribution
	}
	return service.repository.RetryConversionImport(ctx, actor, id, strings.TrimSpace(key))
}

func (service *ConversionService) QueueScheduled(ctx context.Context, limit int) (int, error) {
	if limit < 1 || limit > 100 {
		return 0, ErrInvalidAttribution
	}
	return service.repository.QueueDueConversionImports(ctx, service.now().UTC(), limit)
}

func (service *ConversionService) RecoverStalled(ctx context.Context, lease time.Duration, limit int) (int, error) {
	if lease < time.Minute || limit < 1 || limit > 100 {
		return 0, ErrInvalidAttribution
	}
	now := service.now().UTC()
	return service.repository.RecoverStalledConversionImports(ctx, now.Add(-lease), now, limit)
}

func (service *ConversionService) ProcessNext(ctx context.Context) (bool, error) {
	run, err := service.repository.ClaimNextConversionImport(ctx, service.now().UTC())
	if errors.Is(err, ports.ErrImportNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	adapter := service.providers.Select(run.ProviderConfiguration.AdapterKey)
	if err = adapter.ValidateConversionConfiguration(ctx, run.ProviderConfiguration); err != nil {
		return true, service.failImport(ctx, run, err)
	}
	cursor := run.CursorBefore
	total, rejected := 0, 0
	for page := 0; page < maximumConversionPages; page++ {
		batch, fetchErr := adapter.FetchConversions(ctx, run.ProviderConfiguration, cursor)
		if fetchErr != nil {
			return true, service.failImport(ctx, run, fetchErr)
		}
		if total+len(batch.Events) > maximumConversionEvents {
			return true, service.failImport(ctx, run, errors.New("conversion import exceeded record limit"))
		}
		valid, failures := service.prepareImportEvents(ctx, run, batch.Events)
		applied, applyErr := service.repository.ApplyConversionImport(ctx, run, valid, failures, batch, service.now().UTC())
		if applyErr != nil {
			return true, service.failImport(ctx, run, applyErr)
		}
		total += len(batch.Events)
		rejected += len(failures)
		run.RecordsApplied += applied
		run.RecordsReceived += len(batch.Events)
		run.RecordsRejected += len(failures)
		run.CursorAfter = batch.NextCursor
		if batch.Complete {
			if err = service.repository.CompleteConversionImport(ctx, run, batch, total-rejected, rejected, service.now().UTC()); err != nil {
				return true, err
			}
			if rejected == 0 && batch.CoverageStart != nil && batch.CoverageEnd != nil {
				_, err = service.repository.ReconcileConversionImport(ctx, nil, run.ID, "automatic-"+string(run.ID), service.now().UTC())
			}
			return true, err
		}
		if batch.NextCursor == nil || cursor != nil && *cursor == *batch.NextCursor {
			return true, service.failImport(ctx, run, errors.New("conversion cursor did not advance"))
		}
		cursor = batch.NextCursor
	}
	return true, service.failImport(ctx, run, errors.New("conversion import exceeded page limit"))
}

func (service *ConversionService) ListConversions(ctx context.Context, filter domain.ConversionFilter) ([]domain.Conversion, int64, error) {
	if filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 {
		return nil, 0, ErrInvalidAttribution
	}
	return service.repository.ListConversions(ctx, filter)
}

func (service *ConversionService) ListImports(ctx context.Context, page, pageSize int) ([]domain.ConversionImportRun, int64, error) {
	offset, limit, err := importPagination(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return service.repository.ListConversionImports(ctx, limit, offset)
}

func (service *ConversionService) ListReconciliations(ctx context.Context, page, pageSize int) ([]domain.ReconciliationRun, int64, error) {
	offset, limit, err := importPagination(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return service.repository.ListReconciliations(ctx, limit, offset)
}

func (service *ConversionService) Reconcile(ctx context.Context, actor identity.UserID, runID domain.ConversionImportRunID, key string) (domain.ReconciliationRun, error) {
	if actor == "" || runID == "" || len(strings.TrimSpace(key)) < 8 || len(key) > 200 {
		return domain.ReconciliationRun{}, ErrInvalidAttribution
	}
	return service.repository.ReconcileConversionImport(ctx, &actor, runID, strings.TrimSpace(key), service.now().UTC())
}

func (service *ConversionService) Metrics(ctx context.Context, start, end time.Time) (domain.MonetizationReport, error) {
	start, end = start.UTC(), end.UTC()
	if start.IsZero() || end.IsZero() || !end.After(start) || end.Sub(start) > 366*24*time.Hour || end.After(service.now().UTC().Add(time.Minute)) {
		return domain.MonetizationReport{}, ErrInvalidAttribution
	}
	return service.repository.MonetizationReport(ctx, start, end)
}

func (service *ConversionService) prepareImportEvents(ctx context.Context, run domain.ConversionImportRun, input []domain.ProviderConversionEvent) ([]domain.VerifiedConversionEvent, []domain.ImportRecordFailure) {
	result := make([]domain.VerifiedConversionEvent, 0, len(input))
	failures := make([]domain.ImportRecordFailure, 0)
	seen := make(map[string]bool, len(input))
	for _, raw := range input {
		if seen[strings.TrimSpace(raw.ProviderEventID)] {
			fingerprint := domain.ConversionEventFingerprint(raw)
			failures = append(failures, domain.ImportRecordFailure{ExternalRecordID: rawString(raw.ProviderEventID),
				Code: "conversion.duplicate_in_batch", Message: "Duplicate provider event ID in import page.", RecordFingerprint: &fingerprint})
			continue
		}
		seen[strings.TrimSpace(raw.ProviderEventID)] = true
		validated, err := service.prepareEvents(ctx, run.ProviderConfiguration, []domain.ProviderConversionEvent{raw}, service.now().UTC(), nil, &run.ID)
		if err != nil {
			fingerprint := domain.ConversionEventFingerprint(raw)
			failures = append(failures, domain.ImportRecordFailure{ExternalRecordID: rawString(raw.ProviderEventID),
				Code: "conversion.invalid", Message: "Provider conversion event failed validation.", RecordFingerprint: &fingerprint})
			continue
		}
		result = append(result, validated[0])
	}
	return result, failures
}

func (service *ConversionService) prepareEvents(ctx context.Context, configuration domain.ProviderConfiguration, input []domain.ProviderConversionEvent, receivedAt time.Time, deliveryID *domain.WebhookDeliveryID, importID *domain.ConversionImportRunID) ([]domain.VerifiedConversionEvent, error) {
	result := make([]domain.VerifiedConversionEvent, 0, len(input))
	seen := make(map[string]bool, len(input))
	for _, raw := range input {
		event, err := domain.ValidateProviderConversionEvent(raw, receivedAt)
		if err != nil || seen[event.ProviderEventID] {
			return nil, ErrInvalidAttribution
		}
		seen[event.ProviderEventID] = true
		attribution, err := service.repository.ResolveConversionAttribution(ctx, configuration, event.ClickID,
			event.EventTimestamp, attributionWindow)
		if err != nil {
			return nil, err
		}
		result = append(result, domain.VerifiedConversionEvent{ProviderConversionEvent: event,
			ProviderConfigurationID: configuration.ID, Provider: configuration.ProviderKey,
			MerchantID: configuration.MerchantID, WebhookDeliveryID: deliveryID, ImportRunID: importID,
			ReceivedAt: receivedAt, PayloadFingerprint: domain.ConversionEventFingerprint(event), Attribution: attribution})
	}
	return result, nil
}

func (service *ConversionService) failImport(ctx context.Context, run domain.ConversionImportRun, cause error) error {
	code := "provider.failed"
	if errors.Is(cause, ports.ErrProviderDisabled) {
		code = "provider.disabled"
	} else if errors.Is(cause, ports.ErrProviderUnavailable) {
		code = "provider.unavailable"
	}
	return service.repository.FailConversionImport(ctx, run, code, boundedError(cause), service.now().UTC())
}

func rawString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
