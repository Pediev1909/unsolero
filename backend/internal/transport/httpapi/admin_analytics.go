package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	analyticsapplication "rigmark/internal/modules/analytics/application"
	analytics "rigmark/internal/modules/analytics/domain"
)

type analyticsReportResponse struct {
	Summary struct {
		Users                    int64    `json:"users"`
		RecommendationSessions   int64    `json:"recommendation_sessions"`
		OnboardingStarted        int64    `json:"onboarding_started"`
		OnboardingCompleted      int64    `json:"onboarding_completed"`
		RecommendationCompletion *float64 `json:"recommendation_completion_rate"`
		ProductViews             int64    `json:"product_views"`
		AffiliateClicks          int64    `json:"affiliate_clicks"`
		AffiliateClicksRaw       int64    `json:"affiliate_clicks_raw"`
		AffiliateCTR             *float64 `json:"affiliate_ctr"`
	} `json:"summary"`
	MostRecommended []rankedEntityResponse  `json:"most_recommended_products"`
	MostViewed      []rankedEntityResponse  `json:"most_viewed_products"`
	MostClicked     []rankedEntityResponse  `json:"most_clicked_products"`
	TopMerchants    []rankedEntityResponse  `json:"top_merchants"`
	TopCategories   []rankedEntityResponse  `json:"top_categories"`
	TrafficSources  []trafficSourceResponse `json:"traffic_sources"`
	Window          struct {
		From              time.Time `json:"from"`
		To                time.Time `json:"to"`
		ReportableFrom    time.Time `json:"reportable_from"`
		CompleteThrough   time.Time `json:"complete_through"`
		Coverage          string    `json:"coverage"`
		DataState         string    `json:"data_state"`
		Layer             string    `json:"layer"`
		MinimumSampleSize int64     `json:"minimum_sample_size"`
	} `json:"window"`
	Ingestion struct {
		Received        int64 `json:"received"`
		Accepted        int64 `json:"accepted"`
		Rejected        int64 `json:"rejected"`
		PrivacyFiltered int64 `json:"privacy_filtered"`
		BotFiltered     int64 `json:"bot_filtered"`
		Deduplicated    int64 `json:"deduplicated"`
	} `json:"ingestion"`
}

type rankedEntityResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type trafficSourceResponse struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

func (h *Handler) adminAnalyticsReport(response http.ResponseWriter, request *http.Request) {
	query, err := analyticsReportQuery(request)
	if err != nil {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_analytics_range", "Use a valid RFC3339 date range of at most 366 days and a limit from 1 to 50.", nil, h.logger)
		return
	}
	report, err := h.analyticsReporting.Report(request.Context(), query)
	if errors.Is(err, analyticsapplication.ErrInvalidReportQuery) {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_analytics_range", "Use a valid RFC3339 date range of at most 366 days and a limit from 1 to 50.", nil, h.logger)
		return
	}
	if err != nil {
		h.logger.Error("load admin analytics", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "analytics_report_unavailable", "Analytics reporting is temporarily unavailable.", nil, h.logger)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, analyticsReportDTO(report))
}

func analyticsReportQuery(request *http.Request) (analytics.ReportQuery, error) {
	var result analytics.ReportQuery
	var err error
	if raw := request.URL.Query().Get("from"); raw != "" {
		result.From, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return result, err
		}
	}
	if raw := request.URL.Query().Get("to"); raw != "" {
		result.To, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return result, err
		}
	}
	if raw := request.URL.Query().Get("limit"); raw != "" {
		result.Limit, err = strconv.Atoi(raw)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func analyticsReportDTO(report analytics.Report) analyticsReportResponse {
	var result analyticsReportResponse
	result.Summary.Users = report.Summary.Users
	result.Summary.RecommendationSessions = report.Summary.RecommendationSessions
	result.Summary.OnboardingStarted = report.Summary.OnboardingStarted
	result.Summary.OnboardingCompleted = report.Summary.OnboardingCompleted
	result.Summary.RecommendationCompletion = report.Summary.RecommendationCompletion
	result.Summary.ProductViews = report.Summary.ProductViews
	result.Summary.AffiliateClicks = report.Summary.AffiliateClicks
	result.Summary.AffiliateClicksRaw = report.Summary.AffiliateClicksRaw
	result.Summary.AffiliateCTR = report.Summary.AffiliateCTR
	result.MostRecommended = rankedEntityDTOs(report.MostRecommended)
	result.MostViewed = rankedEntityDTOs(report.MostViewed)
	result.MostClicked = rankedEntityDTOs(report.MostClicked)
	result.TopMerchants = rankedEntityDTOs(report.TopMerchants)
	result.TopCategories = rankedEntityDTOs(report.TopCategories)
	result.TrafficSources = make([]trafficSourceResponse, 0, len(report.TrafficSources))
	for _, source := range report.TrafficSources {
		result.TrafficSources = append(result.TrafficSources, trafficSourceResponse{Source: source.Source, Count: source.Count})
	}
	result.Window.From = report.Window.From
	result.Window.To = report.Window.To
	result.Window.ReportableFrom = report.Window.ReportableFrom
	result.Window.CompleteThrough = report.Window.CompleteThrough
	result.Window.Coverage = report.Window.Coverage
	result.Window.DataState = report.Window.DataState
	result.Window.Layer = report.Window.Layer
	result.Window.MinimumSampleSize = report.Window.MinimumSampleSize
	result.Ingestion.Received = report.Ingestion.Received
	result.Ingestion.Accepted = report.Ingestion.Accepted
	result.Ingestion.Rejected = report.Ingestion.Rejected
	result.Ingestion.PrivacyFiltered = report.Ingestion.PrivacyFiltered
	result.Ingestion.BotFiltered = report.Ingestion.BotFiltered
	result.Ingestion.Deduplicated = report.Ingestion.Deduplicated
	return result
}

func rankedEntityDTOs(values []analytics.RankedEntity) []rankedEntityResponse {
	result := make([]rankedEntityResponse, 0, len(values))
	for _, value := range values {
		result = append(result, rankedEntityResponse{ID: value.ID, Name: value.Name, Count: value.Count})
	}
	return result
}
