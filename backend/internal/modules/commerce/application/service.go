package application

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

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
	"promotion":      true,
}

type Service struct {
	offers         ports.OfferRepository
	redirects      ports.AffiliateRedirectRepository
	clickRetention time.Duration
}

func NewService(offers ports.OfferRepository, redirects ports.AffiliateRedirectRepository, retention ...time.Duration) *Service {
	clickRetention := 397 * 24 * time.Hour
	if len(retention) > 0 && retention[0] > 0 {
		clickRetention = retention[0]
	}
	return &Service{offers: offers, redirects: redirects, clickRetention: clickRetention}
}

func (service *Service) ListOffers(ctx context.Context, productID catalog.ProductID, currency string) ([]domain.Offer, error) {
	return service.offers.ListAvailableByProduct(ctx, productID, currency)
}

func (service *Service) TrackOfferClick(ctx context.Context, click domain.AffiliateClick) (domain.AffiliateRedirectResult, error) {
	click = normalizeAttribution(click)
	if click.OfferID == "" || click.LinkID != "" || click.PromotionSlug != "" || validateAttribution(click) != nil {
		return domain.AffiliateRedirectResult{}, ErrInvalidAttribution
	}
	destination, err := service.redirects.ResolveOfferDestination(ctx, click)
	if err != nil {
		return domain.AffiliateRedirectResult{}, err
	}
	return service.recordResolvedClick(ctx, destination, click)
}

func (service *Service) TrackLegacyLinkClick(ctx context.Context, click domain.AffiliateClick) (domain.AffiliateRedirectResult, error) {
	click = normalizeAttribution(click)
	if click.LinkID == "" || click.OfferID != "" || click.PromotionSlug != "" || validateAttribution(click) != nil {
		return domain.AffiliateRedirectResult{}, ErrInvalidAttribution
	}
	destination, err := service.redirects.ResolveLegacyDestination(ctx, click)
	if err != nil {
		return domain.AffiliateRedirectResult{}, err
	}
	return service.recordResolvedClick(ctx, destination, click)
}

func (service *Service) TrackPromotionClick(ctx context.Context, click domain.AffiliateClick) (domain.AffiliateRedirectResult, error) {
	click = normalizeAttribution(click)
	if click.PromotionSlug == "" || click.OfferID != "" || click.LinkID != "" ||
		click.Source != "promotion" || click.RecommendationID != nil || validateAttribution(click) != nil {
		return domain.AffiliateRedirectResult{}, ErrInvalidAttribution
	}
	destination, err := service.redirects.ResolvePromotionDestination(ctx, click)
	if err != nil {
		return domain.AffiliateRedirectResult{}, err
	}
	validated, err := validatedDestination(destination.DestinationURL)
	if err != nil {
		return domain.AffiliateRedirectResult{}, err
	}
	if click.Classification == "" {
		click.Classification = domain.ClickUnknown
	}
	click.IsCountable = click.Classification == domain.ClickHuman
	if click.RetentionExpires.IsZero() {
		click.RetentionExpires = time.Now().UTC().Add(service.clickRetention)
	}
	trackingErr := service.redirects.RecordPromotionClick(ctx, destination, click)
	return domain.AffiliateRedirectResult{DestinationURL: validated, TrackingError: trackingErr}, nil
}

func (service *Service) recordResolvedClick(ctx context.Context, destination domain.ResolvedAffiliateDestination, click domain.AffiliateClick) (domain.AffiliateRedirectResult, error) {
	validated, err := validatedDestination(destination.DestinationURL)
	if err != nil {
		return domain.AffiliateRedirectResult{}, err
	}
	if click.Classification == "" {
		click.Classification = domain.ClickUnknown
	}
	click.IsCountable = click.Classification == domain.ClickHuman
	if click.RetentionExpires.IsZero() {
		click.RetentionExpires = time.Now().UTC().Add(service.clickRetention)
	}
	trackingErr := service.redirects.RecordClick(ctx, destination, click)
	return domain.AffiliateRedirectResult{DestinationURL: validated, TrackingError: trackingErr}, nil
}

func validateAttribution(click domain.AffiliateClick) error {
	if !allowedSources[click.Source] || click.UserID == nil && click.AnonymousID == nil {
		return ErrInvalidAttribution
	}
	for _, value := range []*string{click.AnonymousID, click.SessionID, click.RequestID, click.IdempotencyKey} {
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
	if click.Referrer != nil {
		parsed, err := url.Parse(strings.TrimSpace(*click.Referrer))
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
			return ErrInvalidAttribution
		}
	}
	if click.ReferrerHost != nil && !referrerHostPattern.MatchString(*click.ReferrerHost) {
		return ErrInvalidAttribution
	}
	return nil
}

func normalizeAttribution(click domain.AffiliateClick) domain.AffiliateClick {
	click.Source = strings.ToLower(strings.TrimSpace(click.Source))
	for _, target := range []**string{&click.Campaign, &click.TrafficSource, &click.TrafficMedium} {
		if *target != nil {
			value := strings.ToLower(strings.TrimSpace(**target))
			*target = &value
		}
	}
	if click.ReferrerHost != nil {
		value := strings.ToLower(strings.TrimSpace(*click.ReferrerHost))
		click.ReferrerHost = &value
	}
	if click.Referrer != nil {
		parsed, _ := url.Parse(strings.TrimSpace(*click.Referrer))
		value := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
		click.Referrer = &value
	}
	return click
}

func validatedDestination(destination string) (string, error) {
	parsed, err := url.Parse(destination)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", errors.New("persisted affiliate destination is invalid")
	}
	return destination, nil
}
