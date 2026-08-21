package application

import (
	"testing"

	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/ports"
)

func goalPointer(value string) *planning.Goal {
	goal := planning.Goal(value)
	return &goal
}

// The goal used to be checked against a hardcoded list of five fitness goals.
// After the vertical changed, picking any real goal made every draft save fail
// with 422, so a signed-in visitor's progress stopped reaching their account
// from question one onwards and the only symptom was a line of small grey text
// beside the form.
func TestDraftAcceptsAnyWellFormedGoal(t *testing.T) {
	for _, goal := range []string{
		"client_services", "sell_products_online", "software_product",
		"solo_consulting", "creator_business",
	} {
		draft := ports.Draft{CurrentStep: 1, Goal: goalPointer(goal)}
		if err := validateDraft(draft); err != nil {
			t.Errorf("validateDraft() with goal %q = %v, want nil", goal, err)
		}
	}
}

// Shape is still enforced, because a draft goal reaches a column with its own
// pattern constraint and a malformed one would fail in the database instead.
func TestDraftRejectsMalformedGoal(t *testing.T) {
	for _, goal := range []string{"Client Services", "9lives", "has-hyphen", "trailing "} {
		draft := ports.Draft{CurrentStep: 1, Goal: goalPointer(goal)}
		if err := validateDraft(draft); err == nil {
			t.Errorf("validateDraft() with goal %q = nil, want an error", goal)
		}
	}
}
