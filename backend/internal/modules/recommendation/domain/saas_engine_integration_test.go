package domain_test

import (
	"context"
	"errors"
	"os"
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

// TestSaaSCatalogProducesRecommendationWithoutSpatialInput exercises the
// non-physical vertical end to end. The engine must produce a complete setup
// without any room measurement, which is the behaviour a software catalog
// depends on and which the fitness path cannot cover.
func TestSaaSCatalogProducesRecommendationWithoutSpatialInput(t *testing.T) {
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

	policy, err := recommendationpostgres.NewForVertical(pool, "saas").ActivePolicy(ctx)
	if errors.Is(err, recommendationports.ErrNotFound) {
		t.Skip("the saas vertical is not seeded in this database")
	}
	if err != nil {
		t.Fatalf("load active saas policy: %v", err)
	}
	if policy.Config.SpatialConstraints {
		t.Fatal("the saas policy must not declare spatial constraints")
	}

	products, err := catalogpostgres.NewForVertical(pool, "saas").
		ListPublished(ctx, ports.ProductFilter{Limit: 100})
	if err != nil {
		t.Fatalf("load saas catalog: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("the saas catalog is empty")
	}
	candidates := make([]recommendation.CandidateSnapshot, 0, len(products))
	for _, product := range products {
		if product.IsPhysical {
			t.Fatalf("product %q in the saas catalog is marked physical", product.Slug)
		}
		candidate, candidateErr := policy.Candidate(product)
		if candidateErr == nil {
			candidates = append(candidates, candidate)
		}
	}

	engine, err := recommendation.NewDeterministicRecommendationEngine(policy.Config)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	// No AvailableSpace is supplied. A non-spatial vertical must not require
	// one, which is the whole point of the spatial_constraints flag.
	input := policy.EnrichInput(recommendation.Input{
		Goal:       planning.Goal("client_services"),
		Experience: planning.ExperienceBeginner,
		Budget:     catalog.Money{AmountMinor: 12_000, Currency: "USD"},
		Priorities: []recommendation.Priority{
			recommendation.Priority("budget"), recommendation.Priority("integrations"),
		},
		TrainingPreferences: []recommendation.TrainingPreference{
			recommendation.TrainingPreference("best_of_breed"),
		},
	})

	result, err := engine.Recommend(input, candidates)
	if err != nil {
		t.Fatalf("saas recommendation: %v", err)
	}
	if result.Status != recommendation.ResultComplete {
		t.Fatalf("result status = %q, want %q", result.Status, recommendation.ResultComplete)
	}
	if len(result.Selected) == 0 {
		t.Fatal("a complete result selected no products")
	}
	if result.TotalCost.AmountMinor > input.Budget.AmountMinor {
		t.Fatalf("total cost %d exceeds budget %d",
			result.TotalCost.AmountMinor, input.Budget.AmountMinor)
	}
	for _, item := range result.Selected {
		if item.Product.Breakdown.SpaceMatch != 100 {
			t.Fatalf("product %q scored space_match %d in a non-spatial vertical",
				item.Product.Candidate.Name, item.Product.Breakdown.SpaceMatch)
		}
	}
}

// TestSaaSPolicyRejectsUndeclaredPriority proves the vocabulary is genuinely
// scoped to the active policy rather than accepted by shape alone.
func TestSaaSPolicyRejectsUndeclaredPriority(t *testing.T) {
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

	policy, err := recommendationpostgres.NewForVertical(pool, "saas").ActivePolicy(ctx)
	if errors.Is(err, recommendationports.ErrNotFound) {
		t.Skip("the saas vertical is not seeded in this database")
	}
	if err != nil {
		t.Fatalf("load active saas policy: %v", err)
	}
	engine, err := recommendation.NewDeterministicRecommendationEngine(policy.Config)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	// "compact" belongs to the fitness vocabulary and is well-formed, so only
	// policy membership can reject it.
	input := policy.EnrichInput(recommendation.Input{
		Goal:       planning.Goal("client_services"),
		Experience: planning.ExperienceBeginner,
		Budget:     catalog.Money{AmountMinor: 12_000, Currency: "USD"},
		Priorities: []recommendation.Priority{recommendation.Priority("compact")},
	})
	if _, err := engine.Recommend(input, nil); !errors.Is(err, recommendation.ErrInvalidInput) {
		t.Fatalf("undeclared priority error = %v, want ErrInvalidInput", err)
	}
}

// TestEveryRequiredRoleHasAProvider guards the failure that made two of the
// five goals unanswerable for weeks without anything logging an error.
//
// A goal declares required setup roles; a role declares the capabilities that
// can fill it. If no published product provides any of them, the engine
// correctly returns no_suitable_products -- for that goal, forever, for every
// visitor. Nothing is broken, so nothing complains. The catalog is simply
// missing a category, and the only symptom is a wizard that says no.
func TestEveryRequiredRoleHasAProvider(t *testing.T) {
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

	policy, err := recommendationpostgres.NewForVertical(pool, "saas").ActivePolicy(ctx)
	if errors.Is(err, recommendationports.ErrNotFound) {
		t.Skip("the saas vertical is not seeded in this database")
	}
	if err != nil {
		t.Fatalf("load active saas policy: %v", err)
	}

	products, err := catalogpostgres.NewForVertical(pool, "saas").
		ListPublished(ctx, ports.ProductFilter{Limit: 200})
	if err != nil {
		t.Fatalf("load saas catalog: %v", err)
	}
	provided := make(map[recommendation.Capability]bool)
	for _, product := range products {
		candidate, candidateErr := policy.Candidate(product)
		if candidateErr != nil {
			continue
		}
		for _, capability := range candidate.Capabilities {
			provided[capability] = true
		}
	}

	for _, goal := range policy.Config.Goals {
		for _, role := range goal.Roles {
			if !role.Required {
				continue
			}
			filled := false
			for _, capability := range role.Capabilities {
				if provided[capability] {
					filled = true
					break
				}
			}
			if !filled {
				t.Errorf("goal %q requires role %q, and no published product provides any of %v",
					goal.Goal, role.Key, role.Capabilities)
			}
		}
	}
}
