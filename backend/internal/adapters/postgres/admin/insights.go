package adminpostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	admin "rigmark/internal/modules/admin/domain"
	"rigmark/internal/modules/admin/ports"
)

func (repository *Repository) ListRecommendations(ctx context.Context, limit, offset int) (admin.Page[admin.Recommendation], error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), recommendations.id, sessions.id, users.email,
			sessions.primary_goal, sessions.experience_level, sessions.status,
			recommendations.objective_score, recommendations.total_price_minor,
			recommendations.currency, recommendations.policy_version,
			recommendations.engine_version, recommendations.created_at
		FROM recommendation.recommendations AS recommendations
		JOIN recommendation.recommendation_sessions AS sessions ON sessions.id=recommendations.session_id
		LEFT JOIN identity.users AS users ON users.id=sessions.user_id
		ORDER BY recommendations.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return admin.Page[admin.Recommendation]{}, fmt.Errorf("list admin recommendations: %w", err)
	}
	defer rows.Close()
	page := admin.Page[admin.Recommendation]{Items: make([]admin.Recommendation, 0)}
	for rows.Next() {
		var item admin.Recommendation
		var email sql.NullString
		if err := rows.Scan(&page.Total, &item.ID, &item.SessionID, &email, &item.Goal,
			&item.Experience, &item.SessionStatus, &item.ObjectiveScore,
			&item.TotalPriceMinor, &item.Currency, &item.PolicyVersion,
			&item.EngineVersion, &item.CreatedAt); err != nil {
			return admin.Page[admin.Recommendation]{}, fmt.Errorf("scan recommendation: %w", err)
		}
		if email.Valid {
			value := email.String
			item.UserEmail = &value
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (repository *Repository) GetRecommendation(ctx context.Context, id string) (admin.RecommendationDetail, error) {
	var detail admin.RecommendationDetail
	var email sql.NullString
	err := repository.pool.QueryRow(ctx, `
		SELECT recommendations.id, sessions.id, users.email, sessions.primary_goal,
			sessions.experience_level, sessions.status, recommendations.objective_score,
			recommendations.total_price_minor, recommendations.currency,
			recommendations.policy_version, recommendations.engine_version,
			recommendations.created_at, recommendations.goal_match_score,
			recommendations.budget_match_score, recommendations.space_match_score,
			recommendations.experience_match_score, recommendations.preference_match_score,
			recommendations.quality_score, recommendations.value_score,
			recommendations.durability_score, recommendations.compatibility_score,
			recommendations.portability_score, recommendations.noise_score
		FROM recommendation.recommendations AS recommendations
		JOIN recommendation.recommendation_sessions AS sessions ON sessions.id=recommendations.session_id
		LEFT JOIN identity.users AS users ON users.id=sessions.user_id
		WHERE recommendations.id=$1`, id).Scan(&detail.Recommendation.ID,
		&detail.Recommendation.SessionID, &email, &detail.Recommendation.Goal,
		&detail.Recommendation.Experience, &detail.Recommendation.SessionStatus,
		&detail.Recommendation.ObjectiveScore, &detail.Recommendation.TotalPriceMinor,
		&detail.Recommendation.Currency, &detail.Recommendation.PolicyVersion,
		&detail.Recommendation.EngineVersion, &detail.Recommendation.CreatedAt,
		&detail.Scores.Goal, &detail.Scores.Budget, &detail.Scores.Space,
		&detail.Scores.Experience, &detail.Scores.Preference, &detail.Scores.Quality,
		&detail.Scores.Value, &detail.Scores.Durability, &detail.Scores.Compatibility,
		&detail.Scores.Portability, &detail.Scores.Noise)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.RecommendationDetail{}, ports.ErrNotFound
	}
	if err != nil {
		return admin.RecommendationDetail{}, fmt.Errorf("get recommendation: %w", err)
	}
	if email.Valid {
		value := email.String
		detail.Recommendation.UserEmail = &value
	}
	rows, err := repository.pool.Query(ctx, `
		SELECT items.id, items.product_id, products.name, items.item_type, items.rank,
			items.quantity, items.objective_score, items.reason_code,
			items.reason_summary, items.rejection_code
		FROM recommendation.recommendation_items AS items
		JOIN catalog.products AS products ON products.id=items.product_id
		WHERE items.recommendation_id=$1 ORDER BY items.item_type, items.rank`, id)
	if err != nil {
		return admin.RecommendationDetail{}, fmt.Errorf("list recommendation items: %w", err)
	}
	defer rows.Close()
	itemIDs := make([]string, 0)
	indexes := make(map[string]int)
	for rows.Next() {
		var itemID string
		var item admin.RecommendationItem
		var rejection sql.NullString
		if err := rows.Scan(&itemID, &item.ProductID, &item.ProductName, &item.ItemType,
			&item.Rank, &item.Quantity, &item.ObjectiveScore, &item.ReasonCode,
			&item.ReasonSummary, &rejection); err != nil {
			return admin.RecommendationDetail{}, fmt.Errorf("scan recommendation item: %w", err)
		}
		if rejection.Valid {
			value := rejection.String
			item.RejectionCode = &value
		}
		indexes[itemID] = len(detail.Items)
		itemIDs = append(itemIDs, itemID)
		detail.Items = append(detail.Items, item)
	}
	if err := rows.Err(); err != nil {
		return admin.RecommendationDetail{}, err
	}
	if len(itemIDs) == 0 {
		return detail, nil
	}
	reasons, err := repository.pool.Query(ctx, `
		SELECT recommendation_item_id, code, message, dimension, score
		FROM recommendation.item_reasons
		WHERE recommendation_item_id=ANY($1::uuid[])
		ORDER BY recommendation_item_id, sort_order`, itemIDs)
	if err != nil {
		return admin.RecommendationDetail{}, fmt.Errorf("list recommendation reasons: %w", err)
	}
	defer reasons.Close()
	for reasons.Next() {
		var itemID string
		var reason admin.RecommendationReason
		if err := reasons.Scan(&itemID, &reason.Code, &reason.Message, &reason.Dimension, &reason.Score); err != nil {
			return admin.RecommendationDetail{}, fmt.Errorf("scan recommendation reason: %w", err)
		}
		detail.Items[indexes[itemID]].Reasons = append(detail.Items[indexes[itemID]].Reasons, reason)
	}
	return detail, reasons.Err()
}

func (repository *Repository) ListUsers(ctx context.Context, limit, offset int) (admin.Page[admin.User], error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), users.id, users.email, users.status,
			ARRAY(SELECT role_key FROM identity.user_roles WHERE user_id=users.id ORDER BY role_key),
			users.last_login_at, users.created_at
		FROM identity.users AS users ORDER BY users.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return admin.Page[admin.User]{}, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()
	page := admin.Page[admin.User]{Items: make([]admin.User, 0)}
	for rows.Next() {
		var item admin.User
		var lastLogin sql.NullTime
		if err := rows.Scan(&page.Total, &item.ID, &item.Email, &item.Status,
			&item.Roles, &lastLogin, &item.CreatedAt); err != nil {
			return admin.Page[admin.User]{}, fmt.Errorf("scan admin user: %w", err)
		}
		if lastLogin.Valid {
			value := lastLogin.Time
			item.LastLoginAt = &value
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (repository *Repository) ListEvents(ctx context.Context, name string, limit, offset int) (admin.Page[admin.Event], error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), id, event_name, user_id, anonymous_id, session_id,
			surface, properties, page_path, traffic_source, traffic_medium, campaign,
			referrer_host, consent_state, occurred_at
		FROM analytics.events WHERE $1='' OR event_name=$1
		ORDER BY occurred_at DESC LIMIT $2 OFFSET $3`, name, limit, offset)
	if err != nil {
		return admin.Page[admin.Event]{}, fmt.Errorf("list admin events: %w", err)
	}
	defer rows.Close()
	page := admin.Page[admin.Event]{Items: make([]admin.Event, 0)}
	for rows.Next() {
		var item admin.Event
		var userID, anonymousID, sessionID, pagePath, trafficSource sql.NullString
		var trafficMedium, campaign, referrerHost sql.NullString
		var properties []byte
		if err := rows.Scan(&page.Total, &item.ID, &item.Name, &userID, &anonymousID,
			&sessionID, &item.Surface, &properties, &pagePath, &trafficSource,
			&trafficMedium, &campaign, &referrerHost, &item.ConsentState,
			&item.OccurredAt); err != nil {
			return admin.Page[admin.Event]{}, fmt.Errorf("scan admin event: %w", err)
		}
		if userID.Valid {
			value := userID.String
			item.UserID = &value
		}
		if anonymousID.Valid {
			value := anonymousID.String
			item.AnonymousID = &value
		}
		if sessionID.Valid {
			value := sessionID.String
			item.SessionID = &value
		}
		item.PagePath = nullableStringValue(pagePath)
		item.TrafficSource = nullableStringValue(trafficSource)
		item.TrafficMedium = nullableStringValue(trafficMedium)
		item.Campaign = nullableStringValue(campaign)
		item.ReferrerHost = nullableStringValue(referrerHost)
		if json.Valid(properties) {
			item.Properties = json.RawMessage(properties)
		} else {
			item.Properties = json.RawMessage(`{}`)
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func nullableStringValue(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
