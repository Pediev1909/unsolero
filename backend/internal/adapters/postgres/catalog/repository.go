package catalogpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
)

type Repository struct {
	pool *pgxpool.Pool
	// vertical scopes every public query. Migrations create each vertical's
	// categories in the same database, so without this a SaaS deployment
	// would list gym equipment categories alongside software ones.
	vertical string
}

// DefaultVertical is used when no vertical is configured.
const DefaultVertical = "saas"

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, vertical: DefaultVertical}
}

// NewForVertical builds a repository scoped to one vertical's catalog.
func NewForVertical(pool *pgxpool.Pool, vertical string) *Repository {
	if vertical == "" {
		vertical = DefaultVertical
	}
	return &Repository{pool: pool, vertical: vertical}
}

func (repository *Repository) ListActiveCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, parent_id, name, slug, description, sort_order, is_active,
			(SELECT count(*) FROM catalog.products
			 WHERE products.category_id = categories.id
			   AND products.status = 'published')
		FROM catalog.categories
		WHERE is_active = true AND vertical_key = $1
		ORDER BY sort_order, name`, repository.vertical)
	if err != nil {
		return nil, fmt.Errorf("list active categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		var parentID sql.NullString
		if err := rows.Scan(
			&category.ID,
			&parentID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.SortOrder,
			&category.IsActive,
			&category.PublishedProducts,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		if parentID.Valid {
			value := domain.CategoryID(parentID.String)
			category.ParentID = &value
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read categories: %w", err)
	}
	return categories, nil
}

func (repository *Repository) GetActiveCategoryBySlug(ctx context.Context, slug string) (domain.Category, error) {
	var category domain.Category
	var parentID sql.NullString
	err := repository.pool.QueryRow(ctx, `
		SELECT id, parent_id, name, slug, description, sort_order, is_active,
			(SELECT count(*) FROM catalog.products
			 WHERE products.category_id = categories.id
			   AND products.status = 'published')
		FROM catalog.categories
		WHERE slug = $1 AND is_active = true AND vertical_key = $2`, slug, repository.vertical).Scan(
		&category.ID,
		&parentID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.SortOrder,
		&category.IsActive,
		&category.PublishedProducts,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Category{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("get active category: %w", err)
	}
	if parentID.Valid {
		value := domain.CategoryID(parentID.String)
		category.ParentID = &value
	}
	return category, nil
}

func (repository *Repository) ListActiveBrands(ctx context.Context) ([]domain.Brand, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, name, slug, description, website_url, country_code, is_active,
			(SELECT count(*) FROM catalog.products
			 WHERE products.brand_id = brands.id
			   AND products.status = 'published')
		FROM catalog.brands
		WHERE is_active = true
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list active brands: %w", err)
	}
	defer rows.Close()

	brands := make([]domain.Brand, 0)
	for rows.Next() {
		var brand domain.Brand
		var websiteURL sql.NullString
		var countryCode sql.NullString
		if err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.Description,
			&websiteURL,
			&countryCode,
			&brand.IsActive,
			&brand.PublishedProducts,
		); err != nil {
			return nil, fmt.Errorf("scan brand: %w", err)
		}
		if websiteURL.Valid {
			brand.WebsiteURL = &websiteURL.String
		}
		if countryCode.Valid {
			brand.CountryCode = &countryCode.String
		}
		brands = append(brands, brand)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read brands: %w", err)
	}
	return brands, nil
}

// ListActiveBrandsInCategory returns brands with at least one published
// product in the category. The count it reports is the count *within* that
// category, not across the catalog, because a filter that says "Zoho (7)" on
// a page holding one Zoho product is lying about what selecting it will do.
func (repository *Repository) ListActiveBrandsInCategory(ctx context.Context, categorySlug string) ([]domain.Brand, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT brands.id, brands.name, brands.slug, brands.description,
			brands.website_url, brands.country_code, brands.is_active,
			count(products.id)
		FROM catalog.brands brands
		JOIN catalog.products products ON products.brand_id = brands.id
		 AND products.status = 'published'
		JOIN catalog.categories categories ON categories.id = products.category_id
		 AND categories.slug = $1 AND categories.vertical_key = $2
		WHERE brands.is_active = true
		GROUP BY brands.id, brands.name, brands.slug, brands.description,
			brands.website_url, brands.country_code, brands.is_active
		ORDER BY brands.name`, categorySlug, repository.vertical)
	if err != nil {
		return nil, fmt.Errorf("list active brands in category: %w", err)
	}
	defer rows.Close()

	brands := make([]domain.Brand, 0)
	for rows.Next() {
		var brand domain.Brand
		var websiteURL, countryCode sql.NullString
		if err := rows.Scan(
			&brand.ID, &brand.Name, &brand.Slug, &brand.Description,
			&websiteURL, &countryCode, &brand.IsActive, &brand.PublishedProducts,
		); err != nil {
			return nil, fmt.Errorf("scan category brand: %w", err)
		}
		if websiteURL.Valid {
			brand.WebsiteURL = &websiteURL.String
		}
		if countryCode.Valid {
			brand.CountryCode = &countryCode.String
		}
		brands = append(brands, brand)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read category brands: %w", err)
	}
	return brands, nil
}

func (repository *Repository) GetActiveBrandBySlug(ctx context.Context, slug string) (domain.Brand, error) {
	var brand domain.Brand
	var websiteURL sql.NullString
	var countryCode sql.NullString
	err := repository.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, website_url, country_code, is_active,
			(SELECT count(*) FROM catalog.products
			 WHERE products.brand_id = brands.id
			   AND products.status = 'published')
		FROM catalog.brands
		WHERE slug = $1 AND is_active = true`, slug).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.Description,
		&websiteURL,
		&countryCode,
		&brand.IsActive,
		&brand.PublishedProducts,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Brand{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Brand{}, fmt.Errorf("get active brand: %w", err)
	}
	if websiteURL.Valid {
		brand.WebsiteURL = &websiteURL.String
	}
	if countryCode.Valid {
		brand.CountryCode = &countryCode.String
	}
	return brand, nil
}

func (repository *Repository) GetPublishedBySlug(ctx context.Context, slug string) (domain.Product, error) {
	page, err := repository.searchProducts(ctx, ports.ProductFilter{Limit: 1}, slug, true)
	if err != nil {
		return domain.Product{}, err
	}
	if len(page.Products) == 0 {
		return domain.Product{}, ports.ErrNotFound
	}
	return page.Products[0], nil
}

func (repository *Repository) ListPublished(ctx context.Context, filter ports.ProductFilter) ([]domain.Product, error) {
	page, err := repository.SearchPublished(ctx, filter)
	return page.Products, err
}

func (repository *Repository) SearchPublished(ctx context.Context, filter ports.ProductFilter) (ports.ProductPage, error) {
	return repository.searchProducts(ctx, filter, "", true)
}

// ListByIDs intentionally includes archived products. Persisted recommendations
// are immutable decision records and must remain readable after catalog status
// changes. Products are loaded in bounded chunks to preserve repository limits.
func (repository *Repository) ListByIDs(
	ctx context.Context,
	ids []domain.ProductID,
) ([]domain.Product, error) {
	products := make([]domain.Product, 0, len(ids))
	for start := 0; start < len(ids); start += 100 {
		end := start + 100
		if end > len(ids) {
			end = len(ids)
		}
		page, err := repository.searchProducts(ctx, ports.ProductFilter{
			ProductIDs: ids[start:end],
			Limit:      end - start,
		}, "", false)
		if err != nil {
			return nil, err
		}
		products = append(products, page.Products...)
	}
	return products, nil
}

func (repository *Repository) searchProducts(
	ctx context.Context,
	filter ports.ProductFilter,
	productSlug string,
	publishedOnly bool,
) (ports.ProductPage, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 24
	}
	if limit > 100 {
		limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	rows, err := repository.pool.Query(ctx, `
		SELECT
			count(*) OVER(),
			products.id,
			products.category_id,
			categories.name,
			categories.slug,
			products.brand_id,
			brands.name,
			brands.slug,
			products.name,
			products.slug,
			products.description,
			products.price_minor,
			products.currency,
			categories.is_physical,
			products.length_mm,
			products.width_mm,
			products.height_mm,
			products.weight_grams,
			products.max_capacity_grams,
			products.material,
			products.warranty_months,
			products.quality_score,
			products.value_score,
			products.durability_score,
			products.beginner_score,
			products.advanced_score,
			products.apartment_score,
			products.noise_score,
			products.portability_score,
			products.published_fact_revision_id,
			products.published_score_revision_id,
			products.status,
			products.created_at,
			products.updated_at
		FROM catalog.products AS products
		JOIN catalog.categories AS categories ON categories.id = products.category_id
		JOIN catalog.brands AS brands ON brands.id = products.brand_id
		WHERE ($12::boolean = false OR (
			products.status = 'published'
			AND products.published_fact_revision_id IS NOT NULL
			AND products.published_score_revision_id IS NOT NULL
			AND EXISTS (
				SELECT 1 FROM evidence.product_fact_revisions facts
				WHERE facts.id = products.published_fact_revision_id
				  AND facts.workflow_status = 'published'
				  AND (facts.valid_until IS NULL OR facts.valid_until > now())
				  AND NOT EXISTS (
					SELECT 1 FROM evidence.fact_provenance provenance
					JOIN evidence.observations observations ON observations.id = provenance.observation_id
					JOIN evidence.sources sources ON sources.id = observations.source_id
					WHERE provenance.fact_revision_id = facts.id
					  AND (sources.review_status <> 'verified'
						OR (observations.expires_at IS NOT NULL AND observations.expires_at <= now()))
				  )
			)
			AND EXISTS (
				SELECT 1 FROM evidence.score_revisions scores
				WHERE scores.id = products.published_score_revision_id
				  AND scores.workflow_status = 'published'
				  AND NOT EXISTS (
					SELECT 1 FROM evidence.score_rationales rationales
					JOIN evidence.observations observations ON observations.id = rationales.observation_id
					JOIN evidence.sources sources ON sources.id = observations.source_id
					WHERE rationales.score_revision_id = scores.id
					  AND (sources.review_status <> 'verified'
						OR (observations.expires_at IS NOT NULL AND observations.expires_at <= now()))
				  )
			)
		))
			AND ($1 = '' OR categories.slug = $1)
			AND ($2 = '' OR brands.slug = $2)
			AND ($3 = '' OR products.name ILIKE '%' || $3 || '%'
				OR products.description ILIKE '%' || $3 || '%'
				OR brands.name ILIKE '%' || $3 || '%')
			AND ($4::bigint IS NULL OR products.price_minor >= $4)
			AND ($5::bigint IS NULL OR products.price_minor <= $5)
			AND ($6 = '' OR products.slug = $6)
			AND ($7 = '' OR products.slug <> $7)
			AND ($11::uuid[] IS NULL OR products.id = ANY($11::uuid[]))
			AND categories.vertical_key = $13
		ORDER BY
			CASE WHEN $11::uuid[] IS NOT NULL THEN array_position($11::uuid[], products.id) END ASC,
			CASE WHEN $8 = 'price_asc' THEN products.price_minor END ASC,
			CASE WHEN $8 = 'price_desc' THEN products.price_minor END DESC,
			CASE WHEN $8 = 'name_asc' THEN lower(products.name) END ASC,
			CASE WHEN $8 = 'quality_desc' THEN products.quality_score END DESC,
			CASE WHEN $8 = 'value_desc' THEN products.value_score END DESC,
			CASE WHEN $8 = 'featured' THEN products.quality_score + products.value_score END DESC,
			products.name ASC,
			products.id ASC
		LIMIT $9 OFFSET $10`,
		filter.CategorySlug,
		filter.BrandSlug,
		filter.Search,
		filter.MinPriceMinor,
		filter.MaxPriceMinor,
		productSlug,
		filter.ExcludeSlug,
		filter.Sort,
		limit,
		filter.Offset,
		productIDs(filter.ProductIDs),
		publishedOnly,
		repository.vertical,
	)
	if err != nil {
		return ports.ProductPage{}, fmt.Errorf("list published products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	productIDs := make([]string, 0)
	total := 0
	for rows.Next() {
		var product domain.Product
		var maxCapacity sql.NullInt64
		var factRevisionID, scoreRevisionID sql.NullString
		// Physical columns are null for a non-physical category, so they are
		// scanned through nullable holders and left at their zero value.
		var lengthMM, widthMM, heightMM, weightGrams sql.NullInt64
		var material sql.NullString
		if err := rows.Scan(
			&total,
			&product.ID,
			&product.CategoryID,
			&product.CategoryName,
			&product.CategorySlug,
			&product.BrandID,
			&product.BrandName,
			&product.BrandSlug,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.Price.AmountMinor,
			&product.Price.Currency,
			&product.IsPhysical,
			&lengthMM,
			&widthMM,
			&heightMM,
			&weightGrams,
			&maxCapacity,
			&material,
			&product.WarrantyMonths,
			&product.Scores.Quality,
			&product.Scores.Value,
			&product.Scores.Durability,
			&product.Scores.Beginner,
			&product.Scores.Advanced,
			&product.Scores.Apartment,
			&product.Scores.Noise,
			&product.Scores.Portability,
			&factRevisionID,
			&scoreRevisionID,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return ports.ProductPage{}, fmt.Errorf("scan product: %w", err)
		}
		product.Dimensions = domain.Dimensions{
			LengthMM: lengthMM.Int64, WidthMM: widthMM.Int64, HeightMM: heightMM.Int64,
		}
		product.WeightGrams = weightGrams.Int64
		product.Material = material.String
		if maxCapacity.Valid {
			product.MaxCapacityGrams = &maxCapacity.Int64
		}
		product.FactRevisionID = factRevisionID.String
		product.ScoreRevisionID = scoreRevisionID.String
		products = append(products, product)
		productIDs = append(productIDs, string(product.ID))
	}
	if err := rows.Err(); err != nil {
		return ports.ProductPage{}, fmt.Errorf("read products: %w", err)
	}
	if len(products) == 0 {
		return ports.ProductPage{Products: products}, nil
	}

	productIndexes := make(map[string]int, len(products))
	for index := range products {
		productIndexes[string(products[index].ID)] = index
	}
	if err := repository.loadAttributes(ctx, productIDs, products, productIndexes); err != nil {
		return ports.ProductPage{}, err
	}
	if err := repository.loadImages(ctx, productIDs, products, productIndexes); err != nil {
		return ports.ProductPage{}, err
	}
	if publishedOnly {
		if err := repository.loadEvidence(ctx, productIDs, products, productIndexes); err != nil {
			return ports.ProductPage{}, err
		}
	}
	for _, product := range products {
		if err := product.Validate(); err != nil {
			return ports.ProductPage{}, fmt.Errorf("validate persisted product %q: %w", product.Slug, err)
		}
	}

	return ports.ProductPage{Products: products, Total: total}, nil
}

func (repository *Repository) loadEvidence(
	ctx context.Context,
	productIDs []string,
	products []domain.Product,
	productIndexes map[string]int,
) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT facts.product_id, provenance.fact_key, provenance.public_classification,
			sources.source_type, sources.title, sources.source_url,
			observations.observed_at, observations.expires_at,
			observations.confidence, sources.is_fictional
		FROM evidence.product_fact_revisions facts
		JOIN evidence.fact_provenance provenance ON provenance.fact_revision_id = facts.id
		JOIN evidence.observations observations ON observations.id = provenance.observation_id
		JOIN evidence.sources sources ON sources.id = observations.source_id
		JOIN catalog.products products ON products.published_fact_revision_id = facts.id
		WHERE facts.product_id = ANY($1::uuid[])
		  AND facts.workflow_status = 'published'
		  AND sources.review_status = 'verified'
		UNION ALL
		SELECT scores.product_id, 'score.' || rationales.score_key, 'editorial_assessment',
			sources.source_type, sources.title, sources.source_url,
			observations.observed_at, observations.expires_at,
			observations.confidence, sources.is_fictional
		FROM evidence.score_revisions scores
		JOIN evidence.score_rationales rationales ON rationales.score_revision_id = scores.id
		JOIN evidence.observations observations ON observations.id = rationales.observation_id
		JOIN evidence.sources sources ON sources.id = observations.source_id
		JOIN catalog.products products ON products.published_score_revision_id = scores.id
		WHERE scores.product_id = ANY($1::uuid[])
		  AND scores.workflow_status = 'published'
		  AND sources.review_status = 'verified'
		ORDER BY 1, 2, 5`, productIDs)
	if err != nil {
		return fmt.Errorf("list product evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var productID string
		var item domain.ProductEvidence
		var sourceURL sql.NullString
		var expiresAt sql.NullTime
		if err := rows.Scan(&productID, &item.FactKey, &item.Classification,
			&item.SourceType, &item.SourceTitle, &sourceURL, &item.ObservedAt,
			&expiresAt, &item.Confidence, &item.IsFictional); err != nil {
			return fmt.Errorf("scan product evidence: %w", err)
		}
		if sourceURL.Valid {
			item.SourceURL = &sourceURL.String
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		index, exists := productIndexes[productID]
		if !exists {
			return errors.New("evidence referenced a product outside the requested set")
		}
		products[index].Evidence = append(products[index].Evidence, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read product evidence: %w", err)
	}
	return nil
}

func productIDs(values []domain.ProductID) []string {
	if values == nil {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func (repository *Repository) loadAttributes(
	ctx context.Context,
	productIDs []string,
	products []domain.Product,
	productIndexes map[string]int,
) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, product_id, attribute_key, attribute_type,
			numeric_value, text_value, boolean_value, unit, is_filterable
		FROM catalog.product_attributes
		WHERE product_id = ANY($1::uuid[])
		ORDER BY product_id, attribute_key`, productIDs)
	if err != nil {
		return fmt.Errorf("list product attributes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var attribute domain.Attribute
		var productID string
		var numericValue sql.NullFloat64
		var textValue sql.NullString
		var booleanValue sql.NullBool
		var unit sql.NullString
		if err := rows.Scan(
			&attribute.ID,
			&productID,
			&attribute.Key,
			&attribute.Type,
			&numericValue,
			&textValue,
			&booleanValue,
			&unit,
			&attribute.IsFilterable,
		); err != nil {
			return fmt.Errorf("scan product attribute: %w", err)
		}
		if numericValue.Valid {
			attribute.NumericValue = &numericValue.Float64
		}
		if textValue.Valid {
			attribute.TextValue = &textValue.String
		}
		if booleanValue.Valid {
			attribute.BooleanValue = &booleanValue.Bool
		}
		if unit.Valid {
			attribute.Unit = &unit.String
		}
		index, exists := productIndexes[productID]
		if !exists {
			return errors.New("attribute referenced a product outside the requested set")
		}
		products[index].Attributes = append(products[index].Attributes, attribute)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read product attributes: %w", err)
	}
	return nil
}

func (repository *Repository) loadImages(
	ctx context.Context,
	productIDs []string,
	products []domain.Product,
	productIndexes map[string]int,
) error {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, product_id, url, alt_text, sort_order, is_primary, width_px, height_px
		FROM catalog.product_images
		WHERE product_id = ANY($1::uuid[])
		ORDER BY product_id, sort_order, id`, productIDs)
	if err != nil {
		return fmt.Errorf("list product images: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var image domain.ProductImage
		var productID string
		var width sql.NullInt32
		var height sql.NullInt32
		if err := rows.Scan(
			&image.ID,
			&productID,
			&image.URL,
			&image.AltText,
			&image.SortOrder,
			&image.IsPrimary,
			&width,
			&height,
		); err != nil {
			return fmt.Errorf("scan product image: %w", err)
		}
		if width.Valid {
			value := int(width.Int32)
			image.WidthPX = &value
		}
		if height.Valid {
			value := int(height.Int32)
			image.HeightPX = &value
		}
		index, exists := productIndexes[productID]
		if !exists {
			return errors.New("image referenced a product outside the requested set")
		}
		products[index].Images = append(products[index].Images, image)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read product images: %w", err)
	}
	return nil
}

var (
	_ ports.CategoryRepository = (*Repository)(nil)
	_ ports.BrandRepository    = (*Repository)(nil)
	_ ports.ProductRepository  = (*Repository)(nil)
)
