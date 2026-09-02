package contentpostgres

import (
	"context"
	"fmt"
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

	nonce := time.Now().UnixNano()
	authorSlug := fmt.Sprintf("integration-author-%d", nonce)
	entrySlug := fmt.Sprintf("saas-stack-integration-%d", nonce)
	stackSlug := fmt.Sprintf("agency-stack-integration-%d", nonce)
	var authorID, entryID, stackID, productID, productSlug, categoryID string
	if err = pool.QueryRow(ctx, `INSERT INTO editorial.authors (name,slug,bio)
		VALUES ('Integration Author',$1,'A fictional author used only to verify the editorial repository integration path.') RETURNING id`,
		authorSlug).Scan(&authorID); err != nil {
		t.Fatalf("insert author fixture: %v", err)
	}
	if err = pool.QueryRow(ctx, `SELECT products.id,products.slug,products.category_id FROM catalog.products AS products
		JOIN catalog.categories AS categories ON categories.id=products.category_id
		WHERE products.status='published' AND categories.vertical_key='saas' ORDER BY products.id LIMIT 1`).Scan(
		&productID, &productSlug, &categoryID); err != nil {
		t.Fatalf("load SaaS relationship fixture: %v", err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO editorial.entries
		(author_id,content_type,status,title,slug,description,hero_image_url,hero_image_alt,
		 content,seo_title,seo_description,published_at)
		VALUES($1,'guide','published','A practical SaaS stack integration guide',$2,
		'A fictional guide that verifies published editorial listing, relationships and sitemap behavior.',
		'/images/saas-stack-planning-v2.svg','A diagram of a fictional software stack',
		'[{"type":"paragraph","text":"Choose software for the job it must perform, then verify the complete stack stays within budget."}]'::jsonb,
		'A practical SaaS stack integration guide',
		'A fictional SaaS guide used to verify repository listing, relationship loading and sitemap behavior.',now()) RETURNING id`,
		authorID, entrySlug).Scan(&entryID); err != nil {
		t.Fatalf("insert entry fixture: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO editorial.entry_products VALUES($1,$2,0)`, entryID, productID); err != nil {
		t.Fatalf("insert editorial product relationship: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO editorial.entry_categories VALUES($1,$2,0)`, entryID, categoryID); err != nil {
		t.Fatalf("insert editorial relationships: %v", err)
	}
	// A stack is the newest content type and the only one whose path segment
	// differs from every earlier one, so it is the row most likely to be filed
	// under /guides/ by a CASE that predates it.
	if err = pool.QueryRow(ctx, `INSERT INTO editorial.entries
		(author_id,content_type,status,title,slug,description,hero_image_url,hero_image_alt,
		 content,seo_title,seo_description,published_at)
		VALUES($1,'stack','published','A fictional agency stack for integration tests',$2,
		'A fictional stack that verifies the stack content type lists, resolves and appears in the sitemap.',
		'/images/saas-agency-stack-v2.svg','A diagram of a fictional software stack',
		'[{"type":"paragraph","text":"Three tools for three people, and what was left out."}]'::jsonb,
		'A fictional agency stack for integration tests',
		'A fictional stack used to verify that the stack content type is stored, listed and mapped to its route.',now()) RETURNING id`,
		authorID, stackSlug).Scan(&stackID); err != nil {
		t.Fatalf("insert stack fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM editorial.entries WHERE id=$1`, entryID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM editorial.entries WHERE id=$1`, stackID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM editorial.authors WHERE id=$1`, authorID)
	})

	repository := New(pool)
	guides, err := repository.ListPublished(ctx, ports.Filter{
		Types: []domain.ContentType{domain.ContentTypeGuide, domain.ContentTypeBuyingGuide},
		Limit: 12,
	})
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if len(guides) != 1 {
		t.Fatalf("published guide count = %d, want 1", len(guides))
	}
	stacks, err := repository.ListPublished(ctx, ports.Filter{Types: []domain.ContentType{domain.ContentTypeStack}, Limit: 24})
	if err != nil {
		t.Fatalf("ListPublished(stacks) error = %v", err)
	}
	foundStack := false
	for _, item := range stacks {
		if item.Slug == stackSlug {
			foundStack = item.Path == "/stacks/"+stackSlug
		}
	}
	if !foundStack {
		t.Fatalf("ListPublished(stacks) did not list the stack at its route: %#v", stacks)
	}

	// The product filter is what a product page uses to find the pieces it
	// appears in. The fixture links exactly one product, so its slug finds the
	// entry and an unrelated slug finds nothing.
	byProduct, err := repository.ListPublished(ctx, ports.Filter{ProductSlug: productSlug, Limit: 12})
	if err != nil {
		t.Fatalf("ListPublished(product) error = %v", err)
	}
	foundByProduct := false
	for _, item := range byProduct {
		if item.Slug == entrySlug {
			foundByProduct = true
		}
	}
	if !foundByProduct {
		t.Fatalf("ListPublished(product=%s) did not include the entry that references it", productSlug)
	}
	unrelated, err := repository.ListPublished(ctx, ports.Filter{ProductSlug: fmt.Sprintf("no-such-product-%d", nonce), Limit: 12})
	if err != nil {
		t.Fatalf("ListPublished(unrelated product) error = %v", err)
	}
	if len(unrelated) != 0 {
		t.Fatalf("ListPublished(unrelated product) = %d entries, want 0", len(unrelated))
	}

	entry, err := repository.GetPublishedBySlug(ctx, entrySlug)
	if err != nil {
		t.Fatalf("GetPublishedBySlug() error = %v", err)
	}
	entry.Path = entry.Type.Path(entry.Slug)
	entry.CanonicalURL = "https://rigmark.example" + entry.Path
	if err := entry.Validate(); err != nil {
		t.Fatalf("seeded entry validation error = %v", err)
	}
	if len(entry.ProductIDs) != 1 || len(entry.RelatedCategories) != 1 || len(entry.RelatedEntries) != 0 {
		t.Fatalf("editorial relationships = products:%d categories:%d entries:%d", len(entry.ProductIDs), len(entry.RelatedCategories), len(entry.RelatedEntries))
	}

	sitemap, err := repository.ListSitemapEntries(ctx)
	if err != nil {
		t.Fatalf("ListSitemapEntries() error = %v", err)
	}
	foundGuide, foundOffers, foundStackHub, foundStack := false, false, false, false
	for _, item := range sitemap {
		switch item.Path {
		case "/guides/" + entrySlug:
			foundGuide = true
		case "/offers":
			foundOffers = true
		case "/stacks":
			foundStackHub = true
		case "/stacks/" + stackSlug:
			foundStack = true
		case "/guides/" + stackSlug:
			t.Fatal("sitemap filed the stack under /guides/")
		}
	}
	if !foundGuide {
		t.Fatal("sitemap did not include published buying guide")
	}
	if !foundOffers {
		t.Fatal("sitemap did not include the /offers page")
	}
	if !foundStackHub || !foundStack {
		t.Fatalf("sitemap stack routes: hub=%v entry=%v", foundStackHub, foundStack)
	}
}
