package main

import "testing"

// The fictional fixture publishes invented products at 'published', which is the
// status the public catalog filters on. Loading it into a deployed environment
// puts invented prices in front of real visitors, so the refusal is a guard
// rather than a convenience and is worth pinning.
func TestFixtureLoadAllowed(t *testing.T) {
	for _, testCase := range []struct {
		environment string
		allowed     bool
	}{
		{environment: "development", allowed: true},
		{environment: "test", allowed: true},
		{environment: "staging", allowed: false},
		{environment: "production", allowed: false},
	} {
		if got := fixtureLoadAllowed(testCase.environment); got != testCase.allowed {
			t.Errorf("fixtureLoadAllowed(%q) = %t, want %t",
				testCase.environment, got, testCase.allowed)
		}
	}
}
