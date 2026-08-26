package domain_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalogpostgres "rigmark/internal/adapters/postgres/catalog"
	recommendationpostgres "rigmark/internal/adapters/postgres/recommendation"
	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
	recommendationports "rigmark/internal/modules/recommendation/ports"
)

func TestSeedCatalogProducesRepeatableRecommendation(t *testing.T) {
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

	products, err := catalogpostgres.NewForVertical(pool, "saas").ListPublished(ctx, ports.ProductFilter{Limit: 100})
	if err != nil {
		t.Fatalf("load catalog candidates: %v", err)
	}
	policy, err := recommendationpostgres.NewForVertical(pool, "saas").ActivePolicy(ctx)
	if errors.Is(err, recommendationports.ErrNotFound) {
		t.Skip("the SaaS vertical is not seeded in this database")
	}
	if err != nil {
		t.Fatalf("load active policy: %v", err)
	}
	candidates := make([]recommendation.CandidateSnapshot, 0, len(products))
	for _, product := range products {
		candidate, candidateErr := policy.Candidate(product)
		if candidateErr == nil {
			candidates = append(candidates, candidate)
		}
	}
	engine, err := recommendation.NewDeterministicRecommendationEngine(policy.Config)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	input := policy.EnrichInput(recommendation.Input{
		Goal: planning.Goal("client_services"), Experience: planning.ExperienceBeginner,
		Budget: catalog.Money{AmountMinor: 12_000, Currency: "USD"},
		Priorities: []recommendation.Priority{
			recommendation.Priority("budget"), recommendation.Priority("integrations"),
		},
	})

	first, err := engine.Recommend(input, candidates)
	if err != nil {
		t.Fatalf("first recommendation: %v", err)
	}
	for left, right := 0, len(candidates)-1; left < right; left, right = left+1, right-1 {
		candidates[left], candidates[right] = candidates[right], candidates[left]
	}
	second, err := engine.Recommend(input, candidates)
	if err != nil {
		t.Fatalf("second recommendation: %v", err)
	}
	if first.Status != recommendation.ResultComplete || len(first.Selected) == 0 {
		t.Fatalf("seed recommendation is incomplete: %#v", first)
	}
	if first.TotalCost.AmountMinor > input.Budget.AmountMinor {
		t.Fatalf("total %d exceeds budget %d", first.TotalCost.AmountMinor, input.Budget.AmountMinor)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("database candidate order changed the recommendation result")
	}
}
