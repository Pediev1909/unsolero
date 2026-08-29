package ports

import (
	"context"
	"errors"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/commerce/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var (
	ErrAffiliateDestinationNotFound = errors.New("affiliate destination not found")
	ErrProviderDisabled             = errors.New("commerce provider is disabled")
	ErrProviderUnavailable          = errors.New("commerce provider is unavailable")
	ErrImportNotFound               = errors.New("commerce import not found")
	ErrImportConflict               = errors.New("commerce import conflicts with current state")
)

type MerchantRepository interface {
	ListActive(context.Context) ([]domain.Merchant, error)
}

type OfferRepository interface {
	ListAvailableByProduct(context.Context, catalog.ProductID, string) ([]domain.Offer, error)
	// ListPurchasableByProducts answers a whole grid in one query. The result
	// holds an entry only for products that have a servable affiliate offer,
	// under exactly the conditions the redirect will later re-check — a card
	// that offers a button the redirect would refuse is worse than a card with
	// no button.
	ListPurchasableByProducts(context.Context, []catalog.ProductID) (map[catalog.ProductID]domain.PurchasableOffer, error)
}

type AffiliateRedirectRepository interface {
	ResolveOfferDestination(context.Context, domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error)
	ResolveLegacyDestination(context.Context, domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error)
	ResolvePromotionDestination(context.Context, domain.AffiliateClick) (domain.ResolvedPromotionDestination, error)
	RecordClick(context.Context, domain.ResolvedAffiliateDestination, domain.AffiliateClick) error
	RecordPromotionClick(context.Context, domain.ResolvedPromotionDestination, domain.AffiliateClick) error
}

type ProviderAdapter interface {
	Key() string
	ValidateConfiguration(context.Context, domain.ProviderConfiguration) error
	FetchOffers(context.Context, domain.ProviderConfiguration, *string) (domain.ProviderBatch, error)
}

type ImportRepository interface {
	CreateProviderConfiguration(context.Context, identity.UserID, domain.ProviderConfigurationInput) (domain.ProviderConfiguration, error)
	ListProviderConfigurations(context.Context) ([]domain.ProviderConfiguration, error)
	GetProviderConfiguration(context.Context, domain.ProviderConfigurationID) (domain.ProviderConfiguration, error)
	SetProviderLifecycle(context.Context, identity.UserID, domain.ProviderConfigurationID, domain.ProviderLifecycle, bool) (domain.ProviderConfiguration, error)
	QueueImport(context.Context, *identity.UserID, domain.ProviderConfigurationID, domain.ImportTrigger, string, int16) (domain.ImportRun, error)
	QueueDueImports(context.Context, time.Time, int) (int, error)
	RecoverStalledImports(context.Context, time.Time, time.Time, int) (int, error)
	ClaimNextImport(context.Context, time.Time) (domain.ImportRun, error)
	ApplyImport(context.Context, domain.ImportRun, []domain.ValidatedOffer, []domain.ImportRecordFailure, domain.ProviderBatch, time.Time) (domain.ImportApplyResult, error)
	CompleteImport(context.Context, domain.ImportRunID, domain.ImportApplyResult, time.Time) error
	FailImport(context.Context, domain.ImportRun, string, string, time.Time) error
	RetryImport(context.Context, identity.UserID, domain.ImportRunID, string) (domain.ImportRun, error)
	ListImports(context.Context, int, int) ([]domain.ImportRun, int64, error)
	ListImportFailures(context.Context, domain.ImportRunID, int, int) ([]domain.ImportFailure, int64, error)
	AnonymizeExpiredClicks(context.Context, time.Time, int) (int64, error)
}
