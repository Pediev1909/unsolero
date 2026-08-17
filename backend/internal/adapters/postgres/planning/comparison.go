package planningpostgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

type ComparisonRepository struct {
	pool *pgxpool.Pool
}

func NewComparisonRepository(pool *pgxpool.Pool) *ComparisonRepository {
	return &ComparisonRepository{pool: pool}
}

func (repository *ComparisonRepository) ListProductIDs(
	ctx context.Context,
	userID identity.UserID,
) ([]catalog.ProductID, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT comparison.product_id
		FROM planning.comparison_items AS comparison
		JOIN catalog.products AS products ON products.id = comparison.product_id
		WHERE comparison.user_id = $1 AND products.status = 'published'
		ORDER BY comparison.position`, userID)
	if err != nil {
		return nil, fmt.Errorf("list comparison products: %w", err)
	}
	defer rows.Close()
	productIDs := make([]catalog.ProductID, 0, 4)
	for rows.Next() {
		var productID catalog.ProductID
		if err := rows.Scan(&productID); err != nil {
			return nil, fmt.Errorf("scan comparison product: %w", err)
		}
		productIDs = append(productIDs, productID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read comparison products: %w", err)
	}
	return productIDs, nil
}

func (repository *ComparisonRepository) Replace(
	ctx context.Context,
	userID identity.UserID,
	productIDs []catalog.ProductID,
) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin comparison update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `DELETE FROM planning.comparison_items WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear comparison: %w", err)
	}
	for position, productID := range productIDs {
		result, insertErr := tx.Exec(ctx, `
			INSERT INTO planning.comparison_items (user_id, product_id, position)
			SELECT $1, products.id, $3
			FROM catalog.products AS products
			WHERE products.id = $2 AND products.status = 'published'`, userID, productID, position)
		if insertErr != nil {
			return fmt.Errorf("insert comparison product: %w", insertErr)
		}
		if result.RowsAffected() == 0 {
			return ports.ErrProductNotFound
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit comparison update: %w", err)
	}
	return nil
}

var _ ports.ComparisonRepository = (*ComparisonRepository)(nil)
