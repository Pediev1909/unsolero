package domain

import (
	"errors"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

func TestPolicyExplicitlyExcludesUnknownAndUnsupportedCategories(t *testing.T) {
	policy := testPolicy()
	product := policyProduct("unknown-category")
	if _, err := policy.Candidate(product); !errors.Is(err, ErrUnsupportedCategory) {
		t.Fatalf("unknown category error = %v, want ErrUnsupportedCategory", err)
	}
	policy.Categories["unknown-category"] = CategoryPolicy{CategorySlug: "unknown-category", Supported: false}
	if _, err := policy.Candidate(product); !errors.Is(err, ErrUnsupportedCategory) {
		t.Fatalf("unsupported category error = %v, want ErrUnsupportedCategory", err)
	}
}

func TestCatalogProductIsNotEligibleWithoutVersionedProductPolicy(t *testing.T) {
	policy := testPolicy()
	product := policyProduct("configured-category")
	if _, err := policy.Candidate(product); !errors.Is(err, ErrProductPolicyMissing) {
		t.Fatalf("unconfigured product error = %v, want ErrProductPolicyMissing", err)
	}
	policy.Products[product.ID] = ProductPolicy{
		ProductID: product.ID, FactRevisionID: "stale-facts", ScoreRevisionID: product.ScoreRevisionID,
	}
	if _, err := policy.Candidate(product); !errors.Is(err, ErrProductPolicyMissing) {
		t.Fatalf("stale revision error = %v, want ErrProductPolicyMissing", err)
	}
}

func TestPolicyBuildsCandidateOnlyFromExplicitPolicyData(t *testing.T) {
	policy := testPolicy()
	product := policyProduct("configured-category")
	policy.Products[product.ID] = ProductPolicy{
		ProductID: product.ID, FactRevisionID: product.FactRevisionID, ScoreRevisionID: product.ScoreRevisionID,
		Capabilities: []Capability{"future_vertical_capability"}, Requires: []Capability{"future_requirement"},
		IncompatibleWith: []Capability{"future_conflict"},
		GoalSupport:      []GoalSupport{{Goal: planning.GoalGeneralFitness, Score: 91}},
		RedundancyGroups: []string{"future_group"},
		Space:            SpaceProfile{Footprint: SpatialEnvelope{LengthMM: 500, WidthMM: 400, HeightMM: 300}},
	}
	candidate, err := policy.Candidate(product)
	if err != nil {
		t.Fatalf("Candidate(): %v", err)
	}
	if candidate.PolicyVersion != "policy-test-v1" ||
		!containsCapability(candidate.Capabilities, "category_capability") ||
		!containsCapability(candidate.Capabilities, "future_vertical_capability") ||
		!containsCapability(candidate.Requires, "future_requirement") ||
		!containsCapability(candidate.IncompatibleWith, "future_conflict") ||
		!intersectsStrings(candidate.RedundancyGroups, []string{"category_group", "future_group"}) {
		t.Fatalf("candidate did not retain explicit policy data: %#v", candidate)
	}
}

func TestEngineRejectsCandidateFromDifferentPolicyVersion(t *testing.T) {
	candidate := testCandidate("candidate", "adjustable-dumbbells", 10000, balancedScores())
	candidate.PolicyVersion = "another-policy"
	_, err := testEngine(t).Recommend(
		testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000),
		[]CandidateSnapshot{candidate},
	)
	if !errors.Is(err, ErrInvalidCandidate) {
		t.Fatalf("error = %v, want ErrInvalidCandidate", err)
	}
}

func TestUnsupportedGoalIsRejectedBeforeScoring(t *testing.T) {
	candidate := testCandidate("candidate", "adjustable-dumbbells", 10000, balancedScores())
	candidate.GoalSupport = []GoalSupport{{Goal: planning.GoalStrength, Score: 90}}
	result, err := testEngine(t).Recommend(
		testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000),
		[]CandidateSnapshot{candidate},
	)
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	rejection, found := rejectionFor(result, candidate.ProductID)
	if !found || rejection.Code != "goal.unsupported" || len(result.Ranked) != 0 {
		t.Fatalf("result = %#v", result)
	}
}

func TestRequiredUnknownSpaceMeasurementsFailClosed(t *testing.T) {
	for name, mutate := range map[string]func(*CandidateSnapshot){
		"storage":   func(candidate *CandidateSnapshot) { candidate.Space.RequiresStorageFootprint = true },
		"operating": func(candidate *CandidateSnapshot) { candidate.Space.RequiresOperatingClearance = true },
		"safety":    func(candidate *CandidateSnapshot) { candidate.Space.RequiresSafetyClearance = true },
		"access":    func(candidate *CandidateSnapshot) { candidate.Space.RequiresAccessWidth = true },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := testCandidate("candidate", "adjustable-dumbbells", 10000, balancedScores())
			mutate(&candidate)
			result, err := testEngine(t).Recommend(
				testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000),
				[]CandidateSnapshot{candidate},
			)
			if err != nil {
				t.Fatalf("Recommend(): %v", err)
			}
			rejection, found := rejectionFor(result, candidate.ProductID)
			if !found || rejection.Code != "space.measurement_unknown" {
				t.Fatalf("rejection = %#v", rejection)
			}
		})
	}
}

func TestRoomHeightAndAccessConstraintsUseKnownMeasurements(t *testing.T) {
	minimumHeight, minimumAccess := int64(2500), int64(900)
	candidate := testCandidate("candidate", "adjustable-dumbbells", 10000, balancedScores())
	candidate.Space.MinimumRoomHeightMM = &minimumHeight
	candidate.Space.MinimumAccessWidthMM = &minimumAccess
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000)
	access := int64(800)
	input.AvailableSpace.AccessWidthMM = &access

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{candidate})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if rejection, found := rejectionFor(result, candidate.ProductID); !found || rejection.Code != "space.access_blocked" {
		t.Fatalf("access rejection = %#v", rejection)
	}

	access = 1000
	input.AvailableSpace.HeightMM = 2400
	result, err = testEngine(t).Recommend(input, []CandidateSnapshot{candidate})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if rejection, found := rejectionFor(result, candidate.ProductID); !found || rejection.Code != "space.height_exceeded" {
		t.Fatalf("height rejection = %#v", rejection)
	}
}

func TestOperatingAndSafetyClearanceAffectRequiredEnvelope(t *testing.T) {
	candidate := testCandidate("candidate", "adjustable-dumbbells", 10000, balancedScores())
	candidate.Space.OperatingClearance = &Clearance{FrontMM: 400, BackMM: 400}
	candidate.Space.SafetyClearance = &Clearance{LeftMM: 400, RightMM: 400}
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000)
	input.AvailableSpace.LengthMM, input.AvailableSpace.WidthMM = 1000, 1000
	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{candidate})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if rejection, found := rejectionFor(result, candidate.ProductID); !found || rejection.Code != "space.does_not_fit" {
		t.Fatalf("clearance rejection = %#v", rejection)
	}
}

func TestOverlapOnlyAppliesWhenExplicitlyModeled(t *testing.T) {
	first := RankedProduct{Candidate: testCandidate("first", "adjustable-dumbbells", 10000, balancedScores())}
	second := RankedProduct{Candidate: testCandidate("second", "cardio-machines", 10000, balancedScores())}
	first.Candidate.Space.Footprint = SpatialEnvelope{LengthMM: 800, WidthMM: 800, HeightMM: 500}
	second.Candidate.Space.Footprint = SpatialEnvelope{LengthMM: 800, WidthMM: 800, HeightMM: 500}
	space := AvailableSpace{LengthMM: 1000, WidthMM: 1000, HeightMM: 2400}
	if fitsWithinTotalFloorArea([]RankedProduct{first, second}, space) {
		t.Fatal("unmodeled overlap reduced combined floor use")
	}
	first.Candidate.Space.OverlapGroup = "shared_zone"
	second.Candidate.Space.OverlapGroup = "shared_zone"
	if !fitsWithinTotalFloorArea([]RankedProduct{first, second}, space) {
		t.Fatal("explicit shared zone was not honored")
	}
}

func testPolicy() Policy {
	return Policy{
		Config: testPolicyConfig("policy-test-v1"),
		Categories: map[string]CategoryPolicy{
			"configured-category": {CategorySlug: "configured-category", Supported: true,
				Capabilities: []Capability{"category_capability"}, RedundancyGroups: []string{"category_group"}},
		},
		Products: make(map[catalog.ProductID]ProductPolicy),
	}
}

func policyProduct(category string) catalog.Product {
	return catalog.Product{
		ID: "policy-product", FactRevisionID: "facts-v1", ScoreRevisionID: "scores-v1",
		Name: "Policy product", CategorySlug: category,
		Price:      catalog.Money{AmountMinor: 10000, Currency: "USD"},
		IsPhysical: true,
		Dimensions: catalog.Dimensions{LengthMM: 500, WidthMM: 400, HeightMM: 300},
		Scores:     balancedScores(),
	}
}
