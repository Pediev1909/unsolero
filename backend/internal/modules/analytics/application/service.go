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

var eventTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,99}$`)
var eventUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var attributionTokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
var referrerHostPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)

type Service struct{ recorder ports.EventRecorder }

func NewService(recorder ports.EventRecorder) *Service { return &Service{recorder: recorder} }

func (service *Service) RecordClientEvent(ctx context.Context, event domain.Event) error {
	if err := validateClientEvent(event); err != nil {
		return err
	}
	event.SchemaVersion = 2
	event.OccurredAt = time.Now().UTC()
	event.PagePath = normalizedOptional(event.PagePath)
	event.TrafficSource = normalizedOptional(event.TrafficSource)
	event.TrafficMedium = normalizedOptional(event.TrafficMedium)
	event.Campaign = normalizedOptional(event.Campaign)
	event.ReferrerHost = normalizedLowerOptional(event.ReferrerHost)
	return service.recorder.Record(ctx, event)
}

func validateClientEvent(event domain.Event) error {
	if event.ConsentState != "granted" || !eventTokenPattern.MatchString(event.Surface) || event.SessionID == nil ||
		!eventUUIDPattern.MatchString(*event.SessionID) || event.UserID == nil && event.AnonymousID == nil {
		return ErrInvalidEvent
	}
	if !validContext(event) {
		return ErrInvalidEvent
	}
	validators, ok := clientEventValidators[event.Name]
	if !ok || len(event.Properties) != len(validators) {
		return ErrInvalidEvent
	}
	for key, validator := range validators {
		value, exists := event.Properties[key]
		if !exists || !validator(value) {
			return ErrInvalidEvent
		}
	}
	return nil
}

type propertyValidator func(json.RawMessage) bool

var clientEventValidators = map[string]map[string]propertyValidator{
	domain.EventPageView: {},
	domain.EventOnboardingStarted: {
		"onboarding_id": uuidProperty,
	},
	domain.EventOnboardingCompleted: {
		"onboarding_id": uuidProperty,
		"outcome":       enumProperty("complete", "no_suitable_products"),
	},
	domain.EventProductViewed: {
		"product_id": uuidProperty,
	},
	domain.EventProductSaved: {
		"product_id":  uuidProperty,
		"persistence": enumProperty("account", "browser"),
	},
	domain.EventRecommendationGenerated: {
		"status":      enumProperty("complete", "no_suitable_products"),
		"persistence": enumProperty("account", "browser"),
	},
	domain.EventComparisonCreated: {
		"product_count": integerRangeProperty(2, 4),
		"persistence":   enumProperty("account", "browser"),
	},
	domain.EventSetupSaved: {
		"setup_id":    uuidProperty,
		"persistence": enumProperty("account", "browser"),
	},
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
			if strings.EqualFold(value, candidate) {
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
