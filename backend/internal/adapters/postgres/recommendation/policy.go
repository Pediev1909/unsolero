package recommendationpostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

func (repository *Repository) ListPolicies(ctx context.Context) ([]domain.PolicySummary, error) {
	rows, err := repository.pool.Query(ctx, `SELECT policies.version,policies.vertical_key,policies.workflow_status,
		count(DISTINCT categories.category_id),count(DISTINCT products.product_id),
		policies.created_at,policies.activated_at,policies.review_note
		FROM recommendation.policy_versions policies
		LEFT JOIN recommendation.category_policies categories ON categories.policy_version=policies.version AND categories.support_status='supported'
		LEFT JOIN recommendation.product_policies products ON products.policy_version=policies.version
		GROUP BY policies.version ORDER BY policies.created_at DESC,policies.version`)
	if err != nil {
		return nil, fmt.Errorf("list recommendation policies: %w", err)
	}
	defer rows.Close()
	result := make([]domain.PolicySummary, 0)
	for rows.Next() {
		var item domain.PolicySummary
		if err = rows.Scan(&item.Version, &item.VerticalKey, &item.Status, &item.CategoryCount,
			&item.ProductCount, &item.CreatedAt, &item.ActivatedAt, &item.ReviewNote); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repository *Repository) TransitionPolicy(ctx context.Context, actor identity.UserID, version string, target domain.PolicyWorkflowStatus, note string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current domain.PolicyWorkflowStatus
	var vertical string
	var creator, submitter, reviewer sql.NullString
	if err = tx.QueryRow(ctx, `SELECT workflow_status,vertical_key,created_by_user_id,submitted_by_user_id,reviewed_by_user_id
		FROM recommendation.policy_versions WHERE version=$1 FOR UPDATE`, version).Scan(
		&current, &vertical, &creator, &submitter, &reviewer); errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}
	allowed := false
	switch target {
	case domain.PolicyInReview:
		allowed = current == domain.PolicyDraft || current == domain.PolicyRejected
		if allowed {
			_, err = tx.Exec(ctx, `UPDATE recommendation.policy_versions SET workflow_status='in_review',submitted_by_user_id=$2,
				submitted_at=now(),reviewed_by_user_id=NULL,reviewed_at=NULL,review_note='' WHERE version=$1`, version, actor)
		}
	case domain.PolicyApproved, domain.PolicyRejected:
		allowed = current == domain.PolicyInReview
		if allowed && ((creator.Valid && creator.String == string(actor)) || (submitter.Valid && submitter.String == string(actor))) {
			return ports.ErrSeparationOfDuties
		}
		if allowed {
			_, err = tx.Exec(ctx, `UPDATE recommendation.policy_versions SET workflow_status=$2,reviewed_by_user_id=$3,
				reviewed_at=now(),review_note=$4 WHERE version=$1`, version, target, actor, note)
		}
	case domain.PolicyActive:
		allowed = current == domain.PolicyApproved
		if allowed && (!reviewer.Valid || reviewer.String == string(actor)) {
			return ports.ErrSeparationOfDuties
		}
		if allowed {
			var invalid int
			err = tx.QueryRow(ctx, `SELECT
				(NOT EXISTS(SELECT 1 FROM recommendation.category_policies WHERE policy_version=$1 AND support_status='supported'))::int +
				(NOT EXISTS(SELECT 1 FROM recommendation.policy_goals WHERE policy_version=$1))::int +
				(SELECT count(*) FROM recommendation.policy_goals goals WHERE goals.policy_version=$1 AND NOT EXISTS (
				 SELECT 1 FROM recommendation.policy_setup_roles roles
				 WHERE roles.policy_version=goals.policy_version AND roles.goal_key=goals.goal_key)) +
				(SELECT count(*) FROM recommendation.policy_setup_roles roles WHERE roles.policy_version=$1 AND NOT EXISTS (
				 SELECT 1 FROM recommendation.policy_setup_role_capabilities capabilities
				 WHERE capabilities.policy_version=roles.policy_version AND capabilities.goal_key=roles.goal_key AND capabilities.role_key=roles.role_key)) +
				(SELECT (goal_match_weight+budget_match_weight+space_match_weight+experience_match_weight+
				 preference_match_weight+quality_weight+value_weight+durability_weight+compatibility_weight+
				 portability_weight+noise_weight=0)::int FROM recommendation.policy_versions WHERE version=$1) +
				(NOT EXISTS(SELECT 1 FROM recommendation.product_policies products
				 JOIN catalog.products catalog_products ON catalog_products.id=products.product_id
				 JOIN recommendation.category_policies categories ON categories.policy_version=products.policy_version
				  AND categories.category_id=catalog_products.category_id AND categories.support_status='supported'
				 WHERE products.policy_version=$1))::int +
				(SELECT count(*) FROM recommendation.product_policies products
				 LEFT JOIN recommendation.product_space_profiles space USING(policy_version,product_id)
				 LEFT JOIN catalog.products catalog_products ON catalog_products.id=products.product_id
				 LEFT JOIN recommendation.category_policies categories ON categories.policy_version=products.policy_version AND categories.category_id=catalog_products.category_id
				 WHERE products.policy_version=$1 AND categories.support_status='supported' AND (space.product_id IS NULL OR products.fact_revision_id<>catalog_products.published_fact_revision_id OR products.score_revision_id<>catalog_products.published_score_revision_id OR NOT EXISTS (
				  SELECT 1 FROM recommendation.product_goal_support goals WHERE goals.policy_version=products.policy_version AND goals.product_id=products.product_id)))`, version).Scan(&invalid)
			if err == nil && invalid != 0 {
				err = ports.ErrConflict
			}
			if err == nil {
				_, err = tx.Exec(ctx, `UPDATE recommendation.policy_versions SET workflow_status='retired',retired_at=now()
					WHERE vertical_key=$1 AND workflow_status='active' AND version<>$2`, vertical, version)
			}
			if err == nil {
				_, err = tx.Exec(ctx, `UPDATE recommendation.policy_versions SET workflow_status='active',activated_by_user_id=$2,
					activated_at=now(),retired_at=NULL WHERE version=$1`, version, actor)
			}
		}
	case domain.PolicyRetired:
		allowed = current == domain.PolicyActive
		if allowed {
			_, err = tx.Exec(ctx, `UPDATE recommendation.policy_versions SET workflow_status='retired',retired_at=now() WHERE version=$1`, version)
		}
	}
	if !allowed {
		return ports.ErrConflict
	}
	if err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"status": string(target), "note": note})
	if _, err = tx.Exec(ctx, `INSERT INTO admin.audit_log(actor_user_id,action,entity_type,entity_id,changes)
		VALUES($1,$2,'recommendation_policy',$3,$4)`, actor, "recommendation.policy."+string(target), version, payload); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (repository *Repository) ActivePolicy(ctx context.Context) (domain.Policy, error) {
	policy := domain.Policy{Categories: make(map[string]domain.CategoryPolicy), Products: make(map[catalog.ProductID]domain.ProductPolicy)}
	err := repository.pool.QueryRow(ctx, `SELECT version,
		goal_match_weight,budget_match_weight,space_match_weight,experience_match_weight,
		preference_match_weight,quality_weight,value_weight,durability_weight,
		compatibility_weight,portability_weight,noise_weight,priority_boost_percent,
		maximum_setup_items,candidates_per_slot,optional_slot_bonus,spatial_constraints
		FROM recommendation.policy_versions
		WHERE vertical_key=$1 AND workflow_status='active'`, repository.vertical).Scan(
		&policy.Config.PolicyVersion, &policy.Config.Weights.GoalMatch,
		&policy.Config.Weights.BudgetMatch, &policy.Config.Weights.SpaceMatch,
		&policy.Config.Weights.ExperienceMatch, &policy.Config.Weights.PreferenceMatch,
		&policy.Config.Weights.Quality, &policy.Config.Weights.Value,
		&policy.Config.Weights.Durability, &policy.Config.Weights.Compatibility,
		&policy.Config.Weights.Portability, &policy.Config.Weights.Noise,
		&policy.Config.PriorityBoostPercent, &policy.Config.MaximumSetupItems,
		&policy.Config.CandidatesPerSlot, &policy.Config.OptionalSlotBonus,
		&policy.Config.SpatialConstraints)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Policy{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Policy{}, fmt.Errorf("load active recommendation policy: %w", err)
	}
	if err = repository.loadPolicyGoals(ctx, &policy); err != nil {
		return domain.Policy{}, err
	}
	if err = repository.loadPolicyVocabulary(ctx, &policy); err != nil {
		return domain.Policy{}, err
	}
	if err = repository.loadCategoryPolicies(ctx, &policy); err != nil {
		return domain.Policy{}, err
	}
	if err = repository.loadProductPolicies(ctx, &policy); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func (repository *Repository) loadPolicyGoals(ctx context.Context, policy *domain.Policy) error {
	rows, err := repository.pool.Query(ctx, `SELECT roles.goal_key,goals.label,roles.role_key,roles.is_required,
		roles.sort_order,capabilities.capability_key
		FROM recommendation.policy_setup_roles roles
		JOIN recommendation.policy_goals goals
		 ON goals.policy_version=roles.policy_version AND goals.goal_key=roles.goal_key
		JOIN recommendation.policy_setup_role_capabilities capabilities
		 ON capabilities.policy_version=roles.policy_version AND capabilities.goal_key=roles.goal_key
		 AND capabilities.role_key=roles.role_key
		WHERE roles.policy_version=$1
		ORDER BY roles.goal_key,roles.sort_order,roles.role_key,capabilities.capability_key`, policy.Config.PolicyVersion)
	if err != nil {
		return fmt.Errorf("load policy setup roles: %w", err)
	}
	defer rows.Close()
	goalIndexes := make(map[planning.Goal]int)
	roleIndexes := make(map[string]int)
	for rows.Next() {
		var goal planning.Goal
		var goalLabel string
		var roleKey string
		var required bool
		var sortOrder int
		var capability domain.Capability
		if err = rows.Scan(&goal, &goalLabel, &roleKey, &required, &sortOrder, &capability); err != nil {
			return err
		}
		goalIndex, exists := goalIndexes[goal]
		if !exists {
			policy.Config.Goals = append(policy.Config.Goals, domain.GoalPolicy{Goal: goal, Label: goalLabel})
			goalIndex = len(policy.Config.Goals) - 1
			goalIndexes[goal] = goalIndex
		}
		key := string(goal) + ":" + roleKey
		roleIndex, exists := roleIndexes[key]
		if !exists {
			policy.Config.Goals[goalIndex].Roles = append(policy.Config.Goals[goalIndex].Roles,
				domain.SetupRole{Key: roleKey, Required: required, SortOrder: sortOrder})
			roleIndex = len(policy.Config.Goals[goalIndex].Roles) - 1
			roleIndexes[key] = roleIndex
		}
		policy.Config.Goals[goalIndex].Roles[roleIndex].Capabilities = append(
			policy.Config.Goals[goalIndex].Roles[roleIndex].Capabilities, capability)
	}
	return rows.Err()
}

func (repository *Repository) loadCategoryPolicies(ctx context.Context, policy *domain.Policy) error {
	rows, err := repository.pool.Query(ctx, `SELECT categories.slug,policies.support_status,
		COALESCE(array_agg(DISTINCT capabilities.capability_key) FILTER (WHERE capabilities.capability_key IS NOT NULL),'{}'),
		COALESCE(array_agg(DISTINCT groups.group_key) FILTER (WHERE groups.group_key IS NOT NULL),'{}')
		FROM recommendation.category_policies policies
		JOIN catalog.categories categories ON categories.id=policies.category_id
		LEFT JOIN recommendation.category_policy_capabilities capabilities
		 ON capabilities.policy_version=policies.policy_version AND capabilities.category_id=policies.category_id
		LEFT JOIN recommendation.category_redundancy_groups groups
		 ON groups.policy_version=policies.policy_version AND groups.category_id=policies.category_id
		WHERE policies.policy_version=$1
		GROUP BY categories.slug,policies.support_status`, policy.Config.PolicyVersion)
	if err != nil {
		return fmt.Errorf("load category policies: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.CategoryPolicy
		var status string
		var capabilities []string
		if err = rows.Scan(&item.CategorySlug, &status, &capabilities, &item.RedundancyGroups); err != nil {
			return err
		}
		for _, capability := range capabilities {
			item.Capabilities = append(item.Capabilities, domain.Capability(capability))
		}
		item.Supported = status == "supported"
		policy.Categories[item.CategorySlug] = item
	}
	return rows.Err()
}

func (repository *Repository) loadProductPolicies(ctx context.Context, policy *domain.Policy) error {
	rows, err := repository.pool.Query(ctx, `SELECT products.product_id,products.fact_revision_id,products.score_revision_id,
		space.footprint_length_mm,space.footprint_width_mm,space.footprint_height_mm,
		space.storage_length_mm,space.storage_width_mm,space.storage_height_mm,
		space.operating_front_mm,space.operating_back_mm,space.operating_left_mm,space.operating_right_mm,space.operating_top_mm,
		space.safety_front_mm,space.safety_back_mm,space.safety_left_mm,space.safety_right_mm,space.safety_top_mm,
		space.minimum_room_height_mm,space.minimum_access_width_mm,space.overlap_group,
		categories.requires_storage_footprint,categories.requires_operating_clearance,
		categories.requires_safety_clearance,categories.requires_access_width
		FROM recommendation.product_policies products
		JOIN catalog.products catalog_products ON catalog_products.id=products.product_id
		JOIN recommendation.category_policies categories
		 ON categories.policy_version=products.policy_version AND categories.category_id=catalog_products.category_id
		LEFT JOIN recommendation.product_space_profiles space
		 ON space.policy_version=products.policy_version AND space.product_id=products.product_id
		WHERE products.policy_version=$1 AND categories.support_status='supported'`, policy.Config.PolicyVersion)
	if err != nil {
		return fmt.Errorf("load product policies: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.ProductPolicy
		// The space profile is optional: a non-physical product has no
		// footprint to declare. An inner join here silently dropped such a
		// product's entire policy, which surfaced as "no active policy" rather
		// than as a missing footprint. A physical vertical still fails closed,
		// because the engine rejects a zero footprint when it validates
		// candidates with spatial constraints enabled.
		var footprintLength, footprintWidth, footprintHeight sql.NullInt64
		var storageLength, storageWidth, storageHeight sql.NullInt64
		var operating [5]sql.NullInt64
		var safety [5]sql.NullInt64
		var roomHeight, accessWidth sql.NullInt64
		var overlap sql.NullString
		if err = rows.Scan(&item.ProductID, &item.FactRevisionID, &item.ScoreRevisionID,
			&footprintLength, &footprintWidth, &footprintHeight,
			&storageLength, &storageWidth, &storageHeight,
			&operating[0], &operating[1], &operating[2], &operating[3], &operating[4],
			&safety[0], &safety[1], &safety[2], &safety[3], &safety[4],
			&roomHeight, &accessWidth, &overlap,
			&item.Space.RequiresStorageFootprint, &item.Space.RequiresOperatingClearance,
			&item.Space.RequiresSafetyClearance, &item.Space.RequiresAccessWidth); err != nil {
			return err
		}
		item.Space.Footprint = domain.SpatialEnvelope{
			LengthMM: footprintLength.Int64, WidthMM: footprintWidth.Int64, HeightMM: footprintHeight.Int64,
		}
		if storageLength.Valid {
			item.Space.StorageFootprint = &domain.SpatialEnvelope{LengthMM: storageLength.Int64, WidthMM: storageWidth.Int64, HeightMM: storageHeight.Int64}
		}
		item.Space.OperatingClearance = nullableClearance(operating)
		item.Space.SafetyClearance = nullableClearance(safety)
		if roomHeight.Valid {
			value := roomHeight.Int64
			item.Space.MinimumRoomHeightMM = &value
		}
		if accessWidth.Valid {
			value := accessWidth.Int64
			item.Space.MinimumAccessWidthMM = &value
		}
		if overlap.Valid {
			item.Space.OverlapGroup = overlap.String
		}
		policy.Products[item.ProductID] = item
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return repository.loadProductPolicyRelations(ctx, policy)
}

func nullableClearance(values [5]sql.NullInt64) *domain.Clearance {
	if !values[0].Valid {
		return nil
	}
	return &domain.Clearance{FrontMM: values[0].Int64, BackMM: values[1].Int64,
		LeftMM: values[2].Int64, RightMM: values[3].Int64, TopMM: values[4].Int64}
}

func (repository *Repository) loadProductPolicyRelations(ctx context.Context, policy *domain.Policy) error {
	rows, err := repository.pool.Query(ctx, `SELECT product_id,capability_key,relation_type
		FROM recommendation.product_policy_capabilities WHERE policy_version=$1
		ORDER BY product_id,relation_type,capability_key`, policy.Config.PolicyVersion)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id catalog.ProductID
		var capability domain.Capability
		var relation string
		if err = rows.Scan(&id, &capability, &relation); err != nil {
			rows.Close()
			return err
		}
		item := policy.Products[id]
		switch relation {
		case "provides":
			item.Capabilities = append(item.Capabilities, capability)
		case "requires":
			item.Requires = append(item.Requires, capability)
		case "compatible_with":
			item.CompatibleWith = append(item.CompatibleWith, capability)
		case "incompatible_with":
			item.IncompatibleWith = append(item.IncompatibleWith, capability)
		}
		policy.Products[id] = item
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	goalRows, err := repository.pool.Query(ctx, `SELECT product_id,goal_key,match_score FROM recommendation.product_goal_support WHERE policy_version=$1 ORDER BY product_id,goal_key`, policy.Config.PolicyVersion)
	if err != nil {
		return err
	}
	for goalRows.Next() {
		var id catalog.ProductID
		var goal planning.Goal
		var score int
		if err = goalRows.Scan(&id, &goal, &score); err != nil {
			goalRows.Close()
			return err
		}
		item := policy.Products[id]
		item.GoalSupport = append(item.GoalSupport, domain.GoalSupport{Goal: goal, Score: score})
		policy.Products[id] = item
	}
	if err = goalRows.Err(); err != nil {
		goalRows.Close()
		return err
	}
	goalRows.Close()
	return repository.loadProductStringRelations(ctx, policy)
}

func (repository *Repository) loadProductStringRelations(ctx context.Context, policy *domain.Policy) error {
	type queryTarget struct {
		query string
		apply func(*domain.ProductPolicy, string)
	}
	targets := []queryTarget{
		{`SELECT product_id,preference_key FROM recommendation.product_preference_tags WHERE policy_version=$1 ORDER BY product_id,preference_key`, func(item *domain.ProductPolicy, value string) {
			item.PreferenceTags = append(item.PreferenceTags, domain.TrainingPreference(value))
		}},
		{`SELECT product_id,group_key FROM recommendation.product_redundancy_groups WHERE policy_version=$1 ORDER BY product_id,group_key`, func(item *domain.ProductPolicy, value string) {
			item.RedundancyGroups = append(item.RedundancyGroups, value)
		}},
	}
	for _, target := range targets {
		rows, err := repository.pool.Query(ctx, target.query, policy.Config.PolicyVersion)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id catalog.ProductID
			var value string
			if err = rows.Scan(&id, &value); err != nil {
				rows.Close()
				return err
			}
			item := policy.Products[id]
			target.apply(&item, value)
			policy.Products[id] = item
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}
	return nil
}

var _ ports.PolicyRepository = (*Repository)(nil)

// loadPolicyVocabulary reads the preference and priority vocabularies the
// active policy declares. Both are ordered deterministically so that two runs
// against the same policy produce identical scores and explanations.
func (repository *Repository) loadPolicyVocabulary(ctx context.Context, policy *domain.Policy) error {
	tagRows, err := repository.pool.Query(ctx, `SELECT tag_key
		FROM recommendation.policy_preference_tags
		WHERE policy_version=$1 ORDER BY tag_key`, policy.Config.PolicyVersion)
	if err != nil {
		return fmt.Errorf("load policy preference tags: %w", err)
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var tag domain.TrainingPreference
		if err = tagRows.Scan(&tag); err != nil {
			return err
		}
		policy.Config.PreferenceTags = append(policy.Config.PreferenceTags, tag)
	}
	if err = tagRows.Err(); err != nil {
		return fmt.Errorf("load policy preference tags: %w", err)
	}

	priorityRows, err := repository.pool.Query(ctx, `SELECT priorities.priority_key,
		priorities.reason_code,priorities.reason_message,priorities.reason_dimension,
		priorities.reason_threshold,dimensions.dimension
		FROM recommendation.policy_priorities priorities
		JOIN recommendation.policy_priority_dimensions dimensions
		 ON dimensions.policy_version=priorities.policy_version
		 AND dimensions.priority_key=priorities.priority_key
		WHERE priorities.policy_version=$1
		ORDER BY priorities.sort_order,priorities.priority_key,dimensions.dimension`,
		policy.Config.PolicyVersion)
	if err != nil {
		return fmt.Errorf("load policy priorities: %w", err)
	}
	defer priorityRows.Close()
	priorityIndexes := make(map[domain.Priority]int)
	for priorityRows.Next() {
		var key domain.Priority
		var reasonCode, reasonMessage string
		var reasonDimension domain.Dimension
		var reasonThreshold int
		var dimension domain.Dimension
		if err = priorityRows.Scan(&key, &reasonCode, &reasonMessage,
			&reasonDimension, &reasonThreshold, &dimension); err != nil {
			return err
		}
		index, exists := priorityIndexes[key]
		if !exists {
			policy.Config.Priorities = append(policy.Config.Priorities, domain.PriorityPolicy{
				Key: key, ReasonCode: reasonCode, ReasonMessage: reasonMessage,
				ReasonDimension: reasonDimension, ReasonThreshold: reasonThreshold,
			})
			index = len(policy.Config.Priorities) - 1
			priorityIndexes[key] = index
		}
		policy.Config.Priorities[index].BoostDimensions = append(
			policy.Config.Priorities[index].BoostDimensions, dimension)
	}
	if err = priorityRows.Err(); err != nil {
		return fmt.Errorf("load policy priorities: %w", err)
	}
	return nil
}
