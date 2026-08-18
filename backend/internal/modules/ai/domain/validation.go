package domain

import (
	"fmt"
	"regexp"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
)

const (
	maxUserText       = 4_000
	maxFreeText       = 1_000
	maxMoneyMinor     = 1_000_000_000_000
	maxDimensionMM    = 100_000
	maxProducts       = 100
	maxCollectionSize = 20
)

var (
	localePattern = regexp.MustCompile(`^[A-Za-z]{2,3}([_-][A-Za-z]{2,4})?$`)
	codePattern   = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,99}$`)
	slugPattern   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

func ValidateUnderstandRequest(value UnderstandUserInputRequest) error {
	return validateTextAndLocale(value.Text, value.Locale)
}

func ValidateUnderstanding(value UserInputUnderstanding) error {
	if !localePattern.MatchString(value.Language) || value.Confidence < 0 || value.Confidence > 100 {
		return invalidOutput("language or confidence is invalid")
	}
	if err := validateFields(value.RecognizedFields, true); err != nil {
		return invalidOutput(err.Error())
	}
	if err := validateAmbiguities(value.Ambiguities); err != nil {
		return invalidOutput(err.Error())
	}
	if value.NeedsClarification != (len(value.Ambiguities) > 0) {
		return invalidOutput("clarification flag does not match ambiguities")
	}
	return nil
}

func ValidateExtractRequest(value ExtractRequirementsRequest) error {
	if err := validateTextAndLocale(value.Text, value.Locale); err != nil {
		return err
	}
	return validateRequirements(value.Known, ErrInvalidInput)
}

func ValidateRequirementsOutput(value RequirementsDraft) error {
	return validateRequirements(value, ErrInvalidProviderOutput)
}

func ValidateClarifyingRequest(value ClarifyingQuestionRequest) error {
	if !localePattern.MatchString(value.Locale) {
		return fmt.Errorf("%w: locale is invalid", ErrInvalidInput)
	}
	if err := validateRequirements(value.Requirements, ErrInvalidInput); err != nil {
		return err
	}
	if err := validateAmbiguities(value.Ambiguities); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if len(value.Ambiguities) == 0 {
		return fmt.Errorf("%w: at least one ambiguity is required", ErrInvalidInput)
	}
	return nil
}

func ValidateClarifyingQuestion(value ClarifyingQuestion, request ClarifyingQuestionRequest) error {
	if !validField(value.Field) || promptForField(value.Field) != value.Prompt {
		return invalidOutput("question field or prompt is invalid")
	}
	allowed := false
	for _, ambiguity := range request.Ambiguities {
		if ambiguity.Field == value.Field {
			allowed = true
		}
	}
	if !allowed {
		return invalidOutput("question targets a field that was not ambiguous")
	}
	return nil
}

func ValidateExplainRequest(value ExplainRecommendationRequest) error {
	if !localePattern.MatchString(value.Locale) {
		return fmt.Errorf("%w: locale is invalid", ErrInvalidInput)
	}
	return validateRecommendation(value.Recommendation, ErrInvalidInput)
}

func ValidateExplanationPlan(value ExplanationPlan, request ExplainRecommendationRequest) error {
	if !map[ExplanationStyle]bool{StyleBalanced: true, StyleBudget: true, StyleSpace: true, StylePerformance: true}[value.Style] {
		return invalidOutput("explanation style is invalid")
	}
	if len(value.Points) == 0 || len(value.Points) > len(request.Recommendation.Selected) {
		return invalidOutput("explanation points are missing or excessive")
	}
	products := selectedProducts(request.Recommendation)
	seenProducts := make(map[catalog.ProductID]bool, len(value.Points))
	for _, point := range value.Points {
		product, exists := products[point.ProductID]
		if !exists || seenProducts[point.ProductID] {
			return invalidOutput("explanation references an unknown or duplicate product")
		}
		seenProducts[point.ProductID] = true
		if err := validateReasonReferences(point.ReasonCodes, product.ReasonCodes); err != nil {
			return err
		}
		if err := validateFactKeys(point.FactKeys, product.Product); err != nil {
			return err
		}
	}
	return nil
}

func ValidateRefineRequest(value RefineRecommendationRequest) error {
	if err := validateTextAndLocale(value.Instruction, value.Locale); err != nil {
		return err
	}
	if err := validateRequirements(value.Current, ErrInvalidInput); err != nil {
		return err
	}
	return validateRecommendation(value.Recommendation, ErrInvalidInput)
}

func ValidateRefinement(value Refinement) error {
	if err := validateRequirements(value.Requirements, ErrInvalidProviderOutput); err != nil {
		return err
	}
	if len(value.ChangedFields) == 0 {
		return invalidOutput("changed fields are required")
	}
	if err := validateFields(value.ChangedFields, false); err != nil {
		return invalidOutput(err.Error())
	}
	return nil
}

func ValidateCompareRequest(value CompareProductsRequest) error {
	if !localePattern.MatchString(value.Locale) || len(value.Products) < 2 || len(value.Products) > 4 {
		return fmt.Errorf("%w: comparison locale or product count is invalid", ErrInvalidInput)
	}
	seen := make(map[catalog.ProductID]bool, len(value.Products))
	for _, product := range value.Products {
		if err := validateProduct(product, ErrInvalidInput); err != nil {
			return err
		}
		if seen[product.ID] {
			return fmt.Errorf("%w: comparison products are duplicated", ErrInvalidInput)
		}
		seen[product.ID] = true
	}
	for _, key := range value.Focus {
		if !validFactKey(key) {
			return fmt.Errorf("%w: comparison focus is invalid", ErrInvalidInput)
		}
	}
	return nil
}

func ValidateComparisonPlan(value ComparisonPlan, request CompareProductsRequest) error {
	if len(value.OrderedProductIDs) != len(request.Products) {
		return invalidOutput("comparison must retain every input product")
	}
	products := make(map[catalog.ProductID]ProductSnapshot, len(request.Products))
	for _, product := range request.Products {
		products[product.ID] = product
	}
	for index, id := range value.OrderedProductIDs {
		if id != request.Products[index].ID {
			return invalidOutput("comparison changed the canonical product order")
		}
	}
	if len(value.FactKeys) == 0 || len(value.FactKeys) > len(allFactKeys()) {
		return invalidOutput("comparison fact keys are missing or excessive")
	}
	if err := validateFactKeysForProducts(value.FactKeys, products); err != nil {
		return err
	}
	if len(value.Highlights) > 12 {
		return invalidOutput("comparison has too many highlights")
	}
	for _, highlight := range value.Highlights {
		product, exists := products[highlight.ProductID]
		if !exists || !containsFactKey(value.FactKeys, highlight.FactKey) || !factAvailable(highlight.FactKey, product) ||
			(highlight.Kind != HighlightStrength && highlight.Kind != HighlightTradeoff) {
			return invalidOutput("comparison highlight is invalid")
		}
	}
	return nil
}

func validateTextAndLocale(text, locale string) error {
	if strings.TrimSpace(text) == "" || len(text) > maxUserText || !localePattern.MatchString(locale) {
		return fmt.Errorf("%w: text or locale is invalid", ErrInvalidInput)
	}
	return nil
}

func validateRequirements(value RequirementsDraft, base error) error {
	if value.Goal != nil && !map[planning.Goal]bool{
		planning.GoalBuildMuscle: true, planning.GoalStrength: true, planning.GoalGeneralFitness: true,
		planning.GoalWeightLoss: true, planning.GoalMobility: true,
	}[*value.Goal] {
		return fmt.Errorf("%w: goal is invalid", base)
	}
	if value.Experience != nil && !map[planning.ExperienceLevel]bool{
		planning.ExperienceBeginner: true, planning.ExperienceIntermediate: true, planning.ExperienceAdvanced: true,
	}[*value.Experience] {
		return fmt.Errorf("%w: experience is invalid", base)
	}
	if value.Budget != nil && (value.Budget.AmountMinor <= 0 || value.Budget.AmountMinor > maxMoneyMinor || value.Budget.Validate() != nil) {
		return fmt.Errorf("%w: budget is invalid", base)
	}
	if value.AvailableSpace != nil && (value.AvailableSpace.LengthMM <= 0 || value.AvailableSpace.WidthMM <= 0 || value.AvailableSpace.HeightMM <= 0 ||
		value.AvailableSpace.LengthMM > maxDimensionMM || value.AvailableSpace.WidthMM > maxDimensionMM || value.AvailableSpace.HeightMM > maxDimensionMM) {
		return fmt.Errorf("%w: available space is invalid", base)
	}
	if len(value.FreeText) > maxFreeText || len(value.ExistingEquipment) > maxCollectionSize ||
		len(value.TrainingPreferences) > maxCollectionSize || len(value.Priorities) > maxCollectionSize {
		return fmt.Errorf("%w: requirements exceed supported limits", base)
	}
	if err := validateEquipment(value.ExistingEquipment); err != nil {
		return fmt.Errorf("%w: %v", base, err)
	}
	if err := validatePreferences(value.TrainingPreferences); err != nil {
		return fmt.Errorf("%w: %v", base, err)
	}
	if err := validatePriorities(value.Priorities); err != nil {
		return fmt.Errorf("%w: %v", base, err)
	}
	return nil
}

func validateEquipment(values []recommendation.ExistingEquipment) error {
	for _, value := range values {
		if strings.TrimSpace(value.Name) == "" || len(value.Name) > 120 || len(value.Capabilities) > maxCollectionSize ||
			(value.CategorySlug != "" && !slugPattern.MatchString(value.CategorySlug)) {
			return errorsForValidation("existing equipment is invalid")
		}
		seen := make(map[recommendation.Capability]bool, len(value.Capabilities))
		for _, capability := range value.Capabilities {
			if !codePattern.MatchString(string(capability)) || seen[capability] {
				return errorsForValidation("equipment capability is invalid")
			}
			seen[capability] = true
		}
	}
	return nil
}

// validatePreferences and validatePriorities check the normalized-code shape
// and reject duplicates. The permitted vocabulary is declared by the active
// recommendation policy and enforced by the engine, so an AI-produced value
// that is well-formed but undeclared is refused there rather than here.
func validatePreferences(values []recommendation.TrainingPreference) error {
	return validateUniqueCodes(values)
}

func validatePriorities(values []recommendation.Priority) error {
	return validateUniqueCodes(values)
}

func validateUniqueCodes[T ~string](values []T) error {
	seen := make(map[T]bool, len(values))
	for _, value := range values {
		if !codePattern.MatchString(string(value)) || seen[value] {
			return errorsForValidation("value is unsupported or duplicated")
		}
		seen[value] = true
	}
	return nil
}

func validateUnique[T comparable](values []T, valid map[T]bool) error {
	seen := make(map[T]bool, len(values))
	for _, value := range values {
		if !valid[value] || seen[value] {
			return errorsForValidation("value is unsupported or duplicated")
		}
		seen[value] = true
	}
	return nil
}

func validateAmbiguities(values []Ambiguity) error {
	if len(values) > len(allFields()) {
		return errorsForValidation("too many ambiguities")
	}
	seen := make(map[RequirementField]bool, len(values))
	for _, value := range values {
		if !validField(value.Field) || seen[value.Field] ||
			(value.Kind != AmbiguityMissing && value.Kind != AmbiguityConflicting && value.Kind != AmbiguityUnclear) {
			return errorsForValidation("ambiguity is invalid or duplicated")
		}
		seen[value.Field] = true
	}
	return nil
}

func validateFields(values []RequirementField, allowEmpty bool) error {
	if (!allowEmpty && len(values) == 0) || len(values) > len(allFields()) {
		return errorsForValidation("requirement fields are missing or excessive")
	}
	seen := make(map[RequirementField]bool, len(values))
	for _, value := range values {
		if !validField(value) || seen[value] {
			return errorsForValidation("requirement field is invalid or duplicated")
		}
		seen[value] = true
	}
	return nil
}

func validateRecommendation(value RecommendationSnapshot, base error) error {
	if err := validateRequirements(value.Requirements, base); err != nil {
		return err
	}
	if len(value.Selected) == 0 || len(value.Selected) > maxProducts || value.ObjectiveScore < 0 || value.ObjectiveScore > 100 ||
		strings.TrimSpace(value.PolicyVersion) == "" || strings.TrimSpace(value.EngineVersion) == "" ||
		value.TotalCost.AmountMinor <= 0 || value.TotalCost.Validate() != nil {
		return fmt.Errorf("%w: recommendation summary is invalid", base)
	}
	seen := make(map[catalog.ProductID]bool, maxProducts)
	for _, item := range value.Selected {
		if err := validateDeterministicProduct(item, base); err != nil || item.Rank < 1 || seen[item.Product.ID] {
			return fmt.Errorf("%w: selected product is invalid", base)
		}
		seen[item.Product.ID] = true
	}
	for _, item := range value.Alternatives {
		if err := validateDeterministicProduct(item, base); err != nil || seen[item.Product.ID] {
			return fmt.Errorf("%w: alternative product is invalid", base)
		}
		seen[item.Product.ID] = true
	}
	for _, item := range value.Rejected {
		if err := validateProduct(item.Product, base); err != nil || !codePattern.MatchString(item.Code) || seen[item.Product.ID] {
			return fmt.Errorf("%w: rejected product is invalid", base)
		}
		seen[item.Product.ID] = true
	}
	return nil
}

func validateDeterministicProduct(value DeterministicProduct, base error) error {
	if err := validateProduct(value.Product, base); err != nil {
		return err
	}
	if value.ObjectiveScore < 0 || value.ObjectiveScore > 100 || len(value.ReasonCodes) == 0 || len(value.ReasonCodes) > maxCollectionSize {
		return fmt.Errorf("%w: deterministic score or reasons are invalid", base)
	}
	seen := make(map[string]bool, len(value.ReasonCodes))
	for _, code := range value.ReasonCodes {
		if !codePattern.MatchString(code) || seen[code] {
			return fmt.Errorf("%w: deterministic reason code is invalid", base)
		}
		seen[code] = true
	}
	return nil
}

func validateProduct(value ProductSnapshot, base error) error {
	if value.ID == "" || strings.TrimSpace(value.Name) == "" || strings.TrimSpace(value.Category) == "" ||
		strings.TrimSpace(value.Brand) == "" || value.Price.AmountMinor <= 0 || value.Price.Validate() != nil ||
		value.Dimensions.Validate() != nil || value.WeightGrams <= 0 || value.WarrantyMonths < 0 || value.Scores.Validate() != nil {
		return fmt.Errorf("%w: product snapshot is invalid", base)
	}
	if value.MaxCapacityGrams != nil && *value.MaxCapacityGrams <= 0 {
		return fmt.Errorf("%w: maximum capacity is invalid", base)
	}
	return nil
}

func validateReasonReferences(values, allowed []string) error {
	if len(values) == 0 || len(values) > len(allowed) {
		return invalidOutput("reason references are missing or excessive")
	}
	allowlist := make(map[string]bool, len(allowed))
	for _, code := range allowed {
		allowlist[code] = true
	}
	seen := make(map[string]bool, len(values))
	for _, code := range values {
		if !allowlist[code] || seen[code] {
			return invalidOutput("reason reference is unknown or duplicated")
		}
		seen[code] = true
	}
	return nil
}

func validateFactKeys(values []FactKey, product ProductSnapshot) error {
	if len(values) == 0 || len(values) > len(allFactKeys()) {
		return invalidOutput("fact references are missing or excessive")
	}
	seen := make(map[FactKey]bool, len(values))
	for _, key := range values {
		if !factAvailable(key, product) || seen[key] {
			return invalidOutput("fact reference is unavailable or duplicated")
		}
		seen[key] = true
	}
	return nil
}

func validateFactKeysForProducts(values []FactKey, products map[catalog.ProductID]ProductSnapshot) error {
	seen := make(map[FactKey]bool, len(values))
	for _, key := range values {
		if !validFactKey(key) || seen[key] {
			return invalidOutput("comparison fact reference is invalid or duplicated")
		}
		for _, product := range products {
			if !factAvailable(key, product) {
				return invalidOutput("comparison fact is not available for every product")
			}
		}
		seen[key] = true
	}
	return nil
}

func selectedProducts(value RecommendationSnapshot) map[catalog.ProductID]DeterministicProduct {
	result := make(map[catalog.ProductID]DeterministicProduct, len(value.Selected))
	for _, item := range value.Selected {
		result[item.Product.ID] = item
	}
	return result
}

func factAvailable(key FactKey, product ProductSnapshot) bool {
	if !validFactKey(key) {
		return false
	}
	if key == FactMaxCapacity {
		return product.MaxCapacityGrams != nil
	}
	if key == FactMaterial {
		return strings.TrimSpace(product.Material) != ""
	}
	return true
}

func validFactKey(value FactKey) bool {
	for _, candidate := range allFactKeys() {
		if value == candidate {
			return true
		}
	}
	return false
}

func containsFactKey(values []FactKey, expected FactKey) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func allFactKeys() []FactKey {
	return []FactKey{FactPrice, FactDimensions, FactWeight, FactMaxCapacity, FactMaterial, FactWarranty,
		FactQuality, FactValue, FactDurability, FactBeginner, FactAdvanced, FactApartment, FactNoise, FactPortability}
}

func validField(value RequirementField) bool {
	for _, candidate := range allFields() {
		if value == candidate {
			return true
		}
	}
	return false
}

func allFields() []RequirementField {
	return []RequirementField{FieldGoal, FieldExperience, FieldBudget, FieldAvailableSpace,
		FieldExistingEquipment, FieldTrainingPreferences, FieldPriorities, FieldFreeText}
}

func promptForField(value RequirementField) ClarifyingPrompt {
	switch value {
	case FieldGoal:
		return PromptPrimaryGoal
	case FieldExperience:
		return PromptExperience
	case FieldBudget:
		return PromptBudget
	case FieldAvailableSpace:
		return PromptAvailableSpace
	case FieldExistingEquipment:
		return PromptExistingEquipment
	case FieldTrainingPreferences:
		return PromptTrainingPreferences
	case FieldPriorities:
		return PromptPriorities
	case FieldFreeText:
		return PromptMoreDetail
	default:
		return ""
	}
}

func invalidOutput(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidProviderOutput, message)
}

func errorsForValidation(message string) error {
	return fmt.Errorf("validation failed: %s", message)
}
