package recommendationpostgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

func TestDraftAndCompletedSetupLifecycle(t *testing.T) {
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

	var userID identity.UserID
	email := fmt.Sprintf("recommendation-repository-%d@example.invalid", time.Now().UnixNano())
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email, status) VALUES ($1, 'active') RETURNING id`, email).Scan(&userID); err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id = $1`, userID)
	})

	var productID catalog.ProductID
	var price int64
	var currency string
	var factRevisionID, scoreRevisionID, productName, categorySlug string
	var dimensions catalog.Dimensions
	var scores catalog.Scores
	if err = pool.QueryRow(ctx, `SELECT products.id, products.price_minor, products.currency,
		products.published_fact_revision_id, products.published_score_revision_id,
		products.name, categories.slug, products.length_mm, products.width_mm, products.height_mm,
		products.quality_score, products.value_score, products.durability_score,
		products.beginner_score, products.advanced_score, products.apartment_score,
		products.noise_score, products.portability_score
		FROM catalog.products products JOIN catalog.categories categories ON categories.id=products.category_id
		WHERE products.status = 'published' AND products.published_fact_revision_id IS NOT NULL
		ORDER BY products.id LIMIT 1`).Scan(&productID, &price, &currency, &factRevisionID,
		&scoreRevisionID, &productName, &categorySlug, &dimensions.LengthMM, &dimensions.WidthMM,
		&dimensions.HeightMM, &scores.Quality, &scores.Value, &scores.Durability,
		&scores.Beginner, &scores.Advanced, &scores.Apartment, &scores.Noise, &scores.Portability); err != nil {
		t.Fatalf("load test product: %v", err)
	}

	repository := New(pool)
	policy, err := repository.ActivePolicy(ctx)
	if err != nil {
		t.Fatalf("ActivePolicy(): %v", err)
	}
	goal := planning.GoalBuildMuscle
	experience := planning.ExperienceBeginner
	budget := int64(70_000)
	draft, err := repository.SaveDraft(ctx, userID, ports.Draft{
		CurrentStep: 5, Goal: &goal, Experience: &experience,
		BudgetMinor: &budget, Currency: stringPointer("USD"),
		AvailableSpace:      &recommendation.AvailableSpace{LengthMM: 2400, WidthMM: 1800, HeightMM: 2400, ApartmentLiving: true},
		ExistingEquipment:   []recommendation.ExistingEquipment{{Name: "Pull-up bar", CategorySlug: "pull-up-bars"}},
		TrainingPreferences: []recommendation.TrainingPreference{recommendation.PreferenceBodyweight},
		Priorities:          []recommendation.Priority{recommendation.PriorityCompact},
	})
	if err != nil || draft.UpdatedAt.IsZero() {
		t.Fatalf("SaveDraft() = %#v, %v", draft, err)
	}
	loadedDraft, err := repository.GetDraft(ctx, userID)
	if err != nil || len(loadedDraft.ExistingEquipment) != 1 || loadedDraft.CurrentStep != 5 {
		t.Fatalf("GetDraft() = %#v, %v", loadedDraft, err)
	}

	input := recommendation.Input{
		Goal: goal, Experience: experience,
		Budget:         catalog.Money{AmountMinor: budget, Currency: "USD"},
		AvailableSpace: *draft.AvailableSpace,
		ExistingEquipment: []recommendation.ExistingEquipment{{
			Name: "Pull-up bar", CategorySlug: "pull-up-bars",
			Capabilities:     []recommendation.Capability{recommendation.CapabilityPullUp, recommendation.CapabilityAnchorPoint},
			RedundancyGroups: []string{"pull_up_station"},
		}},
		TrainingPreferences: loadedDraft.TrainingPreferences,
		Priorities:          loadedDraft.Priorities,
	}
	breakdown := recommendation.ScoreBreakdown{GoalMatch: 90, BudgetMatch: 95, SpaceMatch: 92,
		ExperienceMatch: 94, PreferenceMatch: 80, Quality: 85, Value: 88,
		Durability: 84, Compatibility: 90, Portability: 82, Noise: 91}
	candidate, err := policy.Candidate(catalog.Product{ID: productID, FactRevisionID: factRevisionID,
		ScoreRevisionID: scoreRevisionID, Name: productName, CategorySlug: categorySlug,
		Price: catalog.Money{AmountMinor: price, Currency: currency}, Dimensions: dimensions, Scores: scores})
	if err != nil {
		t.Fatalf("policy Candidate(): %v", err)
	}
	ranked := recommendation.RankedProduct{
		Candidate:      candidate,
		ObjectiveScore: 89, Breakdown: breakdown,
		Reasons: []recommendation.Reason{{Code: "space.fits", Message: "Fits your available space", Dimension: "space_match", Score: 92}},
	}
	result := recommendation.Result{
		Status: recommendation.ResultComplete, PolicyVersion: policy.Config.PolicyVersion, EngineVersion: "test-engine",
		InputFingerprint: fmt.Sprintf("integration-%d", time.Now().UnixNano()), ObjectiveScore: 89,
		Breakdown: breakdown, TotalCost: ranked.Candidate.Price,
		Selected: []recommendation.RecommendedItem{{Rank: 1, Product: ranked, Quantity: 1, UnitPriceMinor: price}},
	}
	saved, err := repository.SaveResult(ctx, userID, input, result, []recommendation.CandidateSnapshot{ranked.Candidate})
	if err != nil {
		t.Fatalf("SaveResult(): %v", err)
	}
	setups, err := repository.ListSetups(ctx, userID)
	if err != nil || len(setups) != 1 || setups[0].ID != saved.SetupID {
		t.Fatalf("ListSetups() = %#v, %v", setups, err)
	}
	loaded, err := repository.GetResultBySetupID(ctx, userID, saved.SetupID)
	if err != nil || loaded.Result.ObjectiveScore != 89 || len(loaded.Result.Selected) != 1 ||
		loaded.Result.Selected[0].Product.Candidate.ProductID != productID || loaded.SetupName != saved.SetupName {
		t.Fatalf("GetResultBySetupID() = %#v, %v", loaded, err)
	}
	if loaded.Result.PolicyVersion != policy.Config.PolicyVersion ||
		loaded.Result.Selected[0].Product.Candidate.PolicyVersion != policy.Config.PolicyVersion ||
		len(loaded.Input.ExistingEquipment) != 1 ||
		len(loaded.Input.ExistingEquipment[0].Capabilities) != 2 ||
		len(loaded.Input.ExistingEquipment[0].RedundancyGroups) != 1 {
		t.Fatalf("historical policy inputs were not preserved: %#v", loaded)
	}
	if err = repository.RenameSetup(ctx, userID, saved.SetupID, "Compact strength plan"); err != nil {
		t.Fatalf("RenameSetup(): %v", err)
	}
	renamed, err := repository.GetResultBySetupID(ctx, userID, saved.SetupID)
	if err != nil || renamed.SetupName != "Compact strength plan" {
		t.Fatalf("renamed setup = %#v, %v", renamed, err)
	}
	if _, err = repository.GetDraft(ctx, userID); err != ports.ErrNotFound {
		t.Fatalf("completed draft still exists: %v", err)
	}
	if err = repository.DeleteSetup(ctx, userID, saved.SetupID); err != nil {
		t.Fatalf("DeleteSetup(): %v", err)
	}
	if _, err = repository.GetResultBySetupID(ctx, userID, saved.SetupID); err != ports.ErrNotFound {
		t.Fatalf("deleted setup still exists: %v", err)
	}
}

func stringPointer(value string) *string { return &value }
