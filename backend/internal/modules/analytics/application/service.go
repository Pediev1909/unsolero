package application

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

var ErrInvalidEvent = errors.New("invalid analytics event")
var ErrConsentRequired = errors.New("valid server-side analytics consent is required")
var ErrInvalidConsent = errors.New("invalid analytics consent")

var eventTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,99}$`)
var eventUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var attributionTokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
var referrerHostPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)

type Config struct {
	AnonymousRetention     time.Duration
	AuthenticatedRetention time.Duration
	ReceiptRetention       time.Duration
}

func DefaultConfig() Config {
	return Config{
		AnonymousRetention:     90 * 24 * time.Hour,
		AuthenticatedRetention: 397 * 24 * time.Hour,
		ReceiptRetention:       30 * 24 * time.Hour,
	}
}

type Service struct {
	repository ports.EventRecorder
	config     Config
	now        func() time.Time
}

func NewService(repository ports.EventRecorder) *Service {
	return NewServiceWithConfig(repository, DefaultConfig())
}

func NewServiceWithConfig(repository ports.EventRecorder, config Config) *Service {
	return &Service{repository: repository, config: config, now: func() time.Time { return time.Now().UTC() }}
}

func (service *Service) RecordClientEvent(ctx context.Context, event domain.Event) (domain.IngestionResult, error) {
	if !eventUUIDPattern.MatchString(string(event.ID)) {
		return domain.IngestionResult{Outcome: domain.IngestionRejected}, ErrInvalidEvent
	}
	if event.Classification == "bot" || event.Classification == "prefetch" {
		if err := service.repository.RecordRejected(ctx, event.ID, receiptEventName(event.Name), domain.IngestionBotFiltered, event.Classification, service.config.ReceiptRetention); err != nil {
			return domain.IngestionResult{}, err
		}
		return domain.IngestionResult{Outcome: domain.IngestionBotFiltered}, nil
	}
	if event.UserID == nil && len(event.AnonymousSubjectHash) == 0 {
		if err := service.repository.RecordRejected(ctx, event.ID, receiptEventName(event.Name), domain.IngestionRejected, "consent_required", service.config.ReceiptRetention); err != nil {
			return domain.IngestionResult{}, err
		}
		return domain.IngestionResult{Outcome: domain.IngestionRejected}, ErrConsentRequired
	}
	if outcome, reason := validateClientEvent(event); outcome != domain.IngestionAccepted {
		if err := service.repository.RecordRejected(ctx, event.ID, receiptEventName(event.Name), outcome, reason, service.config.ReceiptRetention); err != nil {
			return domain.IngestionResult{}, err
		}
		return domain.IngestionResult{Outcome: outcome}, ErrInvalidEvent
	}
	now := service.now()
	event.SchemaVersion = 3
	event.Origin = "client"
	event.ConsentState = "granted"
	event.Reportable = true
	event.OccurredAt = now
	event.PagePath = normalizedOptional(event.PagePath)
	event.TrafficSource = normalizedOptional(event.TrafficSource)
	event.TrafficMedium = normalizedOptional(event.TrafficMedium)
	event.Campaign = normalizedOptional(event.Campaign)
	event.ReferrerHost = normalizedLowerOptional(event.ReferrerHost)
	if event.UserID != nil {
		event.RetentionExpiresAt = now.Add(service.config.AuthenticatedRetention)
	} else {
		event.RetentionExpiresAt = now.Add(service.config.AnonymousRetention)
	}
	result, err := service.repository.Ingest(ctx, event, service.config.ReceiptRetention)
	if err != nil {
		return domain.IngestionResult{}, err
	}
	if result.Outcome == domain.IngestionRejected {
		return result, ErrConsentRequired
	}
	return result, nil
}

func receiptEventName(value string) string {
	if eventTokenPattern.MatchString(value) {
		return value
	}
	return ""
}

func (service *Service) SetConsent(ctx context.Context, decision domain.ConsentDecision) (domain.Consent, error) {
	if (decision.Subject.UserID == nil) == (len(decision.Subject.AnonymousSubjectHash) == 0) ||
		(decision.RequestedState != "granted" && decision.RequestedState != "denied") ||
		decision.PolicyVersion != domain.CurrentConsentPolicyVersion ||
		(decision.Source != "banner" && decision.Source != "preferences" && decision.Source != "account_sync") {
		return domain.Consent{}, ErrInvalidConsent
	}
	decision.DecidedAt = service.now()
	return service.repository.SetConsent(ctx, decision)
}

func (service *Service) GetConsent(ctx context.Context, subject domain.Subject) (domain.Consent, error) {
	return service.repository.GetConsent(ctx, subject)
}

func (service *Service) ClaimIdentity(ctx context.Context, anonymousSubjectHash []byte, userID string) error {
	if len(anonymousSubjectHash) != 32 || !eventUUIDPattern.MatchString(userID) {
		return ports.ErrIdentityClaimNotAllowed
	}
	return service.repository.ClaimIdentity(ctx, anonymousSubjectHash, userID, domain.CurrentConsentPolicyVersion, service.now())
}

func (service *Service) Cleanup(ctx context.Context, batch int) (domain.CleanupResult, error) {
	if batch < 1 || batch > 10_000 {
		return domain.CleanupResult{}, errors.New("analytics cleanup batch must be between 1 and 10000")
	}
	return service.repository.Cleanup(ctx, service.now(), batch)
}

func validateClientEvent(event domain.Event) (domain.IngestionOutcome, string) {
	if event.ConsentPolicyVersion != domain.CurrentConsentPolicyVersion || !eventTokenPattern.MatchString(event.Surface) ||
		event.SessionID == nil || !eventUUIDPattern.MatchString(*event.SessionID) ||
		event.UserID != nil && len(event.AnonymousSubjectHash) != 0 {
		return domain.IngestionRejected, "invalid_envelope"
	}
	if !validContext(event) {
		return domain.IngestionPrivacyFiltered, "unsafe_context"
	}
	validators, ok := clientEventValidators[event.Name]
	if !ok {
		return domain.IngestionRejected, "unsupported_event"
	}
	if len(event.Properties) != len(validators) {
		return domain.IngestionPrivacyFiltered, "property_schema_mismatch"
	}
	for key, validator := range validators {
		value, exists := event.Properties[key]
		if !exists || !validator(value) {
			return domain.IngestionPrivacyFiltered, "invalid_property"
		}
	}
	return domain.IngestionAccepted, "accepted"
}

type propertyValidator func(json.RawMessage) bool

var clientEventValidators = map[string]map[string]propertyValidator{
	domain.EventPageView:                {},
	domain.EventOnboardingStarted:       {"onboarding_id": uuidProperty},
	domain.EventOnboardingCompleted:     {"onboarding_id": uuidProperty, "outcome": enumProperty("complete", "no_suitable_products")},
	domain.EventProductViewed:           {"product_id": uuidProperty},
	domain.EventProductSaved:            {"product_id": uuidProperty, "persistence": enumProperty("account", "browser")},
	domain.EventRecommendationGenerated: {"status": enumProperty("complete", "no_suitable_products"), "persistence": enumProperty("account", "browser")},
	domain.EventComparisonCreated:       {"product_count": integerRangeProperty(2, 4), "persistence": enumProperty("account", "browser")},
	domain.EventSetupSaved:              {"setup_id": uuidProperty, "persistence": enumProperty("account", "browser")},
}

func validContext(event domain.Event) bool {
	if event.PagePath != nil {
		path := strings.TrimSpace(*event.PagePath)
		if len(path) < 1 || len(path) > 500 || !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "?#") {
			return false
		}
	}
	for _, value := range []*string{event.TrafficSource, event.TrafficMedium, event.Campaign} {
		if value != nil && !attributionTokenPattern.MatchString(strings.TrimSpace(*value)) {
			return false
		}
	}
	if event.ReferrerHost != nil && !referrerHostPattern.MatchString(strings.ToLower(strings.TrimSpace(*event.ReferrerHost))) {
		return false
	}
	return true
}

func normalizedOptional(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizedLowerOptional(value *string) *string {
	value = normalizedOptional(value)
	if value == nil {
		return nil
	}
	normalized := strings.ToLower(*value)
	return &normalized
}

func uuidProperty(raw json.RawMessage) bool {
	var value string
	return json.Unmarshal(raw, &value) == nil && eventUUIDPattern.MatchString(value)
}

func enumProperty(values ...string) propertyValidator {
	return func(raw json.RawMessage) bool {
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return false
		}
		for _, candidate := range values {
			if value == candidate {
				return true
			}
		}
		return false
	}
}

func integerRangeProperty(minimum, maximum int) propertyValidator {
	return func(raw json.RawMessage) bool {
		var value int
		return json.Unmarshal(raw, &value) == nil && value >= minimum && value <= maximum
	}
}
