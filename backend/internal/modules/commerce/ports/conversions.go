package ports

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/commerce/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var (
	ErrWebhookRejected    = errors.New("conversion webhook rejected")
	ErrWebhookReplay      = errors.New("conversion webhook replayed")
	ErrConversionConflict = errors.New("conversion event conflicts with stored fact")
	ErrConversionNotFound = errors.New("conversion not found")
)

type ConversionProviderAdapter interface {
	Key() string
	ValidateConversionConfiguration(context.Context, domain.ProviderConfiguration) error
	VerifyWebhook(context.Context, domain.ProviderConfiguration, domain.WebhookRequest) (domain.VerifiedWebhook, error)
	FetchConversions(context.Context, domain.ProviderConfiguration, *string) (domain.ConversionBatch, error)
}

type ConversionRepository interface {
	GetProviderConfiguration(context.Context, domain.ProviderConfigurationID) (domain.ProviderConfiguration, error)
	SetConversionProviderEnabled(context.Context, identity.UserID, domain.ProviderConfigurationID, bool, time.Time) (domain.ProviderConfiguration, error)
	RecordWebhookDelivery(context.Context, domain.ProviderConfigurationID, string, string, string, *time.Time, *string, time.Time) (domain.WebhookDelivery, error)
	ApplyWebhookEvents(context.Context, domain.WebhookDeliveryID, domain.ProviderConfiguration, []domain.VerifiedConversionEvent, time.Time) (int, error)
	ResolveConversionAttribution(context.Context, domain.ProviderConfiguration, *string, time.Time, time.Duration) (domain.ConversionAttribution, error)

	QueueConversionImport(context.Context, *identity.UserID, domain.ProviderConfigurationID, domain.ImportTrigger, string, int16) (domain.ConversionImportRun, error)
	QueueDueConversionImports(context.Context, time.Time, int) (int, error)
	RecoverStalledConversionImports(context.Context, time.Time, time.Time, int) (int, error)
	ClaimNextConversionImport(context.Context, time.Time) (domain.ConversionImportRun, error)
	ApplyConversionImport(context.Context, domain.ConversionImportRun, []domain.VerifiedConversionEvent, []domain.ImportRecordFailure, domain.ConversionBatch, time.Time) (int, error)
	CompleteConversionImport(context.Context, domain.ConversionImportRun, domain.ConversionBatch, int, int, time.Time) error
	FailConversionImport(context.Context, domain.ConversionImportRun, string, string, time.Time) error
	RetryConversionImport(context.Context, identity.UserID, domain.ConversionImportRunID, string) (domain.ConversionImportRun, error)

	ListConversions(context.Context, domain.ConversionFilter) ([]domain.Conversion, int64, error)
	ListConversionImports(context.Context, int, int) ([]domain.ConversionImportRun, int64, error)
	ListReconciliations(context.Context, int, int) ([]domain.ReconciliationRun, int64, error)
	ReconcileConversionImport(context.Context, *identity.UserID, domain.ConversionImportRunID, string, time.Time) (domain.ReconciliationRun, error)
	MonetizationReport(context.Context, time.Time, time.Time) (domain.MonetizationReport, error)
}
