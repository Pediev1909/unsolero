package application

import "testing"

// A caller that builds a query with MaximumPageSize must have it accepted. The
// server-rendered category and brand listings do exactly that, and when they
// asked for 60 against a limit of 48 the query was rejected, the error became
// an empty product list, and every listing page shipped a heading with no
// links — silently, because the failure only reached a log line.
func TestMaximumPageSizeProducesAValidQuery(t *testing.T) {
	normalized, err := normalizeQuery(Query{
		CategorySlug: "crm", Page: 1, PageSize: MaximumPageSize,
	})
	if err != nil {
		t.Fatalf("a query at MaximumPageSize was rejected: %v", err)
	}
	if normalized.PageSize != MaximumPageSize {
		t.Errorf("page size = %d, want %d", normalized.PageSize, MaximumPageSize)
	}
}

func TestOverSizedPageIsRejectedRatherThanClamped(t *testing.T) {
	if _, err := normalizeQuery(Query{PageSize: MaximumPageSize + 1}); err == nil {
		t.Fatal("an over-sized page size was accepted; callers rely on it failing loudly")
	}
}
