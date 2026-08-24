package adminpostgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	admin "rigmark/internal/modules/admin/domain"
	"rigmark/internal/modules/admin/ports"
	identity "rigmark/internal/modules/identity/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) Dashboard(ctx context.Context) (admin.Dashboard, error) {
	var dashboard admin.Dashboard
	err := repository.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM catalog.products),
			(SELECT count(*) FROM catalog.products WHERE status = 'published'),
			(SELECT count(*) FROM commerce.merchant_offers),
			(SELECT count(*) FROM commerce.merchant_offers WHERE is_active),
			(SELECT count(*) FROM identity.users WHERE deleted_at IS NULL),
			(SELECT count(*) FROM recommendation.recommendations),
			(SELECT count(*) FROM recommendation.recommendation_sessions),
			(SELECT count(*) FROM recommendation.recommendation_sessions WHERE status = 'completed'),
			(SELECT count(*) FROM analytics.events WHERE is_reportable AND event_name = 'product_viewed'),
			(SELECT count(*) FROM commerce.affiliate_clicks),
			(SELECT count(*) FROM analytics.events WHERE is_reportable AND event_name = 'product_saved'),
			(SELECT count(*) FROM analytics.events WHERE is_reportable AND event_name = 'setup_saved')`).Scan(
		&dashboard.Counts.Products,
		&dashboard.Counts.Published,
		&dashboard.Counts.Offers,
		&dashboard.Counts.ActiveOffers,
		&dashboard.Counts.Users,
		&dashboard.Counts.Recommendations,
		&dashboard.Analytics.RecommendationStarts,
		&dashboard.Analytics.CompletedRecommendations,
		&dashboard.Analytics.ProductViews,
		&dashboard.Analytics.AffiliateClicks,
		&dashboard.Analytics.SavedProducts,
		&dashboard.Analytics.SavedSetups,
	)
	if err != nil {
		return admin.Dashboard{}, fmt.Errorf("load admin dashboard: %w", err)
	}
	if err := repository.dashboardReadiness(ctx, &dashboard.Readiness); err != nil {
		return admin.Dashboard{}, err
	}
	return dashboard, nil
}

// A published product earns nothing without an active offer carrying an active
// affiliate link. Counting the gap is what turns the dashboard from inventory
// totals into a work list.
func (repository *Repository) dashboardReadiness(ctx context.Context, readiness *admin.Readiness) error {
	const summary = `
		WITH published AS (
			SELECT id FROM catalog.products WHERE status = 'published'
		), offered AS (
			SELECT DISTINCT product_id FROM commerce.merchant_offers WHERE is_active
		), linked AS (
			SELECT DISTINCT offers.product_id
			FROM commerce.merchant_offers offers
			JOIN commerce.affiliate_links links
			  ON links.merchant_offer_id = offers.id AND links.is_active
			WHERE offers.is_active
		)
		SELECT
			(SELECT count(*) FROM published),
			(SELECT count(*) FROM published WHERE id NOT IN (SELECT product_id FROM offered)),
			(SELECT count(*) FROM published WHERE id IN (SELECT product_id FROM offered)
				AND id NOT IN (SELECT product_id FROM linked)),
			(SELECT count(*) FROM published WHERE id IN (SELECT product_id FROM linked)),
			(SELECT count(*) FROM commerce.provider_configurations),
			(SELECT count(*) FROM editorial.entries WHERE published_at IS NOT NULL)`
	if err := repository.pool.QueryRow(ctx, summary).Scan(
		&readiness.PublishedProducts, &readiness.WithoutActiveOffer,
		&readiness.WithoutAffiliateLink, &readiness.EarningReady,
		&readiness.CommerceProviders, &readiness.PublishedContent,
	); err != nil {
		return fmt.Errorf("load monetization readiness: %w", err)
	}

	// Bounded on purpose: the dashboard shows the head of the work list, and
	// the full inventory belongs on the Products page.
	const blocked = `
		WITH offered AS (
			SELECT DISTINCT product_id FROM commerce.merchant_offers WHERE is_active
		), linked AS (
			SELECT DISTINCT offers.product_id
			FROM commerce.merchant_offers offers
			JOIN commerce.affiliate_links links
			  ON links.merchant_offer_id = offers.id AND links.is_active
			WHERE offers.is_active
		)
		SELECT products.id, products.name, products.slug,
			CASE WHEN products.id NOT IN (SELECT product_id FROM offered)
				THEN 'no_active_offer' ELSE 'no_affiliate_link' END
		FROM catalog.products products
		WHERE products.status = 'published'
		  AND products.id NOT IN (SELECT product_id FROM linked)
		ORDER BY products.name
		LIMIT 25`
	rows, err := repository.pool.Query(ctx, blocked)
	if err != nil {
		return fmt.Errorf("load blocked products: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item admin.BlockedProduct
		var reason string
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &reason); err != nil {
			return fmt.Errorf("scan blocked product: %w", err)
		}
		item.Reason = admin.BlockedReason(reason)
		readiness.Blocked = append(readiness.Blocked, item)
	}
	return rows.Err()
}

func audit(ctx context.Context, tx pgx.Tx, actor identity.UserID, action, entityType, entityID string, changes map[string]string) error {
	payload, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("encode audit changes: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO admin.audit_log (actor_user_id, action, entity_type, entity_id, changes)
		VALUES ($1, $2, $3, $4, $5)`, actor, action, entityType, entityID, payload); err != nil {
		return fmt.Errorf("write admin audit log: %w", err)
	}
	return nil
}

func finishMutation(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil {
		_ = tx.Rollback(ctx)
		if isConflict(err) {
			return ports.ErrConflict
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit admin mutation: %w", err)
	}
	return nil
}

func isConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	}
	return err
}

var _ ports.Repository = (*Repository)(nil)
