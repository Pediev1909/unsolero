package commercepostgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

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
	referrer := "https://rigmark.example/products/demo"
	trafficSource, trafficMedium, referrerHost := "newsletter", "email", "rigmark.example"
	click := domain.AffiliateClick{OfferID: offerID, UserID: &userID, SessionID: &sessionID,
		Source: "product_detail", Campaign: &campaign, Referrer: &referrer, RequestID: &requestID,
		TrafficSource: &trafficSource, TrafficMedium: &trafficMedium, ReferrerHost: &referrerHost}
	destination, err := New(pool).TrackOfferClick(ctx, click)
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
	var eventName string
	if err = pool.QueryRow(ctx, `SELECT event_name FROM analytics.events
		WHERE user_id = $1 AND event_name = 'affiliate_clicked' ORDER BY received_at DESC LIMIT 1`, userID).Scan(&eventName); err != nil {
		t.Fatalf("load affiliate analytics event: %v", err)
	}

	missing := click
	missing.OfferID = "00000000-0000-4000-8000-000000000000"
	if _, err = New(pool).TrackOfferClick(ctx, missing); err != ports.ErrAffiliateDestinationNotFound {
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
	if destination, trackErr := New(pool).TrackOfferClick(ctx, owned); trackErr != nil || destination == "" {
		t.Fatalf("owned recommendation click = %q, %v", destination, trackErr)
	}

	unowned := owned
	unowned.UserID = &otherUserID
	unowned.RequestID = &deniedRequestID
	if _, trackErr := New(pool).TrackOfferClick(ctx, unowned); !errors.Is(trackErr, ports.ErrAffiliateDestinationNotFound) {
		t.Fatalf("unowned recommendation click error = %v, want ErrAffiliateDestinationNotFound", trackErr)
	}

	stale := domain.AffiliateClick{OfferID: offerID, UserID: &ownerID, Source: "product_detail",
		RequestID: &staleRequestID}
	if _, trackErr := New(pool, time.Nanosecond).TrackOfferClick(ctx, stale); !errors.Is(trackErr, ports.ErrAffiliateDestinationNotFound) {
		t.Fatalf("stale offer click error = %v, want ErrAffiliateDestinationNotFound", trackErr)
	}
}
