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
	if len(categories) != 8 {
		t.Fatalf("category count = %d, want 8", len(categories))
	}

	brands, err := catalogRepository.ListActiveBrands(ctx)
	if err != nil {
		t.Fatalf("ListActiveBrands() returned an error: %v", err)
	}
	if len(brands) != 10 {
		t.Fatalf("brand count = %d, want 10", len(brands))
	}

	products, err := catalogRepository.ListPublished(ctx, ports.ProductFilter{Limit: 100})
	if err != nil {
		t.Fatalf("ListPublished() returned an error: %v", err)
	}
	if len(products) != 30 {
		t.Fatalf("product count = %d, want 30", len(products))
	}
	for _, product := range products {
		if len(product.Images) == 0 || product.Images[0].URL == "" {
			t.Fatalf("product %q has no catalog image", product.Slug)
		}
	}
	requestedIDs := []catalog.ProductID{products[7].ID, products[2].ID, products[19].ID}
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
		Search: "compact", Sort: "value_desc", Limit: 3,
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

	minimum := int64(15_000)
	maximum := int64(30_000)
	filtered, err := catalogRepository.SearchPublished(ctx, ports.ProductFilter{
		CategorySlug: "adjustable-dumbbells", MinPriceMinor: &minimum,
		MaxPriceMinor: &maximum, Sort: "price_asc", Limit: 20,
	})
	if err != nil {
		t.Fatalf("filtered SearchPublished() returned an error: %v", err)
	}
	if len(filtered.Products) == 0 {
		t.Fatal("expected adjustable dumbbells in the requested price range")
	}
	for _, filteredProduct := range filtered.Products {
		if filteredProduct.CategorySlug != "adjustable-dumbbells" ||
			filteredProduct.Price.AmountMinor < minimum || filteredProduct.Price.AmountMinor > maximum {
			t.Fatalf("product %q did not satisfy the requested filters", filteredProduct.Slug)
		}
	}

	product, err := catalogRepository.GetPublishedBySlug(ctx, "demo-range-lab-adjustable-20-kettlebell")
	if err != nil {
		t.Fatalf("GetPublishedBySlug() returned an error: %v", err)
	}
	if len(product.Attributes) == 0 {
		t.Fatal("expected typed product attributes")
	}

	commerceRepository := commercepostgres.New(pool)
	offers, err := commerceRepository.ListAvailableByProduct(ctx, catalog.ProductID(product.ID), "USD")
	if err != nil {
		t.Fatalf("ListAvailableByProduct() returned an error: %v", err)
	}
	if len(offers) != 3 {
		t.Fatalf("offer count = %d, want 3", len(offers))
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
			length_mm, width_mm, height_mm, weight_grams, material, warranty_months,
			quality_score, value_score, durability_score, beginner_score,
			advanced_score, apartment_score, noise_score, portability_score, status
		)
		SELECT categories.id, brands.id, 'Invalid score', 'integration-invalid-score',
			'Constraint test record that must never be inserted.', 1000, 'USD',
			100, 100, 100, 1000, 'Steel', 12,
			101, 50, 50, 50, 50, 50, 50, 50, 'draft'
		FROM catalog.categories AS categories
		CROSS JOIN catalog.brands AS brands
		WHERE categories.slug = 'barbells' AND brands.slug = 'demo-northline'`)
	if err == nil {
		t.Fatal("database accepted a product score above 100")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO catalog.product_attributes (
			product_id, attribute_key, attribute_type, numeric_value, text_value
		)
		SELECT id, 'integration_invalid_shape', 'number', 1, 'also text'
		FROM catalog.products
		WHERE slug = 'demo-northline-nest-24-pair'`)
	if err == nil {
		t.Fatal("database accepted multiple typed attribute values")
	}
}
