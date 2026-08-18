package commercepostgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/commerce/domain"
	identity "rigmark/internal/modules/identity/domain"
)

func TestProviderImportIsIdempotentAndHistoricallyAuditable(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)

	var merchantID, productID string
	if err = pool.QueryRow(ctx, `SELECT id FROM commerce.merchants ORDER BY id LIMIT 1`).Scan(&merchantID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT id FROM catalog.products ORDER BY id LIMIT 1`).Scan(&productID); err != nil {
		t.Fatal(err)
	}
	var actor identity.UserID
	nonce := time.Now().UnixNano()
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email,status) VALUES ($1,'active') RETURNING id`,
		fmt.Sprintf("commerce-operator-%d@example.invalid", nonce)).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, actor)
	})

	configuration, err := repository.CreateProviderConfiguration(ctx, actor, domain.ProviderConfigurationInput{
		MerchantID: domain.MerchantID(merchantID), ProviderKey: fmt.Sprintf("fixture-%d", nonce),
		AdapterKey: "fixture", ExternalMerchantID: fmt.Sprintf("merchant-%d", nonce),
		ScheduleIntervalMinutes: 360, FreshnessTTLMinutes: 4320,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `UPDATE commerce.provider_configurations SET credential_reference='secret/fixture',
		configuration_verified_at=now(), lifecycle_status='active', next_import_at=now() WHERE id=$1`, configuration.ID)
	if err != nil {
		t.Fatal(err)
	}

	key := fmt.Sprintf("manual-%d", nonce)
	first, err := repository.QueueImport(ctx, &actor, configuration.ID, domain.ImportManual, key, 3)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.QueueImport(ctx, &actor, configuration.ID, domain.ImportManual, key, 3)
	if err != nil || first.ID != second.ID {
		t.Fatalf("idempotent queue = %q/%q, %v", first.ID, second.ID, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.offer_import_runs
		SET status='running', attempt_count=1, started_at=now() WHERE id=$1`, first.ID); err != nil {
		t.Fatal(err)
	}
	run := first
	run.Status = domain.ImportRunning
	run.AttemptCount = 1
	observedAt := time.Now().UTC()
	record, err := domain.ValidateProviderOffer(domain.ProviderOffer{
		ExternalOfferID: fmt.Sprintf("offer-%d", nonce), ProductID: catalog.ProductID(productID),
		MerchantSKU: fmt.Sprintf("fixture-sku-%d", nonce), ProductURL: "https://merchant.example/item",
		PriceMinor: 12900, Currency: "USD", Availability: "in_stock", Condition: "new",
	}, observedAt, 72*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.ApplyImport(ctx, run, []domain.ValidatedOffer{record}, nil,
		domain.ProviderBatch{Complete: true}, observedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.CompleteImport(ctx, run.ID, result, observedAt); err != nil {
		t.Fatal(err)
	}
	var prices, availability, mappings int
	if err = pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM commerce.price_observations WHERE provider_configuration_id=$1),
		(SELECT count(*) FROM commerce.availability_observations WHERE provider_configuration_id=$1),
		(SELECT count(*) FROM commerce.provider_offer_mappings WHERE provider_configuration_id=$1)`,
		configuration.ID).Scan(&prices, &availability, &mappings); err != nil {
		t.Fatal(err)
	}
	if prices != 1 || availability != 1 || mappings != 1 {
		t.Fatalf("observations/mappings = %d/%d/%d", prices, availability, mappings)
	}
	_, err = pool.Exec(ctx, `UPDATE commerce.price_observations SET price_minor=1
		WHERE provider_configuration_id=$1`, configuration.ID)
	if err == nil {
		t.Fatal("immutable price observation was updated")
	}

	failureRun, err := repository.QueueImport(ctx, &actor, configuration.ID, domain.ImportManual,
		fmt.Sprintf("failure-%d", nonce), 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.offer_import_runs
		SET status='running', attempt_count=1, started_at=now() WHERE id=$1`, failureRun.ID); err != nil {
		t.Fatal(err)
	}
	failureRun.Status = domain.ImportRunning
	failureRun.AttemptCount = 1
	failureRun.MaxAttempts = 1
	failedAt := time.Now().UTC()
	if err = repository.FailImport(ctx, failureRun, "provider.disabled", "provider is disabled", failedAt); err != nil {
		t.Fatal(err)
	}
	var failureStatus domain.ImportStatus
	var completedAt *time.Time
	if err = pool.QueryRow(ctx, `SELECT status, completed_at FROM commerce.offer_import_runs WHERE id=$1`,
		failureRun.ID).Scan(&failureStatus, &completedAt); err != nil {
		t.Fatal(err)
	}
	if failureStatus != domain.ImportFailed || completedAt == nil {
		t.Fatalf("failed import state = %q/%v, want failed/completed", failureStatus, completedAt)
	}

	stalledRun, err := repository.QueueImport(ctx, &actor, configuration.ID, domain.ImportManual,
		fmt.Sprintf("stalled-%d", nonce), 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.offer_import_runs SET status='running',
		attempt_count=1, started_at=now()-interval '2 hours', updated_at=now()-interval '2 hours'
		WHERE id=$1`, stalledRun.ID); err != nil {
		t.Fatal(err)
	}
	recoveredAt := time.Now().UTC()
	recovered, err := repository.RecoverStalledImports(ctx, recoveredAt.Add(-time.Hour), recoveredAt, 10)
	if err != nil {
		t.Fatal(err)
	}
	if recovered < 1 {
		t.Fatalf("recovered imports = %d, want at least 1", recovered)
	}
	var stalledStatus domain.ImportStatus
	var stalledCode *string
	completedAt = nil
	if err = pool.QueryRow(ctx, `SELECT status, error_code, completed_at
		FROM commerce.offer_import_runs WHERE id=$1`, stalledRun.ID).Scan(
		&stalledStatus, &stalledCode, &completedAt,
	); err != nil {
		t.Fatal(err)
	}
	if stalledStatus != domain.ImportFailed || stalledCode == nil || *stalledCode != "worker.lease_expired" || completedAt == nil {
		t.Fatalf("stalled import state = %q/%v/%v", stalledStatus, stalledCode, completedAt)
	}
}

func TestAffiliateClickIdempotencyAndFiltering(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)
	var offerID domain.OfferID
	if err = pool.QueryRow(ctx, `SELECT offers.id FROM commerce.merchant_offers offers
		JOIN commerce.affiliate_links links ON links.merchant_offer_id=offers.id
		WHERE offers.is_active AND links.is_active ORDER BY offers.id LIMIT 1`).Scan(&offerID); err != nil {
		t.Fatal(err)
	}
	anonymous := fmt.Sprintf("bot-%d", time.Now().UnixNano())
	key := "prefetch-" + anonymous
	click := domain.AffiliateClick{OfferID: offerID, AnonymousID: &anonymous, Source: "product_detail",
		Classification: domain.ClickPrefetch, IsCountable: false, IdempotencyKey: &key,
		RetentionExpires: time.Now().Add(24 * time.Hour)}
	destination, err := repository.ResolveOfferDestination(ctx, click)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.RecordClick(ctx, destination, click); err != nil {
		t.Fatal(err)
	}
	if err = repository.RecordClick(ctx, destination, click); err != nil {
		t.Fatal(err)
	}
	var raw, countable, events int
	if err = pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM commerce.affiliate_clicks WHERE idempotency_key=$1),
		(SELECT count(*) FROM commerce.affiliate_clicks WHERE idempotency_key=$1 AND is_countable),
		(SELECT count(*) FROM analytics.events WHERE anonymous_id=$2 AND event_name='affiliate_clicked')`,
		key, anonymous).Scan(&raw, &countable, &events); err != nil {
		t.Fatal(err)
	}
	if raw != 1 || countable != 0 || events != 0 {
		t.Fatalf("raw/countable/events = %d/%d/%d", raw, countable, events)
	}
	_, _ = pool.Exec(ctx, `DELETE FROM commerce.affiliate_clicks WHERE idempotency_key=$1`, key)
}
