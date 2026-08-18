package adminpostgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	admin "rigmark/internal/modules/admin/domain"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

func TestAdminProductAndOfferLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)

	var actor identity.UserID
	email := fmt.Sprintf("admin-repository-%d@example.invalid", time.Now().UnixNano())
	if err := pool.QueryRow(ctx, `INSERT INTO identity.users (email, status) VALUES ($1, 'active') RETURNING id`, email).Scan(&actor); err != nil {
		t.Fatalf("create actor: %v", err)
	}
	var categoryID catalog.CategoryID
	var brandID catalog.BrandID
	var merchantID string
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.categories ORDER BY id LIMIT 1`).Scan(&categoryID); err != nil {
		t.Fatalf("load category: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM catalog.brands ORDER BY id LIMIT 1`).Scan(&brandID); err != nil {
		t.Fatalf("load brand: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM commerce.merchants ORDER BY id LIMIT 1`).Scan(&merchantID); err != nil {
		t.Fatalf("load merchant: %v", err)
	}

	slug := fmt.Sprintf("admin-test-product-%d", time.Now().UnixNano())
	capacity := int64(100000)
	product, err := repository.CreateProduct(ctx, actor, admin.ProductInput{
		CategoryID:       categoryID,
		BrandID:          brandID,
		Name:             "Admin integration product",
		Slug:             slug,
		Description:      "A fictional product created only for an administrative integration test.",
		Price:            catalog.Money{AmountMinor: 12500, Currency: "USD"},
		Dimensions:       catalog.Dimensions{LengthMM: 500, WidthMM: 400, HeightMM: 300},
		WeightGrams:      12000,
		MaxCapacityGrams: &capacity,
		Material:         "Demo steel",
		WarrantyMonths:   12,
		Scores:           catalog.Scores{Quality: 70, Value: 71, Durability: 72, Beginner: 73, Advanced: 74, Apartment: 75, Noise: 76, Portability: 77},
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	var offerID string
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.affiliate_clicks WHERE product_id=$1`, product.ID)
		if offerID != "" {
			_, _ = pool.Exec(context.Background(), `DELETE FROM commerce.merchant_offers WHERE id=$1`, offerID)
		}
		_, _ = pool.Exec(context.Background(), `DELETE FROM admin.media_deletion_jobs WHERE product_id=$1`, product.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM catalog.products WHERE id=$1`, product.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM admin.audit_log WHERE actor_user_id=$1`, actor)
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, actor)
	})
	if product.Status != catalog.ProductStatusDraft || product.MaxCapacityGrams == nil || *product.MaxCapacityGrams != capacity {
		t.Fatalf("created product = %#v", product)
	}
	if err := repository.SetProductStatus(ctx, actor, product.ID, catalog.ProductStatusDiscontinued); err != nil {
		t.Fatalf("SetProductStatus() error = %v", err)
	}
	objectName := string(product.ID) + "_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.webp"
	image, err := repository.AddImage(ctx, actor, product.ID, admin.ImageInput{URL: "/api/media/products/" + objectName, AltText: "Demo product", IsPrimary: true})
	if err != nil {
		t.Fatalf("AddImage() error = %v", err)
	}
	textValue := "knurled"
	if _, err := repository.UpsertAttribute(ctx, actor, product.ID, admin.AttributeInput{Key: "handle_finish", Type: catalog.AttributeTypeText, TextValue: &textValue, IsFilterable: true}); err != nil {
		t.Fatalf("UpsertAttribute() error = %v", err)
	}
	affiliate := admin.AffiliateLinkInput{Provider: "direct", DestinationURL: "https://merchant.invalid/admin-test", DisclosureLabel: "Affiliate link", IsActive: true, CommissionType: "unknown"}
	offer, err := repository.CreateOffer(ctx, actor, admin.OfferInput{MerchantID: merchantID, ProductID: string(product.ID), MerchantSKU: slug, ProductURL: "https://merchant.invalid/admin-test-product", PriceMinor: 12400, Currency: "USD", Availability: "out_of_stock", Condition: "new", IsActive: false, Affiliate: &affiliate})
	if err != nil {
		t.Fatalf("CreateOffer() error = %v", err)
	}
	offerID = offer.ID
	if offer.AffiliateLinks != 1 {
		t.Fatalf("affiliate links = %d, want 1", offer.AffiliateLinks)
	}

	loaded, err := repository.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if loaded.Status != catalog.ProductStatusDiscontinued || len(loaded.Images) != 1 || len(loaded.Attributes) != 1 {
		t.Fatalf("loaded product = %#v", loaded)
	}
	state, err := repository.InspectMediaObject(ctx, product.ID, objectName)
	if err != nil || state.ReferenceCount != 1 || state.MatchingProductCount != 1 {
		t.Fatalf("InspectMediaObject() = (%+v, %v)", state, err)
	}
	references, _, err := repository.ListMediaReferences(ctx, "", 500)
	if err != nil {
		t.Fatal(err)
	}
	var referenceFound bool
	for _, reference := range references {
		referenceFound = referenceFound || reference.ObjectName == objectName
	}
	if !referenceFound {
		t.Fatalf("media reference %q was not listed", objectName)
	}
	runID, err := repository.BeginMediaReconciliation(ctx, "dry_run", 50, "", "", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM admin.media_reconciliation_runs WHERE id=$1`, runID)
	})
	if _, err = repository.BeginMediaReconciliation(ctx, "dry_run", 50, "", "", time.Now().UTC()); err != adminports.ErrMediaReconciliationRunning {
		t.Fatalf("concurrent reconciliation error=%v", err)
	}
	safeName := objectName
	if err = repository.RecordMediaReconciliationResult(ctx, runID, adminports.MediaReconciliationResult{
		Classification: "orphan_object", IdentifierHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SafeObjectName: &safeName, Action: "none", DetailCode: "object.unreferenced",
	}); err != nil {
		t.Fatal(err)
	}
	if err = repository.FinishMediaReconciliation(ctx, adminports.MediaReconciliationRun{ID: runID, Mode: "dry_run", BatchSize: 50, Discrepancies: 1}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	deletionName, err := repository.DeleteImage(ctx, actor, product.ID, image.ID)
	if err != nil || deletionName != "/api/media/products/"+objectName {
		t.Fatalf("DeleteImage() = (%q, %v)", deletionName, err)
	}
	deletions, err := repository.ClaimMediaDeletions(ctx, 10, time.Now().UTC())
	if err != nil || len(deletions) != 1 || deletions[0].ObjectName != objectName {
		t.Fatalf("ClaimMediaDeletions() = (%#v, %v)", deletions, err)
	}
	if err := repository.CompleteMediaDeletion(ctx, objectName, time.Now().UTC()); err != nil {
		t.Fatalf("CompleteMediaDeletion() = %v", err)
	}
	var auditEntries int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM admin.audit_log WHERE actor_user_id=$1`, actor).Scan(&auditEntries); err != nil || auditEntries < 4 {
		t.Fatalf("audit entries = %d, error = %v", auditEntries, err)
	}
}
