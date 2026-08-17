package domain

type Report struct {
	Summary         ReportSummary
	MostRecommended []RankedEntity
	MostViewed      []RankedEntity
	MostClicked     []RankedEntity
	TopMerchants    []RankedEntity
	TopCategories   []RankedEntity
	TrafficSources  []TrafficSource
}

type ReportSummary struct {
	Users                    int64
	RecommendationSessions   int64
	OnboardingStarted        int64
	OnboardingCompleted      int64
	RecommendationCompletion *float64
	ProductViews             int64
	AffiliateClicks          int64
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
