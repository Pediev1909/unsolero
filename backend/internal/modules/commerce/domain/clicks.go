package domain

import "strings"

// ClassifyClick is intentionally conservative. It separates obviously
// automated/prefetched requests from raw click history without pretending that
// user-agent heuristics prove a person is human.
func ClassifyClick(userAgent, purpose, secPurpose, mozPrefetch string) ClickClassification {
	prefetchSignals := strings.ToLower(purpose + " " + secPurpose + " " + mozPrefetch)
	if strings.Contains(prefetchSignals, "prefetch") || strings.Contains(prefetchSignals, "prerender") {
		return ClickPrefetch
	}
	agent := strings.ToLower(strings.TrimSpace(userAgent))
	if agent == "" {
		return ClickUnknown
	}
	for _, token := range []string{"bot", "crawler", "spider", "slurp", "headless", "preview", "facebookexternalhit", "curl/", "wget/"} {
		if strings.Contains(agent, token) {
			return ClickBot
		}
	}
	return ClickHuman
}
