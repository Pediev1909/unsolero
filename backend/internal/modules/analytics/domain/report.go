package domain

import "time"

type Report struct {
	Summary         ReportSummary
	MostRecommended []RankedEntity
	MostViewed      []RankedEntity
	MostClicked     []RankedEntity
	TopMerchants    []RankedEntity
	TopCategories   []RankedEntity
	TrafficSources  []TrafficSource
	Window          ReportingWindow
	Ingestion       IngestionSummary
}

type ReportQuery struct {
	From  time.Time
	To    time.Time
	Limit int
}

type ReportingWindow struct {
	From              time.Time
	To                time.Time
	ReportableFrom    time.Time
	CompleteThrough   time.Time
	Coverage          string
	DataState         string
	Layer             string
	MinimumSampleSize int64
}

type IngestionSummary struct {
	Received        int64
	Accepted        int64
	Rejected        int64
	PrivacyFiltered int64
	BotFiltered     int64
	Deduplicated    int64
}

type ReportSummary struct {
	Users                    int64
	RecommendationSessions   int64
	OnboardingStarted        int64
	OnboardingCompleted      int64
	RecommendationCompletion *float64
	ProductViews             int64
	AffiliateClicks          int64
	AffiliateClicksRaw       int64
	AffiliateCTR             *float64
}

type RankedEntity struct {
	ID    string
	Name  string
	Count int64
}

type TrafficSource struct {
	Source string
	Count  int64
}
