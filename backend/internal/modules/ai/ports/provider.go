package ports

import (
	"context"
	"errors"

	aidomain "rigmark/internal/modules/ai/domain"
)

var ErrProviderUnavailable = errors.New("AI provider unavailable")

// AIProvider is the only model-facing port. It receives immutable structured
// snapshots and has no repository, transaction, database, or ranking methods.
type AIProvider interface {
	UnderstandUserInput(context.Context, aidomain.UnderstandUserInputRequest) (aidomain.UserInputUnderstanding, error)
	ExtractRequirements(context.Context, aidomain.ExtractRequirementsRequest) (aidomain.RequirementsDraft, error)
	AskClarifyingQuestion(context.Context, aidomain.ClarifyingQuestionRequest) (aidomain.ClarifyingQuestion, error)
	ExplainRecommendation(context.Context, aidomain.ExplainRecommendationRequest) (aidomain.ExplanationPlan, error)
	RefineRecommendation(context.Context, aidomain.RefineRecommendationRequest) (aidomain.Refinement, error)
	CompareProducts(context.Context, aidomain.CompareProductsRequest) (aidomain.ComparisonPlan, error)
}
