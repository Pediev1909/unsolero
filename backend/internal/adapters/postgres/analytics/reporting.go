package analyticspostgres

import (
	"context"
	"fmt"

	"rigmark/internal/modules/analytics/domain"
)

func (repository *Repository) Report(ctx context.Context, limit int) (domain.Report, error) {
	report := domain.Report{
		MostRecommended: make([]domain.RankedEntity, 0),
		MostViewed:      make([]domain.RankedEntity, 0),
		MostClicked:     make([]domain.RankedEntity, 0),
		TopMerchants:    make([]domain.RankedEntity, 0),
		TopCategories:   make([]domain.RankedEntity, 0),
		TrafficSources:  make([]domain.TrafficSource, 0),
	}
	if err := repository.loadSummary(ctx, &report.Summary); err != nil {
		return domain.Report{}, err
	}
	queries := []struct {
		target *[]domain.RankedEntity
		name   string
		query  string
	}{
		{&report.MostRecommended, "most recommended products", `
			SELECT products.id, products.name, count(*)
			FROM recommendation.recommendation_items items
			JOIN catalog.products products ON products.id = items.product_id
			WHERE items.item_type = 'selected'
			GROUP BY products.id, products.name
			ORDER BY count(*) DESC, products.name
			LIMIT $1`},
		{&report.MostViewed, "most viewed products", `
			SELECT products.id, products.name, count(*)
			FROM analytics.events events
			JOIN catalog.products products ON products.id::text = events.properties->>'product_id'
			WHERE events.event_name = 'product_viewed'
			GROUP BY products.id, products.name
			ORDER BY count(*) DESC, products.name
			LIMIT $1`},
		{&report.MostClicked, "most clicked products", `
			SELECT products.id, products.name, count(*)
			FROM commerce.affiliate_clicks clicks
			JOIN catalog.products products ON products.id = clicks.product_id
			GROUP BY products.id, products.name
			ORDER BY count(*) DESC, products.name
			LIMIT $1`},
		{&report.TopMerchants, "top merchants", `
			SELECT merchants.id, merchants.name, count(*)
			FROM commerce.affiliate_clicks clicks
			JOIN commerce.merchant_offers offers ON offers.id = clicks.merchant_offer_id
			JOIN commerce.merchants merchants ON merchants.id = offers.merchant_id
			GROUP BY merchants.id, merchants.name
			ORDER BY count(*) DESC, merchants.name
			LIMIT $1`},
		{&report.TopCategories, "top categories", `
			SELECT categories.id, categories.name, count(*)
			FROM recommendation.recommendation_items items
			JOIN catalog.products products ON products.id = items.product_id
			JOIN catalog.categories categories ON categories.id = products.category_id
			WHERE items.item_type = 'selected'
			GROUP BY categories.id, categories.name
			ORDER BY count(*) DESC, categories.name
			LIMIT $1`},
	}
	for _, item := range queries {
		if err := repository.loadRanking(ctx, item.target, item.name, item.query, limit); err != nil {
			return domain.Report{}, err
		}
	}
	if err := repository.loadTrafficSources(ctx, &report.TrafficSources, limit); err != nil {
		return domain.Report{}, err
	}
	return report, nil
}

func (repository *Repository) loadSummary(ctx context.Context, summary *domain.ReportSummary) error {
	return repository.pool.QueryRow(ctx, `
		WITH onboarding_starts AS (
			SELECT DISTINCT properties->>'onboarding_id' AS onboarding_id
			FROM analytics.events
			WHERE event_name = 'onboarding_started' AND properties ? 'onboarding_id'
		), onboarding_completions AS (
			SELECT DISTINCT properties->>'onboarding_id' AS onboarding_id
			FROM analytics.events
			WHERE event_name = 'onboarding_completed' AND properties ? 'onboarding_id'
		), viewed_products AS (
			SELECT DISTINCT session_id, properties->>'product_id' AS product_id
			FROM analytics.events
			WHERE event_name = 'product_viewed' AND session_id IS NOT NULL
				AND properties ? 'product_id'
		), clicked_products AS (
			SELECT DISTINCT session_id, product_id::text AS product_id
			FROM commerce.affiliate_clicks
			WHERE source = 'product_detail' AND session_id IS NOT NULL
		), observed_clicks AS (
			SELECT clicked_products.session_id, clicked_products.product_id
			FROM clicked_products
			JOIN viewed_products USING (session_id, product_id)
		), counts AS (
			SELECT
				(SELECT count(*) FROM identity.users WHERE deleted_at IS NULL) AS users,
				(SELECT count(*) FROM recommendation.recommendation_sessions) AS recommendation_sessions,
				(SELECT count(*) FROM onboarding_starts) AS onboarding_started,
				(SELECT count(*) FROM onboarding_completions
					JOIN onboarding_starts USING (onboarding_id)) AS onboarding_completed,
				(SELECT count(*) FROM analytics.events WHERE event_name = 'product_viewed') AS product_views,
				(SELECT count(*) FROM commerce.affiliate_clicks) AS affiliate_clicks,
				(SELECT count(*) FROM viewed_products) AS viewed_product_sessions,
				(SELECT count(*) FROM observed_clicks) AS clicked_product_sessions
		)
		SELECT users, recommendation_sessions, onboarding_started, onboarding_completed,
			CASE WHEN onboarding_started = 0 THEN NULL
				ELSE round(onboarding_completed::numeric * 100 / onboarding_started, 2)::double precision END,
			product_views, affiliate_clicks,
			CASE WHEN viewed_product_sessions = 0 THEN NULL
				ELSE round(clicked_product_sessions::numeric * 100 / viewed_product_sessions, 2)::double precision END
		FROM counts`).Scan(
		&summary.Users, &summary.RecommendationSessions, &summary.OnboardingStarted,
		&summary.OnboardingCompleted, &summary.RecommendationCompletion,
		&summary.ProductViews, &summary.AffiliateClicks, &summary.AffiliateCTR,
	)
}

func (repository *Repository) loadRanking(
	ctx context.Context,
	target *[]domain.RankedEntity,
	name string,
	query string,
	limit int,
) error {
	rows, err := repository.pool.Query(ctx, query, limit)
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

func (repository *Repository) loadTrafficSources(ctx context.Context, target *[]domain.TrafficSource, limit int) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT traffic_source, count(DISTINCT session_id)
		FROM analytics.events
		WHERE event_name = 'page_view' AND traffic_source IS NOT NULL
			AND session_id IS NOT NULL
		GROUP BY traffic_source
		ORDER BY count(*) DESC, traffic_source
		LIMIT $1`, limit)
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
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate traffic sources: %w", err)
	}
	return nil
}
