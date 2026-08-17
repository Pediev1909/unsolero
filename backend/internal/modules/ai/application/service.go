package application

import (
	"context"
	"errors"

	aidomain "rigmark/internal/modules/ai/domain"
	"rigmark/internal/modules/ai/ports"
)

var ErrProviderRequired = errors.New("AI provider is required")

type Service struct {
	provider ports.AIProvider
}

func NewService(provider ports.AIProvider) (*Service, error) {
	if provider == nil {
		return nil, ErrProviderRequired
	}
	return &Service{provider: provider}, nil
}

func (service *Service) UnderstandUserInput(ctx context.Context, request aidomain.UnderstandUserInputRequest) (aidomain.UserInputUnderstanding, error) {
	if err := aidomain.ValidateUnderstandRequest(request); err != nil {
		return aidomain.UserInputUnderstanding{}, err
	}
	result, err := service.provider.UnderstandUserInput(ctx, request)
	if err != nil {
		return aidomain.UserInputUnderstanding{}, err
	}
	if err := aidomain.ValidateUnderstanding(result); err != nil {
		return aidomain.UserInputUnderstanding{}, err
	}
	return result, nil
}

func (service *Service) ExtractRequirements(ctx context.Context, request aidomain.ExtractRequirementsRequest) (aidomain.RequirementsDraft, error) {
	if err := aidomain.ValidateExtractRequest(request); err != nil {
		return aidomain.RequirementsDraft{}, err
	}
	result, err := service.provider.ExtractRequirements(ctx, request)
	if err != nil {
		return aidomain.RequirementsDraft{}, err
	}
	if err := aidomain.ValidateRequirementsOutput(result); err != nil {
		return aidomain.RequirementsDraft{}, err
	}
	return result, nil
}

func (service *Service) AskClarifyingQuestion(ctx context.Context, request aidomain.ClarifyingQuestionRequest) (aidomain.ClarifyingQuestion, error) {
	if err := aidomain.ValidateClarifyingRequest(request); err != nil {
		return aidomain.ClarifyingQuestion{}, err
	}
	result, err := service.provider.AskClarifyingQuestion(ctx, request)
	if err != nil {
		return aidomain.ClarifyingQuestion{}, err
	}
	if err := aidomain.ValidateClarifyingQuestion(result, request); err != nil {
		return aidomain.ClarifyingQuestion{}, err
	}
	return result, nil
}

func (service *Service) ExplainRecommendation(ctx context.Context, request aidomain.ExplainRecommendationRequest) (aidomain.ExplanationPlan, error) {
	if err := aidomain.ValidateExplainRequest(request); err != nil {
		return aidomain.ExplanationPlan{}, err
	}
	result, err := service.provider.ExplainRecommendation(ctx, request)
	if err != nil {
		return aidomain.ExplanationPlan{}, err
	}
	if err := aidomain.ValidateExplanationPlan(result, request); err != nil {
		return aidomain.ExplanationPlan{}, err
	}
	return result, nil
}

func (service *Service) RefineRecommendation(ctx context.Context, request aidomain.RefineRecommendationRequest) (aidomain.Refinement, error) {
	if err := aidomain.ValidateRefineRequest(request); err != nil {
		return aidomain.Refinement{}, err
	}
	result, err := service.provider.RefineRecommendation(ctx, request)
	if err != nil {
		return aidomain.Refinement{}, err
	}
	if err := aidomain.ValidateRefinement(result); err != nil {
		return aidomain.Refinement{}, err
	}
	return result, nil
}

func (service *Service) CompareProducts(ctx context.Context, request aidomain.CompareProductsRequest) (aidomain.ComparisonPlan, error) {
	if err := aidomain.ValidateCompareRequest(request); err != nil {
		return aidomain.ComparisonPlan{}, err
	}
	result, err := service.provider.CompareProducts(ctx, request)
	if err != nil {
		return aidomain.ComparisonPlan{}, err
	}
	if err := aidomain.ValidateComparisonPlan(result, request); err != nil {
		return aidomain.ComparisonPlan{}, err
	}
	return result, nil
}
