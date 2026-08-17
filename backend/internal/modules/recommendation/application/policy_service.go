package application

import (
	"context"
	"errors"
	"strings"

	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

var ErrInvalidPolicyTransition = errors.New("invalid recommendation policy transition")

type PolicyService struct{ repository ports.PolicyRepository }

func NewPolicyService(repository ports.PolicyRepository) *PolicyService {
	return &PolicyService{repository: repository}
}

func (service *PolicyService) List(ctx context.Context) ([]domain.PolicySummary, error) {
	return service.repository.ListPolicies(ctx)
}

func (service *PolicyService) Transition(ctx context.Context, actor identity.UserID, version string, target domain.PolicyWorkflowStatus, note string) error {
	version, note = strings.TrimSpace(version), strings.TrimSpace(note)
	if actor == "" || version == "" || len(note) > 2000 ||
		(target != domain.PolicyInReview && target != domain.PolicyApproved && target != domain.PolicyActive && target != domain.PolicyRejected && target != domain.PolicyRetired) ||
		(target == domain.PolicyRejected && note == "") {
		return ErrInvalidPolicyTransition
	}
	return service.repository.TransitionPolicy(ctx, actor, version, target, note)
}
