package recommendationpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

type Repository struct {
	pool *pgxpool.Pool
	// vertical selects which policy this deployment serves. The schema allows
	// one active policy per vertical, so the value is what makes a pivot a
	// configuration change rather than a code change.
	vertical string
}

// DefaultVertical is used when no vertical is configured.
const DefaultVertical = "saas"

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, vertical: DefaultVertical}
}

// NewForVertical builds a repository serving the named vertical's active
// policy. An empty name falls back to DefaultVertical rather than matching no
// policy at all, which would fail every recommendation at runtime.
func NewForVertical(pool *pgxpool.Pool, vertical string) *Repository {
	if vertical == "" {
		vertical = DefaultVertical
	}
	return &Repository{pool: pool, vertical: vertical}
}

func (repository *Repository) GetDraft(ctx context.Context, userID identity.UserID) (ports.Draft, error) {
	var draft ports.Draft
	var goal, experience, currency sql.NullString
	var budget, length, width, height, accessWidth sql.NullInt64
	var apartment sql.NullBool
	err := repository.pool.QueryRow(ctx, `
		SELECT current_step, primary_goal, experience_level, budget_minor, currency,
			space_length_mm, space_width_mm, space_height_mm, access_width_mm, apartment_living,
			training_preferences, priorities, free_text, updated_at
		FROM recommendation.drafts WHERE user_id = $1`, userID).Scan(
		&draft.CurrentStep, &goal, &experience, &budget, &currency,
		&length, &width, &height, &accessWidth, &apartment,
		&draft.TrainingPreferences, &draft.Priorities, &draft.FreeText, &draft.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Draft{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.Draft{}, fmt.Errorf("get recommendation draft: %w", err)
	}
	assignDraftOptionals(&draft, goal, experience, budget, currency, length, width, height, accessWidth, apartment)
	equipment, err := repository.listDraftEquipment(ctx, userID)
	if err != nil {
		return ports.Draft{}, err
	}
	draft.ExistingEquipment = equipment
	return draft, nil
}

func (repository *Repository) SaveDraft(
	ctx context.Context,
	userID identity.UserID,
	draft ports.Draft,
) (ports.Draft, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return ports.Draft{}, fmt.Errorf("begin save recommendation draft: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var goal, experience any
	if draft.Goal != nil {
		goal = string(*draft.Goal)
	}
	if draft.Experience != nil {
		experience = string(*draft.Experience)
	}
	var length, width, height, accessWidth, apartment any
	if draft.AvailableSpace != nil {
		length, width, height = draft.AvailableSpace.LengthMM, draft.AvailableSpace.WidthMM, draft.AvailableSpace.HeightMM
		accessWidth = draft.AvailableSpace.AccessWidthMM
		apartment = draft.AvailableSpace.ApartmentLiving
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO recommendation.drafts (
			user_id, current_step, primary_goal, experience_level, budget_minor, currency,
			space_length_mm, space_width_mm, space_height_mm, access_width_mm, apartment_living,
			training_preferences, priorities, free_text
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (user_id) DO UPDATE SET
			current_step = EXCLUDED.current_step, primary_goal = EXCLUDED.primary_goal,
			experience_level = EXCLUDED.experience_level, budget_minor = EXCLUDED.budget_minor,
			currency = EXCLUDED.currency, space_length_mm = EXCLUDED.space_length_mm,
			space_width_mm = EXCLUDED.space_width_mm, space_height_mm = EXCLUDED.space_height_mm,
			access_width_mm = EXCLUDED.access_width_mm,
			apartment_living = EXCLUDED.apartment_living,
			training_preferences = EXCLUDED.training_preferences, priorities = EXCLUDED.priorities,
			free_text = EXCLUDED.free_text, updated_at = now()
		RETURNING updated_at`, userID, draft.CurrentStep, goal, experience,
		draft.BudgetMinor, draft.Currency, length, width, height, accessWidth, apartment,
		draft.TrainingPreferences, draft.Priorities, draft.FreeText).Scan(&draft.UpdatedAt)
	if err != nil {
		return ports.Draft{}, fmt.Errorf("upsert recommendation draft: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM recommendation.draft_existing_equipment WHERE user_id = $1`, userID); err != nil {
		return ports.Draft{}, fmt.Errorf("replace draft equipment: %w", err)
	}
	for index, equipment := range draft.ExistingEquipment {
		if _, err = tx.Exec(ctx, `
			INSERT INTO recommendation.draft_existing_equipment (user_id, name, category_slug, sort_order)
			VALUES ($1,$2,$3,$4)`, userID, equipment.Name, equipment.CategorySlug, index); err != nil {
			return ports.Draft{}, fmt.Errorf("insert draft equipment: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return ports.Draft{}, fmt.Errorf("commit recommendation draft: %w", err)
	}
	return draft, nil
}

func (repository *Repository) DeleteDraft(ctx context.Context, userID identity.UserID) error {
	_, err := repository.pool.Exec(ctx, `DELETE FROM recommendation.drafts WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete recommendation draft: %w", err)
	}
	return nil
}

func (repository *Repository) SaveResult(
	ctx context.Context,
	userID identity.UserID,
	input domain.Input,
	result domain.Result,
	candidates []domain.CandidateSnapshot,
) (ports.SavedResult, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return ports.SavedResult{}, fmt.Errorf("begin save recommendation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	spaceLength, spaceWidth, spaceHeight := optionalDimensionValues(
		input.AvailableSpace.LengthMM,
		input.AvailableSpace.WidthMM,
		input.AvailableSpace.HeightMM,
	)
	var sessionID domain.SessionID
	err = tx.QueryRow(ctx, `
		INSERT INTO recommendation.recommendation_sessions (
			user_id, status, primary_goal, experience_level, budget_minor, currency,
			space_length_mm, space_width_mm, space_height_mm, access_width_mm, apartment_living,
			training_preferences, priorities, free_text, completed_at
		) VALUES ($1,'completed',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,now()) RETURNING id`,
		userID, input.Goal, input.Experience, input.Budget.AmountMinor, input.Budget.Currency,
		spaceLength, spaceWidth, spaceHeight,
		input.AvailableSpace.AccessWidthMM, input.AvailableSpace.ApartmentLiving,
		input.TrainingPreferences, input.Priorities, input.FreeText,
	).Scan(&sessionID)
	if err != nil {
		return ports.SavedResult{}, fmt.Errorf("insert recommendation session: %w", err)
	}
	for index, equipment := range input.ExistingEquipment {
		if _, err = tx.Exec(ctx, `INSERT INTO recommendation.session_existing_equipment
			(session_id, name, category_slug, capabilities, redundancy_groups, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			sessionID, equipment.Name, equipment.CategorySlug,
			capabilityStrings(equipment.Capabilities), equipment.RedundancyGroups, index); err != nil {
			return ports.SavedResult{}, fmt.Errorf("insert session equipment: %w", err)
		}
	}

	var recommendationID domain.RecommendationID
	b := result.Breakdown
	err = tx.QueryRow(ctx, `
		INSERT INTO recommendation.recommendations (
			session_id, policy_version, engine_version, objective_score, total_price_minor,
			currency, result_fingerprint, goal_match_score, budget_match_score, space_match_score,
			experience_match_score, preference_match_score, quality_score, value_score,
			durability_score, compatibility_score, portability_score, noise_score
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING id`,
		sessionID, result.PolicyVersion, result.EngineVersion, result.ObjectiveScore,
		result.TotalCost.AmountMinor, result.TotalCost.Currency, result.InputFingerprint,
		b.GoalMatch, b.BudgetMatch, b.SpaceMatch, b.ExperienceMatch, b.PreferenceMatch,
		b.Quality, b.Value, b.Durability, b.Compatibility, b.Portability, b.Noise,
	).Scan(&recommendationID)
	if err != nil {
		return ports.SavedResult{}, fmt.Errorf("insert recommendation: %w", err)
	}
	if err = insertCandidateSnapshots(ctx, tx, recommendationID, candidates); err != nil {
		return ports.SavedResult{}, err
	}
	if err = insertResultItems(ctx, tx, recommendationID, result); err != nil {
		return ports.SavedResult{}, err
	}

	var setupID planning.SetupID
	setupName := fmt.Sprintf("Personalized software stack · %.8s", recommendationID)
	err = tx.QueryRow(ctx, `INSERT INTO planning.setups
		(user_id, source_recommendation_id, name, description, currency)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`, userID, recommendationID, setupName,
		"A deterministic UNSOLERO stack based on your goals, constraints and budget.", result.TotalCost.Currency,
	).Scan(&setupID)
	if err != nil {
		return ports.SavedResult{}, fmt.Errorf("insert setup: %w", err)
	}
	for _, item := range result.Selected {
		if _, err = tx.Exec(ctx, `INSERT INTO planning.setup_items
			(setup_id, product_id, quantity, purchase_status) VALUES ($1,$2,$3,'planned')`,
			setupID, item.Product.Candidate.ProductID, item.Quantity); err != nil {
			return ports.SavedResult{}, fmt.Errorf("insert setup item: %w", err)
		}
	}
	if _, err = tx.Exec(ctx, `DELETE FROM recommendation.drafts WHERE user_id = $1`, userID); err != nil {
		return ports.SavedResult{}, fmt.Errorf("clear completed draft: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return ports.SavedResult{}, fmt.Errorf("commit recommendation: %w", err)
	}
	return ports.SavedResult{RecommendationID: recommendationID, SetupID: setupID, SetupName: setupName}, nil
}

func insertCandidateSnapshots(
	ctx context.Context,
	tx pgx.Tx,
	recommendationID domain.RecommendationID,
	candidates []domain.CandidateSnapshot,
) error {
	for _, candidate := range candidates {
		scores := candidate.Scores
		space := candidate.Space
		length, width, height := optionalDimensionValues(
			candidate.Dimensions.LengthMM,
			candidate.Dimensions.WidthMM,
			candidate.Dimensions.HeightMM,
		)
		storageLength, storageWidth, storageHeight := envelopeValues(space.StorageFootprint)
		operating := clearanceValues(space.OperatingClearance)
		safety := clearanceValues(space.SafetyClearance)
		var overlap any
		if space.OverlapGroup != "" {
			overlap = space.OverlapGroup
		}
		_, err := tx.Exec(ctx, `INSERT INTO recommendation.candidate_snapshots (
			recommendation_id, product_id, fact_revision_id, score_revision_id,
			name, category_slug, price_minor, currency, length_mm, width_mm, height_mm,
			quality_score, value_score, durability_score, beginner_score, advanced_score,
			apartment_score, noise_score, portability_score, policy_version,
			capabilities,requirements,compatible_with,incompatible_with,preference_tags,redundancy_groups,
			storage_length_mm,storage_width_mm,storage_height_mm,operating_clearance_mm,safety_clearance_mm,
			minimum_room_height_mm,minimum_access_width_mm,overlap_group,
			requires_storage_footprint,requires_operating_clearance,requires_safety_clearance,requires_access_width
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38)`,
			recommendationID, candidate.ProductID, candidate.FactRevisionID,
			candidate.ScoreRevisionID, candidate.Name, candidate.CategorySlug,
			candidate.Price.AmountMinor, candidate.Price.Currency,
			length, width, height, scores.Quality, scores.Value,
			scores.Durability, scores.Beginner, scores.Advanced, scores.Apartment,
			scores.Noise, scores.Portability, candidate.PolicyVersion,
			capabilityStrings(candidate.Capabilities), capabilityStrings(candidate.Requires),
			capabilityStrings(candidate.CompatibleWith), capabilityStrings(candidate.IncompatibleWith),
			preferenceStrings(candidate.PreferenceTags), candidate.RedundancyGroups,
			storageLength, storageWidth, storageHeight, operating, safety,
			space.MinimumRoomHeightMM, space.MinimumAccessWidthMM, overlap,
			space.RequiresStorageFootprint, space.RequiresOperatingClearance,
			space.RequiresSafetyClearance, space.RequiresAccessWidth)
		if err != nil {
			return fmt.Errorf("insert recommendation candidate snapshot: %w", err)
		}
		for _, support := range candidate.GoalSupport {
			if _, err = tx.Exec(ctx, `INSERT INTO recommendation.candidate_snapshot_goal_support
				(recommendation_id,product_id,goal_key,match_score) VALUES ($1,$2,$3,$4)`,
				recommendationID, candidate.ProductID, support.Goal, support.Score); err != nil {
				return fmt.Errorf("insert candidate goal support: %w", err)
			}
		}
	}
	return nil
}

// optionalDimensionValues translates the domain's zero-value representation
// for non-physical products and non-spatial recommendation inputs to SQL NULL.
// Partial or negative values are intentionally passed through so the database
// constraints reject corrupt physical data rather than silently erasing it.
func optionalDimensionValues(length, width, height int64) (any, any, any) {
	if length == 0 && width == 0 && height == 0 {
		return nil, nil, nil
	}
	return length, width, height
}

func envelopeValues(value *domain.SpatialEnvelope) (any, any, any) {
	if value == nil {
		return nil, nil, nil
	}
	return value.LengthMM, value.WidthMM, value.HeightMM
}

func clearanceValues(value *domain.Clearance) any {
	if value == nil {
		return nil
	}
	return []int64{value.FrontMM, value.BackMM, value.LeftMM, value.RightMM, value.TopMM}
}

func capabilityStrings(values []domain.Capability) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func preferenceStrings(values []domain.TrainingPreference) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func capabilitiesFromStrings(values []string) []domain.Capability {
	result := make([]domain.Capability, len(values))
	for index, value := range values {
		result[index] = domain.Capability(value)
	}
	return result
}

func clearanceFromValues(values []int64) *domain.Clearance {
	if len(values) != 5 {
		return nil
	}
	return &domain.Clearance{FrontMM: values[0], BackMM: values[1], LeftMM: values[2], RightMM: values[3], TopMM: values[4]}
}

func (repository *Repository) loadCandidateGoalSupport(
	ctx context.Context,
	recommendationID domain.RecommendationID,
) (map[catalog.ProductID][]domain.GoalSupport, error) {
	rows, err := repository.pool.Query(ctx, `SELECT product_id,goal_key,match_score
		FROM recommendation.candidate_snapshot_goal_support
		WHERE recommendation_id=$1 ORDER BY product_id,goal_key`, recommendationID)
	if err != nil {
		return nil, fmt.Errorf("load candidate goal support: %w", err)
	}
	defer rows.Close()
	result := make(map[catalog.ProductID][]domain.GoalSupport)
	for rows.Next() {
		var productID catalog.ProductID
		var support domain.GoalSupport
		if err = rows.Scan(&productID, &support.Goal, &support.Score); err != nil {
			return nil, err
		}
		result[productID] = append(result[productID], support)
	}
	return result, rows.Err()
}

func (repository *Repository) ListSetups(ctx context.Context, userID identity.UserID, limit, offset int) (ports.SetupPage, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), s.id, s.name, count(si.id), r.total_price_minor, s.currency,
			r.objective_score, s.created_at, s.updated_at
		FROM planning.setups s
		JOIN recommendation.recommendations r ON r.id = s.source_recommendation_id
		LEFT JOIN planning.setup_items si ON si.setup_id = s.id AND si.purchase_status <> 'removed'
		WHERE s.user_id = $1
		GROUP BY s.id, r.total_price_minor, r.objective_score
		ORDER BY s.updated_at DESC, s.id DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return ports.SetupPage{}, fmt.Errorf("list saved setups: %w", err)
	}
	defer rows.Close()
	page := ports.SetupPage{Setups: make([]ports.SetupSummary, 0)}
	for rows.Next() {
		var setup ports.SetupSummary
		if err = rows.Scan(&page.Total, &setup.ID, &setup.Name, &setup.ItemCount, &setup.TotalCostMinor,
			&setup.Currency, &setup.ObjectiveScore, &setup.CreatedAt, &setup.UpdatedAt); err != nil {
			return ports.SetupPage{}, fmt.Errorf("scan saved setup: %w", err)
		}
		page.Setups = append(page.Setups, setup)
	}
	if err = rows.Err(); err != nil {
		return ports.SetupPage{}, fmt.Errorf("read saved setups: %w", err)
	}
	if len(page.Setups) == 0 && offset > 0 {
		if err := repository.pool.QueryRow(ctx, `SELECT count(*) FROM planning.setups WHERE user_id=$1`, userID).Scan(&page.Total); err != nil {
			return ports.SetupPage{}, fmt.Errorf("count saved setups: %w", err)
		}
	}
	return page, nil
}

func (repository *Repository) GetResultBySetupID(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
) (ports.PersistedResult, error) {
	var stored ports.PersistedResult
	stored.SetupID = setupID
	var input domain.Input
	var result domain.Result
	var goal, experience string
	var spaceLength, spaceWidth, spaceHeight sql.NullInt64
	b := &result.Breakdown
	err := repository.pool.QueryRow(ctx, `
		SELECT r.id, s.name, rs.primary_goal, rs.experience_level, rs.budget_minor, rs.currency,
			rs.space_length_mm, rs.space_width_mm, rs.space_height_mm, rs.access_width_mm, rs.apartment_living,
			rs.training_preferences, rs.priorities, rs.free_text,
			r.policy_version, r.engine_version, r.objective_score, r.total_price_minor,
			r.currency, r.result_fingerprint, r.goal_match_score, r.budget_match_score,
			r.space_match_score, r.experience_match_score, r.preference_match_score,
			r.quality_score, r.value_score, r.durability_score, r.compatibility_score,
			r.portability_score, r.noise_score, r.created_at
		FROM planning.setups s
		JOIN recommendation.recommendations r ON r.id = s.source_recommendation_id
		JOIN recommendation.recommendation_sessions rs ON rs.id = r.session_id
		WHERE s.id = $1 AND s.user_id = $2`, setupID, userID).Scan(
		&stored.RecommendationID, &stored.SetupName, &goal, &experience, &input.Budget.AmountMinor, &input.Budget.Currency,
		&spaceLength, &spaceWidth, &spaceHeight,
		&input.AvailableSpace.AccessWidthMM, &input.AvailableSpace.ApartmentLiving,
		&input.TrainingPreferences, &input.Priorities, &input.FreeText,
		&result.PolicyVersion, &result.EngineVersion, &result.ObjectiveScore,
		&result.TotalCost.AmountMinor, &result.TotalCost.Currency, &result.InputFingerprint,
		&b.GoalMatch, &b.BudgetMatch, &b.SpaceMatch, &b.ExperienceMatch, &b.PreferenceMatch,
		&b.Quality, &b.Value, &b.Durability, &b.Compatibility, &b.Portability, &b.Noise, &stored.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.PersistedResult{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.PersistedResult{}, fmt.Errorf("get saved setup: %w", err)
	}
	input.AvailableSpace.LengthMM = spaceLength.Int64
	input.AvailableSpace.WidthMM = spaceWidth.Int64
	input.AvailableSpace.HeightMM = spaceHeight.Int64
	input.Goal, input.Experience = planning.Goal(goal), planning.ExperienceLevel(experience)
	equipment, err := repository.listSessionEquipment(ctx, stored.RecommendationID)
	if err != nil {
		return ports.PersistedResult{}, err
	}
	input.ExistingEquipment = equipment
	if err = repository.loadResultItems(ctx, stored.RecommendationID, &result); err != nil {
		return ports.PersistedResult{}, err
	}
	if len(result.Selected) > 0 {
		result.Status = domain.ResultComplete
	} else {
		result.Status = domain.ResultNoSuitableProducts
	}
	stored.Input, stored.Result = input, result
	return stored, nil
}

func (repository *Repository) RenameSetup(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
	name string,
) error {
	result, err := repository.pool.Exec(ctx, `
		UPDATE planning.setups
		SET name = $3, updated_at = now()
		WHERE id = $1 AND user_id = $2`, setupID, userID, name)
	if err != nil {
		return fmt.Errorf("rename setup: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (repository *Repository) DeleteSetup(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM planning.setups WHERE id = $1 AND user_id = $2`, setupID, userID)
	if err != nil {
		return fmt.Errorf("delete setup: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func insertResultItems(ctx context.Context, tx pgx.Tx, recommendationID domain.RecommendationID, result domain.Result) error {
	for _, item := range result.Selected {
		if err := insertRankedItem(ctx, tx, recommendationID, "selected", item.Rank, item.Quantity,
			item.Product, nil, nil); err != nil {
			return err
		}
	}
	typeRanks := map[string]int{}
	for _, alternative := range result.Alternatives {
		itemType := "premium_alternative"
		if alternative.Type == domain.AlternativeCheaper {
			itemType = "cheaper_alternative"
		}
		typeRanks[itemType]++
		target := alternative.ForProductID
		if err := insertRankedItem(ctx, tx, recommendationID, itemType, typeRanks[itemType], 1,
			alternative.Product, &target, nil); err != nil {
			return err
		}
	}
	for index, rejected := range result.Rejected {
		ranked := domain.RankedProduct{Candidate: rejected.Candidate}
		code := rejected.Code
		ranked.Reasons = []domain.Reason{{Code: rejected.Code, Message: rejected.Message, Dimension: "eligibility", Score: 0}}
		if err := insertRankedItem(ctx, tx, recommendationID, "rejected", index+1, 1,
			ranked, nil, &code); err != nil {
			return err
		}
	}
	return nil
}

func insertRankedItem(
	ctx context.Context, tx pgx.Tx, recommendationID domain.RecommendationID,
	itemType string, rank, quantity int, product domain.RankedProduct,
	alternativeFor *catalog.ProductID, rejectionCode *string,
) error {
	reasonCode, reasonSummary := "recommendation.match", "Selected by the deterministic recommendation policy"
	if len(product.Reasons) > 0 {
		reasonCode, reasonSummary = product.Reasons[0].Code, product.Reasons[0].Message
	}
	b := product.Breakdown
	var itemID domain.ItemID
	err := tx.QueryRow(ctx, `INSERT INTO recommendation.recommendation_items (
		recommendation_id, product_id, item_type, rank, quantity, unit_price_minor, currency,
		objective_score, reason_code, reason_summary, rejection_code, alternative_for_product_id,
		goal_match_score, budget_match_score, space_match_score, experience_match_score,
		preference_match_score, quality_score, value_score, durability_score,
		compatibility_score, portability_score, noise_score
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
	RETURNING id`, recommendationID, product.Candidate.ProductID, itemType, rank, quantity,
		product.Candidate.Price.AmountMinor, product.Candidate.Price.Currency, product.ObjectiveScore,
		reasonCode, reasonSummary, rejectionCode, alternativeFor,
		b.GoalMatch, b.BudgetMatch, b.SpaceMatch, b.ExperienceMatch, b.PreferenceMatch,
		b.Quality, b.Value, b.Durability, b.Compatibility, b.Portability, b.Noise).Scan(&itemID)
	if err != nil {
		return fmt.Errorf("insert recommendation item: %w", err)
	}
	for index, reason := range product.Reasons {
		if _, err = tx.Exec(ctx, `INSERT INTO recommendation.item_reasons
			(recommendation_item_id, sort_order, code, message, dimension, score)
			VALUES ($1,$2,$3,$4,$5,$6)`, itemID, index, reason.Code, reason.Message,
			reason.Dimension, reason.Score); err != nil {
			return fmt.Errorf("insert recommendation reason: %w", err)
		}
	}
	return nil
}

func (repository *Repository) loadResultItems(
	ctx context.Context,
	recommendationID domain.RecommendationID,
	result *domain.Result,
) error {
	rows, err := repository.pool.Query(ctx, `SELECT items.id, items.product_id, items.item_type, items.rank, items.quantity,
		items.unit_price_minor, items.currency, items.objective_score, items.rejection_code, items.alternative_for_product_id,
		items.goal_match_score, items.budget_match_score, items.space_match_score, items.experience_match_score,
		items.preference_match_score, items.quality_score, items.value_score, items.durability_score,
		items.compatibility_score, items.portability_score, items.noise_score,
		snapshots.fact_revision_id, snapshots.score_revision_id, snapshots.name,
		snapshots.category_slug, snapshots.length_mm, snapshots.width_mm, snapshots.height_mm,
		snapshots.quality_score, snapshots.value_score, snapshots.durability_score,
		snapshots.beginner_score, snapshots.advanced_score, snapshots.apartment_score,
		snapshots.noise_score, snapshots.portability_score, snapshots.policy_version,
		snapshots.capabilities,snapshots.requirements,snapshots.compatible_with,snapshots.incompatible_with,
		snapshots.preference_tags,snapshots.redundancy_groups,
		snapshots.storage_length_mm,snapshots.storage_width_mm,snapshots.storage_height_mm,
		snapshots.operating_clearance_mm,snapshots.safety_clearance_mm,
		snapshots.minimum_room_height_mm,snapshots.minimum_access_width_mm,snapshots.overlap_group,
		snapshots.requires_storage_footprint,snapshots.requires_operating_clearance,
		snapshots.requires_safety_clearance,snapshots.requires_access_width
		FROM recommendation.recommendation_items items
		LEFT JOIN recommendation.candidate_snapshots snapshots
		  ON snapshots.recommendation_id = items.recommendation_id
		 AND snapshots.product_id = items.product_id
		WHERE items.recommendation_id = $1
		ORDER BY items.item_type, items.rank`, recommendationID)
	if err != nil {
		return fmt.Errorf("list recommendation items: %w", err)
	}
	defer rows.Close()
	type loadedItem struct {
		id                     string
		itemType               string
		rank, quantity         int
		rejection, alternative sql.NullString
		product                domain.RankedProduct
	}
	loaded := make([]loadedItem, 0)
	itemIDs := make([]string, 0)
	for rows.Next() {
		var item loadedItem
		var factRevisionID, scoreRevisionID, name, categorySlug sql.NullString
		var length, width, height sql.NullInt64
		var quality, value, durability, beginner, advanced, apartment, noise, portability sql.NullInt16
		var policyVersion sql.NullString
		var capabilities, requirements, compatible, incompatible, preferences, redundancy []string
		var storageLength, storageWidth, storageHeight, roomHeight, accessWidth sql.NullInt64
		var operating, safety []int64
		var overlap sql.NullString
		var requiresStorage, requiresOperating, requiresSafety, requiresAccess bool
		b := &item.product.Breakdown
		if err = rows.Scan(&item.id, &item.product.Candidate.ProductID, &item.itemType, &item.rank, &item.quantity,
			&item.product.Candidate.Price.AmountMinor, &item.product.Candidate.Price.Currency,
			&item.product.ObjectiveScore, &item.rejection, &item.alternative,
			&b.GoalMatch, &b.BudgetMatch, &b.SpaceMatch, &b.ExperienceMatch, &b.PreferenceMatch,
			&b.Quality, &b.Value, &b.Durability, &b.Compatibility, &b.Portability, &b.Noise,
			&factRevisionID, &scoreRevisionID, &name, &categorySlug, &length, &width, &height,
			&quality, &value, &durability, &beginner, &advanced, &apartment, &noise, &portability,
			&policyVersion, &capabilities, &requirements, &compatible, &incompatible,
			&preferences, &redundancy, &storageLength, &storageWidth, &storageHeight,
			&operating, &safety, &roomHeight, &accessWidth, &overlap,
			&requiresStorage, &requiresOperating, &requiresSafety, &requiresAccess); err != nil {
			return fmt.Errorf("scan recommendation item: %w", err)
		}
		candidate := &item.product.Candidate
		candidate.FactRevisionID, candidate.ScoreRevisionID = factRevisionID.String, scoreRevisionID.String
		candidate.Name, candidate.CategorySlug = name.String, categorySlug.String
		candidate.PolicyVersion = policyVersion.String
		candidate.Dimensions = catalog.Dimensions{LengthMM: length.Int64, WidthMM: width.Int64, HeightMM: height.Int64}
		candidate.Space = domain.SpaceProfile{Footprint: domain.SpatialEnvelope{LengthMM: length.Int64, WidthMM: width.Int64, HeightMM: height.Int64},
			RequiresStorageFootprint: requiresStorage, RequiresOperatingClearance: requiresOperating,
			RequiresSafetyClearance: requiresSafety, RequiresAccessWidth: requiresAccess}
		candidate.Capabilities = capabilitiesFromStrings(capabilities)
		candidate.Requires = capabilitiesFromStrings(requirements)
		candidate.CompatibleWith = capabilitiesFromStrings(compatible)
		candidate.IncompatibleWith = capabilitiesFromStrings(incompatible)
		for _, value := range preferences {
			candidate.PreferenceTags = append(candidate.PreferenceTags, domain.TrainingPreference(value))
		}
		candidate.RedundancyGroups = redundancy
		if storageLength.Valid {
			candidate.Space.StorageFootprint = &domain.SpatialEnvelope{LengthMM: storageLength.Int64, WidthMM: storageWidth.Int64, HeightMM: storageHeight.Int64}
		}
		candidate.Space.OperatingClearance = clearanceFromValues(operating)
		candidate.Space.SafetyClearance = clearanceFromValues(safety)
		if roomHeight.Valid {
			value := roomHeight.Int64
			candidate.Space.MinimumRoomHeightMM = &value
		}
		if accessWidth.Valid {
			value := accessWidth.Int64
			candidate.Space.MinimumAccessWidthMM = &value
		}
		if overlap.Valid {
			candidate.Space.OverlapGroup = overlap.String
		}
		candidate.Scores = catalog.Scores{Quality: quality.Int16, Value: value.Int16,
			Durability: durability.Int16, Beginner: beginner.Int16, Advanced: advanced.Int16,
			Apartment: apartment.Int16, Noise: noise.Int16, Portability: portability.Int16}
		loaded, itemIDs = append(loaded, item), append(itemIDs, item.id)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("read recommendation items: %w", err)
	}
	goalSupport, err := repository.loadCandidateGoalSupport(ctx, recommendationID)
	if err != nil {
		return err
	}
	for index := range loaded {
		loaded[index].product.Candidate.GoalSupport = goalSupport[loaded[index].product.Candidate.ProductID]
	}
	reasons, err := repository.loadReasons(ctx, itemIDs)
	if err != nil {
		return err
	}
	selectedPrices := make(map[catalog.ProductID]int64)
	for index := range loaded {
		item := &loaded[index]
		item.product.Reasons = reasons[item.id]
		if item.itemType == "selected" {
			selectedPrices[item.product.Candidate.ProductID] = item.product.Candidate.Price.AmountMinor
		}
	}
	for _, item := range loaded {
		switch item.itemType {
		case "selected":
			result.Selected = append(result.Selected, domain.RecommendedItem{Rank: item.rank, Quantity: item.quantity,
				UnitPriceMinor: item.product.Candidate.Price.AmountMinor, Product: item.product})
		case "cheaper_alternative", "premium_alternative":
			alternativeType := domain.AlternativePremium
			if item.itemType == "cheaper_alternative" {
				alternativeType = domain.AlternativeCheaper
			}
			target := catalog.ProductID(item.alternative.String)
			result.Alternatives = append(result.Alternatives, domain.Alternative{ForProductID: target,
				Type: alternativeType, Product: item.product,
				PriceDifferenceMinor: item.product.Candidate.Price.AmountMinor - selectedPrices[target]})
		case "rejected":
			message := "Not selected for this setup"
			if len(item.product.Reasons) > 0 {
				message = item.product.Reasons[0].Message
			}
			result.Rejected = append(result.Rejected, domain.RejectedProduct{Candidate: item.product.Candidate,
				Code: item.rejection.String, Message: message})
		}
	}
	return nil
}

func (repository *Repository) loadReasons(ctx context.Context, itemIDs []string) (map[string][]domain.Reason, error) {
	result := make(map[string][]domain.Reason)
	if len(itemIDs) == 0 {
		return result, nil
	}
	rows, err := repository.pool.Query(ctx, `SELECT recommendation_item_id, code, message, dimension, score
		FROM recommendation.item_reasons WHERE recommendation_item_id = ANY($1::uuid[])
		ORDER BY recommendation_item_id, sort_order`, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("list recommendation reasons: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var itemID string
		var reason domain.Reason
		if err = rows.Scan(&itemID, &reason.Code, &reason.Message, &reason.Dimension, &reason.Score); err != nil {
			return nil, fmt.Errorf("scan recommendation reason: %w", err)
		}
		result[itemID] = append(result[itemID], reason)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read recommendation reasons: %w", err)
	}
	return result, nil
}

func (repository *Repository) listDraftEquipment(ctx context.Context, userID identity.UserID) ([]domain.ExistingEquipment, error) {
	rows, err := repository.pool.Query(ctx, `SELECT name, category_slug FROM recommendation.draft_existing_equipment
		WHERE user_id = $1 ORDER BY sort_order`, userID)
	if err != nil {
		return nil, fmt.Errorf("list draft equipment: %w", err)
	}
	defer rows.Close()
	return scanEquipment(rows)
}

func (repository *Repository) listSessionEquipment(ctx context.Context, recommendationID domain.RecommendationID) ([]domain.ExistingEquipment, error) {
	rows, err := repository.pool.Query(ctx, `SELECT e.name, e.category_slug, e.capabilities, e.redundancy_groups
		FROM recommendation.session_existing_equipment e
		JOIN recommendation.recommendations r ON r.session_id = e.session_id
		WHERE r.id = $1 ORDER BY e.sort_order`, recommendationID)
	if err != nil {
		return nil, fmt.Errorf("list session equipment: %w", err)
	}
	defer rows.Close()
	result := make([]domain.ExistingEquipment, 0)
	for rows.Next() {
		var equipment domain.ExistingEquipment
		var capabilities []string
		if err = rows.Scan(&equipment.Name, &equipment.CategorySlug, &capabilities, &equipment.RedundancyGroups); err != nil {
			return nil, fmt.Errorf("scan session equipment: %w", err)
		}
		equipment.Capabilities = capabilitiesFromStrings(capabilities)
		result = append(result, equipment)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read session equipment: %w", err)
	}
	return result, nil
}

type equipmentRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func scanEquipment(rows equipmentRows) ([]domain.ExistingEquipment, error) {
	result := make([]domain.ExistingEquipment, 0)
	for rows.Next() {
		var equipment domain.ExistingEquipment
		if err := rows.Scan(&equipment.Name, &equipment.CategorySlug); err != nil {
			return nil, fmt.Errorf("scan existing equipment: %w", err)
		}
		result = append(result, equipment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read existing equipment: %w", err)
	}
	return result, nil
}

func assignDraftOptionals(draft *ports.Draft, goal, experience sql.NullString, budget sql.NullInt64,
	currency sql.NullString, length, width, height, accessWidth sql.NullInt64, apartment sql.NullBool) {
	if goal.Valid {
		value := planning.Goal(goal.String)
		draft.Goal = &value
	}
	if experience.Valid {
		value := planning.ExperienceLevel(experience.String)
		draft.Experience = &value
	}
	if budget.Valid {
		value := budget.Int64
		draft.BudgetMinor = &value
	}
	if currency.Valid {
		value := currency.String
		draft.Currency = &value
	}
	if length.Valid && width.Valid && height.Valid && apartment.Valid {
		draft.AvailableSpace = &domain.AvailableSpace{LengthMM: length.Int64, WidthMM: width.Int64,
			HeightMM: height.Int64, ApartmentLiving: apartment.Bool}
		if accessWidth.Valid {
			value := accessWidth.Int64
			draft.AvailableSpace.AccessWidthMM = &value
		}
	}
}

var _ ports.Repository = (*Repository)(nil)
