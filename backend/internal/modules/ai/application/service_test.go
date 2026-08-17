package application

import (
	"context"
	"errors"
	"testing"

	aidomain "rigmark/internal/modules/ai/domain"
	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
)

type mockProvider struct {
	understanding aidomain.UserInputUnderstanding
	requirements  aidomain.RequirementsDraft
	question      aidomain.ClarifyingQuestion
	explanation   aidomain.ExplanationPlan
	refinement    aidomain.Refinement
	comparison    aidomain.ComparisonPlan
	err           error
}

func (provider *mockProvider) UnderstandUserInput(context.Context, aidomain.UnderstandUserInputRequest) (aidomain.UserInputUnderstanding, error) {
	return provider.understanding, provider.err
}

func (provider *mockProvider) ExtractRequirements(context.Context, aidomain.ExtractRequirementsRequest) (aidomain.RequirementsDraft, error) {
	return provider.requirements, provider.err
}

func (provider *mockProvider) AskClarifyingQuestion(context.Context, aidomain.ClarifyingQuestionRequest) (aidomain.ClarifyingQuestion, error) {
	return provider.question, provider.err
}

func (provider *mockProvider) ExplainRecommendation(context.Context, aidomain.ExplainRecommendationRequest) (aidomain.ExplanationPlan, error) {
	return provider.explanation, provider.err
}

func (provider *mockProvider) RefineRecommendation(context.Context, aidomain.RefineRecommendationRequest) (aidomain.Refinement, error) {
	return provider.refinement, provider.err
}

func (provider *mockProvider) CompareProducts(context.Context, aidomain.CompareProductsRequest) (aidomain.ComparisonPlan, error) {
	return provider.comparison, provider.err
}

func TestServiceAcceptsValidatedMockProviderOutputs(t *testing.T) {
	fixture := newFixture()
	service, err := NewService(fixture.provider)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	ctx := context.Background()

	if _, err = service.UnderstandUserInput(ctx, fixture.understandRequest); err != nil {
		t.Errorf("UnderstandUserInput() error = %v", err)
	}
	if _, err = service.ExtractRequirements(ctx, fixture.extractRequest); err != nil {
		t.Errorf("ExtractRequirements() error = %v", err)
	}
	if _, err = service.AskClarifyingQuestion(ctx, fixture.questionRequest); err != nil {
		t.Errorf("AskClarifyingQuestion() error = %v", err)
	}
	if _, err = service.ExplainRecommendation(ctx, fixture.explainRequest); err != nil {
		t.Errorf("ExplainRecommendation() error = %v", err)
	}
	if _, err = service.RefineRecommendation(ctx, fixture.refineRequest); err != nil {
		t.Errorf("RefineRecommendation() error = %v", err)
	}
	if _, err = service.CompareProducts(ctx, fixture.compareRequest); err != nil {
		t.Errorf("CompareProducts() error = %v", err)
	}
}

func TestServiceRejectsInventedProductReference(t *testing.T) {
	fixture := newFixture()
	fixture.provider.explanation.Points[0].ProductID = "invented-product"
	service, _ := NewService(fixture.provider)

	_, err := service.ExplainRecommendation(context.Background(), fixture.explainRequest)
	if !errors.Is(err, aidomain.ErrInvalidProviderOutput) {
		t.Fatalf("ExplainRecommendation() error = %v, want ErrInvalidProviderOutput", err)
	}
}

func TestServiceRejectsInventedFactReference(t *testing.T) {
	fixture := newFixture()
	fixture.provider.comparison.FactKeys = []aidomain.FactKey{"review_rating"}
	service, _ := NewService(fixture.provider)

	_, err := service.CompareProducts(context.Background(), fixture.compareRequest)
	if !errors.Is(err, aidomain.ErrInvalidProviderOutput) {
		t.Fatalf("CompareProducts() error = %v, want ErrInvalidProviderOutput", err)
	}
}

func TestServiceRejectsComparisonReordering(t *testing.T) {
	fixture := newFixture()
	fixture.provider.comparison.OrderedProductIDs[0], fixture.provider.comparison.OrderedProductIDs[1] =
		fixture.provider.comparison.OrderedProductIDs[1], fixture.provider.comparison.OrderedProductIDs[0]
	service, _ := NewService(fixture.provider)

	_, err := service.CompareProducts(context.Background(), fixture.compareRequest)
	if !errors.Is(err, aidomain.ErrInvalidProviderOutput) {
		t.Fatalf("CompareProducts() error = %v, want ErrInvalidProviderOutput", err)
	}
}

func TestServicePassesProviderFailuresWithoutFabricatingFallback(t *testing.T) {
	fixture := newFixture()
	providerFailure := errors.New("provider timeout")
	fixture.provider.err = providerFailure
	service, _ := NewService(fixture.provider)

	_, err := service.ExtractRequirements(context.Background(), fixture.extractRequest)
	if !errors.Is(err, providerFailure) {
		t.Fatalf("ExtractRequirements() error = %v, want provider failure", err)
	}
}

type fixture struct {
	provider          *mockProvider
	understandRequest aidomain.UnderstandUserInputRequest
	extractRequest    aidomain.ExtractRequirementsRequest
	questionRequest   aidomain.ClarifyingQuestionRequest
	explainRequest    aidomain.ExplainRecommendationRequest
	refineRequest     aidomain.RefineRecommendationRequest
	compareRequest    aidomain.CompareProductsRequest
}

func newFixture() fixture {
	goal := planning.GoalBuildMuscle
	experience := planning.ExperienceBeginner
	requirements := aidomain.RequirementsDraft{
		Goal: &goal, Experience: &experience,
		Budget:              &catalog.Money{AmountMinor: 70_000, Currency: "USD"},
		AvailableSpace:      &recommendation.AvailableSpace{LengthMM: 2400, WidthMM: 1800, HeightMM: 2400, ApartmentLiving: true},
		ExistingEquipment:   []recommendation.ExistingEquipment{{Name: "pull-up bar", CategorySlug: "pull-up-bars", Capabilities: []recommendation.Capability{recommendation.CapabilityPullUp}}},
		TrainingPreferences: []recommendation.TrainingPreference{recommendation.PreferenceDumbbells},
		Priorities:          []recommendation.Priority{recommendation.PriorityCompact},
	}
	first := product("product-1", "Demo Adjustable Dumbbells", 29_900)
	second := product("product-2", "Demo Folding Bench", 19_900)
	recommendationSnapshot := aidomain.RecommendationSnapshot{
		Requirements: requirements,
		Selected: []aidomain.DeterministicProduct{
			{Product: first, Rank: 1, ObjectiveScore: 92, ReasonCodes: []string{"space.fits", "budget.within"}},
		},
		Alternatives: []aidomain.DeterministicProduct{
			{Product: second, Rank: 2, ObjectiveScore: 84, ReasonCodes: []string{"space.fits", "budget.within"}},
		},
		TotalCost: catalog.Money{AmountMinor: 29_900, Currency: "USD"}, ObjectiveScore: 92,
		PolicyVersion: "home-gym-v1", EngineVersion: "deterministic-v1",
	}
	provider := &mockProvider{
		understanding: aidomain.UserInputUnderstanding{Language: "en-US", Confidence: 96,
			RecognizedFields: []aidomain.RequirementField{aidomain.FieldGoal, aidomain.FieldBudget},
			Ambiguities:      []aidomain.Ambiguity{{Field: aidomain.FieldAvailableSpace, Kind: aidomain.AmbiguityMissing}}, NeedsClarification: true},
		requirements: requirements,
		question:     aidomain.ClarifyingQuestion{Field: aidomain.FieldAvailableSpace, Prompt: aidomain.PromptAvailableSpace},
		explanation: aidomain.ExplanationPlan{Style: aidomain.StyleSpace, Points: []aidomain.ProductExplanationPlan{{
			ProductID: first.ID, ReasonCodes: []string{"space.fits", "budget.within"}, FactKeys: []aidomain.FactKey{aidomain.FactPrice, aidomain.FactDimensions},
		}}},
		refinement: aidomain.Refinement{Requirements: requirements, ChangedFields: []aidomain.RequirementField{aidomain.FieldPriorities}},
		comparison: aidomain.ComparisonPlan{OrderedProductIDs: []catalog.ProductID{first.ID, second.ID},
			FactKeys:   []aidomain.FactKey{aidomain.FactPrice, aidomain.FactDimensions, aidomain.FactQuality},
			Highlights: []aidomain.ComparisonHighlight{{ProductID: first.ID, FactKey: aidomain.FactQuality, Kind: aidomain.HighlightStrength}}},
	}
	return fixture{
		provider:          provider,
		understandRequest: aidomain.UnderstandUserInputRequest{Text: "I have $700 and a small apartment.", Locale: "en-US"},
		extractRequest:    aidomain.ExtractRequirementsRequest{Text: "I have $700 and a small apartment.", Locale: "en-US"},
		questionRequest: aidomain.ClarifyingQuestionRequest{Requirements: requirements,
			Ambiguities: []aidomain.Ambiguity{{Field: aidomain.FieldAvailableSpace, Kind: aidomain.AmbiguityMissing}}, Locale: "en-US"},
		explainRequest: aidomain.ExplainRecommendationRequest{Recommendation: recommendationSnapshot, Locale: "en-US"},
		refineRequest: aidomain.RefineRecommendationRequest{Instruction: "Make it easier to store.", Current: requirements,
			Recommendation: recommendationSnapshot, Locale: "en-US"},
		compareRequest: aidomain.CompareProductsRequest{Products: []aidomain.ProductSnapshot{first, second},
			Focus: []aidomain.FactKey{aidomain.FactPrice}, Locale: "en-US"},
	}
}

func product(id catalog.ProductID, name string, price int64) aidomain.ProductSnapshot {
	capacity := int64(50_000)
	return aidomain.ProductSnapshot{
		ID: id, Name: name, Category: "Strength", Brand: "Demo Forge", Price: catalog.Money{AmountMinor: price, Currency: "USD"},
		Dimensions: catalog.Dimensions{LengthMM: 400, WidthMM: 200, HeightMM: 200}, WeightGrams: 20_000,
		MaxCapacityGrams: &capacity, Material: "steel", WarrantyMonths: 24,
		Scores: catalog.Scores{Quality: 90, Value: 85, Durability: 90, Beginner: 90, Advanced: 80, Apartment: 88, Noise: 90, Portability: 75},
	}
}
