package httpapi

import (
	"net/http"

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
		AffiliateCTR             *float64 `json:"affiliate_ctr"`
	} `json:"summary"`
	MostRecommended []rankedEntityResponse  `json:"most_recommended_products"`
	MostViewed      []rankedEntityResponse  `json:"most_viewed_products"`
	MostClicked     []rankedEntityResponse  `json:"most_clicked_products"`
	TopMerchants    []rankedEntityResponse  `json:"top_merchants"`
	TopCategories   []rankedEntityResponse  `json:"top_categories"`
	TrafficSources  []trafficSourceResponse `json:"traffic_sources"`
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
	report, err := h.analyticsReporting.Report(request.Context())
	if err != nil {
		h.logger.Error("load admin analytics", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "analytics_report_unavailable", "Analytics reporting is temporarily unavailable.", nil, h.logger)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, analyticsReportDTO(report))
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
	return result
}

func rankedEntityDTOs(values []analytics.RankedEntity) []rankedEntityResponse {
	result := make([]rankedEntityResponse, 0, len(values))
	for _, value := range values {
		result = append(result, rankedEntityResponse{ID: value.ID, Name: value.Name, Count: value.Count})
	}
	return result
}
