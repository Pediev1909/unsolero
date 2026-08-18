package domain

import (
	"testing"
	"time"
)

func TestConversionLifecycleTransitions(t *testing.T) {
	orderCases := []struct {
		from, to OrderStatus
		want     bool
	}{
		{OrderPending, OrderConfirmed, true},
		{OrderConfirmed, OrderReversed, true},
		{OrderReversed, OrderConfirmed, false},
		{OrderRejected, OrderPending, false},
	}
	for _, test := range orderCases {
		if got := ValidateOrderTransition(test.from, test.to, false); got != test.want {
			t.Errorf("order transition %s -> %s = %t", test.from, test.to, got)
		}
	}
	commissionCases := []struct {
		from, to CommissionStatus
		want     bool
	}{
		{CommissionPending, CommissionApproved, true},
		{CommissionApproved, CommissionPaid, true},
		{CommissionPaid, CommissionReversed, true},
		{CommissionReversed, CommissionPaid, false},
		{CommissionRejected, CommissionApproved, false},
	}
	for _, test := range commissionCases {
		if got := ValidateCommissionTransition(test.from, test.to, false); got != test.want {
			t.Errorf("commission transition %s -> %s = %t", test.from, test.to, got)
		}
	}
}

func TestValidateProviderConversionPreservesOriginalCurrency(t *testing.T) {
	now := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	orderCurrency, commissionCurrency := "bgn", "eur"
	orderValue, commission := int64(24999), int64(1250)
	status := CommissionApproved
	event, err := ValidateProviderConversionEvent(ProviderConversionEvent{
		ProviderEventID: "evt-1", EventType: EventConversionCreated,
		ExternalConversionID: "conversion-1", OrderStatus: OrderConfirmed,
		OrderValueMinor: &orderValue, OrderCurrency: &orderCurrency,
		CommissionMinor: &commission, CommissionCurrency: &commissionCurrency,
		CommissionStatus: &status, EventTimestamp: now.Add(-time.Minute),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if *event.OrderCurrency != "BGN" || *event.CommissionCurrency != "EUR" ||
		*event.OrderValueMinor != orderValue || *event.CommissionMinor != commission {
		t.Fatalf("provider money was not preserved: %#v", event)
	}
}

func TestValidateProviderConversionRejectsMalformedMoneyAndFutureEvents(t *testing.T) {
	now := time.Now().UTC()
	currency := "USD"
	negative := int64(-1)
	base := ProviderConversionEvent{ProviderEventID: "evt", EventType: EventConversionCreated,
		ExternalConversionID: "conversion", OrderStatus: OrderPending, EventTimestamp: now}
	badMoney := base
	badMoney.CommissionMinor = &negative
	badMoney.CommissionCurrency = &currency
	if _, err := ValidateProviderConversionEvent(badMoney, now); err == nil {
		t.Fatal("negative commission was accepted")
	}
	oversized := MaximumVerifiedMoneyMinor + 1
	badMoney.CommissionMinor = &oversized
	if _, err := ValidateProviderConversionEvent(badMoney, now); err == nil {
		t.Fatal("unbounded commission was accepted")
	}
	future := base
	future.EventTimestamp = now.Add(6 * time.Minute)
	if _, err := ValidateProviderConversionEvent(future, now); err == nil {
		t.Fatal("future event was accepted")
	}
}

func TestConversionFingerprintIsDeterministicAndSensitive(t *testing.T) {
	event := ProviderConversionEvent{ProviderEventID: "evt", EventType: EventConversionCreated,
		ExternalConversionID: "conversion", OrderStatus: OrderPending, EventTimestamp: time.Unix(100, 0).UTC()}
	first := ConversionEventFingerprint(event)
	if first != ConversionEventFingerprint(event) {
		t.Fatal("same event produced different fingerprints")
	}
	event.OrderStatus = OrderConfirmed
	if first == ConversionEventFingerprint(event) {
		t.Fatal("conflicting event produced the same fingerprint")
	}
}
