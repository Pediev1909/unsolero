package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

type ProviderConfigurationID string
type ImportRunID string

type ProviderLifecycle string

const (
	ProviderDisabled   ProviderLifecycle = "disabled"
	ProviderConfigured ProviderLifecycle = "configured"
	ProviderActive     ProviderLifecycle = "active"
	ProviderDegraded   ProviderLifecycle = "degraded"
	ProviderSuspended  ProviderLifecycle = "suspended"
)

type ProviderConfiguration struct {
	ID                      ProviderConfigurationID `json:"id"`
	MerchantID              MerchantID              `json:"merchant_id"`
	MerchantName            string                  `json:"merchant_name"`
	ProviderKey             string                  `json:"provider_key"`
	AdapterKey              string                  `json:"adapter_key"`
	ExternalMerchantID      string                  `json:"external_merchant_id"`
	CredentialReference     *string                 `json:"credential_reference"`
	LifecycleStatus         ProviderLifecycle       `json:"lifecycle_status"`
	ConfigurationVerifiedAt *time.Time              `json:"configuration_verified_at"`
	ScheduleIntervalMinutes int                     `json:"schedule_interval_minutes"`
	FreshnessTTLMinutes     int                     `json:"freshness_ttl_minutes"`
	Cursor                  *string                 `json:"cursor"`
	NextImportAt            *time.Time              `json:"next_import_at"`
	LastImportStartedAt     *time.Time              `json:"last_import_started_at"`
	LastImportSucceededAt   *time.Time              `json:"last_import_succeeded_at"`
	LastImportFailedAt      *time.Time              `json:"last_import_failed_at"`
	ConsecutiveFailures     int                     `json:"consecutive_failures"`
	LastErrorCode           *string                 `json:"last_error_code"`
	ConversionCursor        *string                 `json:"conversion_cursor"`
	NextConversionImportAt  *time.Time              `json:"next_conversion_import_at"`
	LastConversionSucceeded *time.Time              `json:"last_conversion_import_succeeded_at"`
	LastConversionFailed    *time.Time              `json:"last_conversion_import_failed_at"`
	ConversionFailures      int                     `json:"conversion_consecutive_failures"`
	LastConversionError     *string                 `json:"last_conversion_error_code"`
	ConversionEnabled       bool                    `json:"conversion_ingestion_enabled"`
	ConversionVerifiedAt    *time.Time              `json:"conversion_configuration_verified_at"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               time.Time               `json:"updated_at"`
}

type ProviderConfigurationInput struct {
	MerchantID              MerchantID
	ProviderKey             string
	AdapterKey              string
	ExternalMerchantID      string
	CredentialReference     *string
	ScheduleIntervalMinutes int
	FreshnessTTLMinutes     int
}

type ImportTrigger string

const (
	ImportScheduled ImportTrigger = "scheduled"
	ImportManual    ImportTrigger = "manual"
	ImportRetry     ImportTrigger = "retry"
)

type ImportStatus string

const (
	ImportQueued    ImportStatus = "queued"
	ImportRunning   ImportStatus = "running"
	ImportRetryWait ImportStatus = "retry_wait"
	ImportSucceeded ImportStatus = "succeeded"
	ImportPartial   ImportStatus = "partial"
	ImportFailed    ImportStatus = "failed"
	ImportCancelled ImportStatus = "cancelled"
)

type ImportRun struct {
	ID                    ImportRunID           `json:"id"`
	ProviderConfiguration ProviderConfiguration `json:"provider_configuration"`
	Trigger               ImportTrigger         `json:"trigger"`
	Status                ImportStatus          `json:"status"`
	IdempotencyKey        string                `json:"idempotency_key"`
	RequestedBy           *string               `json:"requested_by"`
	CursorBefore          *string               `json:"cursor_before"`
	CursorAfter           *string               `json:"cursor_after"`
	AttemptCount          int16                 `json:"attempt_count"`
	MaxAttempts           int16                 `json:"max_attempts"`
	RecordsReceived       int                   `json:"records_received"`
	RecordsApplied        int                   `json:"records_applied"`
	RecordsRejected       int                   `json:"records_rejected"`
	OffersDeactivated     int                   `json:"offers_deactivated"`
	ErrorCode             *string               `json:"error_code"`
	ErrorMessage          *string               `json:"error_message"`
	NextRetryAt           *time.Time            `json:"next_retry_at"`
	StartedAt             *time.Time            `json:"started_at"`
	CompletedAt           *time.Time            `json:"completed_at"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
}

type ImportFailure struct {
	ID                string    `json:"id"`
	ImportRunID       string    `json:"import_run_id"`
	ExternalRecordID  *string   `json:"external_record_id"`
	ErrorCode         string    `json:"error_code"`
	ErrorMessage      string    `json:"error_message"`
	RecordFingerprint *string   `json:"record_fingerprint"`
	CreatedAt         time.Time `json:"created_at"`
}

type ProviderOffer struct {
	ExternalOfferID    string
	ProductID          catalog.ProductID
	MerchantSKU        string
	ProductURL         string
	AffiliateURL       *string
	PriceMinor         int64
	ShippingMinor      int64
	Currency           string
	Availability       string
	Condition          string
	ProviderObservedAt *time.Time
	ExpiresAt          *time.Time
}

type ProviderBatch struct {
	Offers     []ProviderOffer
	NextCursor *string
	Complete   bool
	RateLimit  *RateLimitState
}

type RateLimitState struct {
	Remaining int
	ResetAt   time.Time
}

type ValidatedOffer struct {
	ProviderOffer
	ObservedAt              time.Time
	ExpiresAt               time.Time
	PriceFingerprint        string
	AvailabilityFingerprint string
}

type ImportApplyResult struct {
	Received          int
	Applied           int
	Rejected          int
	OffersDeactivated int
	Failures          []ImportRecordFailure
	NextCursor        *string
	Complete          bool
}

type ImportRecordFailure struct {
	ExternalRecordID  *string
	Code              string
	Message           string
	RecordFingerprint *string
}

var providerKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)

func (input *ProviderConfigurationInput) NormalizeAndValidate() error {
	input.ProviderKey = strings.ToLower(strings.TrimSpace(input.ProviderKey))
	input.AdapterKey = strings.ToLower(strings.TrimSpace(input.AdapterKey))
	input.ExternalMerchantID = strings.TrimSpace(input.ExternalMerchantID)
	if input.CredentialReference != nil {
		value := strings.TrimSpace(*input.CredentialReference)
		if value == "" {
			input.CredentialReference = nil
		} else {
			input.CredentialReference = &value
		}
	}
	if input.MerchantID == "" || !providerKeyPattern.MatchString(input.ProviderKey) ||
		!providerKeyPattern.MatchString(input.AdapterKey) || len(input.ExternalMerchantID) < 1 ||
		len(input.ExternalMerchantID) > 200 || input.ScheduleIntervalMinutes < 5 ||
		input.ScheduleIntervalMinutes > 10080 || input.FreshnessTTLMinutes < 60 ||
		input.FreshnessTTLMinutes > 43200 {
		return errors.New("invalid provider configuration")
	}
	if input.CredentialReference != nil && len(*input.CredentialReference) > 200 {
		return errors.New("invalid credential reference")
	}
	return nil
}

func ValidateProviderOffer(record ProviderOffer, observedAt time.Time, freshnessTTL time.Duration) (ValidatedOffer, error) {
	record.ExternalOfferID = strings.TrimSpace(record.ExternalOfferID)
	record.MerchantSKU = strings.TrimSpace(record.MerchantSKU)
	record.Currency = strings.ToUpper(strings.TrimSpace(record.Currency))
	record.Availability = strings.ToLower(strings.TrimSpace(record.Availability))
	record.Condition = strings.ToLower(strings.TrimSpace(record.Condition))
	if len(record.ExternalOfferID) < 1 || len(record.ExternalOfferID) > 200 || record.ProductID == "" ||
		len(record.MerchantSKU) < 1 || len(record.MerchantSKU) > 120 || record.PriceMinor < 0 ||
		record.ShippingMinor < 0 || !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(record.Currency) ||
		freshnessTTL < time.Hour || freshnessTTL > 30*24*time.Hour {
		return ValidatedOffer{}, errors.New("malformed provider offer")
	}
	if !stringIn(record.Availability, "in_stock", "backorder", "out_of_stock", "discontinued") ||
		!stringIn(record.Condition, "new", "refurbished", "used") || !SafeMerchantURL(record.ProductURL) {
		return ValidatedOffer{}, errors.New("malformed provider offer")
	}
	if record.AffiliateURL != nil && !SafeMerchantURL(*record.AffiliateURL) {
		return ValidatedOffer{}, errors.New("unsafe affiliate destination")
	}
	effectiveObservedAt := observedAt.UTC()
	if record.ProviderObservedAt != nil {
		providerObserved := record.ProviderObservedAt.UTC()
		if providerObserved.After(observedAt.Add(5*time.Minute)) || providerObserved.Before(observedAt.Add(-30*24*time.Hour)) {
			return ValidatedOffer{}, errors.New("provider timestamp is outside the accepted window")
		}
		effectiveObservedAt = providerObserved
	}
	expiresAt := effectiveObservedAt.Add(freshnessTTL)
	if record.ExpiresAt != nil {
		expiresAt = record.ExpiresAt.UTC()
		if !expiresAt.After(effectiveObservedAt) || expiresAt.After(effectiveObservedAt.Add(30*24*time.Hour)) {
			return ValidatedOffer{}, errors.New("provider expiry is invalid")
		}
	}
	price := fingerprint(record.ExternalOfferID, fmt.Sprint(record.PriceMinor), fmt.Sprint(record.ShippingMinor), record.Currency, effectiveObservedAt.Format(time.RFC3339Nano))
	availability := fingerprint(record.ExternalOfferID, record.Availability, effectiveObservedAt.Format(time.RFC3339Nano))
	return ValidatedOffer{ProviderOffer: record, ObservedAt: effectiveObservedAt, ExpiresAt: expiresAt,
		PriceFingerprint: price, AvailabilityFingerprint: availability}, nil
}

func SafeMerchantURL(raw string) bool {
	if len(raw) < 9 || len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\x00") {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".invalid") {
		return false
	}
	if address, err := netip.ParseAddr(host); err == nil {
		return address.IsGlobalUnicast() && !address.IsPrivate() && !address.IsLoopback() && !address.IsLinkLocalUnicast()
	}
	if strings.Contains(host, ":") || net.ParseIP(host) != nil || !strings.Contains(host, ".") {
		return false
	}
	return true
}

func NextImportRetry(attempt, maximum int16, failedAt time.Time) (*time.Time, bool) {
	if attempt < 1 || maximum < 1 || attempt >= maximum {
		return nil, false
	}
	delay := time.Duration(1<<max(0, int(attempt)-1)) * time.Minute
	if delay > 30*time.Minute {
		delay = 30 * time.Minute
	}
	next := failedAt.Add(delay)
	return &next, true
}

func fingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:])
}

func stringIn(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
