package httpapi

import "testing"

// Both array columns on recommendation.drafts are NOT NULL. Appending to a nil
// slice that never receives an element leaves it nil, and pgx sends nil as
// NULL, so every draft save before the preference and priority questions failed
// with a constraint violation. A signed-in visitor saw "your latest change
// could not be saved" on every single answer.
func TestDraftFromRequestNeverProducesNilArrays(t *testing.T) {
	for _, name := range []string{"absent", "empty"} {
		body := draftRequest{CurrentStep: 1}
		if name == "empty" {
			body.TrainingPreferences = []string{}
			body.Priorities = []string{}
		}

		draft := draftFromRequest(body)
		if draft.TrainingPreferences == nil {
			t.Errorf("%s: training preferences are nil; the column rejects NULL", name)
		}
		if draft.Priorities == nil {
			t.Errorf("%s: priorities are nil; the column rejects NULL", name)
		}
	}
}

func TestDraftFromRequestKeepsSuppliedValues(t *testing.T) {
	draft := draftFromRequest(draftRequest{
		CurrentStep:         4,
		TrainingPreferences: []string{"automation_first"},
		Priorities:          []string{"lowest_total_cost", "fewest_tools"},
	})
	if len(draft.TrainingPreferences) != 1 || len(draft.Priorities) != 2 {
		t.Fatalf("values were dropped: %v %v", draft.TrainingPreferences, draft.Priorities)
	}
}
