package contentpostgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

func TestPublishedEditorialContentAndSitemap(t *testing.T) {
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

	repository := New(pool)
	guides, err := repository.ListPublished(ctx, ports.Filter{
		Types: []domain.ContentType{domain.ContentTypeGuide, domain.ContentTypeBuyingGuide},
		Limit: 12,
	})
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if len(guides) != 2 {
		t.Fatalf("published guide count = %d, want 2", len(guides))
	}

	entry, err := repository.GetPublishedBySlug(ctx, "best-adjustable-dumbbells")
	if err != nil {
		t.Fatalf("GetPublishedBySlug() error = %v", err)
	}
	entry.Path = entry.Type.Path(entry.Slug)
	entry.CanonicalURL = "https://rigmark.example" + entry.Path
	if err := entry.Validate(); err != nil {
		t.Fatalf("seeded entry validation error = %v", err)
	}
	if len(entry.ProductIDs) != 3 || len(entry.RelatedCategories) != 1 || len(entry.RelatedEntries) != 2 {
		t.Fatalf("editorial relationships = products:%d categories:%d entries:%d", len(entry.ProductIDs), len(entry.RelatedCategories), len(entry.RelatedEntries))
	}

	sitemap, err := repository.ListSitemapEntries(ctx)
	if err != nil {
		t.Fatalf("ListSitemapEntries() error = %v", err)
	}
	foundGuide := false
	for _, item := range sitemap {
		if item.Path == "/guides/best-adjustable-dumbbells" {
			foundGuide = true
			break
		}
	}
	if !foundGuide {
		t.Fatal("sitemap did not include published buying guide")
	}
}
