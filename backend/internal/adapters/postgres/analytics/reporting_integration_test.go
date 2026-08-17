package analyticspostgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReportQueriesExecuteAgainstMigratedSchema(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	report, err := New(pool).Report(ctx, 10)
	if err != nil {
		t.Fatalf("Report(): %v", err)
	}
	if report.Summary.Users < 0 || report.Summary.RecommendationSessions < 0 ||
		report.Summary.ProductViews < 0 || report.Summary.AffiliateClicks < 0 {
		t.Fatalf("report contains an invalid count: %#v", report.Summary)
	}
	if report.MostRecommended == nil || report.MostViewed == nil || report.MostClicked == nil ||
		report.TopMerchants == nil || report.TopCategories == nil || report.TrafficSources == nil {
		t.Fatal("report collections must encode as arrays, not null")
	}
}
