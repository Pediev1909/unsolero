package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	domain "rigmark/internal/modules/recommendation/domain"
)

// The browser parses this response with a strict schema, so the field names are
// part of the contract. Serialising the domain summary directly emitted
// PascalCase while every other admin response is snake_case, and nothing caught
// it because no client consumed the endpoint.
func TestPresentPolicyUsesSnakeCaseFieldNames(t *testing.T) {
	activated := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	encoded, err := json.Marshal(presentPolicy(domain.PolicySummary{
		Version:       "saas-v1",
		VerticalKey:   "saas",
		Status:        domain.PolicyActive,
		CategoryCount: 12,
		ProductCount:  53,
		CreatedAt:     time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC),
		ActivatedAt:   &activated,
		ReviewNote:    "approved by reviewer",
	}))
	if err != nil {
		t.Fatalf("marshal policy: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal policy: %v", err)
	}

	for _, field := range []string{
		"version", "vertical_key", "status", "category_count",
		"product_count", "created_at", "activated_at", "review_note",
	} {
		if _, ok := decoded[field]; !ok {
			t.Errorf("response is missing %q; got %v", field, decoded)
		}
	}
	if len(decoded) != 8 {
		t.Errorf("expected exactly 8 fields, got %d: %v", len(decoded), decoded)
	}
	if decoded["status"] != "active" {
		t.Errorf("status = %v, want active", decoded["status"])
	}
	if decoded["vertical_key"] != "saas" {
		t.Errorf("vertical_key = %v, want saas", decoded["vertical_key"])
	}
}

// A version that was never activated has to reach the browser as an explicit
// null rather than a zero timestamp, which the page renders as "Never".
func TestPresentPolicyLeavesActivatedAtNull(t *testing.T) {
	encoded, err := json.Marshal(presentPolicy(domain.PolicySummary{
		Version: "saas-v2", VerticalKey: "saas", Status: domain.PolicyDraft,
	}))
	if err != nil {
		t.Fatalf("marshal policy: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal policy: %v", err)
	}
	if decoded["activated_at"] != nil {
		t.Errorf("activated_at = %v, want null", decoded["activated_at"])
	}
}
