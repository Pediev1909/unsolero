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
	// Campaign attribution: what a UTM-tagged link brought. Sessions and page
	// views come from consented page_view events; affiliate clicks from the
	// server-authored affiliate_clicked event, which needs no analytics
	// consent, so a campaign can legitimately show clicks with zero sessions.
	Campaigns       []CampaignPerformance
	LandingPages    []CampaignLandingPage
	SourcesByMedium []TrafficSourceMedium
	// Daily carries one point per day in the window. A total answers "how
	// many"; only a series answers "is this growing", which is the question
	// anyone watching a new site actually has.
	Daily     []DailyPoint
	Window    ReportingWindow
	Ingestion IngestionSummary
}

type DailyPoint struct {
	Day             time.Time
	ProductViews    int64
	AffiliateClicks int64
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

// CampaignPerformance is one (utm_campaign, utm_source, utm_medium) triple as
// the visit arrived with it. Source and Medium are nil when the link carried a
// campaign without that parameter.
type CampaignPerformance struct {
	Campaign        string
	Source          *string
	Medium          *string
	Sessions        int64
	PageViews       int64
	AffiliateClicks int64
}

// CampaignLandingPage is the page a campaign session opened first.
type CampaignLandingPage struct {
	Campaign string
	PagePath string
	Sessions int64
}

// TrafficSourceMedium splits a source by medium, so youtube/shorts and
// youtube/video stop sharing one row. Medium is nil when the link had none.
type TrafficSourceMedium struct {
	Source   string
	Medium   *string
	Sessions int64
}
