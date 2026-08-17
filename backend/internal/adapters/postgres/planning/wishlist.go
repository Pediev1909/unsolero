package planningpostgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

type WishlistRepository struct {
	pool *pgxpool.Pool
}

func NewWishlistRepository(pool *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{pool: pool}
}

func (repository *WishlistRepository) ListProductIDs(
	ctx context.Context,
	userID identity.UserID,
) ([]catalog.ProductID, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT wishlists.product_id
		FROM planning.wishlists AS wishlists
		JOIN catalog.products AS products ON products.id = wishlists.product_id
		WHERE wishlists.user_id = $1 AND products.status = 'published'
		ORDER BY wishlists.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list wishlist: %w", err)
	}
	defer rows.Close()

	productIDs := make([]catalog.ProductID, 0)
	for rows.Next() {
		var productID catalog.ProductID
		if err := rows.Scan(&productID); err != nil {
			return nil, fmt.Errorf("scan wishlist product: %w", err)
		}
		productIDs = append(productIDs, productID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read wishlist: %w", err)
	}
	return productIDs, nil
}

func (repository *WishlistRepository) Save(
	ctx context.Context,
	userID identity.UserID,
	productID catalog.ProductID,
) error {
	result, err := repository.pool.Exec(ctx, `
		INSERT INTO planning.wishlists (user_id, product_id)
		SELECT $1, products.id
		FROM catalog.products AS products
		WHERE products.id = $2 AND products.status = 'published'
		ON CONFLICT (user_id, product_id) DO UPDATE SET updated_at = now()`,
		userID,
		productID,
	)
	if err != nil {
		return fmt.Errorf("save wishlist product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ports.ErrProductNotFound
	}
	return nil
}

func (repository *WishlistRepository) Delete(
	ctx context.Context,
	userID identity.UserID,
	productID catalog.ProductID,
) error {
	_, err := repository.pool.Exec(ctx, `
		DELETE FROM planning.wishlists
		WHERE user_id = $1 AND product_id = $2`, userID, productID)
	if err != nil {
		return fmt.Errorf("delete wishlist product: %w", err)
	}
	return nil
}

var _ ports.WishlistRepository = (*WishlistRepository)(nil)
