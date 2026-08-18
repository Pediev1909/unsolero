package commercepostgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

func TestVerifiedConversionImportLifecycleMetricsAndImmutability(t *testing.T) {
	pool, ctx := conversionTestPool(t)
	repository := New(pool)
	configuration := createConversionConfiguration(t, ctx, pool)
	clickID := createCountableClick(t, ctx, pool, configuration.MerchantID)
	now := time.Now().UTC().Truncate(time.Microsecond)
	coverageStart, coverageEnd := now.Add(-time.Hour), now.Add(time.Hour)
	run, err := repository.QueueConversionImport(ctx, nil, configuration.ID, domain.ImportManual,
		fmt.Sprintf("conversion-%d", now.UnixNano()), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.conversion_import_runs SET status='running',
		attempt_count=1,started_at=$2 WHERE id=$1`, run.ID, now); err != nil {
		t.Fatal(err)
	}
	run.Status, run.AttemptCount = domain.ImportRunning, 1
	commission, currency := int64(1250), "USD"
	commissionStatus := domain.CommissionPending
	providerEvent := domain.ProviderConversionEvent{ProviderEventID: fmt.Sprintf("event-%d", now.UnixNano()),
		EventType: domain.EventConversionCreated, ExternalConversionID: fmt.Sprintf("conversion-%d", now.UnixNano()),
		OrderStatus: domain.OrderConfirmed, CommissionMinor: &commission,
		CommissionCurrency: &currency, CommissionStatus: &commissionStatus,
		ClickID: &clickID, EventTimestamp: now}
	attribution, err := repository.ResolveConversionAttribution(ctx, configuration, &clickID, now, 30*24*time.Hour)
	if err != nil || attribution.Status != "attributed" {
		t.Fatalf("attribution=%#v err=%v", attribution, err)
	}
	event := domain.VerifiedConversionEvent{ProviderConversionEvent: providerEvent,
		ProviderConfigurationID: configuration.ID, Provider: configuration.ProviderKey,
		MerchantID: configuration.MerchantID, ImportRunID: &run.ID, ReceivedAt: now,
		PayloadFingerprint: domain.ConversionEventFingerprint(providerEvent), Attribution: attribution}
	batch := domain.ConversionBatch{Complete: true, CoverageStart: &coverageStart, CoverageEnd: &coverageEnd}
	applied, err := repository.ApplyConversionImport(ctx, run, []domain.VerifiedConversionEvent{event}, nil, batch, now)
	if err != nil || applied != 1 {
		t.Fatalf("apply=%d err=%v", applied, err)
	}
	if err = repository.CompleteConversionImport(ctx, run, batch, 1, 0, now); err != nil {
		t.Fatal(err)
	}
	reconciliation, err := repository.ReconcileConversionImport(ctx, nil, run.ID,
		"automatic-"+string(run.ID), now)
	if err != nil || reconciliation.Status != "succeeded" || reconciliation.Matched != 1 {
		t.Fatalf("reconciliation=%#v err=%v", reconciliation, err)
	}
	report, err := repository.MonetizationReport(ctx, coverageStart, coverageEnd)
	if err != nil {
		t.Fatal(err)
	}
	if report.ConversionRate.Status != domain.MetricAvailable || report.ConversionRate.Numerator != 1 ||
		report.ConversionRate.Denominator < 1 {
		t.Fatalf("conversion rate=%#v", report.ConversionRate)
	}
	if report.Commission.Status != domain.MetricAvailable || len(report.Commission.Values) != 0 {
		t.Fatalf("pending commission was reported as earned: %#v", report.Commission)
	}
	approved := event
	approved.ProviderEventID = fmt.Sprintf("approved-%d", now.UnixNano())
	approved.EventType = domain.EventCommissionChanged
	approvedStatus := domain.CommissionApproved
	approved.CommissionStatus = &approvedStatus
	approved.EventTimestamp = now.Add(time.Minute)
	approved.ReceivedAt = approved.EventTimestamp
	approved.PayloadFingerprint = domain.ConversionEventFingerprint(approved.ProviderConversionEvent)
	approved.ImportRunID = nil
	approvedDelivery := createVerifiedDelivery(t, ctx, repository, configuration.ID, now, 601)
	approved.WebhookDeliveryID = &approvedDelivery
	if applied, applyErr := repository.ApplyWebhookEvents(ctx, approvedDelivery, configuration,
		[]domain.VerifiedConversionEvent{approved}, approved.ReceivedAt); applyErr != nil || applied != 1 {
		t.Fatalf("approved apply=%d err=%v", applied, applyErr)
	}
	report, err = repository.MonetizationReport(ctx, coverageStart, coverageEnd)
	if err != nil || len(report.Commission.Values) != 1 || report.Commission.Values[0].AmountMinor != commission {
		t.Fatalf("approved commission report=%#v err=%v", report.Commission, err)
	}
	reversed := approved
	reversed.ProviderEventID = fmt.Sprintf("reversed-%d", now.UnixNano())
	reversed.EventType = domain.EventCommissionChanged
	reversedStatus := domain.CommissionReversed
	reversed.CommissionStatus = &reversedStatus
	reversed.EventTimestamp = now.Add(2 * time.Minute)
	reversed.ReceivedAt = reversed.EventTimestamp
	reversed.PayloadFingerprint = domain.ConversionEventFingerprint(reversed.ProviderConversionEvent)
	reversedDelivery := createVerifiedDelivery(t, ctx, repository, configuration.ID, now, 701)
	reversed.WebhookDeliveryID = &reversedDelivery
	if applied, applyErr := repository.ApplyWebhookEvents(ctx, reversedDelivery, configuration,
		[]domain.VerifiedConversionEvent{reversed}, reversed.ReceivedAt); applyErr != nil || applied != 1 {
		t.Fatalf("reversed apply=%d err=%v", applied, applyErr)
	}
	report, err = repository.MonetizationReport(ctx, coverageStart, coverageEnd)
	if err != nil || len(report.Commission.Values) != 0 || report.ReversalRate.Status != domain.MetricAvailable ||
		report.ReversalRate.Numerator != 1 || report.ReversalRate.Denominator != 1 {
		t.Fatalf("reversed report commission=%#v reversal=%#v err=%v", report.Commission, report.ReversalRate, err)
	}

	duplicateDelivery, err := repository.RecordWebhookDelivery(ctx, configuration.ID,
		fmt.Sprintf("%064x", now.UnixNano()+501), fmt.Sprintf("%064x", now.UnixNano()+502),
		"verified", &now, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	duplicateEvent := event
	duplicateEvent.ImportRunID = nil
	duplicateEvent.WebhookDeliveryID = &duplicateDelivery.ID
	duplicate, err := repository.ApplyWebhookEvents(ctx, duplicateDelivery.ID, configuration,
		[]domain.VerifiedConversionEvent{duplicateEvent}, now)
	if err != nil || duplicate != 0 {
		t.Fatalf("duplicate apply=%d err=%v", duplicate, err)
	}
	conflictDelivery, err := repository.RecordWebhookDelivery(ctx, configuration.ID,
		fmt.Sprintf("%064x", now.UnixNano()+503), fmt.Sprintf("%064x", now.UnixNano()+504),
		"verified", &now, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	conflicting := duplicateEvent
	conflicting.WebhookDeliveryID = &conflictDelivery.ID
	conflicting.CommissionMinor = pointerInt64(9999)
	conflicting.PayloadFingerprint = domain.ConversionEventFingerprint(conflicting.ProviderConversionEvent)
	if _, err = repository.ApplyWebhookEvents(ctx, conflictDelivery.ID, configuration,
		[]domain.VerifiedConversionEvent{conflicting}, now); !errors.Is(err, ports.ErrConversionConflict) {
		t.Fatalf("conflicting duplicate error=%v", err)
	}
	var eventID string
	if err = pool.QueryRow(ctx, `SELECT id FROM commerce.conversion_events
		WHERE provider_configuration_id=$1 AND provider_event_id=$2`, configuration.ID,
		providerEvent.ProviderEventID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.conversion_events SET order_status='rejected' WHERE id=$1`, eventID); err == nil {
		t.Fatal("immutable verified conversion event was updated")
	}
}

func TestConcurrentDuplicateWebhookEventCreatesOneFact(t *testing.T) {
	pool, ctx := conversionTestPool(t)
	repository := New(pool)
	configuration := createConversionConfiguration(t, ctx, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	providerEvent := domain.ProviderConversionEvent{ProviderEventID: fmt.Sprintf("concurrent-%d", now.UnixNano()),
		EventType: domain.EventConversionCreated, ExternalConversionID: fmt.Sprintf("order-%d", now.UnixNano()),
		OrderStatus: domain.OrderPending, EventTimestamp: now}
	var deliveries [2]domain.WebhookDeliveryID
	for index := range deliveries {
		delivery, err := repository.RecordWebhookDelivery(ctx, configuration.ID,
			fmt.Sprintf("%064x", now.UnixNano()+int64(index)+1), fmt.Sprintf("%064x", now.UnixNano()+100+int64(index)),
			"verified", &now, nil, now)
		if err != nil {
			t.Fatal(err)
		}
		deliveries[index] = delivery.ID
	}
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, delivery := range deliveries {
		wait.Add(1)
		go func(deliveryID domain.WebhookDeliveryID) {
			defer wait.Done()
			event := domain.VerifiedConversionEvent{ProviderConversionEvent: providerEvent,
				ProviderConfigurationID: configuration.ID, Provider: configuration.ProviderKey,
				MerchantID: configuration.MerchantID, WebhookDeliveryID: &deliveryID, ReceivedAt: now,
				PayloadFingerprint: domain.ConversionEventFingerprint(providerEvent),
				Attribution:        domain.ConversionAttribution{Status: "unattributed"}}
			_, err := repository.ApplyWebhookEvents(context.Background(), deliveryID, configuration,
				[]domain.VerifiedConversionEvent{event}, now)
			results <- err
		}(delivery)
	}
	wait.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM commerce.conversion_events
		WHERE provider_configuration_id=$1 AND provider_event_id=$2`, configuration.ID,
		providerEvent.ProviderEventID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("fact count=%d err=%v", count, err)
	}
}

func TestWebhookRetryResumesUnprocessedAndAcknowledgesProcessed(t *testing.T) {
	pool, ctx := conversionTestPool(t)
	repository := New(pool)
	configuration := createConversionConfiguration(t, ctx, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	requestFingerprint := fmt.Sprintf("%064x", now.UnixNano()+801)
	bodyFingerprint := fmt.Sprintf("%064x", now.UnixNano()+802)
	delivery, err := repository.RecordWebhookDelivery(ctx, configuration.ID, requestFingerprint,
		bodyFingerprint, "verified", &now, nil, now)
	if err != nil || delivery.Processed {
		t.Fatalf("initial delivery=%#v err=%v", delivery, err)
	}
	retry, err := repository.RecordWebhookDelivery(ctx, configuration.ID, requestFingerprint,
		bodyFingerprint, "verified", &now, nil, now)
	if err != nil || retry.ID != delivery.ID || retry.Processed {
		t.Fatalf("unprocessed retry=%#v err=%v", retry, err)
	}
	providerEvent := domain.ProviderConversionEvent{ProviderEventID: fmt.Sprintf("retry-%d", now.UnixNano()),
		EventType: domain.EventConversionCreated, ExternalConversionID: fmt.Sprintf("retry-order-%d", now.UnixNano()),
		OrderStatus: domain.OrderPending, EventTimestamp: now}
	verified := domain.VerifiedConversionEvent{ProviderConversionEvent: providerEvent,
		ProviderConfigurationID: configuration.ID, Provider: configuration.ProviderKey,
		MerchantID: configuration.MerchantID, WebhookDeliveryID: &delivery.ID, ReceivedAt: now,
		PayloadFingerprint: domain.ConversionEventFingerprint(providerEvent),
		Attribution:        domain.ConversionAttribution{Status: "unattributed"}}
	if _, err = repository.ApplyWebhookEvents(ctx, delivery.ID, configuration,
		[]domain.VerifiedConversionEvent{verified}, now); err != nil {
		t.Fatal(err)
	}
	processed, err := repository.RecordWebhookDelivery(ctx, configuration.ID, requestFingerprint,
		bodyFingerprint, "verified", &now, nil, now)
	if err != nil || !processed.Processed || processed.ID != delivery.ID {
		t.Fatalf("processed retry=%#v err=%v", processed, err)
	}
}

func conversionTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool, ctx
}

func createConversionConfiguration(t *testing.T, ctx context.Context, pool *pgxpool.Pool) domain.ProviderConfiguration {
	t.Helper()
	var merchantID string
	if err := pool.QueryRow(ctx, `SELECT id FROM commerce.merchants ORDER BY id LIMIT 1`).Scan(&merchantID); err != nil {
		t.Fatal(err)
	}
	var id domain.ProviderConfigurationID
	nonce := time.Now().UnixNano()
	if err := pool.QueryRow(ctx, `INSERT INTO commerce.provider_configurations
		(merchant_id,provider_key,adapter_key,external_merchant_id,credential_reference,
		 lifecycle_status,configuration_verified_at,conversion_ingestion_enabled,
		 conversion_configuration_verified_at)
		VALUES($1,$2,'fixture',$3,'secret/test','active',now(),true,now()) RETURNING id`,
		merchantID, fmt.Sprintf("fixture-%d", nonce), fmt.Sprintf("merchant-%d", nonce)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	configuration, err := New(pool).GetProviderConfiguration(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return configuration
}

func createCountableClick(t *testing.T, ctx context.Context, pool *pgxpool.Pool, merchantID domain.MerchantID) string {
	t.Helper()
	var linkID, offerID, productID string
	if err := pool.QueryRow(ctx, `SELECT links.id,offers.id,offers.product_id
		FROM commerce.affiliate_links links JOIN commerce.merchant_offers offers ON offers.id=links.merchant_offer_id
		WHERE offers.merchant_id=$1 ORDER BY offers.id LIMIT 1`, merchantID).Scan(&linkID, &offerID, &productID); err != nil {
		t.Fatal(err)
	}
	anonymous := fmt.Sprintf("conversion-click-%d", time.Now().UnixNano())
	var clickID string
	if err := pool.QueryRow(ctx, `INSERT INTO commerce.affiliate_clicks
		(affiliate_link_id,merchant_offer_id,product_id,anonymous_id,session_id,source,
		 classification,is_countable,retention_expires_at)
		VALUES($1,$2,$3,$4,$4,'product_detail','human',true,now()+interval '30 days') RETURNING id`,
		linkID, offerID, productID, anonymous).Scan(&clickID); err != nil {
		t.Fatal(err)
	}
	return clickID
}

func createVerifiedDelivery(t *testing.T, ctx context.Context, repository *Repository,
	configurationID domain.ProviderConfigurationID, now time.Time, offset int64) domain.WebhookDeliveryID {
	t.Helper()
	delivery, err := repository.RecordWebhookDelivery(ctx, configurationID,
		fmt.Sprintf("%064x", now.UnixNano()+offset), fmt.Sprintf("%064x", now.UnixNano()+offset+1),
		"verified", &now, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	return delivery.ID
}

func pointerInt64(value int64) *int64 { return &value }
