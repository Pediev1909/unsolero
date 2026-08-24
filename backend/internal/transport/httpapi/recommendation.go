package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/application"
	recommendationdomain "rigmark/internal/modules/recommendation/domain"
	recommendationports "rigmark/internal/modules/recommendation/ports"
)

const maximumRecommendationBodyBytes = 64 * 1024

var uuidPathPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type recommendationService interface {
	Generate(context.Context, *identity.UserID, recommendationdomain.Input) (recommendation.Generated, error)
	GetDraft(context.Context, identity.UserID) (recommendationports.Draft, error)
	SaveDraft(context.Context, identity.UserID, recommendationports.Draft) (recommendationports.Draft, error)
	DeleteDraft(context.Context, identity.UserID) error
	ListSetups(context.Context, identity.UserID, int, int) (recommendationports.SetupPage, error)
	GetSetup(context.Context, identity.UserID, planning.SetupID) (recommendation.Generated, error)
	RenameSetup(context.Context, identity.UserID, planning.SetupID, string) error
	DeleteSetup(context.Context, identity.UserID, planning.SetupID) error
}

type spaceRequest struct {
	LengthMM        int64  `json:"length_mm"`
	WidthMM         int64  `json:"width_mm"`
	HeightMM        int64  `json:"height_mm"`
	AccessWidthMM   *int64 `json:"access_width_mm,omitempty"`
	ApartmentLiving bool   `json:"apartment_living"`
}

type equipmentRequest struct {
	Name         string `json:"name"`
	CategorySlug string `json:"category_slug"`
}

type recommendationInputRequest struct {
	Goal                string             `json:"goal"`
	Experience          string             `json:"experience"`
	BudgetMinor         int64              `json:"budget_minor"`
	Currency            string             `json:"currency"`
	AvailableSpace      spaceRequest       `json:"available_space"`
	ExistingEquipment   []equipmentRequest `json:"existing_equipment"`
	TrainingPreferences []string           `json:"training_preferences"`
	Priorities          []string           `json:"priorities"`
	FreeText            string             `json:"free_text"`
}

type draftRequest struct {
	CurrentStep         int                `json:"current_step"`
	Goal                *string            `json:"goal"`
	Experience          *string            `json:"experience"`
	BudgetMinor         *int64             `json:"budget_minor"`
	Currency            *string            `json:"currency"`
	AvailableSpace      *spaceRequest      `json:"available_space"`
	ExistingEquipment   []equipmentRequest `json:"existing_equipment"`
	TrainingPreferences []string           `json:"training_preferences"`
	Priorities          []string           `json:"priorities"`
	FreeText            string             `json:"free_text"`
}

type reasonResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
}

type scoreBreakdownResponse struct {
	GoalMatch       int `json:"goal_match"`
	BudgetMatch     int `json:"budget_match"`
	SpaceMatch      int `json:"space_match"`
	ExperienceMatch int `json:"experience_match"`
	PreferenceMatch int `json:"preference_match"`
	Quality         int `json:"quality"`
	Value           int `json:"value"`
	Durability      int `json:"durability"`
	Compatibility   int `json:"compatibility"`
	Portability     int `json:"portability"`
	Noise           int `json:"noise"`
}

type recommendedProductResponse struct {
	Rank      int                    `json:"rank"`
	Quantity  int                    `json:"quantity"`
	Score     int                    `json:"score"`
	Breakdown scoreBreakdownResponse `json:"breakdown"`
	Reasons   []reasonResponse       `json:"reasons"`
	Product   productSummaryResponse `json:"product"`
}

type alternativeResponse struct {
	ForProductID         string                 `json:"for_product_id"`
	Type                 string                 `json:"type"`
	PriceDifferenceMinor int64                  `json:"price_difference_minor"`
	Score                int                    `json:"score"`
	Reasons              []reasonResponse       `json:"reasons"`
	Product              productSummaryResponse `json:"product"`
}

type rejectedProductResponse struct {
	Code    string                 `json:"code"`
	Reason  string                 `json:"reason"`
	Product productSummaryResponse `json:"product"`
}

type recommendationResponse struct {
	RecommendationID    *string                      `json:"recommendation_id"`
	SetupID             *string                      `json:"setup_id"`
	SetupName           *string                      `json:"setup_name"`
	Saved               bool                         `json:"saved"`
	Status              string                       `json:"status"`
	TotalCost           moneyResponse                `json:"total_cost"`
	RecommendationScore int                          `json:"recommendation_score"`
	Fit                 scoreBreakdownResponse       `json:"fit"`
	Products            []recommendedProductResponse `json:"recommended_products"`
	Alternatives        []alternativeResponse        `json:"alternatives"`
	Rejected            []rejectedProductResponse    `json:"rejected_alternatives"`
	PolicyVersion       string                       `json:"policy_version"`
	EngineVersion       string                       `json:"engine_version"`
	Input               recommendationInputRequest   `json:"input"`
}

type setupNameRequest struct {
	Name string `json:"name"`
}

type setupsResponse struct {
	Setups     []setupSummaryResponse `json:"setups"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"total_pages"`
}
type setupSummaryResponse struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	ItemCount           int           `json:"item_count"`
	TotalCost           moneyResponse `json:"total_cost"`
	RecommendationScore int           `json:"recommendation_score"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// previewRecommendation answers the same question as generateRecommendation and
// deliberately persists nothing.
//
// The builder shows a live suggestion from the second question onward, so this
// runs on every change a visitor makes. Routing that through the saving path
// would file a recommendation and a named setup in their account on every
// keystroke, and they would return to a saved-setups page full of rubbish they
// never asked to keep. Passing a nil user is the whole difference.
func (h *Handler) previewRecommendation(response http.ResponseWriter, request *http.Request) {
	var body recommendationInputRequest
	if !h.decodeRecommendationJSON(response, request, &body) {
		return
	}
	generated, err := h.recommendations.Generate(request.Context(), nil, recommendationInput(body))
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, recommendationDTO(generated), h.logger)
}

func (h *Handler) generateRecommendation(response http.ResponseWriter, request *http.Request) {
	var body recommendationInputRequest
	if !h.decodeRecommendationJSON(response, request, &body) {
		return
	}
	input := recommendationInput(body)
	var userID *identity.UserID
	if principal, ok := principalFromContext(request.Context()); ok {
		userID = &principal.UserID
	}
	generated, err := h.recommendations.Generate(request.Context(), userID, input)
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, recommendationDTO(generated), h.logger)
}

func (h *Handler) getRecommendationDraft(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	draft, err := h.recommendations.GetDraft(request.Context(), principal.UserID)
	if errors.Is(err, recommendationports.ErrNotFound) {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, draftDTO(draft), h.logger)
}

func (h *Handler) saveRecommendationDraft(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body draftRequest
	if !h.decodeRecommendationJSON(response, request, &body) {
		return
	}
	draft, err := h.recommendations.SaveDraft(request.Context(), principal.UserID, draftFromRequest(body))
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, draftDTO(draft), h.logger)
}

func (h *Handler) deleteRecommendationDraft(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	if err := h.recommendations.DeleteDraft(request.Context(), principal.UserID); err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listSetups(response http.ResponseWriter, request *http.Request) {
	page, err := optionalPositiveInt(request.URL.Query().Get("page"), 1)
	if err != nil {
		h.writeRecommendationError(response, recommendation.ErrInvalidSetupPagination)
		return
	}
	pageSize, err := optionalPositiveInt(request.URL.Query().Get("page_size"), 50)
	if err != nil || page > 10_000 || pageSize > 100 {
		h.writeRecommendationError(response, recommendation.ErrInvalidSetupPagination)
		return
	}
	principal, _ := principalFromContext(request.Context())
	setupPage, err := h.recommendations.ListSetups(request.Context(), principal.UserID, page, pageSize)
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	result := make([]setupSummaryResponse, 0, len(setupPage.Setups))
	for _, setup := range setupPage.Setups {
		result = append(result, setupSummaryResponse{ID: string(setup.ID), Name: setup.Name,
			ItemCount: setup.ItemCount, TotalCost: moneyResponse{AmountMinor: setup.TotalCostMinor, Currency: setup.Currency},
			RecommendationScore: setup.ObjectiveScore, CreatedAt: setup.CreatedAt, UpdatedAt: setup.UpdatedAt})
	}
	totalPages := 0
	if setupPage.Total > 0 {
		totalPages = (setupPage.Total + pageSize - 1) / pageSize
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, setupsResponse{Setups: result, Page: page, PageSize: pageSize,
		Total: setupPage.Total, TotalPages: totalPages}, h.logger)
}

func (h *Handler) getSetup(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	setupID := strings.TrimSpace(request.PathValue("setupID"))
	if !uuidPathPattern.MatchString(setupID) {
		h.writeRecommendationError(response, recommendationports.ErrNotFound)
		return
	}
	generated, err := h.recommendations.GetSetup(request.Context(), principal.UserID, planning.SetupID(setupID))
	if err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, recommendationDTO(generated), h.logger)
}

func (h *Handler) renameSetup(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	setupID := strings.TrimSpace(request.PathValue("setupID"))
	if !uuidPathPattern.MatchString(setupID) {
		h.writeRecommendationError(response, recommendationports.ErrNotFound)
		return
	}
	var body setupNameRequest
	if !h.decodeRecommendationJSON(response, request, &body) {
		return
	}
	if err := h.recommendations.RenameSetup(request.Context(), principal.UserID, planning.SetupID(setupID), body.Name); err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteSetup(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	setupID := strings.TrimSpace(request.PathValue("setupID"))
	if !uuidPathPattern.MatchString(setupID) {
		h.writeRecommendationError(response, recommendationports.ErrNotFound)
		return
	}
	if err := h.recommendations.DeleteSetup(request.Context(), principal.UserID, planning.SetupID(setupID)); err != nil {
		h.writeRecommendationError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func recommendationInput(body recommendationInputRequest) recommendationdomain.Input {
	input := recommendationdomain.Input{Goal: planning.Goal(body.Goal), Experience: planning.ExperienceLevel(body.Experience),
		Budget: catalog.Money{AmountMinor: body.BudgetMinor, Currency: body.Currency},
		AvailableSpace: recommendationdomain.AvailableSpace{LengthMM: body.AvailableSpace.LengthMM,
			WidthMM: body.AvailableSpace.WidthMM, HeightMM: body.AvailableSpace.HeightMM,
			AccessWidthMM:   body.AvailableSpace.AccessWidthMM,
			ApartmentLiving: body.AvailableSpace.ApartmentLiving}, FreeText: strings.TrimSpace(body.FreeText)}
	for _, value := range body.TrainingPreferences {
		input.TrainingPreferences = append(input.TrainingPreferences, recommendationdomain.TrainingPreference(value))
	}
	for _, value := range body.Priorities {
		input.Priorities = append(input.Priorities, recommendationdomain.Priority(value))
	}
	for _, equipment := range body.ExistingEquipment {
		input.ExistingEquipment = append(input.ExistingEquipment,
			recommendationdomain.ExistingEquipment{Name: strings.TrimSpace(equipment.Name), CategorySlug: equipment.CategorySlug})
	}
	return input
}

func draftFromRequest(body draftRequest) recommendationports.Draft {
	draft := recommendationports.Draft{CurrentStep: body.CurrentStep, BudgetMinor: body.BudgetMinor, Currency: body.Currency,
		FreeText: strings.TrimSpace(body.FreeText)}
	if body.Goal != nil {
		value := planning.Goal(*body.Goal)
		draft.Goal = &value
	}
	if body.Experience != nil {
		value := planning.ExperienceLevel(*body.Experience)
		draft.Experience = &value
	}
	// A room of nothing by nothing is not a room, it is the absence of one.
	// A non-spatial vertical never asks for measurements, so its client sends
	// the field zeroed rather than omitted -- and a zeroed space was reaching
	// the database, where a check constraint written for the physical vertical
	// rejected it and turned every draft save into a 500. Treating it as
	// absent here fixes it for any client, not just the one we ship.
	if space := body.AvailableSpace; space != nil &&
		(space.LengthMM > 0 || space.WidthMM > 0 || space.HeightMM > 0) {
		draft.AvailableSpace = &recommendationdomain.AvailableSpace{
			LengthMM: space.LengthMM, WidthMM: space.WidthMM,
			HeightMM: space.HeightMM, AccessWidthMM: space.AccessWidthMM,
			ApartmentLiving: space.ApartmentLiving}
	}
	// Empty, not nil. Both columns are NOT NULL, and appending to a nil slice
	// that never receives an element leaves it nil, which pgx sends as NULL.
	// Nobody has chosen a preference or a priority until the questions that ask
	// for them, so every draft save from question one onwards failed with a
	// constraint violation, and a signed-in visitor watched "your latest change
	// could not be saved" appear on each answer. Same class as the zeroed space
	// above: a value the non-physical vertical never fills, reaching a column
	// that was written expecting it.
	draft.TrainingPreferences = make([]recommendationdomain.TrainingPreference, 0, len(body.TrainingPreferences))
	for _, value := range body.TrainingPreferences {
		draft.TrainingPreferences = append(draft.TrainingPreferences, recommendationdomain.TrainingPreference(value))
	}
	draft.Priorities = make([]recommendationdomain.Priority, 0, len(body.Priorities))
	for _, value := range body.Priorities {
		draft.Priorities = append(draft.Priorities, recommendationdomain.Priority(value))
	}
	for _, equipment := range body.ExistingEquipment {
		draft.ExistingEquipment = append(draft.ExistingEquipment,
			recommendationdomain.ExistingEquipment{Name: strings.TrimSpace(equipment.Name), CategorySlug: equipment.CategorySlug})
	}
	return draft
}

func draftDTO(draft recommendationports.Draft) draftRequest {
	result := draftRequest{CurrentStep: draft.CurrentStep, BudgetMinor: draft.BudgetMinor,
		Currency: draft.Currency, FreeText: draft.FreeText,
		ExistingEquipment: []equipmentRequest{}, TrainingPreferences: []string{}, Priorities: []string{}}
	if draft.Goal != nil {
		value := string(*draft.Goal)
		result.Goal = &value
	}
	if draft.Experience != nil {
		value := string(*draft.Experience)
		result.Experience = &value
	}
	if draft.AvailableSpace != nil {
		result.AvailableSpace = &spaceRequest{LengthMM: draft.AvailableSpace.LengthMM,
			WidthMM: draft.AvailableSpace.WidthMM, HeightMM: draft.AvailableSpace.HeightMM,
			AccessWidthMM:   draft.AvailableSpace.AccessWidthMM,
			ApartmentLiving: draft.AvailableSpace.ApartmentLiving}
	}
	for _, item := range draft.ExistingEquipment {
		result.ExistingEquipment = append(result.ExistingEquipment,
			equipmentRequest{Name: item.Name, CategorySlug: item.CategorySlug})
	}
	for _, value := range draft.TrainingPreferences {
		result.TrainingPreferences = append(result.TrainingPreferences, string(value))
	}
	for _, value := range draft.Priorities {
		result.Priorities = append(result.Priorities, string(value))
	}
	return result
}

func recommendationDTO(generated recommendation.Generated) recommendationResponse {
	result := generated.Result
	response := recommendationResponse{Saved: generated.Saved, Status: string(result.Status),
		TotalCost:           moneyResponse{AmountMinor: result.TotalCost.AmountMinor, Currency: result.TotalCost.Currency},
		RecommendationScore: result.ObjectiveScore, Fit: breakdownDTO(result.Breakdown),
		PolicyVersion: result.PolicyVersion, EngineVersion: result.EngineVersion,
		Input:    recommendationInputDTO(generated.Input),
		Products: []recommendedProductResponse{}, Alternatives: []alternativeResponse{}, Rejected: []rejectedProductResponse{}}
	if generated.RecommendationID != nil {
		value := string(*generated.RecommendationID)
		response.RecommendationID = &value
	}
	if generated.SetupID != nil {
		value := string(*generated.SetupID)
		response.SetupID = &value
	}
	response.SetupName = generated.SetupName
	for _, item := range result.Selected {
		if product, ok := generated.Products[item.Product.Candidate.ProductID]; ok {
			response.Products = append(response.Products, recommendedProductResponse{Rank: item.Rank,
				Quantity: item.Quantity, Score: item.Product.ObjectiveScore, Breakdown: breakdownDTO(item.Product.Breakdown),
				Reasons: reasonDTOs(item.Product.Reasons), Product: productSummaryDTO(product)})
		}
	}
	for _, item := range result.Alternatives {
		if product, ok := generated.Products[item.Product.Candidate.ProductID]; ok {
			response.Alternatives = append(response.Alternatives, alternativeResponse{ForProductID: string(item.ForProductID),
				Type: string(item.Type), PriceDifferenceMinor: item.PriceDifferenceMinor, Score: item.Product.ObjectiveScore,
				Reasons: reasonDTOs(item.Product.Reasons), Product: productSummaryDTO(product)})
		}
	}
	for _, item := range result.Rejected {
		if product, ok := generated.Products[item.Candidate.ProductID]; ok {
			response.Rejected = append(response.Rejected, rejectedProductResponse{Code: item.Code, Reason: item.Message,
				Product: productSummaryDTO(product)})
		}
	}
	return response
}

func recommendationInputDTO(input recommendationdomain.Input) recommendationInputRequest {
	result := recommendationInputRequest{
		Goal: string(input.Goal), Experience: string(input.Experience),
		BudgetMinor: input.Budget.AmountMinor, Currency: input.Budget.Currency,
		AvailableSpace: spaceRequest{LengthMM: input.AvailableSpace.LengthMM,
			WidthMM: input.AvailableSpace.WidthMM, HeightMM: input.AvailableSpace.HeightMM,
			AccessWidthMM:   input.AvailableSpace.AccessWidthMM,
			ApartmentLiving: input.AvailableSpace.ApartmentLiving},
		FreeText: input.FreeText, ExistingEquipment: []equipmentRequest{},
		TrainingPreferences: []string{}, Priorities: []string{},
	}
	for _, equipment := range input.ExistingEquipment {
		result.ExistingEquipment = append(result.ExistingEquipment, equipmentRequest{Name: equipment.Name, CategorySlug: equipment.CategorySlug})
	}
	for _, preference := range input.TrainingPreferences {
		result.TrainingPreferences = append(result.TrainingPreferences, string(preference))
	}
	for _, priority := range input.Priorities {
		result.Priorities = append(result.Priorities, string(priority))
	}
	return result
}

func breakdownDTO(value recommendationdomain.ScoreBreakdown) scoreBreakdownResponse {
	return scoreBreakdownResponse{GoalMatch: value.GoalMatch, BudgetMatch: value.BudgetMatch,
		SpaceMatch: value.SpaceMatch, ExperienceMatch: value.ExperienceMatch,
		PreferenceMatch: value.PreferenceMatch, Quality: value.Quality, Value: value.Value,
		Durability: value.Durability, Compatibility: value.Compatibility,
		Portability: value.Portability, Noise: value.Noise}
}

func reasonDTOs(reasons []recommendationdomain.Reason) []reasonResponse {
	result := make([]reasonResponse, 0, len(reasons))
	for _, reason := range reasons {
		result = append(result, reasonResponse{Code: reason.Code,
			Message: reason.Message, Dimension: reason.Dimension, Score: reason.Score})
	}
	return result
}

func (h *Handler) decodeRecommendationJSON(response http.ResponseWriter, request *http.Request, destination any) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(response, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use application/json for this request.", nil, h.logger)
		return false
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumRecommendationBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(destination); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body is invalid.", nil, h.logger)
		return false
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body must contain one JSON object.", nil, h.logger)
		return false
	}
	return true
}

func (h *Handler) writeRecommendationError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, recommendation.ErrInvalidSetupPagination):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_pagination", "Saved setup pagination is invalid.", nil, h.logger)
	case errors.Is(err, recommendationdomain.ErrInvalidInput), errors.Is(err, recommendation.ErrInvalidDraft), errors.Is(err, recommendation.ErrInvalidSetupName):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_recommendation_input", "Check your answers and try again.", nil, h.logger)
	case errors.Is(err, recommendationports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "setup_not_found", "This saved setup could not be found.", nil, h.logger)
	case errors.Is(err, recommendationdomain.ErrInvalidCandidate):
		// A 500 is right -- the visitor did nothing wrong, our own catalog data
		// is what the policy rejected -- but the reason has to survive into the
		// log. The redaction rule reduces an error attribute to its type name,
		// which turned a bad price on one product into "*fmt.wrapError" and
		// nothing else. This message is assembled from product identifiers and
		// fixed strings only, so it carries nothing private.
		h.logger.Error("recommendation rejected a catalog candidate", "reason", err.Error())
		writeAPIError(response, http.StatusInternalServerError, "recommendation_unavailable", "Recommendations are temporarily unavailable.", nil, h.logger)
	default:
		h.logger.Error("recommendation request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "recommendation_unavailable", "Recommendations are temporarily unavailable.", nil, h.logger)
	}
}
