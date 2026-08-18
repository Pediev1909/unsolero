package database

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBoundedLookupQueriesHaveIndexPlans(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	connection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	if _, err = connection.Exec(context.Background(), "SET enable_seqscan=off"); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, relation, query string
		args                  []any
	}{
		{"published product slug", "products", `SELECT id FROM catalog.products WHERE slug=$1 AND status='published'`, []any{"not-present"}},
		{"active session token", "sessions", `SELECT id FROM identity.sessions WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>now()`, []any{make([]byte, 32)}},
		{"wishlist stable page", "wishlists", `SELECT product_id FROM planning.wishlists WHERE user_id=$1 ORDER BY created_at DESC,product_id DESC LIMIT 25`, []any{"12345678-1234-4234-8234-123456789abc"}},
		{"pending media deletion", "media_deletion_jobs", `SELECT object_name FROM admin.media_deletion_jobs WHERE status='pending' AND next_attempt_at<=now() ORDER BY next_attempt_at,id LIMIT 25`, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var raw []byte
			query := "EXPLAIN (FORMAT JSON, COSTS OFF) " + test.query
			if err := connection.QueryRow(context.Background(), query, test.args...).Scan(&raw); err != nil {
				t.Fatal(err)
			}
			var plan any
			if err := json.Unmarshal(raw, &plan); err != nil {
				t.Fatal(err)
			}
			if !containsIndexNode(plan) {
				t.Fatalf("bounded lookup has no index-backed %s node: %s", test.relation, raw)
			}
		})
	}
}

func containsIndexNode(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, child := range typed {
			if containsIndexNode(child) {
				return true
			}
		}
	case map[string]any:
		nodeType, _ := typed["Node Type"].(string)
		if strings.Contains(nodeType, "Index") {
			return true
		}
		for _, child := range typed {
			if containsIndexNode(child) {
				return true
			}
		}
	}
	return false
}
