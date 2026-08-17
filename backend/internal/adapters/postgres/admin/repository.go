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
			(SELECT count(*) FROM analytics.events WHERE event_name = 'product_viewed'),
			(SELECT count(*) FROM commerce.affiliate_clicks),
			(SELECT count(*) FROM analytics.events WHERE event_name = 'product_saved'),
			(SELECT count(*) FROM analytics.events WHERE event_name = 'setup_saved')`).Scan(
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
	return dashboard, nil
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
