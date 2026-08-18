package commercepostgres

import (
	"testing"

	"rigmark/internal/modules/commerce/domain"
)

func TestRatioMetricDistinguishesZeroFromInsufficientData(t *testing.T) {
	zero := ratioMetric(0, 10, "fixture")
	if zero.Status != domain.MetricAvailable || zero.Value == nil || *zero.Value != 0 {
		t.Fatalf("zero metric=%#v", zero)
	}
	insufficient := ratioMetric(0, 0, "fixture")
	if insufficient.Status != domain.MetricInsufficient || insufficient.Value != nil {
		t.Fatalf("insufficient metric=%#v", insufficient)
	}
	missing := ratioNoData("fixture")
	if missing.Status != domain.MetricNoData || missing.Value != nil {
		t.Fatalf("missing metric=%#v", missing)
	}
}
