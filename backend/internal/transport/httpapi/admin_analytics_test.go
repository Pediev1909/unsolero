package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	analytics "rigmark/internal/modules/analytics/domain"
)

type analyticsReportingStub struct {
	report analytics.Report
	err    error
}

func (stub analyticsReportingStub) Report(context.Context, analytics.ReportQuery) (analytics.Report, error) {
	return stub.report, stub.err
}

func TestAdminAnalyticsReportPreservesUnavailableRates(t *testing.T) {
	now := time.Now().UTC()
	handler := &Handler{
		analyticsReporting: analyticsReportingStub{report: analytics.Report{
			Summary:         analytics.ReportSummary{Users: 4},
			MostRecommended: []analytics.RankedEntity{},
			MostViewed:      []analytics.RankedEntity{},
			MostClicked:     []analytics.RankedEntity{},
			TopMerchants:    []analytics.RankedEntity{},
			TopCategories:   []analytics.RankedEntity{},
			TrafficSources:  []analytics.TrafficSource{},
			Window:          analytics.ReportingWindow{From: now.Add(-24 * time.Hour), To: now, ReportableFrom: now.Add(-time.Hour), CompleteThrough: now, Coverage: "partial", DataState: "no_data", Layer: "validated_filtered", MinimumSampleSize: 20},
		}},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	response := httptest.NewRecorder()
	handler.adminAnalyticsReport(response, httptest.NewRequest(http.MethodGet, "/api/admin/analytics", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body struct {
		Summary struct {
			Users          int64    `json:"users"`
			CompletionRate *float64 `json:"recommendation_completion_rate"`
			AffiliateCTR   *float64 `json:"affiliate_ctr"`
		} `json:"summary"`
		MostViewed      []rankedEntityResponse `json:"most_viewed_products"`
		Campaigns       []campaignResponse     `json:"campaigns"`
		LandingPages    []landingPageResponse  `json:"landing_pages"`
		SourcesByMedium []sourceMediumResponse `json:"sources_by_medium"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Summary.Users != 4 || body.Summary.CompletionRate != nil || body.Summary.AffiliateCTR != nil {
		t.Fatalf("unexpected summary: %#v", body.Summary)
	}
	if body.MostViewed == nil || body.Campaigns == nil || body.LandingPages == nil || body.SourcesByMedium == nil {
		t.Fatal("empty rankings must encode as an array")
	}
}

func TestAdminAnalyticsReportRejectsInvalidWindow(t *testing.T) {
	handler := &Handler{analyticsReporting: analyticsReportingStub{}, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()
	handler.adminAnalyticsReport(response, httptest.NewRequest(http.MethodGet, "/api/admin/analytics?from=not-a-date&limit=500", nil))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func TestAdminAnalyticsReportSerialisesCampaignAttribution(t *testing.T) {
	youtube, shorts := "youtube", "shorts"
	handler := &Handler{
		analyticsReporting: analyticsReportingStub{report: analytics.Report{
			Campaigns: []analytics.CampaignPerformance{{Campaign: "2026-09-crm-shootout", Source: &youtube, Medium: &shorts,
				Sessions: 12, PageViews: 31, AffiliateClicks: 3}},
			LandingPages:    []analytics.CampaignLandingPage{{Campaign: "2026-09-crm-shootout", PagePath: "/tools/crm", Sessions: 9}},
			SourcesByMedium: []analytics.TrafficSourceMedium{{Source: "youtube", Sessions: 4}},
		}},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	response := httptest.NewRecorder()
	handler.adminAnalyticsReport(response, httptest.NewRequest(http.MethodGet, "/api/admin/analytics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
	var body struct {
		Campaigns       []campaignResponse     `json:"campaigns"`
		LandingPages    []landingPageResponse  `json:"landing_pages"`
		SourcesByMedium []sourceMediumResponse `json:"sources_by_medium"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Campaigns) != 1 || body.Campaigns[0].Campaign != "2026-09-crm-shootout" ||
		body.Campaigns[0].TrafficSource == nil || *body.Campaigns[0].TrafficSource != "youtube" ||
		body.Campaigns[0].TrafficMedium == nil || *body.Campaigns[0].TrafficMedium != "shorts" ||
		body.Campaigns[0].Sessions != 12 || body.Campaigns[0].PageViews != 31 || body.Campaigns[0].AffiliateClicks != 3 {
		t.Fatalf("campaigns = %+v", body.Campaigns)
	}
	if len(body.LandingPages) != 1 || body.LandingPages[0].Campaign != "2026-09-crm-shootout" ||
		body.LandingPages[0].PagePath != "/tools/crm" || body.LandingPages[0].Sessions != 9 {
		t.Fatalf("landing pages = %+v", body.LandingPages)
	}
	if len(body.SourcesByMedium) != 1 || body.SourcesByMedium[0].TrafficSource != "youtube" ||
		body.SourcesByMedium[0].TrafficMedium != nil || body.SourcesByMedium[0].Sessions != 4 {
		t.Fatalf("sources by medium = %+v", body.SourcesByMedium)
	}
	// A link without utm_medium is an explicit null, not an omitted key: the
	// client schema treats the field as required-nullable, and an omitted key
	// would read as an older API rather than as a visit with no medium.
	if !strings.Contains(response.Body.String(), `"traffic_medium":null`) {
		t.Fatalf("missing medium must serialise as null: %s", response.Body)
	}
}
