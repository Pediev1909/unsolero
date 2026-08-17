package planningpostgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

func TestComparisonRepositoryPersistsOrderAndRollsBackInvalidReplacement(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var userID identity.UserID
	email := fmt.Sprintf("comparison-repository-%d@example.invalid", time.Now().UnixNano())
	if err = pool.QueryRow(ctx, `INSERT INTO identity.users (email, status) VALUES ($1, 'active') RETURNING id`, email).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id = $1`, userID) })

	rows, err := pool.Query(ctx, `SELECT id FROM catalog.products WHERE status = 'published' ORDER BY id LIMIT 3`)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	var productIDs []catalog.ProductID
	for rows.Next() {
		var productID catalog.ProductID
		if err = rows.Scan(&productID); err != nil {
			t.Fatalf("scan product: %v", err)
		}
		productIDs = append(productIDs, productID)
	}
	rows.Close()
	if len(productIDs) != 3 {
		t.Fatalf("published product count = %d, want at least 3", len(productIDs))
	}

	repository := NewComparisonRepository(pool)
	ordered := []catalog.ProductID{productIDs[2], productIDs[0], productIDs[1]}
	if err = repository.Replace(ctx, userID, ordered); err != nil {
		t.Fatalf("Replace(): %v", err)
	}
	loaded, err := repository.ListProductIDs(ctx, userID)
	if err != nil || fmt.Sprint(loaded) != fmt.Sprint(ordered) {
		t.Fatalf("ListProductIDs() = %#v, %v; want %#v", loaded, err, ordered)
	}
	if err = repository.Replace(ctx, userID, []catalog.ProductID{"00000000-0000-4000-8000-000000000000"}); err != ports.ErrProductNotFound {
		t.Fatalf("invalid Replace() error = %v, want ErrProductNotFound", err)
	}
	loaded, err = repository.ListProductIDs(ctx, userID)
	if err != nil || fmt.Sprint(loaded) != fmt.Sprint(ordered) {
		t.Fatalf("failed replacement changed comparison: %#v, %v", loaded, err)
	}
}
