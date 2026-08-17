package domain

import (
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
)

func TestOfferLandedPriceMinor(t *testing.T) {
	offer := Offer{Price: catalog.Money{AmountMinor: 10_000, Currency: "USD"}, ShippingMinor: 799}
	if got := offer.LandedPriceMinor(); got != 10_799 {
		t.Fatalf("LandedPriceMinor() = %d, want 10799", got)
	}
}
