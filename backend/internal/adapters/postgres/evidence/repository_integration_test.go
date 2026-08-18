package evidencepostgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalogpostgres "rigmark/internal/adapters/postgres/catalog"
	catalog "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	evidence "rigmark/internal/modules/evidence/domain"
	"rigmark/internal/modules/evidence/ports"
	identity "rigmark/internal/modules/identity/domain"
)

func TestGovernedPublicationLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	editor := insertEvidenceUser(t, ctx, pool, "editor")
	reviewer := insertEvidenceUser(t, ctx, pool, "reviewer")
	publisher := insertEvidenceUser(t, ctx, pool, "publisher")
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=ANY($1::uuid[])`,
			[]string{string(editor), string(reviewer), string(publisher)})
	})

	var categoryID catalog.CategoryID
	var brandID catalog.BrandID
	if err = pool.QueryRow(ctx, `SELECT id FROM catalog.categories WHERE is_physical ORDER BY id LIMIT 1`).Scan(&categoryID); err != nil {
		t.Fatalf("load category: %v", err)
	}
	if err = pool.QueryRow(ctx, `SELECT id FROM catalog.brands ORDER BY id LIMIT 1`).Scan(&brandID); err != nil {
		t.Fatalf("load brand: %v", err)
	}
	slug := fmt.Sprintf("evidence-integration-%d", time.Now().UnixNano())
	var productID catalog.ProductID
	err = pool.QueryRow(ctx, `INSERT INTO catalog.products (
		category_id, brand_id, name, slug, description, price_minor, currency,
		length_mm, width_mm, height_mm, weight_grams, material, warranty_months,
		quality_score, value_score, durability_score, beginner_score, advanced_score,
		apartment_score, noise_score, portability_score
	) VALUES ($1,$2,'Evidence integration product',$3,'Integration-only governed product.',
		10000,'USD',500,400,300,10000,'Steel',12,80,80,80,80,80,80,80,80)
	RETURNING id`, categoryID, brandID, slug).Scan(&productID)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM evidence.score_rationales
			WHERE score_revision_id IN (SELECT id FROM evidence.score_revisions WHERE product_id=$1)`, productID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM evidence.fact_provenance
			WHERE fact_revision_id IN (SELECT id FROM evidence.product_fact_revisions WHERE product_id=$1)`, productID)
		_, _ = pool.Exec(context.Background(), `UPDATE catalog.products SET
			published_fact_revision_id=NULL, published_score_revision_id=NULL WHERE id=$1`, productID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM catalog.products WHERE id=$1`, productID)
	})

	repository := New(pool)
	sourceURL := "https://evidence-integration.example.invalid/specification"
	source, err := repository.CreateSource(ctx, editor, evidence.SourceInput{
		Type: evidence.SourceIndependent, Title: "Integration test evidence",
		Publisher: "UNSOLERO test suite", URL: &sourceURL,
	})
	if err != nil {
		t.Fatalf("CreateSource(): %v", err)
	}
	if _, err = repository.ReviewSource(ctx, reviewer, source.ID, evidence.ReviewVerified, "Independent test source verified"); err != nil {
		t.Fatalf("ReviewSource(): %v", err)
	}
	expires := time.Now().Add(24 * time.Hour)
	observation, err := repository.CreateObservation(ctx, editor, evidence.ObservationInput{
		SourceID: source.ID, ProductID: productID, ObservedAt: time.Now().Add(-time.Hour),
		ExpiresAt: &expires, Confidence: 90, Notes: "Integration observation",
	})
	if err != nil {
		t.Fatalf("CreateObservation(): %v", err)
	}

	product := catalog.Product{ID: productID, CategoryID: categoryID, BrandID: brandID,
		Name: "Published evidence product", Slug: slug,
		Description: "Published through the evidence workflow.",
		Price:       catalog.Money{AmountMinor: 10000, Currency: "USD"},
		Dimensions:  catalog.Dimensions{LengthMM: 500, WidthMM: 400, HeightMM: 300},
		WeightGrams: 10000, Material: "Steel", WarrantyMonths: 12,
		Scores: catalog.Scores{Quality: 81, Value: 82, Durability: 83, Beginner: 84,
			Advanced: 85, Apartment: 86, Noise: 87, Portability: 88},
		Status: catalog.ProductStatusDraft}
	factLinks := make([]evidence.FactLink, 0, 10)
	for _, key := range []string{"category", "brand", "name", "slug", "description", "price", "dimensions", "weight", "material", "warranty"} {
		factLinks = append(factLinks, evidence.FactLink{FactKey: key,
			ObservationID: observation.ID, Classification: evidence.ClassificationVerified})
	}
	rationales := make([]evidence.ScoreRationale, 0, 8)
	for _, key := range []string{"quality", "value", "durability", "beginner", "advanced", "apartment", "noise", "portability"} {
		rationales = append(rationales, evidence.ScoreRationale{ScoreKey: key,
			Rationale: "Evidence-backed integration score rationale.", ObservationID: observation.ID})
	}
	misclassifiedLinks := append([]evidence.FactLink(nil), factLinks...)
	for index := range misclassifiedLinks {
		misclassifiedLinks[index].Classification = evidence.ClassificationManufacturer
	}
	misclassifiedRevision, err := repository.CreateRevision(ctx, editor, evidence.RevisionInput{
		Product: product, FactLinks: misclassifiedLinks, Scores: product.Scores, Rationales: rationales,
	})
	if err != nil {
		t.Fatalf("CreateRevision(misclassified): %v", err)
	}
	if _, err = repository.TransitionRevision(ctx, editor, misclassifiedRevision.FactRevisionID, evidence.WorkflowInReview, ""); err != nil {
		t.Fatalf("submit misclassified revision: %v", err)
	}
	if _, err = repository.TransitionRevision(ctx, reviewer, misclassifiedRevision.FactRevisionID, evidence.WorkflowApproved, "Classification review fixture"); err != nil {
		t.Fatalf("approve misclassified revision: %v", err)
	}
	if _, err = repository.PublishRevision(ctx, publisher, misclassifiedRevision.FactRevisionID); !errors.Is(err, ports.ErrIncompleteProvenance) {
		t.Fatalf("misclassified publication error = %v", err)
	}
	revision, err := repository.CreateRevision(ctx, editor, evidence.RevisionInput{
		Product: product, FactLinks: factLinks, Scores: product.Scores, Rationales: rationales,
	})
	if err != nil {
		t.Fatalf("CreateRevision(): %v", err)
	}
	if _, err = repository.TransitionRevision(ctx, editor, revision.FactRevisionID, evidence.WorkflowInReview, ""); err != nil {
		t.Fatalf("submit revision: %v", err)
	}
	if _, err = repository.TransitionRevision(ctx, editor, revision.FactRevisionID, evidence.WorkflowApproved, "self approval"); !errors.Is(err, ports.ErrSeparationOfDuties) {
		t.Fatalf("self approval error = %v", err)
	}
	if _, err = repository.TransitionRevision(ctx, reviewer, revision.FactRevisionID, evidence.WorkflowApproved, "Reviewed against source"); err != nil {
		t.Fatalf("approve revision: %v", err)
	}
	if _, err = repository.PublishRevision(ctx, publisher, revision.FactRevisionID); err != nil {
		t.Fatalf("PublishRevision(): %v", err)
	}

	var status, factID, scoreID, name string
	if err = pool.QueryRow(ctx, `SELECT status, published_fact_revision_id,
		published_score_revision_id, name FROM catalog.products WHERE id=$1`, productID).Scan(
		&status, &factID, &scoreID, &name); err != nil {
		t.Fatalf("load published projection: %v", err)
	}
	if status != "published" || factID != revision.FactRevisionID ||
		scoreID != revision.ScoreRevisionID || name != product.Name {
		t.Fatalf("published projection = status=%s fact=%s score=%s name=%s", status, factID, scoreID, name)
	}
	governance, err := repository.GetProductGovernance(ctx, productID)
	if err != nil || len(governance.Revisions) != 2 || len(governance.Provenance) != 36 || len(governance.Audit) < 7 {
		t.Fatalf("GetProductGovernance() = %#v, %v", governance, err)
	}
	if _, err = repository.ReviewSource(ctx, reviewer, source.ID, evidence.ReviewRejected, "Source withdrawn"); err != nil {
		t.Fatalf("reject published source: %v", err)
	}
	if _, err = catalogpostgres.New(pool).GetPublishedBySlug(ctx, slug); !errors.Is(err, catalogports.ErrNotFound) {
		t.Fatalf("revoked source product lookup error = %v, want not found", err)
	}
}

func insertEvidenceUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) identity.UserID {
	t.Helper()
	var id identity.UserID
	email := fmt.Sprintf("evidence-%s-%d@example.invalid", label, time.Now().UnixNano())
	if err := pool.QueryRow(ctx, `INSERT INTO identity.users (email, status) VALUES ($1,'active') RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("insert %s: %v", label, err)
	}
	return id
}
