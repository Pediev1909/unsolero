package catalogpostgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	commercepostgres "rigmark/internal/adapters/postgres/commerce"
	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
)

func TestDemoCatalogRepositories(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	catalogRepository := New(pool)
	categories, err := catalogRepository.ListActiveCategories(ctx)
	if err != nil {
		t.Fatalf("ListActiveCategories() returned an error: %v", err)
	}
	if len(categories) != 15 {
		t.Fatalf("category count = %d, want 15", len(categories))
	}

	brands, err := catalogRepository.ListActiveBrands(ctx)
	if err != nil {
		t.Fatalf("ListActiveBrands() returned an error: %v", err)
	}
	if len(brands) != 8 {
		t.Fatalf("brand count = %d, want 8", len(brands))
	}

	products, err := catalogRepository.ListPublished(ctx, ports.ProductFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListPublished() returned an error: %v", err)
	}
	if len(products) != 14 {
		t.Fatalf("product count = %d, want 14", len(products))
	}
	for _, product := range products {
		if product.IsPhysical {
			t.Fatalf("SaaS fixture product %q is marked physical", product.Slug)
		}
	}
	requestedIDs := []catalog.ProductID{products[7].ID, products[2].ID, products[12].ID}
	selected, err := catalogRepository.SearchPublished(ctx, ports.ProductFilter{
		ProductIDs: requestedIDs, Limit: len(requestedIDs),
	})
	if err != nil {
		t.Fatalf("ID-filtered SearchPublished() returned an error: %v", err)
	}
	for index, productID := range requestedIDs {
		if len(selected.Products) != len(requestedIDs) || selected.Products[index].ID != productID {
			t.Fatalf("ID-filtered product order = %#v, want %#v", selected.Products, requestedIDs)
		}
	}

	page, err := catalogRepository.SearchPublished(ctx, ports.ProductFilter{
		Search: "fictional demo", Sort: "value_desc", Limit: 3,
	})
	if err != nil {
		t.Fatalf("SearchPublished() returned an error: %v", err)
	}
	if page.Total < len(page.Products) || len(page.Products) != 3 {
		t.Fatalf("search page = %d of %d, want 3 products with a matching total", len(page.Products), page.Total)
	}
	for index := 1; index < len(page.Products); index++ {
		if page.Products[index-1].Scores.Value < page.Products[index].Scores.Value {
			t.Fatal("value_desc search was not sorted in descending score order")
		}
	}

	minimum := int64(900)
	maximum := int64(2_900)
	filtered, err := catalogRepository.SearchPublished(ctx, ports.ProductFilter{
		CategorySlug: "crm", MinPriceMinor: &minimum,
		MaxPriceMinor: &maximum, Sort: "price_asc", Limit: 20,
	})
	if err != nil {
		t.Fatalf("filtered SearchPublished() returned an error: %v", err)
	}
	if len(filtered.Products) == 0 {
		t.Fatal("expected CRM products in the requested price range")
	}
	for _, filteredProduct := range filtered.Products {
		if filteredProduct.CategorySlug != "crm" ||
			filteredProduct.Price.AmountMinor < minimum || filteredProduct.Price.AmountMinor > maximum {
			t.Fatalf("product %q did not satisfy the requested filters", filteredProduct.Slug)
		}
	}

	product, err := catalogRepository.GetPublishedBySlug(ctx, "saas-northwind-crm")
	if err != nil {
		t.Fatalf("GetPublishedBySlug() returned an error: %v", err)
	}
	if product.IsPhysical || product.Price.AmountMinor != 2_900 {
		t.Fatalf("unexpected SaaS product projection: %#v", product)
	}

	commerceRepository := commercepostgres.New(pool)
	offers, err := commerceRepository.ListAvailableByProduct(ctx, catalog.ProductID(product.ID), "USD")
	if err != nil {
		t.Fatalf("ListAvailableByProduct() returned an error: %v", err)
	}
	if len(offers) != 1 {
		t.Fatalf("offer count = %d, want 1", len(offers))
	}
	for _, offer := range offers {
		if len(offer.AffiliateLinks) != 1 {
			t.Fatalf("affiliate link count for offer %s = %d, want 1", offer.ID, len(offer.AffiliateLinks))
		}
	}
	staleOffers, err := commercepostgres.New(pool, time.Nanosecond).
		ListAvailableByProduct(ctx, catalog.ProductID(product.ID), "USD")
	if err != nil {
		t.Fatalf("stale ListAvailableByProduct() returned an error: %v", err)
	}
	if len(staleOffers) != 0 {
		t.Fatalf("stale offer count = %d, want 0", len(staleOffers))
	}
}

func TestCoreProductConstraints(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `
		INSERT INTO catalog.products (
			category_id, brand_id, name, slug, description, price_minor, currency,
			warranty_months,
			quality_score, value_score, durability_score, beginner_score,
			advanced_score, apartment_score, noise_score, portability_score, status
		)
		SELECT categories.id, brands.id, 'Invalid score', 'integration-invalid-score',
			'Constraint test record that must never be inserted.', 1000, 'USD',
			0, 101, 50, 50, 50, 50, 0, 0, 50, 'draft'
		FROM catalog.categories AS categories
		CROSS JOIN catalog.brands AS brands
		WHERE categories.slug = 'crm' AND brands.slug = 'northwind-software'`)
	if err == nil {
		t.Fatal("database accepted a product score above 100")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO catalog.product_attributes (
			product_id, attribute_key, attribute_type, numeric_value, text_value
		)
		SELECT id, 'integration_invalid_shape', 'number', 1, 'also text'
		FROM catalog.products
		WHERE slug = 'saas-northwind-crm'`)
	if err == nil {
		t.Fatal("database accepted multiple typed attribute values")
	}
}
