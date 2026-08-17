package adminpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	admin "rigmark/internal/modules/admin/domain"
	"rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

const adminProductColumns = `
	products.id, products.category_id, categories.name, categories.slug,
	products.brand_id, brands.name, brands.slug, products.name, products.slug,
	products.description, products.price_minor, products.currency,
	products.length_mm, products.width_mm, products.height_mm,
	products.weight_grams, products.max_capacity_grams, products.material,
	products.warranty_months, products.quality_score, products.value_score,
	products.durability_score, products.beginner_score, products.advanced_score,
	products.apartment_score, products.noise_score, products.portability_score,
	products.status, products.created_at, products.updated_at`

type scanner interface {
	Scan(...interface{}) error
}

func (repository *Repository) ListProducts(ctx context.Context, search string, limit, offset int) (admin.ProductPage, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), `+adminProductColumns+`
		FROM catalog.products AS products
		JOIN catalog.categories AS categories ON categories.id = products.category_id
		JOIN catalog.brands AS brands ON brands.id = products.brand_id
		WHERE $1 = '' OR products.name ILIKE '%' || $1 || '%'
			OR products.slug ILIKE '%' || $1 || '%'
			OR brands.name ILIKE '%' || $1 || '%'
		ORDER BY products.updated_at DESC, products.name
		LIMIT $2 OFFSET $3`, search, limit, offset)
	if err != nil {
		return admin.ProductPage{}, fmt.Errorf("list admin products: %w", err)
	}
	defer rows.Close()
	page := admin.ProductPage{Items: make([]catalog.Product, 0)}
	for rows.Next() {
		var product catalog.Product
		if err := scanAdminProduct(rows, &page.Total, &product); err != nil {
			return admin.ProductPage{}, err
		}
		page.Items = append(page.Items, product)
	}
	if err := rows.Err(); err != nil {
		return admin.ProductPage{}, fmt.Errorf("read admin products: %w", err)
	}
	if err := repository.loadProductChildren(ctx, page.Items); err != nil {
		return admin.ProductPage{}, err
	}
	return page, nil
}

func (repository *Repository) GetProduct(ctx context.Context, id catalog.ProductID) (catalog.Product, error) {
	var product catalog.Product
	err := scanAdminProduct(repository.pool.QueryRow(ctx, `
		SELECT `+adminProductColumns+`
		FROM catalog.products AS products
		JOIN catalog.categories AS categories ON categories.id = products.category_id
		JOIN catalog.brands AS brands ON brands.id = products.brand_id
		WHERE products.id = $1`, id), nil, &product)
	if err = notFound(err); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return catalog.Product{}, err
		}
		return catalog.Product{}, fmt.Errorf("get admin product: %w", err)
	}
	products := []catalog.Product{product}
	if err := repository.loadProductChildren(ctx, products); err != nil {
		return catalog.Product{}, err
	}
	return products[0], nil
}

func scanAdminProduct(row scanner, total *int64, product *catalog.Product) error {
	var capacity sql.NullInt64
	targets := []interface{}{
		&product.ID, &product.CategoryID, &product.CategoryName, &product.CategorySlug,
		&product.BrandID, &product.BrandName, &product.BrandSlug, &product.Name, &product.Slug,
		&product.Description, &product.Price.AmountMinor, &product.Price.Currency,
		&product.Dimensions.LengthMM, &product.Dimensions.WidthMM, &product.Dimensions.HeightMM,
		&product.WeightGrams, &capacity, &product.Material, &product.WarrantyMonths,
		&product.Scores.Quality, &product.Scores.Value, &product.Scores.Durability,
		&product.Scores.Beginner, &product.Scores.Advanced, &product.Scores.Apartment,
		&product.Scores.Noise, &product.Scores.Portability, &product.Status,
		&product.CreatedAt, &product.UpdatedAt,
	}
	if total != nil {
		targets = append([]interface{}{total}, targets...)
	}
	if err := row.Scan(targets...); err != nil {
		return fmt.Errorf("scan admin product: %w", err)
	}
	if capacity.Valid {
		value := capacity.Int64
		product.MaxCapacityGrams = &value
	}
	return nil
}

func (repository *Repository) loadProductChildren(ctx context.Context, products []catalog.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]string, len(products))
	indexes := make(map[string]int, len(products))
	for index := range products {
		ids[index] = string(products[index].ID)
		indexes[ids[index]] = index
	}
	if err := repository.loadImages(ctx, ids, indexes, products); err != nil {
		return err
	}
	return repository.loadAttributes(ctx, ids, indexes, products)
}

func (repository *Repository) loadImages(ctx context.Context, ids []string, indexes map[string]int, products []catalog.Product) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, product_id, url, alt_text, sort_order, is_primary, width_px, height_px
		FROM catalog.product_images WHERE product_id = ANY($1::uuid[])
		ORDER BY product_id, is_primary DESC, sort_order, id`, ids)
	if err != nil {
		return fmt.Errorf("load admin product images: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var image catalog.ProductImage
		var productID string
		var width, height sql.NullInt32
		if err := rows.Scan(&image.ID, &productID, &image.URL, &image.AltText, &image.SortOrder, &image.IsPrimary, &width, &height); err != nil {
			return fmt.Errorf("scan admin image: %w", err)
		}
		if width.Valid {
			value := int(width.Int32)
			image.WidthPX = &value
		}
		if height.Valid {
			value := int(height.Int32)
			image.HeightPX = &value
		}
		products[indexes[productID]].Images = append(products[indexes[productID]].Images, image)
	}
	return rows.Err()
}

func (repository *Repository) loadAttributes(ctx context.Context, ids []string, indexes map[string]int, products []catalog.Product) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, product_id, attribute_key, attribute_type, numeric_value,
			text_value, boolean_value, unit, is_filterable
		FROM catalog.product_attributes WHERE product_id = ANY($1::uuid[])
		ORDER BY product_id, attribute_key`, ids)
	if err != nil {
		return fmt.Errorf("load admin product attributes: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var attribute catalog.Attribute
		var productID string
		var numeric sql.NullFloat64
		var textValue, unit sql.NullString
		var boolean sql.NullBool
		if err := rows.Scan(&attribute.ID, &productID, &attribute.Key, &attribute.Type, &numeric, &textValue, &boolean, &unit, &attribute.IsFilterable); err != nil {
			return fmt.Errorf("scan admin attribute: %w", err)
		}
		if numeric.Valid {
			value := numeric.Float64
			attribute.NumericValue = &value
		}
		if textValue.Valid {
			value := textValue.String
			attribute.TextValue = &value
		}
		if boolean.Valid {
			value := boolean.Bool
			attribute.BooleanValue = &value
		}
		if unit.Valid {
			value := unit.String
			attribute.Unit = &value
		}
		products[indexes[productID]].Attributes = append(products[indexes[productID]].Attributes, attribute)
	}
	return rows.Err()
}

func (repository *Repository) CreateProduct(ctx context.Context, actor identity.UserID, input admin.ProductInput) (catalog.Product, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return catalog.Product{}, fmt.Errorf("begin create product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id catalog.ProductID
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.products (
			category_id, brand_id, name, slug, description, price_minor, currency,
			length_mm, width_mm, height_mm, weight_grams, max_capacity_grams,
			material, warranty_months, quality_score, value_score, durability_score,
			beginner_score, advanced_score, apartment_score, noise_score, portability_score
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id`, input.CategoryID, input.BrandID, input.Name, input.Slug, input.Description,
		input.Price.AmountMinor, input.Price.Currency, input.Dimensions.LengthMM, input.Dimensions.WidthMM,
		input.Dimensions.HeightMM, input.WeightGrams, input.MaxCapacityGrams, input.Material,
		input.WarrantyMonths, input.Scores.Quality, input.Scores.Value, input.Scores.Durability,
		input.Scores.Beginner, input.Scores.Advanced, input.Scores.Apartment, input.Scores.Noise,
		input.Scores.Portability).Scan(&id)
	if err == nil {
		err = audit(ctx, tx, actor, "product.create", "product", string(id), map[string]string{"slug": input.Slug})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return catalog.Product{}, fmt.Errorf("create product: %w", err)
	}
	return repository.GetProduct(ctx, id)
}

func (repository *Repository) UpdateProduct(ctx context.Context, actor identity.UserID, id catalog.ProductID, input admin.ProductInput) (catalog.Product, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return catalog.Product{}, fmt.Errorf("begin update product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE catalog.products SET category_id=$2, brand_id=$3, name=$4, slug=$5,
			description=$6, price_minor=$7, currency=$8, length_mm=$9, width_mm=$10,
			height_mm=$11, weight_grams=$12, max_capacity_grams=$13, material=$14,
			warranty_months=$15, quality_score=$16, value_score=$17, durability_score=$18,
			beginner_score=$19, advanced_score=$20, apartment_score=$21, noise_score=$22,
			portability_score=$23, updated_at=now()
		WHERE id=$1 AND status <> 'published'`, id, input.CategoryID, input.BrandID, input.Name, input.Slug, input.Description,
		input.Price.AmountMinor, input.Price.Currency, input.Dimensions.LengthMM, input.Dimensions.WidthMM,
		input.Dimensions.HeightMM, input.WeightGrams, input.MaxCapacityGrams, input.Material,
		input.WarrantyMonths, input.Scores.Quality, input.Scores.Value, input.Scores.Durability,
		input.Scores.Beginner, input.Scores.Advanced, input.Scores.Apartment, input.Scores.Noise,
		input.Scores.Portability)
	if err == nil && tag.RowsAffected() == 0 {
		err = ports.ErrConflict
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.update", "product", string(id), map[string]string{"slug": input.Slug})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return catalog.Product{}, fmt.Errorf("update product: %w", err)
	}
	return repository.GetProduct(ctx, id)
}

func (repository *Repository) SetProductStatus(ctx context.Context, actor identity.UserID, id catalog.ProductID, status catalog.ProductStatus) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE catalog.products SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	if err == nil && tag.RowsAffected() == 0 {
		err = ports.ErrNotFound
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.status", "product", string(id), map[string]string{"status": string(status)})
	}
	return finishMutation(ctx, tx, err)
}

func (repository *Repository) AddImage(ctx context.Context, actor identity.UserID, productID catalog.ProductID, input admin.ImageInput) (catalog.ProductImage, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return catalog.ProductImage{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if input.IsPrimary {
		_, err = tx.Exec(ctx, `UPDATE catalog.product_images SET is_primary=false WHERE product_id=$1 AND is_primary`, productID)
	}
	var image catalog.ProductImage
	if err == nil {
		err = tx.QueryRow(ctx, `
			INSERT INTO catalog.product_images (product_id, url, alt_text, sort_order, is_primary)
			SELECT $1,$2,$3,$4,$5 WHERE EXISTS (SELECT 1 FROM catalog.products WHERE id=$1)
			RETURNING id, url, alt_text, sort_order, is_primary`, productID, input.URL, input.AltText, input.SortOrder, input.IsPrimary).Scan(
			&image.ID, &image.URL, &image.AltText, &image.SortOrder, &image.IsPrimary)
		if errors.Is(err, pgx.ErrNoRows) {
			err = ports.ErrNotFound
		}
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.image.add", "product", string(productID), map[string]string{"image_id": image.ID})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return catalog.ProductImage{}, err
	}
	return image, nil
}

func (repository *Repository) DeleteImage(ctx context.Context, actor identity.UserID, productID catalog.ProductID, imageID string) (string, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var imageURL string
	err = tx.QueryRow(ctx, `DELETE FROM catalog.product_images WHERE id=$1 AND product_id=$2 RETURNING url`, imageID, productID).Scan(&imageURL)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ports.ErrNotFound
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.image.delete", "product", string(productID), map[string]string{"image_id": imageID})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return "", err
	}
	return imageURL, nil
}

func (repository *Repository) UpsertAttribute(ctx context.Context, actor identity.UserID, productID catalog.ProductID, input admin.AttributeInput) (catalog.Attribute, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return catalog.Attribute{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var attribute catalog.Attribute
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.product_attributes (
			product_id, attribute_key, attribute_type, numeric_value, text_value,
			boolean_value, unit, is_filterable
		) SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE EXISTS (SELECT 1 FROM catalog.products WHERE id=$1)
		ON CONFLICT (product_id, attribute_key) DO UPDATE SET
			attribute_type=EXCLUDED.attribute_type, numeric_value=EXCLUDED.numeric_value,
			text_value=EXCLUDED.text_value, boolean_value=EXCLUDED.boolean_value,
			unit=EXCLUDED.unit, is_filterable=EXCLUDED.is_filterable, updated_at=now()
		RETURNING id, attribute_key, attribute_type, numeric_value, text_value,
			boolean_value, unit, is_filterable`, productID, input.Key, input.Type,
		input.NumericValue, input.TextValue, input.BooleanValue, input.Unit, input.IsFilterable).Scan(
		&attribute.ID, &attribute.Key, &attribute.Type, &attribute.NumericValue,
		&attribute.TextValue, &attribute.BooleanValue, &attribute.Unit, &attribute.IsFilterable)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ports.ErrNotFound
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.attribute.upsert", "product", string(productID), map[string]string{"attribute_key": input.Key})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return catalog.Attribute{}, err
	}
	return attribute, nil
}

func (repository *Repository) DeleteAttribute(ctx context.Context, actor identity.UserID, productID catalog.ProductID, key string) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `DELETE FROM catalog.product_attributes WHERE product_id=$1 AND attribute_key=$2`, productID, key)
	if err == nil && tag.RowsAffected() == 0 {
		err = ports.ErrNotFound
	}
	if err == nil {
		err = audit(ctx, tx, actor, "product.attribute.delete", "product", string(productID), map[string]string{"attribute_key": key})
	}
	return finishMutation(ctx, tx, err)
}
