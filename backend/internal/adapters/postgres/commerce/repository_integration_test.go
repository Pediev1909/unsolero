package commercepostgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

func TestConcurrentAffiliateClickIdempotencyCreatesOneClickAndEvent(t *testing.T) {
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
	requestID := fmt.Sprintf("affiliate-concurrent-%d", time.Now().UnixNano())
	anonymousID := "phase8-" + requestID
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics.events WHERE request_id=$1`, requestID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_clicks WHERE idempotency_key=$1`, requestID)
	})

	var offerID domain.OfferID
	if err := pool.QueryRow(ctx, `SELECT offers.id FROM commerce.merchant_offers offers
		JOIN commerce.affiliate_links links ON links.merchant_offer_id=offers.id
		JOIN commerce.merchants merchants ON merchants.id=offers.merchant_id
		WHERE offers.is_active AND links.is_active AND merchants.status='active'
		AND offers.availability IN ('in_stock','backorder') AND offers.last_checked_at>=now()-interval '72 hours'
		AND (offers.expires_at IS NULL OR offers.expires_at>now()) ORDER BY offers.id LIMIT 1`).Scan(&offerID); err != nil {
		t.Fatal(err)
	}
	click := domain.AffiliateClick{OfferID: offerID, AnonymousID: &anonymousID, Source: "product_detail", RequestID: &requestID,
		IdempotencyKey: &requestID, Classification: domain.ClickHuman, IsCountable: true,
		RetentionExpires: time.Now().Add(24 * time.Hour)}
	destination, err := repository.ResolveOfferDestination(ctx, click)
	if err != nil {
		t.Fatal(err)
	}

	const contenders = 12
	start := make(chan struct{})
	errorsChannel := make(chan error, contenders)
	var wait sync.WaitGroup
	for range contenders {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			errorsChannel <- repository.RecordClick(context.Background(), destination, click)
		}()
	}
	close(start)
	wait.Wait()
	close(errorsChannel)
	for recordErr := range errorsChannel {
		if recordErr != nil {
			t.Fatalf("concurrent RecordClick() error = %v", recordErr)
		}
	}

	var clicks, events int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM commerce.affiliate_clicks WHERE idempotency_key=$1`, requestID).Scan(&clicks); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics.events WHERE request_id=$1 AND event_name='affiliate_clicked'`, requestID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || events != 1 {
		t.Fatalf("concurrent idempotency clicks=%d events=%d, want 1/1", clicks, events)
	}
}

func TestAffiliateClickPersistsAttributionAndAnalyticsEvent(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var userID string
	email := fmt.Sprintf("affiliate-click-%d@example.invalid", time.Now().UnixNano())
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email, status) VALUES ($1, 'active') RETURNING id`, email).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	requestID := fmt.Sprintf("affiliate-integration-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics.events WHERE request_id = $1`, requestID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_clicks WHERE request_id = $1`, requestID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id = $1`, userID)
	})

	var offerID domain.OfferID
	if err = pool.QueryRow(ctx, `
		SELECT offers.id
		FROM commerce.merchant_offers offers
		JOIN commerce.affiliate_links links ON links.merchant_offer_id = offers.id
		WHERE offers.is_active AND links.is_active
		ORDER BY offers.id LIMIT 1`).Scan(&offerID); err != nil {
		t.Fatalf("load offer: %v", err)
	}
	sessionID := "integration-session"
	campaign := "strength_launch"
	referrer := "https://rigmark.example"
	trafficSource, trafficMedium, referrerHost := "newsletter", "email", "rigmark.example"
	click := domain.AffiliateClick{OfferID: offerID, UserID: &userID, SessionID: &sessionID,
		Source: "product_detail", Campaign: &campaign, Referrer: &referrer, RequestID: &requestID,
		TrafficSource: &trafficSource, TrafficMedium: &trafficMedium, ReferrerHost: &referrerHost}
	destination, err := resolveAndRecord(ctx, New(pool), click)
	if err != nil || destination == "" {
		t.Fatalf("TrackOfferClick() = %q, %v", destination, err)
	}

	var storedUserID, storedSession, storedSource, storedCampaign, storedReferrer string
	var storedTrafficSource, storedTrafficMedium, storedReferrerHost string
	if err = pool.QueryRow(ctx, `SELECT user_id, session_id, source, campaign, referrer,
		traffic_source, traffic_medium, referrer_host
		FROM commerce.affiliate_clicks WHERE request_id = $1`, requestID).Scan(
		&storedUserID, &storedSession, &storedSource, &storedCampaign, &storedReferrer,
		&storedTrafficSource, &storedTrafficMedium, &storedReferrerHost); err != nil {
		t.Fatalf("load affiliate click: %v", err)
	}
	if storedUserID != userID || storedSession != sessionID || storedSource != "product_detail" ||
		storedCampaign != campaign || storedReferrer != referrer || storedTrafficSource != trafficSource ||
		storedTrafficMedium != trafficMedium || storedReferrerHost != referrerHost {
		t.Fatalf("unexpected stored attribution: %q %q %q %q %q %q %q %q", storedUserID,
			storedSession, storedSource, storedCampaign, storedReferrer, storedTrafficSource,
			storedTrafficMedium, storedReferrerHost)
	}
	var eventName, eventProperties string
	if err = pool.QueryRow(ctx, `SELECT event_name,properties::text FROM analytics.events
		WHERE user_id = $1 AND event_name = 'affiliate_clicked' ORDER BY received_at DESC LIMIT 1`, userID).Scan(&eventName, &eventProperties); err != nil {
		t.Fatalf("load affiliate analytics event: %v", err)
	}
	for _, forbidden := range []string{"https://", "user_agent", "commission", "destination", "referrer"} {
		if strings.Contains(eventProperties, forbidden) {
			t.Fatalf("affiliate analytics leaked %q in %s", forbidden, eventProperties)
		}
	}

	missing := click
	missing.OfferID = "00000000-0000-4000-8000-000000000000"
	if _, err = resolveAndRecord(ctx, New(pool), missing); err != ports.ErrAffiliateDestinationNotFound {
		t.Fatalf("missing offer error = %v, want ErrAffiliateDestinationNotFound", err)
	}
}

func TestAffiliateClickRequiresOwnedRecommendationAndFreshOffer(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var ownerID, otherUserID string
	nonce := time.Now().UnixNano()
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email, status)
		VALUES ($1, 'active') RETURNING id`, fmt.Sprintf("affiliate-owner-%d@example.invalid", nonce)).Scan(&ownerID); err != nil {
		t.Fatalf("insert recommendation owner: %v", err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email, status)
		VALUES ($1, 'active') RETURNING id`, fmt.Sprintf("affiliate-other-%d@example.invalid", nonce)).Scan(&otherUserID); err != nil {
		t.Fatalf("insert unrelated user: %v", err)
	}

	var offerID domain.OfferID
	var productID string
	var price int64
	var currency string
	if err = pool.QueryRow(ctx, `
		SELECT offers.id, offers.product_id, offers.price_minor, offers.currency
		FROM commerce.merchant_offers offers
		JOIN commerce.affiliate_links links ON links.merchant_offer_id = offers.id
		WHERE offers.is_active AND links.is_active
		ORDER BY offers.id LIMIT 1`).Scan(&offerID, &productID, &price, &currency); err != nil {
		t.Fatalf("load offer: %v", err)
	}

	var sessionID, recommendationID string
	if err = pool.QueryRow(ctx, `INSERT INTO recommendation.recommendation_sessions
		(user_id, status, primary_goal, experience_level, budget_minor, currency, completed_at)
		VALUES ($1, 'completed', 'strength', 'intermediate', $2, $3, now()) RETURNING id`,
		ownerID, price, currency).Scan(&sessionID); err != nil {
		t.Fatalf("insert recommendation session: %v", err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO recommendation.recommendations
		(session_id, policy_version, engine_version, objective_score, total_price_minor,
		 currency, result_fingerprint)
		VALUES ($1, 'home-gym-v1', 'integration-engine', 80, $2, $3, $4) RETURNING id`,
		sessionID, price, currency, fmt.Sprintf("affiliate-ownership-%d", nonce)).Scan(&recommendationID); err != nil {
		t.Fatalf("insert recommendation: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO recommendation.recommendation_items
		(recommendation_id, product_id, item_type, rank, unit_price_minor, currency,
		 objective_score, reason_code, reason_summary)
		VALUES ($1, $2, 'selected', 1, $3, $4, 80, 'integration.match',
		        'Integration-test recommendation match')`, recommendationID, productID, price, currency); err != nil {
		t.Fatalf("insert recommendation item: %v", err)
	}

	ownedRequestID := fmt.Sprintf("affiliate-owned-%d", nonce)
	deniedRequestID := fmt.Sprintf("affiliate-denied-%d", nonce)
	staleRequestID := fmt.Sprintf("affiliate-stale-%d", nonce)
	t.Cleanup(func() {
		for _, requestID := range []string{ownedRequestID, deniedRequestID, staleRequestID} {
			_, _ = pool.Exec(context.Background(), `DELETE FROM analytics.events WHERE request_id = $1`, requestID)
			_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_clicks WHERE request_id = $1`, requestID)
		}
		_, _ = pool.Exec(context.Background(), `DELETE FROM recommendation.recommendation_sessions WHERE id = $1`, sessionID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id = ANY($1::uuid[])`, []string{ownerID, otherUserID})
	})

	source := "recommendation"
	owned := domain.AffiliateClick{OfferID: offerID, UserID: &ownerID, Source: source,
		RecommendationID: &recommendationID, RequestID: &ownedRequestID}
	if destination, trackErr := resolveAndRecord(ctx, New(pool), owned); trackErr != nil || destination == "" {
		t.Fatalf("owned recommendation click = %q, %v", destination, trackErr)
	}

	unowned := owned
	unowned.UserID = &otherUserID
	unowned.RequestID = &deniedRequestID
	if _, trackErr := resolveAndRecord(ctx, New(pool), unowned); !errors.Is(trackErr, ports.ErrAffiliateDestinationNotFound) {
		t.Fatalf("unowned recommendation click error = %v, want ErrAffiliateDestinationNotFound", trackErr)
	}

	stale := domain.AffiliateClick{OfferID: offerID, UserID: &ownerID, Source: "product_detail",
		RequestID: &staleRequestID}
	if _, trackErr := resolveAndRecord(ctx, New(pool, time.Nanosecond), stale); !errors.Is(trackErr, ports.ErrAffiliateDestinationNotFound) {
		t.Fatalf("stale offer click error = %v, want ErrAffiliateDestinationNotFound", trackErr)
	}
	var originalLastChecked time.Time
	var originalProviderObserved *time.Time
	var originalExpiresAt *time.Time
	if err = pool.QueryRow(ctx, `SELECT last_checked_at, provider_observed_at, expires_at
		FROM commerce.merchant_offers WHERE id=$1`, offerID).Scan(
		&originalLastChecked, &originalProviderObserved, &originalExpiresAt,
	); err != nil {
		t.Fatalf("read offer freshness: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.merchant_offers
		SET last_checked_at=now()-interval '2 minutes',
		    provider_observed_at=now()-interval '2 minutes',
		    expires_at=now()-interval '1 minute'
		WHERE id=$1`, offerID); err != nil {
		t.Fatalf("expire offer: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `UPDATE commerce.merchant_offers
			SET last_checked_at=$2, provider_observed_at=$3, expires_at=$4 WHERE id=$1`,
			offerID, originalLastChecked, originalProviderObserved, originalExpiresAt)
	})
	expired := stale
	expired.RequestID = nil
	if _, trackErr := New(pool).ResolveOfferDestination(ctx, expired); !errors.Is(trackErr, ports.ErrAffiliateDestinationNotFound) {
		t.Fatalf("expired offer click error = %v, want ErrAffiliateDestinationNotFound", trackErr)
	}
}

func resolveAndRecord(ctx context.Context, repository *Repository, click domain.AffiliateClick) (string, error) {
	destination, err := repository.ResolveOfferDestination(ctx, click)
	if err != nil {
		return "", err
	}
	click.Classification = domain.ClickHuman
	click.IsCountable = true
	click.RetentionExpires = time.Now().Add(24 * time.Hour)
	click.IdempotencyKey = click.RequestID
	if err := repository.RecordClick(ctx, destination, click); err != nil {
		return "", err
	}
	return destination.DestinationURL, nil
}
