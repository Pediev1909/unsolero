package recommendationpostgres

import (
	"context"
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
	if err = pool.QueryRow(ctx, `SELECT id, priority, commission_type, commission_rate_bps
		FROM commerce.affiliate_links ORDER BY id LIMIT 1`).Scan(
		&linkID, &originalPriority, &originalCommissionType, &originalCommission); err != nil {
		t.Fatalf("load commercial fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `UPDATE commerce.affiliate_links
			SET priority=$2, commission_type=$3, commission_rate_bps=$4 WHERE id=$1`,
			linkID, originalPriority, originalCommissionType, originalCommission)
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
