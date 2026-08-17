package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestRecommendationInputsCannotRepresentCommercialIncentives(t *testing.T) {
	forbidden := []string{"commission", "sponsor", "payout", "revenue", "affiliate", "conversion", "click"}
	for _, value := range []any{CandidateSnapshot{}, Config{}, Weights{}} {
		typeOf := reflect.TypeOf(value)
		for index := 0; index < typeOf.NumField(); index++ {
			name := strings.ToLower(typeOf.Field(index).Name)
			for _, term := range forbidden {
				if strings.Contains(name, term) {
					t.Fatalf("%s exposes forbidden commercial field %q", typeOf.Name(), name)
				}
			}
		}
	}
}
