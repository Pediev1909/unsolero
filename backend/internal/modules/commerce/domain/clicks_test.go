package domain

import "testing"

func TestClassifyClick(t *testing.T) {
	for name, test := range map[string]struct {
		agent, purpose string
		want           ClickClassification
	}{
		"browser":  {agent: "Mozilla/5.0 AppleWebKit/537.36 Chrome/128", want: ClickHuman},
		"bot":      {agent: "ExampleBot/1.0", want: ClickBot},
		"prefetch": {agent: "Mozilla/5.0", purpose: "prefetch", want: ClickPrefetch},
		"unknown":  {want: ClickUnknown},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ClassifyClick(test.agent, test.purpose, "", ""); got != test.want {
				t.Fatalf("classification = %q, want %q", got, test.want)
			}
		})
	}
}
