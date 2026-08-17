package analyticspostgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) Record(ctx context.Context, event domain.Event) error {
	properties, err := json.Marshal(event.Properties)
	if err != nil {
		return fmt.Errorf("encode analytics properties: %w", err)
	}
	_, err = repository.pool.Exec(ctx, `
		INSERT INTO analytics.events (
			event_name, schema_version, user_id, recommendation_session_id,
			anonymous_id, session_id, request_id, surface, properties,
			page_path, traffic_source, traffic_medium, campaign, referrer_host,
			consent_state, occurred_at
		) VALUES ($1,$2,$3::uuid,$4::uuid,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13,$14,$15,$16)`,
		event.Name, event.SchemaVersion, event.UserID, event.RecommendationSessionID,
		event.AnonymousID, event.SessionID, event.RequestID, event.Surface,
		properties, event.PagePath, event.TrafficSource, event.TrafficMedium,
		event.Campaign, event.ReferrerHost, event.ConsentState, event.OccurredAt)
	if err != nil {
		return fmt.Errorf("record analytics event: %w", err)
	}
	return nil
}

var _ ports.EventRecorder = (*Repository)(nil)
var _ ports.ReportingRepository = (*Repository)(nil)
