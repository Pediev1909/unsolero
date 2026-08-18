package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ConversionID string
type ConversionEventID string
type ConversionImportRunID string
type WebhookDeliveryID string
type ReconciliationRunID string

const MaximumVerifiedMoneyMinor int64 = 1_000_000_000_000_000

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderConfirmed OrderStatus = "confirmed"
	OrderCancelled OrderStatus = "cancelled"
	OrderReversed  OrderStatus = "reversed"
	OrderRejected  OrderStatus = "rejected"
)

type CommissionStatus string

const (
	CommissionPending  CommissionStatus = "pending"
	CommissionApproved CommissionStatus = "approved"
	CommissionReversed CommissionStatus = "reversed"
	CommissionRejected CommissionStatus = "rejected"
	CommissionPaid     CommissionStatus = "paid"
)

type ConversionEventType string

const (
	EventConversionCreated ConversionEventType = "conversion_created"
	EventOrderChanged      ConversionEventType = "order_status_changed"
	EventCommissionChanged ConversionEventType = "commission_status_changed"
	EventCancelled         ConversionEventType = "cancelled"
	EventReversed          ConversionEventType = "reversed"
	EventCorrection        ConversionEventType = "correction"
)

type ProviderConversionEvent struct {
	ProviderEventID      string
	EventType            ConversionEventType
	ExternalConversionID string
	OrderReference       *string
	OrderStatus          OrderStatus
	OrderValueMinor      *int64
	OrderCurrency        *string
	CommissionMinor      *int64
	CommissionCurrency   *string
	CommissionStatus     *CommissionStatus
	ClickID              *string
	RawProviderReference *string
	EventTimestamp       time.Time
}

type VerifiedConversionEvent struct {
	ProviderConversionEvent
	ProviderConfigurationID ProviderConfigurationID
	Provider                string
	MerchantID              MerchantID
	WebhookDeliveryID       *WebhookDeliveryID
	ImportRunID             *ConversionImportRunID
	ReceivedAt              time.Time
	PayloadFingerprint      string
	Attribution             ConversionAttribution
}

type ConversionAttribution struct {
	Status               string
	ClickID              *string
	RecommendationID     *string
	RecommendationItemID *string
	Source               *string
	Campaign             *string
}

type WebhookRequest struct {
	Headers    map[string][]string
	Body       []byte
	ReceivedAt time.Time
}

type VerifiedWebhook struct {
	SignatureTimestamp time.Time
	Events             []ProviderConversionEvent
}

type WebhookDelivery struct {
	ID                WebhookDeliveryID
	VerificationState string
	Processed         bool
}

type ConversionBatch struct {
	Events        []ProviderConversionEvent
	NextCursor    *string
	Complete      bool
	CoverageStart *time.Time
	CoverageEnd   *time.Time
	RateLimit     *RateLimitState
}

type Conversion struct {
	ID                      ConversionID      `json:"id"`
	ProviderConfigurationID string            `json:"provider_configuration_id"`
	Provider                string            `json:"provider"`
	MerchantID              string            `json:"merchant_id"`
	MerchantName            string            `json:"merchant_name"`
	ExternalConversionID    string            `json:"external_conversion_id"`
	OrderReference          *string           `json:"order_reference"`
	OrderStatus             OrderStatus       `json:"order_status"`
	OrderValueMinor         *int64            `json:"order_value_minor"`
	OrderCurrency           *string           `json:"order_currency"`
	CommissionMinor         *int64            `json:"commission_amount_minor"`
	CommissionCurrency      *string           `json:"commission_currency"`
	CommissionStatus        *CommissionStatus `json:"commission_status"`
	AttributionStatus       string            `json:"attribution_status"`
	ClickID                 *string           `json:"click_id"`
	RecommendationID        *string           `json:"recommendation_id"`
	Source                  *string           `json:"source"`
	Campaign                *string           `json:"campaign"`
	VerificationState       string            `json:"verification_state"`
	EventTimestamp          time.Time         `json:"event_timestamp"`
	ReceivedAt              time.Time         `json:"received_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
	ReconciliationStatus    *string           `json:"reconciliation_status"`
}

type ConversionFilter struct {
	Provider             string
	OrderStatus          OrderStatus
	CommissionStatus     CommissionStatus
	AttributionStatus    string
	ReconciliationStatus string
	Currency             string
	Limit                int
	Offset               int
}

type ConversionImportRun struct {
	ID                    ConversionImportRunID `json:"id"`
	ProviderConfiguration ProviderConfiguration `json:"provider_configuration"`
	Trigger               ImportTrigger         `json:"trigger"`
	Status                ImportStatus          `json:"status"`
	AttemptCount          int16                 `json:"attempt_count"`
	MaxAttempts           int16                 `json:"max_attempts"`
	RecordsReceived       int                   `json:"records_received"`
	RecordsApplied        int                   `json:"records_applied"`
	RecordsRejected       int                   `json:"records_rejected"`
	CursorBefore          *string               `json:"cursor_before"`
	CursorAfter           *string               `json:"cursor_after"`
	CoverageStart         *time.Time            `json:"coverage_start"`
	CoverageEnd           *time.Time            `json:"coverage_end"`
	ErrorCode             *string               `json:"error_code"`
	ErrorMessage          *string               `json:"error_message"`
	CreatedAt             time.Time             `json:"created_at"`
	StartedAt             *time.Time            `json:"started_at"`
	CompletedAt           *time.Time            `json:"completed_at"`
}

type ReconciliationResult string

const (
	ReconcileMatched     ReconciliationResult = "matched"
	ReconcileMissing     ReconciliationResult = "missing"
	ReconcileConflicting ReconciliationResult = "conflicting"
	ReconcileStale       ReconciliationResult = "stale"
	ReconcileUnresolved  ReconciliationResult = "unresolved"
)

type ReconciliationItem struct {
	ConversionID          *ConversionID
	ProviderEventID       *string
	Result                ReconciliationResult
	ReasonCode            string
	ComparisonFingerprint *string
}

type ReconciliationRun struct {
	ID                    ReconciliationRunID   `json:"id"`
	ProviderConfiguration ProviderConfiguration `json:"provider_configuration"`
	Status                string                `json:"status"`
	CoverageStart         time.Time             `json:"coverage_start"`
	CoverageEnd           time.Time             `json:"coverage_end"`
	Matched               int                   `json:"matched"`
	Missing               int                   `json:"missing"`
	Conflicting           int                   `json:"conflicting"`
	Stale                 int                   `json:"stale"`
	Unresolved            int                   `json:"unresolved"`
	ErrorCode             *string               `json:"error_code"`
	StartedAt             time.Time             `json:"started_at"`
	CompletedAt           *time.Time            `json:"completed_at"`
}

type MetricStatus string

const (
	MetricAvailable    MetricStatus = "available"
	MetricNoData       MetricStatus = "no_data"
	MetricInsufficient MetricStatus = "insufficient_data"
)

type RatioMetric struct {
	Status      MetricStatus `json:"status"`
	Value       *float64     `json:"value"`
	Numerator   int64        `json:"numerator"`
	Denominator int64        `json:"denominator"`
	Definition  string       `json:"definition"`
}

type CurrencyMetric struct {
	Currency    string   `json:"currency"`
	AmountMinor int64    `json:"amount_minor"`
	Denominator int64    `json:"denominator"`
	ValueMinor  *float64 `json:"value_minor"`
}

type CurrencyMetricGroup struct {
	Status     MetricStatus     `json:"status"`
	Values     []CurrencyMetric `json:"values"`
	Definition string           `json:"definition"`
}

type MonetizationReport struct {
	WindowStart              time.Time           `json:"window_start"`
	WindowEnd                time.Time           `json:"window_end"`
	FreshThrough             *time.Time          `json:"fresh_through"`
	ConversionRate           RatioMetric         `json:"affiliate_conversion_rate"`
	EarningsPerClick         CurrencyMetricGroup `json:"earnings_per_click"`
	RevenuePerVisitor        CurrencyMetricGroup `json:"revenue_per_visitor"`
	RevenuePerRecommendation CurrencyMetricGroup `json:"revenue_per_recommendation"`
	Commission               CurrencyMetricGroup `json:"commission"`
	ReversalRate             RatioMetric         `json:"reversal_rate"`
	RepeatUserRate           RatioMetric         `json:"repeat_user_rate"`
	CurrencyPolicy           string              `json:"currency_policy"`
}

func ValidateProviderConversionEvent(event ProviderConversionEvent, receivedAt time.Time) (ProviderConversionEvent, error) {
	event.ProviderEventID = strings.TrimSpace(event.ProviderEventID)
	event.ExternalConversionID = strings.TrimSpace(event.ExternalConversionID)
	if event.OrderReference != nil {
		value := strings.TrimSpace(*event.OrderReference)
		event.OrderReference = optionalString(value)
	}
	if event.OrderCurrency != nil {
		value := strings.ToUpper(strings.TrimSpace(*event.OrderCurrency))
		event.OrderCurrency = &value
	}
	if event.CommissionCurrency != nil {
		value := strings.ToUpper(strings.TrimSpace(*event.CommissionCurrency))
		event.CommissionCurrency = &value
	}
	if event.ClickID != nil {
		value := strings.TrimSpace(*event.ClickID)
		event.ClickID = optionalString(value)
	}
	if event.RawProviderReference != nil {
		value := strings.TrimSpace(*event.RawProviderReference)
		event.RawProviderReference = optionalString(value)
	}
	if len(event.ProviderEventID) < 1 || len(event.ProviderEventID) > 200 ||
		len(event.ExternalConversionID) < 1 || len(event.ExternalConversionID) > 200 ||
		!validConversionEventType(event.EventType) || !validOrderStatus(event.OrderStatus) ||
		event.EventTimestamp.IsZero() || event.EventTimestamp.After(receivedAt.Add(5*time.Minute)) ||
		event.EventTimestamp.Before(receivedAt.AddDate(-2, 0, 0)) {
		return ProviderConversionEvent{}, errors.New("invalid conversion event")
	}
	if (event.OrderValueMinor == nil) != (event.OrderCurrency == nil) ||
		(event.CommissionMinor == nil) != (event.CommissionCurrency == nil) ||
		(event.OrderValueMinor != nil && (*event.OrderValueMinor < 0 || *event.OrderValueMinor > MaximumVerifiedMoneyMinor)) ||
		(event.CommissionMinor != nil && (*event.CommissionMinor < 0 || *event.CommissionMinor > MaximumVerifiedMoneyMinor)) ||
		(event.OrderCurrency != nil && !validCurrency(*event.OrderCurrency)) ||
		(event.CommissionCurrency != nil && !validCurrency(*event.CommissionCurrency)) ||
		(event.CommissionStatus != nil && !validCommissionStatus(*event.CommissionStatus)) {
		return ProviderConversionEvent{}, errors.New("invalid conversion money or status")
	}
	if event.OrderReference != nil && len(*event.OrderReference) > 200 ||
		event.RawProviderReference != nil && len(*event.RawProviderReference) > 200 {
		return ProviderConversionEvent{}, errors.New("conversion reference is too long")
	}
	return event, nil
}

func ValidateOrderTransition(previous, next OrderStatus, correction bool) bool {
	if previous == next || correction {
		return validOrderStatus(next)
	}
	switch previous {
	case OrderPending:
		return next == OrderConfirmed || next == OrderCancelled || next == OrderRejected
	case OrderConfirmed:
		return next == OrderCancelled || next == OrderReversed
	default:
		return false
	}
}

func ValidateCommissionTransition(previous, next CommissionStatus, correction bool) bool {
	if previous == next || correction {
		return validCommissionStatus(next)
	}
	switch previous {
	case CommissionPending:
		return next == CommissionApproved || next == CommissionRejected || next == CommissionReversed
	case CommissionApproved:
		return next == CommissionPaid || next == CommissionReversed || next == CommissionRejected
	case CommissionPaid:
		return next == CommissionReversed
	default:
		return false
	}
}

func ConversionEventFingerprint(event ProviderConversionEvent) string {
	parts := []string{event.ProviderEventID, string(event.EventType), event.ExternalConversionID,
		string(event.OrderStatus), event.EventTimestamp.UTC().Format(time.RFC3339Nano)}
	parts = append(parts, pointerValue(event.OrderReference), moneyValue(event.OrderValueMinor, event.OrderCurrency),
		moneyValue(event.CommissionMinor, event.CommissionCurrency), commissionValue(event.CommissionStatus),
		pointerValue(event.ClickID), pointerValue(event.RawProviderReference))
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:])
}

func RequestFingerprint(configurationID ProviderConfigurationID, body []byte, signatureTimestamp time.Time) string {
	bodyHash := sha256.Sum256(body)
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x1f%x\x1f%s", configurationID, bodyHash,
		signatureTimestamp.UTC().Format(time.RFC3339Nano))))
	return hex.EncodeToString(sum[:])
}

func BodyFingerprint(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func validConversionEventType(value ConversionEventType) bool {
	return value == EventConversionCreated || value == EventOrderChanged || value == EventCommissionChanged ||
		value == EventCancelled || value == EventReversed || value == EventCorrection
}

func validOrderStatus(value OrderStatus) bool {
	return value == OrderPending || value == OrderConfirmed || value == OrderCancelled ||
		value == OrderReversed || value == OrderRejected
}

func validCommissionStatus(value CommissionStatus) bool {
	return value == CommissionPending || value == CommissionApproved || value == CommissionReversed ||
		value == CommissionRejected || value == CommissionPaid
}

func validCurrency(value string) bool {
	return len(value) == 3 && value == strings.ToUpper(value) && regexp.MustCompile(`^[A-Z]{3}$`).MatchString(value)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func moneyValue(amount *int64, currency *string) string {
	if amount == nil || currency == nil {
		return ""
	}
	return fmt.Sprintf("%d:%s", *amount, *currency)
}

func commissionValue(value *CommissionStatus) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
