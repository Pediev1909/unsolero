package analyticspostgres

import (
	"context"
	"fmt"
	"time"

	"rigmark/internal/modules/analytics/domain"
)

const minimumRateSampleSize int64 = 20

func (repository *Repository) Report(ctx context.Context, query domain.ReportQuery) (domain.Report, error) {
	report := domain.Report{
		MostRecommended: []domain.RankedEntity{}, MostViewed: []domain.RankedEntity{},
		MostClicked: []domain.RankedEntity{}, TopMerchants: []domain.RankedEntity{},
		Daily: []domain.DailyPoint{},
		TopCategories: []domain.RankedEntity{}, TrafficSources: []domain.TrafficSource{},
		Window: domain.ReportingWindow{From: query.From, To: query.To, Layer: "validated_filtered", MinimumSampleSize: minimumRateSampleSize},
	}
	if err := repository.loadSummary(ctx, &report.Summary, query); err != nil {
		return domain.Report{}, err
	}
	if err := repository.loadIngestion(ctx, &report.Ingestion, query); err != nil {
		return domain.Report{}, err
	}
	if err := repository.loadCoverage(ctx, &report.Window); err != nil {
		return domain.Report{}, err
	}
	if err := repository.loadDaily(ctx, &report.Daily, query); err != nil {
		return domain.Report{}, err
	}
	queries := []struct {
		target    *[]domain.RankedEntity
		name, sql string
	}{
		{&report.MostRecommended, "most recommended products", `
			SELECT products.id,products.name,count(*) FROM recommendation.recommendation_items items
			JOIN recommendation.recommendations recommendations ON recommendations.id=items.recommendation_id
			JOIN catalog.products products ON products.id=items.product_id
			WHERE items.item_type='selected' AND recommendations.created_at >= $1 AND recommendations.created_at < $2
			GROUP BY products.id,products.name ORDER BY count(*) DESC,products.name LIMIT $3`},
		{&report.MostViewed, "most viewed products", `
			SELECT products.id,products.name,count(*) FROM analytics.events events
			JOIN catalog.products products ON products.id::text=events.properties->>'product_id'
			WHERE events.is_reportable AND events.event_name='product_viewed' AND events.occurred_at >= $1 AND events.occurred_at < $2
			GROUP BY products.id,products.name ORDER BY count(*) DESC,products.name LIMIT $3`},
		{&report.MostClicked, "most clicked products", `
			SELECT products.id,products.name,count(*) FROM commerce.affiliate_clicks clicks
			JOIN catalog.products products ON products.id=clicks.product_id
			WHERE clicks.is_countable AND clicks.occurred_at >= $1 AND clicks.occurred_at < $2
			GROUP BY products.id,products.name ORDER BY count(*) DESC,products.name LIMIT $3`},
		{&report.TopMerchants, "top merchants", `
			SELECT merchants.id,merchants.name,count(*) FROM commerce.affiliate_clicks clicks
			JOIN commerce.merchant_offers offers ON offers.id=clicks.merchant_offer_id
			JOIN commerce.merchants merchants ON merchants.id=offers.merchant_id
			WHERE clicks.is_countable AND clicks.occurred_at >= $1 AND clicks.occurred_at < $2
			GROUP BY merchants.id,merchants.name ORDER BY count(*) DESC,merchants.name LIMIT $3`},
		{&report.TopCategories, "top categories", `
			SELECT categories.id,categories.name,count(*) FROM recommendation.recommendation_items items
			JOIN recommendation.recommendations recommendations ON recommendations.id=items.recommendation_id
			JOIN catalog.products products ON products.id=items.product_id
			JOIN catalog.categories categories ON categories.id=products.category_id
			WHERE items.item_type='selected' AND recommendations.created_at >= $1 AND recommendations.created_at < $2
			GROUP BY categories.id,categories.name ORDER BY count(*) DESC,categories.name LIMIT $3`},
	}
	for _, item := range queries {
		if err := repository.loadRanking(ctx, item.target, item.name, item.sql, query); err != nil {
			return domain.Report{}, err
		}
	}
	if err := repository.loadTrafficSources(ctx, &report.TrafficSources, query); err != nil {
		return domain.Report{}, err
	}
	activity := report.Summary.OnboardingStarted + report.Summary.ProductViews + report.Summary.AffiliateClicks + report.Summary.RecommendationSessions
	switch {
	case activity == 0:
		report.Window.DataState = "no_data"
	case activity < minimumRateSampleSize:
		report.Window.DataState = "insufficient_data"
	default:
		report.Window.DataState = "available"
	}
	return report, nil
}

func (repository *Repository) loadSummary(ctx context.Context, summary *domain.ReportSummary, query domain.ReportQuery) error {
	return repository.pool.QueryRow(ctx, `WITH onboarding_starts AS (
		SELECT DISTINCT properties->>'onboarding_id' onboarding_id FROM analytics.events
		WHERE is_reportable AND event_name='onboarding_started' AND properties?'onboarding_id' AND occurred_at >= $1 AND occurred_at < $2
	), onboarding_completions AS (
		SELECT DISTINCT properties->>'onboarding_id' onboarding_id FROM analytics.events
		WHERE is_reportable AND event_name='onboarding_completed' AND properties?'onboarding_id' AND occurred_at >= $1 AND occurred_at < $2
	), viewed_products AS (
		SELECT DISTINCT session_id,properties->>'product_id' product_id FROM analytics.events
		WHERE is_reportable AND event_name='product_viewed' AND session_id IS NOT NULL AND properties?'product_id' AND occurred_at >= $1 AND occurred_at < $2
	), clicked_products AS (
		SELECT DISTINCT session_id,product_id::text product_id FROM commerce.affiliate_clicks
		WHERE source='product_detail' AND session_id IS NOT NULL AND is_countable AND occurred_at >= $1 AND occurred_at < $2
	), observed_clicks AS (SELECT clicked_products.session_id,clicked_products.product_id FROM clicked_products JOIN viewed_products USING(session_id,product_id)),
	counts AS (SELECT
		(SELECT count(*) FROM identity.users WHERE deleted_at IS NULL AND created_at >= $1 AND created_at < $2) users,
		(SELECT count(*) FROM recommendation.recommendation_sessions WHERE started_at >= $1 AND started_at < $2) recommendation_sessions,
		(SELECT count(*) FROM onboarding_starts) onboarding_started,
		(SELECT count(*) FROM onboarding_completions JOIN onboarding_starts USING(onboarding_id)) onboarding_completed,
		(SELECT count(*) FROM analytics.events WHERE is_reportable AND event_name='product_viewed' AND occurred_at >= $1 AND occurred_at < $2) product_views,
		(SELECT count(*) FROM commerce.affiliate_clicks WHERE is_countable AND occurred_at >= $1 AND occurred_at < $2) affiliate_clicks,
		(SELECT count(*) FROM commerce.affiliate_clicks WHERE occurred_at >= $1 AND occurred_at < $2) affiliate_clicks_raw,
		(SELECT count(*) FROM viewed_products) viewed_product_sessions,
		(SELECT count(*) FROM observed_clicks) clicked_product_sessions)
	SELECT users,recommendation_sessions,onboarding_started,onboarding_completed,
		CASE WHEN onboarding_started < $3 THEN NULL ELSE round(onboarding_completed::numeric*100/onboarding_started,2)::double precision END,
		product_views,affiliate_clicks,affiliate_clicks_raw,
		CASE WHEN viewed_product_sessions < $3 THEN NULL ELSE round(clicked_product_sessions::numeric*100/viewed_product_sessions,2)::double precision END
	FROM counts`, query.From, query.To, minimumRateSampleSize).Scan(&summary.Users, &summary.RecommendationSessions,
		&summary.OnboardingStarted, &summary.OnboardingCompleted, &summary.RecommendationCompletion,
		&summary.ProductViews, &summary.AffiliateClicks, &summary.AffiliateClicksRaw, &summary.AffiliateCTR)
}

func (repository *Repository) loadIngestion(ctx context.Context, summary *domain.IngestionSummary, query domain.ReportQuery) error {
	return repository.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE outcome='accepted'),
		count(*) FILTER(WHERE outcome='rejected'),count(*) FILTER(WHERE outcome='privacy_filtered'),
		count(*) FILTER(WHERE outcome='bot_filtered'),count(*) FILTER(WHERE outcome='deduplicated')
		FROM analytics.ingestion_receipts WHERE received_at >= $1 AND received_at < $2`, query.From, query.To).
		Scan(&summary.Received, &summary.Accepted, &summary.Rejected, &summary.PrivacyFiltered, &summary.BotFiltered, &summary.Deduplicated)
}

func (repository *Repository) loadCoverage(ctx context.Context, window *domain.ReportingWindow) error {
	if err := repository.pool.QueryRow(ctx, `SELECT reportable_from,complete_through FROM analytics.reporting_coverage WHERE pipeline_key='first_party_events_v3'`).Scan(&window.ReportableFrom, &window.CompleteThrough); err != nil {
		return err
	}
	window.Coverage = "complete"
	if window.From.Before(window.ReportableFrom) || window.To.After(window.CompleteThrough.Add(5*time.Minute)) {
		window.Coverage = "partial"
	}
	return nil
}

func (repository *Repository) loadRanking(ctx context.Context, target *[]domain.RankedEntity, name, sql string, query domain.ReportQuery) error {
	rows, err := repository.pool.Query(ctx, sql, query.From, query.To, query.Limit)
	if err != nil {
		return fmt.Errorf("load %s: %w", name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.RankedEntity
		if err := rows.Scan(&item.ID, &item.Name, &item.Count); err != nil {
			return fmt.Errorf("scan %s: %w", name, err)
		}
		*target = append(*target, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s: %w", name, err)
	}
	return nil
}

func (repository *Repository) loadTrafficSources(ctx context.Context, target *[]domain.TrafficSource, query domain.ReportQuery) error {
	rows, err := repository.pool.Query(ctx, `SELECT traffic_source,count(DISTINCT session_id) FROM analytics.events
		WHERE is_reportable AND event_name='page_view' AND traffic_source IS NOT NULL AND session_id IS NOT NULL
			AND occurred_at >= $1 AND occurred_at < $2
		GROUP BY traffic_source ORDER BY count(*) DESC,traffic_source LIMIT $3`, query.From, query.To, query.Limit)
	if err != nil {
		return fmt.Errorf("load traffic sources: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.TrafficSource
		if err := rows.Scan(&item.Source, &item.Count); err != nil {
			return fmt.Errorf("scan traffic source: %w", err)
		}
		*target = append(*target, item)
	}
	return rows.Err()
}

// loadDaily returns one row per day in the window, including days with no
// activity. Generating the calendar in SQL rather than filling gaps afterwards
// keeps a quiet day visible as a zero instead of vanishing from the series,
// which is the difference between a flat line and a misleadingly short one.
func (repository *Repository) loadDaily(ctx context.Context, target *[]domain.DailyPoint, query domain.ReportQuery) error {
	rows, err := repository.pool.Query(ctx, `
		WITH days AS (
			SELECT generate_series(
				date_trunc('day', $1::timestamptz),
				date_trunc('day', $2::timestamptz - interval '1 microsecond'),
				interval '1 day') AS day
		), views AS (
			SELECT date_trunc('day', occurred_at) AS day, count(*) AS total
			FROM analytics.events
			WHERE is_reportable AND event_name = 'product_viewed'
				AND occurred_at >= $1 AND occurred_at < $2
			GROUP BY 1
		), clicks AS (
			SELECT date_trunc('day', occurred_at) AS day, count(*) AS total
			FROM commerce.affiliate_clicks
			WHERE is_countable AND occurred_at >= $1 AND occurred_at < $2
			GROUP BY 1
		)
		SELECT days.day, COALESCE(views.total, 0), COALESCE(clicks.total, 0)
		FROM days
		LEFT JOIN views ON views.day = days.day
		LEFT JOIN clicks ON clicks.day = days.day
		ORDER BY days.day`, query.From, query.To)
	if err != nil {
		return fmt.Errorf("load daily analytics: %w", err)
	}
	defer rows.Close()
	points := make([]domain.DailyPoint, 0)
	for rows.Next() {
		var point domain.DailyPoint
		if err := rows.Scan(&point.Day, &point.ProductViews, &point.AffiliateClicks); err != nil {
			return fmt.Errorf("scan daily analytics: %w", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate daily analytics: %w", err)
	}
	*target = points
	return nil
}
