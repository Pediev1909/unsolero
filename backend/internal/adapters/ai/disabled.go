package ai

import (
	"context"

	aidomain "rigmark/internal/modules/ai/domain"
	"rigmark/internal/modules/ai/ports"
)

// DisabledProvider preserves deterministic fallbacks when AI is not configured.
type DisabledProvider struct{}

func (DisabledProvider) UnderstandUserInput(context.Context, aidomain.UnderstandUserInputRequest) (aidomain.UserInputUnderstanding, error) {
	return aidomain.UserInputUnderstanding{}, ports.ErrProviderUnavailable
}

func (DisabledProvider) ExtractRequirements(context.Context, aidomain.ExtractRequirementsRequest) (aidomain.RequirementsDraft, error) {
	return aidomain.RequirementsDraft{}, ports.ErrProviderUnavailable
}

func (DisabledProvider) AskClarifyingQuestion(context.Context, aidomain.ClarifyingQuestionRequest) (aidomain.ClarifyingQuestion, error) {
	return aidomain.ClarifyingQuestion{}, ports.ErrProviderUnavailable
}

func (DisabledProvider) ExplainRecommendation(context.Context, aidomain.ExplainRecommendationRequest) (aidomain.ExplanationPlan, error) {
	return aidomain.ExplanationPlan{}, ports.ErrProviderUnavailable
}

func (DisabledProvider) RefineRecommendation(context.Context, aidomain.RefineRecommendationRequest) (aidomain.Refinement, error) {
	return aidomain.Refinement{}, ports.ErrProviderUnavailable
}

func (DisabledProvider) CompareProducts(context.Context, aidomain.CompareProductsRequest) (aidomain.ComparisonPlan, error) {
	return aidomain.ComparisonPlan{}, ports.ErrProviderUnavailable
}
