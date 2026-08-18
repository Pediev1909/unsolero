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
	limit int,
	offset int,
) (ports.WishlistPage, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), wishlists.product_id
		FROM planning.wishlists AS wishlists
		JOIN catalog.products AS products ON products.id = wishlists.product_id
		WHERE wishlists.user_id = $1 AND products.status = 'published'
		ORDER BY wishlists.created_at DESC, wishlists.product_id DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return ports.WishlistPage{}, fmt.Errorf("list wishlist: %w", err)
	}
	defer rows.Close()

	page := ports.WishlistPage{ProductIDs: make([]catalog.ProductID, 0)}
	for rows.Next() {
		var productID catalog.ProductID
		if err := rows.Scan(&page.Total, &productID); err != nil {
			return ports.WishlistPage{}, fmt.Errorf("scan wishlist product: %w", err)
		}
		page.ProductIDs = append(page.ProductIDs, productID)
	}
	if err := rows.Err(); err != nil {
		return ports.WishlistPage{}, fmt.Errorf("read wishlist: %w", err)
	}
	if len(page.ProductIDs) == 0 && offset > 0 {
		if err := repository.pool.QueryRow(ctx, `SELECT count(*) FROM planning.wishlists wishlists
			JOIN catalog.products products ON products.id=wishlists.product_id
			WHERE wishlists.user_id=$1 AND products.status='published'`, userID).Scan(&page.Total); err != nil {
			return ports.WishlistPage{}, fmt.Errorf("count wishlist: %w", err)
		}
	}
	return page, nil
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
