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
		report.TopMerchants == nil || report.TopCategories == nil || report.TrafficSources == nil ||
		report.Campaigns == nil || report.LandingPages == nil || report.SourcesByMedium == nil {
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

// One campaign posted in two formats. Session one lands on /pricing, moves to
// /tools, then clicks a merchant; session two lands on /pricing from the other
// format; session three is identical to session one but never passed
// reportability and must not appear anywhere.
func TestReportAttributesCampaignSessionsLandingPagesAndClicks(t *testing.T) {
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
	// Tokens unique to this run, so rows already in a shared database cannot
	// be mistaken for fixtures, and cleanup can key on the campaign alone.
	var campaign, source, first, second, third string
	if err := pool.QueryRow(ctx, `SELECT 'test-'||replace(gen_random_uuid()::text,'-',''),'src-'||replace(gen_random_uuid()::text,'-',''),
		gen_random_uuid()::text,gen_random_uuid()::text,gen_random_uuid()::text`).Scan(&campaign, &source, &first, &second, &third); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics.events WHERE campaign=$1`, campaign)
	})
	now := time.Now().UTC()
	fixtures := []struct {
		name, session, path, medium string
		reportable                  bool
		at                          time.Time
	}{
		{"page_view", first, "/pricing", "shorts", true, now},
		{"page_view", first, "/tools", "shorts", true, now.Add(time.Second)},
		{"affiliate_clicked", first, "", "shorts", true, now.Add(2 * time.Second)},
		{"page_view", second, "/pricing", "video", true, now},
		{"page_view", third, "/pricing", "shorts", false, now},
		{"affiliate_clicked", third, "", "shorts", false, now},
	}
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, `INSERT INTO analytics.events (event_name,schema_version,session_id,surface,properties,page_path,
			traffic_source,traffic_medium,campaign,consent_state,origin,classification,is_reportable,occurred_at,retention_expires_at)
			VALUES ($1,3,$2,'test','{}'::jsonb,NULLIF($3,''),$4,$5,$6,'granted','client','human',$7,$8,$9)`,
			fixture.name, fixture.session, fixture.path, source, fixture.medium, campaign, fixture.reportable, fixture.at, now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	report, err := New(pool).Report(ctx, domain.ReportQuery{From: now.Add(-time.Minute), To: now.Add(time.Minute), Limit: 50})
	if err != nil {
		t.Fatal(err)
	}

	var campaigns []domain.CampaignPerformance
	for _, row := range report.Campaigns {
		if row.Campaign == campaign {
			campaigns = append(campaigns, row)
		}
	}
	if len(campaigns) != 2 ||
		stringValue(campaigns[0].Source) != source || stringValue(campaigns[0].Medium) != "shorts" ||
		campaigns[0].Sessions != 1 || campaigns[0].PageViews != 2 || campaigns[0].AffiliateClicks != 1 ||
		stringValue(campaigns[1].Medium) != "video" || campaigns[1].Sessions != 1 || campaigns[1].PageViews != 1 || campaigns[1].AffiliateClicks != 0 {
		t.Fatalf("campaigns = %+v", campaigns)
	}

	var landings []domain.CampaignLandingPage
	for _, row := range report.LandingPages {
		if row.Campaign == campaign {
			landings = append(landings, row)
		}
	}
	if len(landings) != 1 || landings[0].PagePath != "/pricing" || landings[0].Sessions != 2 {
		t.Fatalf("landing pages = %+v", landings)
	}

	var mediums []domain.TrafficSourceMedium
	for _, row := range report.SourcesByMedium {
		if row.Source == source {
			mediums = append(mediums, row)
		}
	}
	if len(mediums) != 2 || stringValue(mediums[0].Medium) != "shorts" || mediums[0].Sessions != 1 ||
		stringValue(mediums[1].Medium) != "video" || mediums[1].Sessions != 1 {
		t.Fatalf("sources by medium = %+v", mediums)
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
