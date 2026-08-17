package recommendationpostgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	identity "rigmark/internal/modules/identity/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

func TestPolicyApprovalActivationImmutabilityAndRetirement(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	editor := insertPolicyUser(t, ctx, pool, "editor")
	reviewer := insertPolicyUser(t, ctx, pool, "reviewer")
	publisher := insertPolicyUser(t, ctx, pool, "publisher")
	version := fmt.Sprintf("policy-integration-%d", time.Now().UnixNano())
	vertical := fmt.Sprintf("integration_%d", time.Now().UnixNano())

	var productID, categoryID, factRevisionID, scoreRevisionID string
	var length, width, height int64
	err = pool.QueryRow(ctx, `SELECT products.id,products.category_id,
		products.published_fact_revision_id,products.published_score_revision_id,
		products.length_mm,products.width_mm,products.height_mm
		FROM catalog.products products
		WHERE products.status='published' AND products.published_fact_revision_id IS NOT NULL
		ORDER BY products.id LIMIT 1`).Scan(&productID, &categoryID, &factRevisionID,
		&scoreRevisionID, &length, &width, &height)
	if err != nil {
		t.Fatalf("load governed product: %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO recommendation.policy_versions (
		version,vertical_key,workflow_status,created_by_user_id,
		goal_match_weight,budget_match_weight,space_match_weight,experience_match_weight,
		preference_match_weight,quality_weight,value_weight,durability_weight,
		compatibility_weight,portability_weight,noise_weight,priority_boost_percent,
		maximum_setup_items,candidates_per_slot,optional_slot_bonus,published_at
	) VALUES ($1,$2,'draft',$3,20,12,12,10,8,8,9,7,10,2,2,150,4,12,8,now())`,
		version, vertical, editor)
	if err != nil {
		t.Fatalf("insert policy: %v", err)
	}
	configuration := []struct {
		statement string
		arguments []any
	}{
		{`INSERT INTO recommendation.policy_capabilities VALUES ($1,'integration_capability','Integration capability')`, []any{version}},
		{`INSERT INTO recommendation.policy_goals VALUES ($1,'general_fitness','General fitness')`, []any{version}},
		{`INSERT INTO recommendation.policy_setup_roles VALUES ($1,'general_fitness','primary','Primary',true,0)`, []any{version}},
		{`INSERT INTO recommendation.policy_setup_role_capabilities VALUES ($1,'general_fitness','primary','integration_capability')`, []any{version}},
		{`INSERT INTO recommendation.category_policies (policy_version,category_id,support_status) VALUES ($1,$2,'supported')`, []any{version, categoryID}},
		{`INSERT INTO recommendation.product_policies VALUES ($1,$2,$3,$4)`, []any{version, productID, factRevisionID, scoreRevisionID}},
		{`INSERT INTO recommendation.product_space_profiles
			(policy_version,product_id,footprint_length_mm,footprint_width_mm,footprint_height_mm)
			VALUES ($1,$2,$3,$4,$5)`, []any{version, productID, length, width, height}},
		{`INSERT INTO recommendation.product_goal_support VALUES ($1,$2,'general_fitness',90)`, []any{version, productID}},
	}
	for _, item := range configuration {
		if _, err = pool.Exec(ctx, item.statement, item.arguments...); err != nil {
			t.Fatalf("configure policy: %v", err)
		}
	}

	repository := New(pool)
	if err = repository.TransitionPolicy(ctx, editor, version, recommendation.PolicyInReview, ""); err != nil {
		t.Fatalf("submit policy: %v", err)
	}
	if err = repository.TransitionPolicy(ctx, editor, version, recommendation.PolicyApproved, "self review"); !errors.Is(err, ports.ErrSeparationOfDuties) {
		t.Fatalf("self approval error = %v", err)
	}
	if err = repository.TransitionPolicy(ctx, reviewer, version, recommendation.PolicyApproved, "verified test policy"); err != nil {
		t.Fatalf("approve policy: %v", err)
	}
	if err = repository.TransitionPolicy(ctx, reviewer, version, recommendation.PolicyActive, ""); !errors.Is(err, ports.ErrSeparationOfDuties) {
		t.Fatalf("reviewer activation error = %v", err)
	}
	if err = repository.TransitionPolicy(ctx, publisher, version, recommendation.PolicyActive, ""); err != nil {
		t.Fatalf("activate policy: %v", err)
	}
	var status string
	if err = pool.QueryRow(ctx, `SELECT workflow_status FROM recommendation.policy_versions WHERE version=$1`, version).Scan(&status); err != nil || status != "active" {
		t.Fatalf("activated status = %q, %v", status, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE recommendation.product_goal_support SET match_score=91
		WHERE policy_version=$1 AND product_id=$2`, version, productID); err == nil {
		t.Fatal("active policy behavior was mutable")
	}
	if err = repository.TransitionPolicy(ctx, publisher, version, recommendation.PolicyRetired, "superseded test policy"); err != nil {
		t.Fatalf("retire policy: %v", err)
	}

	cleanupPolicyFixture(t, ctx, pool, version, []identity.UserID{editor, reviewer, publisher})
}

func insertPolicyUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) identity.UserID {
	t.Helper()
	var id identity.UserID
	email := fmt.Sprintf("policy-%s-%d@example.invalid", label, time.Now().UnixNano())
	if err := pool.QueryRow(ctx, `INSERT INTO identity.users (email,status) VALUES ($1,'active') RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("insert policy %s: %v", label, err)
	}
	return id
}

func cleanupPolicyFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, version string, users []identity.UserID) {
	t.Helper()
	for _, statement := range []string{
		`DELETE FROM recommendation.product_goal_support WHERE policy_version=$1`,
		`DELETE FROM recommendation.product_space_profiles WHERE policy_version=$1`,
		`DELETE FROM recommendation.product_policies WHERE policy_version=$1`,
		`DELETE FROM recommendation.category_policies WHERE policy_version=$1`,
		`DELETE FROM recommendation.policy_setup_role_capabilities WHERE policy_version=$1`,
		`DELETE FROM recommendation.policy_setup_roles WHERE policy_version=$1`,
		`DELETE FROM recommendation.policy_goals WHERE policy_version=$1`,
		`DELETE FROM recommendation.policy_capabilities WHERE policy_version=$1`,
	} {
		if _, err := pool.Exec(ctx, statement, version); err != nil {
			t.Fatalf("clean policy fixture: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, `DELETE FROM admin.audit_log WHERE entity_type='recommendation_policy' AND entity_id=$1`, version); err != nil {
		t.Fatalf("clean policy audit: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM recommendation.policy_versions WHERE version=$1`, version); err != nil {
		t.Fatalf("clean policy version: %v", err)
	}
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, string(user))
	}
	if _, err := pool.Exec(ctx, `DELETE FROM identity.users WHERE id=ANY($1::uuid[])`, ids); err != nil {
		t.Fatalf("clean policy users: %v", err)
	}
}
