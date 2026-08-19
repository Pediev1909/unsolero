package contentpostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) ListPublished(ctx context.Context, filter ports.Filter) ([]domain.Summary, error) {
	types := make([]string, len(filter.Types))
	for index, contentType := range filter.Types {
		types[index] = string(contentType)
	}
	var typeFilter []string
	if len(types) > 0 {
		typeFilter = types
	}
	rows, err := repository.pool.Query(ctx, `
		SELECT entries.id, entries.content_type, entries.title, entries.slug,
			entries.description, entries.hero_image_url, entries.hero_image_alt,
			authors.name, entries.published_at, entries.updated_at
		FROM editorial.entries AS entries
		JOIN editorial.authors AS authors ON authors.id = entries.author_id
		WHERE entries.status = 'published'
			AND ($1::text[] IS NULL OR entries.content_type = ANY($1))
			AND ($2 = '' OR EXISTS (
				SELECT 1
				FROM editorial.entry_categories
				JOIN catalog.categories ON categories.id = entry_categories.category_id
				WHERE entry_categories.entry_id = entries.id AND categories.slug = $2
			))
			AND ($3 = '' OR entries.id::text <> $3)
		ORDER BY entries.published_at DESC, entries.id
		LIMIT $4`, typeFilter, filter.CategorySlug, filter.ExcludeID, filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("list published editorial entries: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.Summary, 0)
	for rows.Next() {
		summary, err := scanSummary(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read published editorial entries: %w", err)
	}
	return entries, nil
}

func (repository *Repository) GetPublishedBySlug(ctx context.Context, slug string) (domain.Entry, error) {
	var entry domain.Entry
	var rawContent []byte
	var canonical sql.NullString
	var avatar sql.NullString
	err := repository.pool.QueryRow(ctx, `
		SELECT entries.id, entries.content_type, entries.title, entries.slug,
			entries.description, entries.hero_image_url, entries.hero_image_alt,
			authors.name, entries.published_at, entries.updated_at,
			authors.id, authors.slug, authors.bio, authors.avatar_url,
			entries.content, entries.seo_title, entries.seo_description,
			entries.canonical_url
		FROM editorial.entries AS entries
		JOIN editorial.authors AS authors ON authors.id = entries.author_id
		WHERE entries.slug = $1 AND entries.status = 'published'`, slug).Scan(
		&entry.ID, &entry.Type, &entry.Title, &entry.Slug,
		&entry.Description, &entry.HeroImageURL, &entry.HeroImageAlt,
		&entry.AuthorName, &entry.PublishedAt, &entry.UpdatedAt,
		&entry.Author.ID, &entry.Author.Slug, &entry.Author.Bio, &avatar,
		&rawContent, &entry.SEOTitle, &entry.SEODescription, &canonical,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Entry{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Entry{}, fmt.Errorf("get published editorial entry: %w", err)
	}
	entry.Author.Name = entry.AuthorName
	if avatar.Valid {
		entry.Author.AvatarURL = &avatar.String
	}
	if canonical.Valid {
		entry.CanonicalURL = canonical.String
	}
	if err := json.Unmarshal(rawContent, &entry.Content); err != nil {
		return domain.Entry{}, fmt.Errorf("decode editorial content: %w", err)
	}

	if err := repository.loadRelationships(ctx, &entry); err != nil {
		return domain.Entry{}, err
	}
	return entry, nil
}

func (repository *Repository) loadRelationships(ctx context.Context, entry *domain.Entry) error {
	productRows, err := repository.pool.Query(ctx, `
		SELECT product_id
		FROM editorial.entry_products
		WHERE entry_id = $1
		ORDER BY position`, entry.ID)
	if err != nil {
		return fmt.Errorf("list editorial products: %w", err)
	}
	for productRows.Next() {
		var id catalog.ProductID
		if err := productRows.Scan(&id); err != nil {
			productRows.Close()
			return fmt.Errorf("scan editorial product: %w", err)
		}
		entry.ProductIDs = append(entry.ProductIDs, id)
	}
	if err := productRows.Err(); err != nil {
		productRows.Close()
		return fmt.Errorf("read editorial products: %w", err)
	}
	productRows.Close()

	categoryRows, err := repository.pool.Query(ctx, `
		SELECT categories.id, categories.name, categories.slug, categories.description
		FROM editorial.entry_categories
		JOIN catalog.categories ON categories.id = entry_categories.category_id
		WHERE entry_categories.entry_id = $1 AND categories.is_active = true
		ORDER BY entry_categories.position`, entry.ID)
	if err != nil {
		return fmt.Errorf("list editorial categories: %w", err)
	}
	for categoryRows.Next() {
		var category domain.CategoryReference
		if err := categoryRows.Scan(&category.ID, &category.Name, &category.Slug, &category.Description); err != nil {
			categoryRows.Close()
			return fmt.Errorf("scan editorial category: %w", err)
		}
		entry.RelatedCategories = append(entry.RelatedCategories, category)
	}
	if err := categoryRows.Err(); err != nil {
		categoryRows.Close()
		return fmt.Errorf("read editorial categories: %w", err)
	}
	categoryRows.Close()

	relatedRows, err := repository.pool.Query(ctx, `
		SELECT related.id, related.content_type, related.title, related.slug,
			related.description, related.hero_image_url, related.hero_image_alt,
			authors.name, related.published_at, related.updated_at
		FROM editorial.related_entries
		JOIN editorial.entries AS related ON related.id = related_entries.related_entry_id
		JOIN editorial.authors AS authors ON authors.id = related.author_id
		WHERE related_entries.entry_id = $1 AND related.status = 'published'
		ORDER BY related_entries.position`, entry.ID)
	if err != nil {
		return fmt.Errorf("list related editorial entries: %w", err)
	}
	for relatedRows.Next() {
		summary, err := scanSummary(relatedRows)
		if err != nil {
			relatedRows.Close()
			return err
		}
		entry.RelatedEntries = append(entry.RelatedEntries, summary)
	}
	if err := relatedRows.Err(); err != nil {
		relatedRows.Close()
		return fmt.Errorf("read related editorial entries: %w", err)
	}
	relatedRows.Close()
	return nil
}

func (repository *Repository) ListSitemapEntries(ctx context.Context) ([]domain.SitemapEntry, error) {
	rows, err := repository.pool.Query(ctx, `
		WITH public_updates AS (
			SELECT
				GREATEST(
					COALESCE((SELECT max(updated_at) FROM editorial.entries WHERE status = 'published'), to_timestamp(0)),
					COALESCE((SELECT max(updated_at) FROM catalog.products WHERE status = 'published'), to_timestamp(0)),
					COALESCE((SELECT max(updated_at) FROM catalog.categories WHERE is_active = true), to_timestamp(0))
				) AS site_updated_at,
				COALESCE((SELECT max(updated_at) FROM editorial.entries WHERE status = 'published'), to_timestamp(0)) AS editorial_updated_at,
				COALESCE((SELECT max(updated_at) FROM catalog.products WHERE status = 'published'), to_timestamp(0)) AS catalog_updated_at
		), routes AS (
			SELECT '/' AS path, site_updated_at AS updated_at FROM public_updates
			UNION ALL SELECT '/products', catalog_updated_at FROM public_updates
			UNION ALL SELECT '/guides', editorial_updated_at FROM public_updates
			UNION ALL SELECT '/articles', editorial_updated_at FROM public_updates
			-- The legal pages are static, so they carry the site timestamp.
			-- They belong here because affiliate programme reviews and search
			-- engines both look for them on a commercial site.
			UNION ALL SELECT '/privacy', site_updated_at FROM public_updates
			UNION ALL SELECT '/affiliate-disclosure', site_updated_at FROM public_updates
			UNION ALL
			SELECT '/categories/' || slug, updated_at
			FROM catalog.categories WHERE is_active = true
			UNION ALL
			SELECT '/brands/' || slug, updated_at
			FROM catalog.brands WHERE is_active = true AND slug NOT LIKE 'demo-%'
			UNION ALL
			SELECT '/products/' || slug, updated_at
			FROM catalog.products WHERE status = 'published' AND slug NOT LIKE 'demo-%'
			UNION ALL
			SELECT CASE content_type
				WHEN 'article' THEN '/articles/' || slug
				WHEN 'comparison' THEN '/compare/' || slug
				ELSE '/guides/' || slug
			END, updated_at
			FROM editorial.entries WHERE status = 'published'
		)
		SELECT path, updated_at FROM routes ORDER BY path`)
	if err != nil {
		return nil, fmt.Errorf("list sitemap entries: %w", err)
	}
	defer rows.Close()
	entries := make([]domain.SitemapEntry, 0)
	for rows.Next() {
		var entry domain.SitemapEntry
		if err := rows.Scan(&entry.Path, &entry.ModifiedAt); err != nil {
			return nil, fmt.Errorf("scan sitemap entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read sitemap entries: %w", err)
	}
	return entries, nil
}

type summaryScanner interface {
	Scan(...any) error
}

func scanSummary(scanner summaryScanner) (domain.Summary, error) {
	var summary domain.Summary
	if err := scanner.Scan(
		&summary.ID, &summary.Type, &summary.Title, &summary.Slug,
		&summary.Description, &summary.HeroImageURL, &summary.HeroImageAlt,
		&summary.AuthorName, &summary.PublishedAt, &summary.UpdatedAt,
	); err != nil {
		return domain.Summary{}, fmt.Errorf("scan editorial summary: %w", err)
	}
	summary.Path = summary.Type.Path(summary.Slug)
	return summary, nil
}
