package recommendationpostgres

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalogpostgres "rigmark/internal/adapters/postgres/catalog"
	catalog "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
)

func TestCommercialDataCannotChangeRecommendationOutput(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var linkID string
	var originalPriority int16
	var originalCommissionType string
	var originalCommission *int
	var offerID, merchantID, offerProductID string
	var originalPrice, originalShipping int64
	var originalAvailability string
	var originalOfferActive bool
	var originalMerchantStatus string
	var originalTrust int16
	if err = pool.QueryRow(ctx, `SELECT links.id, links.priority, links.commission_type,
		links.commission_rate_bps, offers.id, offers.product_id, offers.price_minor, offers.shipping_minor,
		offers.availability, offers.is_active, merchants.id, merchants.status, merchants.trust_score
		FROM commerce.affiliate_links links
		JOIN commerce.merchant_offers offers ON offers.id=links.merchant_offer_id
		JOIN commerce.merchants merchants ON merchants.id=offers.merchant_id
		ORDER BY links.id LIMIT 1`).Scan(
		&linkID, &originalPriority, &originalCommissionType, &originalCommission,
		&offerID, &offerProductID, &originalPrice, &originalShipping, &originalAvailability, &originalOfferActive,
		&merchantID, &originalMerchantStatus, &originalTrust); err != nil {
		t.Fatalf("load commercial fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `UPDATE commerce.affiliate_links
			SET priority=$2, commission_type=$3, commission_rate_bps=$4 WHERE id=$1`,
			linkID, originalPriority, originalCommissionType, originalCommission)
		_, _ = pool.Exec(context.Background(), `UPDATE commerce.merchant_offers SET
			price_minor=$2, shipping_minor=$3, availability=$4, is_active=$5 WHERE id=$1`,
			offerID, originalPrice, originalShipping, originalAvailability, originalOfferActive)
		_, _ = pool.Exec(context.Background(), `UPDATE commerce.merchants SET status=$2,
			trust_score=$3 WHERE id=$1`, merchantID, originalMerchantStatus, originalTrust)
	})

	catalogRepository := catalogpostgres.New(pool)
	products, err := catalogRepository.ListPublished(ctx, catalogProductFilter())
	if err != nil || len(products) == 0 {
		t.Fatalf("load governed products: %d, %v", len(products), err)
	}
	policy, err := New(pool).ActivePolicy(ctx)
	if err != nil {
		t.Fatalf("load active policy: %v", err)
	}
	engine, err := recommendation.NewDeterministicRecommendationEngine(policy.Config)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	input := recommendation.Input{
		Goal: planning.GoalBuildMuscle, Experience: planning.ExperienceBeginner,
		Budget:         catalog.Money{AmountMinor: 250000, Currency: "USD"},
		AvailableSpace: recommendation.AvailableSpace{LengthMM: 3000, WidthMM: 3000, HeightMM: 2500},
		Priorities:     []recommendation.Priority{recommendation.PriorityBudget},
	}
	before, err := engine.Recommend(input, candidateSnapshots(t, policy, products))
	if err != nil {
		t.Fatalf("recommend before commercial change: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.affiliate_links
		SET priority=100, commission_type='percentage', commission_rate_bps=9999,
			updated_at=now() WHERE id=$1`, linkID); err != nil {
		t.Fatalf("change commercial data: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.merchant_offers SET price_minor=99999999,
		shipping_minor=999999, availability='out_of_stock', is_active=false, updated_at=now()
		WHERE id=$1`, offerID); err != nil {
		t.Fatalf("change offer data: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.merchants SET status='paused', trust_score=0,
		updated_at=now() WHERE id=$1`, merchantID); err != nil {
		t.Fatalf("change merchant data: %v", err)
	}
	nonce := time.Now().UnixNano()
	var configurationID, clickID, conversionID string
	if err = pool.QueryRow(ctx, `INSERT INTO commerce.provider_configurations
		(merchant_id,provider_key,adapter_key,external_merchant_id,credential_reference,
		 lifecycle_status,configuration_verified_at,conversion_ingestion_enabled,
		 conversion_configuration_verified_at)
		VALUES($1,$2,'test_fixture',$3,'test/fixture','active',now(),true,now()) RETURNING id`,
		merchantID, fmt.Sprintf("commercial-test-%d", nonce), fmt.Sprintf("merchant-%d", nonce)).Scan(&configurationID); err != nil {
		t.Fatalf("create commercial provider fixture: %v", err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO commerce.affiliate_clicks
		(affiliate_link_id,merchant_offer_id,product_id,anonymous_id,session_id,source,campaign,
		 classification,is_countable,retention_expires_at)
		VALUES($1,$2,$3,$4,$4,'recommendation','initial','human',true,now()+interval '1 day') RETURNING id`,
		linkID, offerID, offerProductID, fmt.Sprintf("commercial-test-%d", nonce)).Scan(&clickID); err != nil {
		t.Fatalf("create commercial click fixture: %v", err)
	}
	commission := int64(4321)
	if err = pool.QueryRow(ctx, `INSERT INTO commerce.affiliate_conversions
		(affiliate_click_id,provider,external_conversion_id,order_status,converted_at,
		 provider_configuration_id,merchant_id,event_type,event_received_at,commission_amount_minor,
		 commission_currency,commission_status,attribution_status,source,campaign,verification_state,
		 latest_event_at)
		VALUES($1,$2,$3,'confirmed',now(),$4,$5,'conversion_created',now(),$6,'USD','approved',
		'attributed','recommendation','initial','verified',now()) RETURNING id`, clickID,
		fmt.Sprintf("commercial-test-%d", nonce), fmt.Sprintf("order-%d", nonce), configurationID,
		merchantID, commission).Scan(&conversionID); err != nil {
		t.Fatalf("create verified conversion fixture: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.affiliate_conversions SET
		commission_amount_minor=999999,order_status='reversed',commission_status='reversed',
		attribution_status='attributed',source='comparison',campaign='mutated' WHERE id=$1`, conversionID); err != nil {
		t.Fatalf("mutate conversion fields: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE commerce.affiliate_clicks SET source='comparison',
		campaign='mutated',traffic_source='test',traffic_medium='fixture' WHERE id=$1`, clickID); err != nil {
		t.Fatalf("mutate click metadata: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_conversions WHERE id=$1`, conversionID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_clicks WHERE id=$1`, clickID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.provider_configurations WHERE id=$1`, configurationID)
	})
	productIDs := make([]catalog.ProductID, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	productsAfter, err := catalogRepository.ListByIDs(ctx, productIDs)
	if err != nil {
		t.Fatalf("reload governed products: %v", err)
	}
	after, err := engine.Recommend(input, candidateSnapshots(t, policy, productsAfter))
	if err != nil {
		t.Fatalf("recommend after commercial change: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("commercial-only mutation changed deterministic output\nbefore=%#v\nafter=%#v", before, after)
	}
}

func catalogProductFilter() catalogports.ProductFilter {
	return catalogports.ProductFilter{Limit: 100}
}

func candidateSnapshots(t *testing.T, policy recommendation.Policy, products []catalog.Product) []recommendation.CandidateSnapshot {
	t.Helper()
	result := make([]recommendation.CandidateSnapshot, 0, len(products))
	for _, product := range products {
		candidate, err := policy.Candidate(product)
		if err != nil {
			t.Fatalf("map policy candidate %s: %v", product.ID, err)
		}
		result = append(result, candidate)
	}
	return result
}
