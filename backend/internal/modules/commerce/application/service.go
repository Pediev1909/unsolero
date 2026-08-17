package application

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

var ErrInvalidAttribution = errors.New("invalid affiliate attribution")

var campaignPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
var referrerHostPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var allowedSources = map[string]bool{
	"product_detail": true,
	"wishlist":       true,
	"recommendation": true,
	"comparison":     true,
	"setup":          true,
}

type Service struct {
	offers    ports.OfferRepository
	redirects ports.AffiliateRedirectRepository
}

func NewService(offers ports.OfferRepository, redirects ports.AffiliateRedirectRepository) *Service {
	return &Service{offers: offers, redirects: redirects}
}

func (service *Service) ListOffers(ctx context.Context, productID catalog.ProductID, currency string) ([]domain.Offer, error) {
	return service.offers.ListAvailableByProduct(ctx, productID, currency)
}

func (service *Service) TrackOfferClick(ctx context.Context, click domain.AffiliateClick) (string, error) {
	if click.OfferID == "" || click.LinkID != "" || validateAttribution(click) != nil {
		return "", ErrInvalidAttribution
	}
	destination, err := service.redirects.TrackOfferClick(ctx, click)
	if err != nil {
		return "", err
	}
	return validatedDestination(destination)
}

func (service *Service) TrackLegacyLinkClick(ctx context.Context, click domain.AffiliateClick) (string, error) {
	if click.LinkID == "" || click.OfferID != "" || validateAttribution(click) != nil {
		return "", ErrInvalidAttribution
	}
	destination, err := service.redirects.TrackLegacyLinkClick(ctx, click)
	if err != nil {
		return "", err
	}
	return validatedDestination(destination)
}

func validateAttribution(click domain.AffiliateClick) error {
	if !allowedSources[click.Source] || click.UserID == nil && click.AnonymousID == nil {
		return ErrInvalidAttribution
	}
	for _, value := range []*string{click.AnonymousID, click.SessionID, click.RequestID} {
		if value != nil && (strings.TrimSpace(*value) == "" || len(*value) > 128) {
			return ErrInvalidAttribution
		}
	}
	if click.RecommendationID != nil {
		if click.UserID == nil ||
			(click.Source != "recommendation" && click.Source != "setup") ||
			!uuidPattern.MatchString(*click.RecommendationID) {
			return ErrInvalidAttribution
		}
	}
	for _, value := range []*string{click.Campaign, click.TrafficSource, click.TrafficMedium} {
		if value != nil && !campaignPattern.MatchString(*value) {
			return ErrInvalidAttribution
		}
	}
	if click.Referrer != nil && len(*click.Referrer) > 500 {
		return ErrInvalidAttribution
	}
	if click.ReferrerHost != nil && !referrerHostPattern.MatchString(*click.ReferrerHost) {
		return ErrInvalidAttribution
	}
	return nil
}

func validatedDestination(destination string) (string, error) {
	parsed, err := url.Parse(destination)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", errors.New("persisted affiliate destination is invalid")
	}
	return destination, nil
}
