package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	evidenceapp "rigmark/internal/modules/evidence/application"
	"rigmark/internal/modules/evidence/domain"
	"rigmark/internal/modules/evidence/ports"
)

type evidenceSourceRequest struct {
	Type        string  `json:"source_type"`
	Title       string  `json:"title"`
	Publisher   string  `json:"publisher"`
	SourceURL   *string `json:"source_url"`
	IsFictional bool    `json:"is_fictional"`
}

type evidenceObservationRequest struct {
	SourceID   string     `json:"source_id"`
	ProductID  string     `json:"product_id"`
	ObservedAt time.Time  `json:"observed_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Confidence int16      `json:"confidence"`
	Notes      string     `json:"notes"`
}

type evidenceRevisionRequest struct {
	Product    productInputRequest       `json:"product"`
	FactLinks  []evidenceFactLinkRequest `json:"fact_links"`
	Rationales []scoreRationaleRequest   `json:"score_rationales"`
}

type evidenceFactLinkRequest struct {
	FactKey        string `json:"fact_key"`
	ObservationID  string `json:"observation_id"`
	Classification string `json:"classification"`
}

type scoreRationaleRequest struct {
	ScoreKey      string `json:"score_key"`
	Rationale     string `json:"rationale"`
	ObservationID string `json:"observation_id"`
}

type evidenceSourceResponse struct {
	ID           string     `json:"id"`
	SourceType   string     `json:"source_type"`
	Title        string     `json:"title"`
	Publisher    string     `json:"publisher"`
	SourceURL    *string    `json:"source_url"`
	IsFictional  bool       `json:"is_fictional"`
	ReviewStatus string     `json:"review_status"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNote   string     `json:"review_note"`
	CreatedAt    time.Time  `json:"created_at"`
}

type evidenceRevisionResponse struct {
	FactRevisionID  string     `json:"fact_revision_id"`
	ScoreRevisionID string     `json:"score_revision_id"`
	FactVersion     int        `json:"fact_version"`
	ScoreVersion    int        `json:"score_version"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	PublishedAt     *time.Time `json:"published_at"`
	ValidUntil      *time.Time `json:"valid_until"`
	ReviewNote      string     `json:"review_note"`
}

type productGovernanceResponse struct {
	ProductID                string                     `json:"product_id"`
	ProductName              string                     `json:"product_name"`
	Status                   string                     `json:"status"`
	PublishedFactRevisionID  *string                    `json:"published_fact_revision_id"`
	PublishedScoreRevisionID *string                    `json:"published_score_revision_id"`
	Revisions                []evidenceRevisionResponse `json:"revisions"`
	Provenance               []provenanceResponse       `json:"provenance"`
	Audit                    []auditEventResponse       `json:"audit"`
}

type provenanceResponse struct {
	FactKey        string                 `json:"fact_key"`
	ScoreKey       string                 `json:"score_key"`
	Classification string                 `json:"classification"`
	Rationale      string                 `json:"rationale"`
	Observation    observationResponse    `json:"observation"`
	Source         evidenceSourceResponse `json:"source"`
}

type observationResponse struct {
	ID         string     `json:"id"`
	ObservedAt time.Time  `json:"observed_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Confidence int16      `json:"confidence"`
	Notes      string     `json:"notes"`
}

type auditEventResponse struct {
	Action     string            `json:"action"`
	ActorEmail *string           `json:"actor_email"`
	Changes    map[string]string `json:"changes"`
	OccurredAt time.Time         `json:"occurred_at"`
}

type governancePageResponse struct {
	Items      []productGovernanceResponse `json:"items"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Total      int64                       `json:"total"`
	TotalPages int                         `json:"total_pages"`
}

func (h *Handler) adminListProductGovernance(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	items, total, err := h.evidence.ListProducts(request.Context(), page, pageSize)
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	dtos := make([]productGovernanceResponse, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, governanceDTO(item))
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	h.writeAdminJSON(response, http.StatusOK, governancePageResponse{
		Items: dtos, Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages,
	})
}

func (h *Handler) adminGetProductGovernance(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	result, err := h.evidence.GetProduct(request.Context(), catalog.ProductID(id))
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, governanceDTO(result))
}

func (h *Handler) adminCreateEvidenceSource(response http.ResponseWriter, request *http.Request) {
	var body evidenceSourceRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.evidence.CreateSource(request.Context(), principal.UserID, domain.SourceInput{
		Type: domain.SourceType(body.Type), Title: body.Title, Publisher: body.Publisher,
		URL: body.SourceURL, IsFictional: body.IsFictional,
	})
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, sourceDTO(result))
}

// adminListEvidenceSources and adminListEvidenceObservations back the forms
// that record evidence. Without them an operator can create a source but never
// see one again, which is why recording evidence was only possible in SQL.
func (h *Handler) adminListEvidenceSources(response http.ResponseWriter, request *http.Request) {
	limit, _ := strconv.Atoi(strings.TrimSpace(request.URL.Query().Get("limit")))
	sources, err := h.evidence.ListSources(request.Context(), limit)
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	items := make([]evidenceSourceResponse, 0, len(sources))
	for _, source := range sources {
		items = append(items, sourceDTO(source))
	}
	h.writeAdminJSON(response, http.StatusOK, map[string]any{"sources": items})
}

func (h *Handler) adminListEvidenceObservations(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	observations, err := h.evidence.ListObservations(request.Context(), id)
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	items := make([]productObservationResponse, 0, len(observations))
	for _, observation := range observations {
		items = append(items, productObservationResponse{
			ID: observation.ID, SourceID: observation.SourceID,
			ObservedAt: observation.ObservedAt, ExpiresAt: observation.ExpiresAt,
			Confidence: observation.Confidence, Notes: observation.Notes,
		})
	}
	h.writeAdminJSON(response, http.StatusOK, map[string]any{"observations": items})
}

// productObservationResponse carries source_id, which observationResponse does
// not: the revision form has to show which source each observation came from.
type productObservationResponse struct {
	ID         string     `json:"id"`
	SourceID   string     `json:"source_id"`
	ObservedAt time.Time  `json:"observed_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Confidence int16      `json:"confidence"`
	Notes      string     `json:"notes"`
}

func (h *Handler) adminReviewEvidenceSource(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("sourceID"), h)
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.evidence.ReviewSource(request.Context(), principal.UserID,
		id, domain.ReviewStatus(body.Status), body.Note)
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, sourceDTO(result))
}

func (h *Handler) adminCreateEvidenceObservation(response http.ResponseWriter, request *http.Request) {
	var body evidenceObservationRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.evidence.CreateObservation(request.Context(), principal.UserID, domain.ObservationInput{
		SourceID: body.SourceID, ProductID: catalog.ProductID(body.ProductID),
		ObservedAt: body.ObservedAt, ExpiresAt: body.ExpiresAt,
		Confidence: body.Confidence, Notes: body.Notes,
	})
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, observationResponse{
		ID: result.ID, ObservedAt: result.ObservedAt, ExpiresAt: result.ExpiresAt,
		Confidence: result.Confidence, Notes: result.Notes,
	})
}

func (h *Handler) adminCreateEvidenceRevision(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	var body evidenceRevisionRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	input := body.Product.domain()
	product := catalog.Product{ID: catalog.ProductID(id), CategoryID: input.CategoryID,
		BrandID: input.BrandID, Name: input.Name, Slug: input.Slug,
		Description: input.Description, Price: input.Price, Billing: input.Billing,
		Dimensions:  input.Dimensions,
		WeightGrams: input.WeightGrams, MaxCapacityGrams: input.MaxCapacityGrams,
		Material: input.Material, WarrantyMonths: input.WarrantyMonths,
		Scores: input.Scores, Status: catalog.ProductStatusDraft}
	factLinks := make([]domain.FactLink, 0, len(body.FactLinks))
	for _, link := range body.FactLinks {
		factLinks = append(factLinks, domain.FactLink{FactKey: link.FactKey,
			ObservationID: link.ObservationID, Classification: domain.Classification(link.Classification)})
	}
	rationales := make([]domain.ScoreRationale, 0, len(body.Rationales))
	for _, rationale := range body.Rationales {
		rationales = append(rationales, domain.ScoreRationale{ScoreKey: rationale.ScoreKey,
			Rationale: rationale.Rationale, ObservationID: rationale.ObservationID})
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.evidence.CreateRevision(request.Context(), principal.UserID,
		domain.RevisionInput{Product: product, FactLinks: factLinks, Scores: input.Scores, Rationales: rationales})
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, revisionDTO(result))
}

func (h *Handler) adminSubmitEvidenceRevision(response http.ResponseWriter, request *http.Request) {
	h.transitionEvidenceRevision(response, request, "submit")
}
func (h *Handler) adminApproveEvidenceRevision(response http.ResponseWriter, request *http.Request) {
	h.transitionEvidenceRevision(response, request, "approve")
}
func (h *Handler) adminRejectEvidenceRevision(response http.ResponseWriter, request *http.Request) {
	h.transitionEvidenceRevision(response, request, "reject")
}
func (h *Handler) adminPublishEvidenceRevision(response http.ResponseWriter, request *http.Request) {
	h.transitionEvidenceRevision(response, request, "publish")
}

func (h *Handler) transitionEvidenceRevision(response http.ResponseWriter, request *http.Request, action string) {
	id, ok := adminUUIDPath(response, request.PathValue("revisionID"), h)
	if !ok {
		return
	}
	note := ""
	if action == "approve" || action == "reject" {
		var body struct {
			Note string `json:"note"`
		}
		if !h.decodeAdminJSON(response, request, &body) {
			return
		}
		note = body.Note
	}
	principal, _ := principalFromContext(request.Context())
	var result domain.Revision
	var err error
	switch action {
	case "submit":
		result, err = h.evidence.Submit(request.Context(), principal.UserID, id)
	case "approve":
		result, err = h.evidence.Approve(request.Context(), principal.UserID, id, note)
	case "reject":
		result, err = h.evidence.Reject(request.Context(), principal.UserID, id, note)
	case "publish":
		result, err = h.evidence.Publish(request.Context(), principal.UserID, id)
	}
	if err != nil {
		h.writeEvidenceError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, revisionDTO(result))
}

func (h *Handler) writeEvidenceError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, evidenceapp.ErrInvalidInput):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_evidence_input", "Check the evidence fields and provenance coverage.", fieldErrors(err), h.logger)
	case errors.Is(err, ports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "evidence_not_found", "The evidence record was not found.", nil, h.logger)
	case errors.Is(err, ports.ErrSeparationOfDuties):
		writeAPIError(response, http.StatusConflict, "evidence_separation_required", "A different reviewer or publisher must perform this action.", nil, h.logger)
	case errors.Is(err, ports.ErrIncompleteProvenance):
		writeAPIError(response, http.StatusConflict, "evidence_incomplete", "Evidence is incomplete, unverified, stale, or belongs to another product.", nil, h.logger)
	case errors.Is(err, ports.ErrConflict):
		writeAPIError(response, http.StatusConflict, "evidence_state_conflict", "The evidence record is not in a valid state for this action.", nil, h.logger)
	default:
		h.logger.Error("evidence request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "evidence_unavailable", "Evidence governance is temporarily unavailable.", nil, h.logger)
	}
}

func sourceDTO(value domain.Source) evidenceSourceResponse {
	return evidenceSourceResponse{ID: value.ID, SourceType: string(value.Type),
		Title: value.Title, Publisher: value.Publisher, SourceURL: value.URL,
		IsFictional: value.IsFictional, ReviewStatus: string(value.ReviewStatus),
		ReviewedAt: value.ReviewedAt, ReviewNote: value.ReviewNote, CreatedAt: value.CreatedAt}
}

func revisionDTO(value domain.Revision) evidenceRevisionResponse {
	return evidenceRevisionResponse{FactRevisionID: value.FactRevisionID,
		ScoreRevisionID: value.ScoreRevisionID, FactVersion: value.FactVersion,
		ScoreVersion: value.ScoreVersion, Status: string(value.Status),
		CreatedAt: value.CreatedAt, SubmittedAt: value.SubmittedAt,
		ReviewedAt: value.ReviewedAt, PublishedAt: value.PublishedAt,
		ValidUntil: value.ValidUntil, ReviewNote: value.ReviewNote}
}

func governanceDTO(value domain.ProductGovernance) productGovernanceResponse {
	result := productGovernanceResponse{ProductID: string(value.ProductID),
		ProductName: value.ProductName, Status: string(value.Status),
		PublishedFactRevisionID:  value.PublishedFactRevisionID,
		PublishedScoreRevisionID: value.PublishedScoreRevisionID,
		Revisions:                []evidenceRevisionResponse{}, Provenance: []provenanceResponse{},
		Audit: []auditEventResponse{}}
	for _, revision := range value.Revisions {
		result.Revisions = append(result.Revisions, revisionDTO(revision))
	}
	for _, item := range value.Provenance {
		result.Provenance = append(result.Provenance, provenanceResponse{
			FactKey: item.FactKey, ScoreKey: item.ScoreKey,
			Classification: string(item.Classification), Rationale: item.Rationale,
			Observation: observationResponse{ID: item.Observation.ID,
				ObservedAt: item.Observation.ObservedAt, ExpiresAt: item.Observation.ExpiresAt,
				Confidence: item.Observation.Confidence, Notes: item.Observation.Notes},
			Source: sourceDTO(item.Source),
		})
	}
	for _, item := range value.Audit {
		result.Audit = append(result.Audit, auditEventResponse{Action: item.Action,
			ActorEmail: item.ActorEmail, Changes: item.Changes, OccurredAt: item.OccurredAt})
	}
	return result
}
