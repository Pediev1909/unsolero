package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	recommendation "rigmark/internal/modules/recommendation/application"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

type policyTransitionRequest struct {
	Note string `json:"note"`
}

// The domain summary carries no JSON tags, so serialising it directly emitted
// PascalCase while every other admin response is snake_case. Nothing consumed
// this endpoint, so nothing caught it.
type adminPolicyResponse struct {
	Version       string     `json:"version"`
	VerticalKey   string     `json:"vertical_key"`
	Status        string     `json:"status"`
	CategoryCount int        `json:"category_count"`
	ProductCount  int        `json:"product_count"`
	CreatedAt     time.Time  `json:"created_at"`
	ActivatedAt   *time.Time `json:"activated_at"`
	ReviewNote    string     `json:"review_note"`
}

func presentPolicy(summary domain.PolicySummary) adminPolicyResponse {
	return adminPolicyResponse{
		Version:       summary.Version,
		VerticalKey:   summary.VerticalKey,
		Status:        string(summary.Status),
		CategoryCount: summary.CategoryCount,
		ProductCount:  summary.ProductCount,
		CreatedAt:     summary.CreatedAt,
		ActivatedAt:   summary.ActivatedAt,
		ReviewNote:    summary.ReviewNote,
	}
}

func (h *Handler) adminListRecommendationPolicies(response http.ResponseWriter, request *http.Request) {
	items, err := h.recommendationPolicy.List(request.Context())
	if err != nil {
		writeAPIError(response, http.StatusInternalServerError, "policy_unavailable", "Recommendation policies are unavailable.", nil, h.logger)
		return
	}
	policies := make([]adminPolicyResponse, 0, len(items))
	for _, item := range items {
		policies = append(policies, presentPolicy(item))
	}
	writeJSON(response, http.StatusOK, map[string]any{"policies": policies}, h.logger)
}

func (h *Handler) adminSubmitRecommendationPolicy(w http.ResponseWriter, r *http.Request) {
	h.transitionRecommendationPolicy(w, r, domain.PolicyInReview)
}
func (h *Handler) adminApproveRecommendationPolicy(w http.ResponseWriter, r *http.Request) {
	h.transitionRecommendationPolicy(w, r, domain.PolicyApproved)
}
func (h *Handler) adminRejectRecommendationPolicy(w http.ResponseWriter, r *http.Request) {
	h.transitionRecommendationPolicy(w, r, domain.PolicyRejected)
}
func (h *Handler) adminActivateRecommendationPolicy(w http.ResponseWriter, r *http.Request) {
	h.transitionRecommendationPolicy(w, r, domain.PolicyActive)
}
func (h *Handler) adminDeactivateRecommendationPolicy(w http.ResponseWriter, r *http.Request) {
	h.transitionRecommendationPolicy(w, r, domain.PolicyRetired)
}

func (h *Handler) transitionRecommendationPolicy(response http.ResponseWriter, request *http.Request, target domain.PolicyWorkflowStatus) {
	var body policyTransitionRequest
	if request.Body != nil && request.ContentLength != 0 {
		decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeAPIError(response, http.StatusBadRequest, "invalid_request", "The policy transition request is invalid.", nil, h.logger)
			return
		}
	}
	principal, _ := principalFromContext(request.Context())
	err := h.recommendationPolicy.Transition(request.Context(), principal.UserID, strings.TrimSpace(request.PathValue("version")), target, body.Note)
	switch {
	case err == nil:
		response.WriteHeader(http.StatusNoContent)
	case errors.Is(err, recommendation.ErrInvalidPolicyTransition):
		writeAPIError(response, http.StatusBadRequest, "invalid_policy_transition", "The policy transition is invalid.", nil, h.logger)
	case errors.Is(err, ports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "policy_not_found", "Recommendation policy not found.", nil, h.logger)
	case errors.Is(err, ports.ErrConflict), errors.Is(err, ports.ErrSeparationOfDuties):
		writeAPIError(response, http.StatusConflict, "policy_transition_conflict", "The policy cannot enter that state.", nil, h.logger)
	default:
		writeAPIError(response, http.StatusInternalServerError, "policy_unavailable", "Recommendation policy is unavailable.", nil, h.logger)
	}
}
