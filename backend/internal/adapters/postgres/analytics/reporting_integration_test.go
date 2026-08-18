package analyticspostgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/analytics/domain"
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

	report, err := New(pool).Report(ctx, domain.ReportQuery{From: time.Now().Add(-24 * time.Hour), To: time.Now().Add(time.Minute), Limit: 10})
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

func TestReportUsesOnlyValidatedReportableEvents(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var productID, acceptedID, filteredID string
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.products ORDER BY id LIMIT 1`).Scan(&productID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT gen_random_uuid(),gen_random_uuid()`).Scan(&acceptedID, &filteredID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, item := range []struct {
		id         string
		reportable bool
	}{{acceptedID, true}, {filteredID, false}} {
		if _, err := pool.Exec(ctx, `INSERT INTO analytics.events (public_event_id,event_name,schema_version,session_id,surface,properties,consent_state,
			origin,classification,is_reportable,occurred_at,retention_expires_at) VALUES ($1,'product_viewed',3,$2,'test',jsonb_build_object('product_id',$3::text),
			'granted','client','human',$4,$5,$6)`, item.id, "1191bb26-a9a2-41df-9346-74d693350ce8", productID, item.reportable, now, now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics.events WHERE public_event_id=ANY($1::uuid[])`, []string{acceptedID, filteredID})
	})
	report, err := New(pool).Report(ctx, domain.ReportQuery{From: now.Add(-time.Minute), To: now.Add(time.Minute), Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.ProductViews != 1 || len(report.MostViewed) != 1 || report.MostViewed[0].Count != 1 {
		t.Fatalf("filtered report=%#v rankings=%#v", report.Summary, report.MostViewed)
	}
}
