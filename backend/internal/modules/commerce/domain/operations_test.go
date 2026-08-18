package domain

import (
	"testing"
	"time"
)

func TestValidateProviderOfferRejectsMalformedAndUnsafeData(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	valid := ProviderOffer{ExternalOfferID: "external-1", ProductID: "56c11ce4-d2b3-4d3b-994c-a04afe3b9b16",
		MerchantSKU: "sku-1", ProductURL: "https://merchant.example/products/one", PriceMinor: 1000,
		Currency: "USD", Availability: "in_stock", Condition: "new"}
	if _, err := ValidateProviderOffer(valid, now, 72*time.Hour); err != nil {
		t.Fatalf("valid offer rejected: %v", err)
	}
	for name, mutate := range map[string]func(*ProviderOffer){
		"private destination":  func(record *ProviderOffer) { record.ProductURL = "https://127.0.0.1/product" },
		"userinfo":             func(record *ProviderOffer) { record.ProductURL = "https://user:secret@merchant.example/product" },
		"invalid currency":     func(record *ProviderOffer) { record.Currency = "usdollars" },
		"invalid availability": func(record *ProviderOffer) { record.Availability = "maybe" },
	} {
		t.Run(name, func(t *testing.T) {
			record := valid
			mutate(&record)
			if _, err := ValidateProviderOffer(record, now, 72*time.Hour); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateProviderOfferIsDeterministic(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	record := ProviderOffer{ExternalOfferID: "external-1", ProductID: "56c11ce4-d2b3-4d3b-994c-a04afe3b9b16",
		MerchantSKU: "sku-1", ProductURL: "https://merchant.example/products/one", PriceMinor: 1000,
		Currency: "USD", Availability: "in_stock", Condition: "new"}
	first, err := ValidateProviderOffer(record, now, 72*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ValidateProviderOffer(record, now, 72*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if first.PriceFingerprint != second.PriceFingerprint || first.AvailabilityFingerprint != second.AvailabilityFingerprint {
		t.Fatal("equivalent observations must produce stable fingerprints")
	}
}

func TestImportRetriesAreBounded(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	first, retry := NextImportRetry(1, 3, now)
	if !retry || first == nil || !first.Equal(now.Add(time.Minute)) {
		t.Fatalf("first retry = %v, %v", first, retry)
	}
	second, retry := NextImportRetry(2, 3, now)
	if !retry || second == nil || !second.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("second retry = %v, %v", second, retry)
	}
	if next, retry := NextImportRetry(3, 3, now); retry || next != nil {
		t.Fatalf("retry beyond maximum = %v, %v", next, retry)
	}
}
