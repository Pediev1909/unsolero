package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		MostViewed []rankedEntityResponse `json:"most_viewed_products"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Summary.Users != 4 || body.Summary.CompletionRate != nil || body.Summary.AffiliateCTR != nil {
		t.Fatalf("unexpected summary: %#v", body.Summary)
	}
	if body.MostViewed == nil {
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
