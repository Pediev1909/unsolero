package application

import (
	"context"
	"errors"
	"testing"

	"rigmark/internal/modules/commerce/domain"
)

type redirectStub struct {
	destination string
	recordErr   error
}

func (stub redirectStub) ResolveOfferDestination(context.Context, domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error) {
	return domain.ResolvedAffiliateDestination{DestinationURL: stub.destination}, nil
}
func (stub redirectStub) ResolveLegacyDestination(context.Context, domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error) {
	return domain.ResolvedAffiliateDestination{DestinationURL: stub.destination}, nil
}
func (stub redirectStub) RecordClick(context.Context, domain.ResolvedAffiliateDestination, domain.AffiliateClick) error {
	return stub.recordErr
}

func validClick() domain.AffiliateClick {
	anonymousID := "anonymous"
	return domain.AffiliateClick{OfferID: "offer-id", Source: "product_detail", AnonymousID: &anonymousID}
}

func TestTrackOfferClickRejectsUnsafePersistedURL(t *testing.T) {
	service := &Service{redirects: redirectStub{destination: "javascript:alert(1)"}}
	if _, err := service.TrackOfferClick(context.Background(), validClick()); err == nil {
		t.Fatal("expected unsafe destination to be rejected")
	}
}

func TestTrackOfferClickAcceptsHTTPS(t *testing.T) {
	service := &Service{redirects: redirectStub{destination: "https://merchant.example/product"}}
	result, err := service.TrackOfferClick(context.Background(), validClick())
	if err != nil {
		t.Fatalf("track and resolve: %v", err)
	}
	if result.DestinationURL != "https://merchant.example/product" {
		t.Fatalf("unexpected destination %q", result.DestinationURL)
	}
}

func TestTrackOfferClickPreservesNavigationWhenRecordingFails(t *testing.T) {
	recordErr := errors.New("analytics unavailable")
	service := &Service{redirects: redirectStub{destination: "https://merchant.example/product", recordErr: recordErr}}
	result, err := service.TrackOfferClick(context.Background(), validClick())
	if err != nil || result.DestinationURL == "" || !errors.Is(result.TrackingError, recordErr) {
		t.Fatalf("result = %#v, error = %v", result, err)
	}
}

func TestNormalizeAttributionUsesCanonicalSourceAndCampaign(t *testing.T) {
	campaign := " Summer_Launch "
	click := validClick()
	click.Source = " Product_Detail "
	click.Campaign = &campaign
	normalized := normalizeAttribution(click)
	if normalized.Source != "product_detail" || normalized.Campaign == nil || *normalized.Campaign != "summer_launch" {
		t.Fatalf("normalized attribution = %#v", normalized)
	}
}

func TestTrackOfferClickRejectsCommissionLikeUnrecognizedSource(t *testing.T) {
	service := &Service{redirects: redirectStub{destination: "https://merchant.example/product"}}
	click := validClick()
	click.Source = "highest_commission"
	if _, err := service.TrackOfferClick(context.Background(), click); err != ErrInvalidAttribution {
		t.Fatalf("error = %v, want ErrInvalidAttribution", err)
	}
}

func TestTrackOfferClickRejectsAnonymousRecommendationAttribution(t *testing.T) {
	service := &Service{redirects: redirectStub{destination: "https://merchant.example/product"}}
	click := validClick()
	recommendationID := "8f045a40-37d8-4f83-a2d1-8953153dd3e9"
	click.Source = "recommendation"
	click.RecommendationID = &recommendationID
	if _, err := service.TrackOfferClick(context.Background(), click); err != ErrInvalidAttribution {
		t.Fatalf("error = %v, want ErrInvalidAttribution", err)
	}
}

func TestTrackOfferClickRejectsRecommendationOnUnrelatedSurface(t *testing.T) {
	service := &Service{redirects: redirectStub{destination: "https://merchant.example/product"}}
	click := validClick()
	userID := "8c806415-dd48-46bc-a976-4df457c4a1c8"
	recommendationID := "8f045a40-37d8-4f83-a2d1-8953153dd3e9"
	click.UserID = &userID
	click.Source = "wishlist"
	click.RecommendationID = &recommendationID
	if _, err := service.TrackOfferClick(context.Background(), click); err != ErrInvalidAttribution {
		t.Fatalf("error = %v, want ErrInvalidAttribution", err)
	}
}

func TestNormalizeAttributionDropsReferrerPathAndQuery(t *testing.T) {
	referrer := "https://Example.COM/account/person@example.com?token=secret"
	click := normalizeAttribution(domain.AffiliateClick{Referrer: &referrer})
	if click.Referrer == nil || *click.Referrer != "https://example.com" {
		t.Fatalf("normalized referrer = %v", click.Referrer)
	}
}
