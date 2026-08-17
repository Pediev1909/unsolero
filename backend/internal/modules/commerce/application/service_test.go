package application

import (
	"context"
	"testing"

	"rigmark/internal/modules/commerce/domain"
)

type redirectStub struct{ destination string }

func (stub redirectStub) TrackOfferClick(context.Context, domain.AffiliateClick) (string, error) {
	return stub.destination, nil
}
func (stub redirectStub) TrackLegacyLinkClick(context.Context, domain.AffiliateClick) (string, error) {
	return stub.destination, nil
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
	destination, err := service.TrackOfferClick(context.Background(), validClick())
	if err != nil {
		t.Fatalf("track and resolve: %v", err)
	}
	if destination != "https://merchant.example/product" {
		t.Fatalf("unexpected destination %q", destination)
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
