package ports

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/commerce/domain"
)

var ErrAffiliateDestinationNotFound = errors.New("affiliate destination not found")

type MerchantRepository interface {
	ListActive(context.Context) ([]domain.Merchant, error)
}

type OfferRepository interface {
	ListAvailableByProduct(context.Context, catalog.ProductID, string) ([]domain.Offer, error)
}

type AffiliateRedirectRepository interface {
	TrackOfferClick(context.Context, domain.AffiliateClick) (string, error)
	TrackLegacyLinkClick(context.Context, domain.AffiliateClick) (string, error)
}
