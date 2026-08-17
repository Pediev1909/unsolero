package adminpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	admin "rigmark/internal/modules/admin/domain"
	"rigmark/internal/modules/admin/ports"
	identity "rigmark/internal/modules/identity/domain"
)

func (repository *Repository) References(ctx context.Context) (admin.References, error) {
	categories, err := repository.ListCategories(ctx)
	if err != nil {
		return admin.References{}, err
	}
	brands, err := repository.ListBrands(ctx)
	if err != nil {
		return admin.References{}, err
	}
	merchants, err := repository.ListMerchants(ctx)
	if err != nil {
		return admin.References{}, err
	}
	rows, err := repository.pool.Query(ctx, `SELECT id, name, slug FROM catalog.products ORDER BY name`)
	if err != nil {
		return admin.References{}, fmt.Errorf("list product references: %w", err)
	}
	defer rows.Close()
	products := make([]admin.ProductReference, 0)
	for rows.Next() {
		var product admin.ProductReference
		if err := rows.Scan(&product.ID, &product.Name, &product.Slug); err != nil {
			return admin.References{}, fmt.Errorf("scan product reference: %w", err)
		}
		products = append(products, product)
	}
	return admin.References{Categories: categories, Brands: brands, Merchants: merchants, Products: products}, rows.Err()
}

func (repository *Repository) ListCategories(ctx context.Context) ([]admin.Category, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT categories.id, categories.name, categories.slug, categories.is_active,
			count(products.id), categories.updated_at
		FROM catalog.categories AS categories
		LEFT JOIN catalog.products AS products ON products.category_id=categories.id
		GROUP BY categories.id ORDER BY categories.sort_order, categories.name`)
	if err != nil {
		return nil, fmt.Errorf("list admin categories: %w", err)
	}
	defer rows.Close()
	items := make([]admin.Category, 0)
	for rows.Next() {
		var item admin.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.IsActive, &item.Products, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin category: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) ListBrands(ctx context.Context) ([]admin.Brand, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT brands.id, brands.name, brands.slug, brands.is_active,
			count(products.id), brands.updated_at
		FROM catalog.brands AS brands
		LEFT JOIN catalog.products AS products ON products.brand_id=brands.id
		GROUP BY brands.id ORDER BY brands.name`)
	if err != nil {
		return nil, fmt.Errorf("list admin brands: %w", err)
	}
	defer rows.Close()
	items := make([]admin.Brand, 0)
	for rows.Next() {
		var item admin.Brand
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.IsActive, &item.Products, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin brand: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) ListMerchants(ctx context.Context) ([]admin.Merchant, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT merchants.id, merchants.name, merchants.slug, merchants.website_url,
			merchants.country_code, merchants.trust_score, merchants.status,
			count(offers.id), merchants.updated_at
		FROM commerce.merchants AS merchants
		LEFT JOIN commerce.merchant_offers AS offers ON offers.merchant_id=merchants.id
		GROUP BY merchants.id ORDER BY merchants.name`)
	if err != nil {
		return nil, fmt.Errorf("list admin merchants: %w", err)
	}
	defer rows.Close()
	items := make([]admin.Merchant, 0)
	for rows.Next() {
		var item admin.Merchant
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.WebsiteURL, &item.CountryCode, &item.TrustScore, &item.Status, &item.Offers, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin merchant: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) ListOffers(ctx context.Context, limit, offset int) (admin.Page[admin.Offer], error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), offers.id, offers.merchant_id, merchants.name,
			offers.product_id, products.name, offers.merchant_sku, offers.product_url,
			offers.price_minor, offers.shipping_minor, offers.currency, offers.availability,
			offers.condition, offers.is_active, offers.last_checked_at,
			(SELECT count(*) FROM commerce.affiliate_links WHERE merchant_offer_id=offers.id),
			offers.updated_at
		FROM commerce.merchant_offers AS offers
		JOIN commerce.merchants AS merchants ON merchants.id=offers.merchant_id
		JOIN catalog.products AS products ON products.id=offers.product_id
		ORDER BY offers.updated_at DESC, products.name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return admin.Page[admin.Offer]{}, fmt.Errorf("list admin offers: %w", err)
	}
	defer rows.Close()
	page := admin.Page[admin.Offer]{Items: make([]admin.Offer, 0)}
	for rows.Next() {
		var item admin.Offer
		if err := rows.Scan(&page.Total, &item.ID, &item.MerchantID, &item.MerchantName,
			&item.ProductID, &item.ProductName, &item.MerchantSKU, &item.ProductURL,
			&item.PriceMinor, &item.ShippingMinor, &item.Currency, &item.Availability,
			&item.Condition, &item.IsActive, &item.LastCheckedAt, &item.AffiliateLinks,
			&item.UpdatedAt); err != nil {
			return admin.Page[admin.Offer]{}, fmt.Errorf("scan admin offer: %w", err)
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (repository *Repository) CreateOffer(ctx context.Context, actor identity.UserID, input admin.OfferInput) (admin.Offer, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return admin.Offer{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO commerce.merchant_offers (
			merchant_id, product_id, merchant_sku, product_url, price_minor,
			shipping_minor, currency, availability, condition, last_checked_at, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now(),$10) RETURNING id`,
		input.MerchantID, input.ProductID, input.MerchantSKU, input.ProductURL,
		input.PriceMinor, input.ShippingMinor, input.Currency, input.Availability,
		input.Condition, input.IsActive).Scan(&id)
	if err == nil && input.Affiliate != nil {
		err = insertAffiliate(ctx, tx, id, *input.Affiliate)
	}
	if err == nil {
		err = audit(ctx, tx, actor, "offer.create", "merchant_offer", id, map[string]string{"merchant_sku": input.MerchantSKU})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return admin.Offer{}, fmt.Errorf("create offer: %w", err)
	}
	return repository.getOffer(ctx, id)
}

func (repository *Repository) UpdateOffer(ctx context.Context, actor identity.UserID, id string, input admin.OfferInput) (admin.Offer, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return admin.Offer{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE commerce.merchant_offers SET merchant_id=$2, product_id=$3,
			merchant_sku=$4, product_url=$5, price_minor=$6, shipping_minor=$7,
			currency=$8, availability=$9, condition=$10, is_active=$11,
			last_checked_at=now(), updated_at=now() WHERE id=$1`, id, input.MerchantID,
		input.ProductID, input.MerchantSKU, input.ProductURL, input.PriceMinor,
		input.ShippingMinor, input.Currency, input.Availability, input.Condition, input.IsActive)
	if err == nil && tag.RowsAffected() == 0 {
		err = ports.ErrNotFound
	}
	if err == nil && input.Affiliate != nil {
		err = upsertAffiliate(ctx, tx, id, *input.Affiliate)
	}
	if err == nil {
		err = audit(ctx, tx, actor, "offer.update", "merchant_offer", id, map[string]string{"merchant_sku": input.MerchantSKU})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return admin.Offer{}, fmt.Errorf("update offer: %w", err)
	}
	return repository.getOffer(ctx, id)
}

func (repository *Repository) getOffer(ctx context.Context, id string) (admin.Offer, error) {
	var item admin.Offer
	err := repository.pool.QueryRow(ctx, `
		SELECT offers.id, offers.merchant_id, merchants.name, offers.product_id,
			products.name, offers.merchant_sku, offers.product_url, offers.price_minor,
			offers.shipping_minor, offers.currency, offers.availability, offers.condition,
			offers.is_active, offers.last_checked_at,
			(SELECT count(*) FROM commerce.affiliate_links WHERE merchant_offer_id=offers.id),
			offers.updated_at
		FROM commerce.merchant_offers AS offers
		JOIN commerce.merchants AS merchants ON merchants.id=offers.merchant_id
		JOIN catalog.products AS products ON products.id=offers.product_id
		WHERE offers.id=$1`, id).Scan(&item.ID, &item.MerchantID, &item.MerchantName,
		&item.ProductID, &item.ProductName, &item.MerchantSKU, &item.ProductURL,
		&item.PriceMinor, &item.ShippingMinor, &item.Currency, &item.Availability,
		&item.Condition, &item.IsActive, &item.LastCheckedAt, &item.AffiliateLinks,
		&item.UpdatedAt)
	if err = notFound(err); err != nil {
		return admin.Offer{}, err
	}
	return item, nil
}

func insertAffiliate(ctx context.Context, tx pgx.Tx, offerID string, input admin.AffiliateLinkInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO commerce.affiliate_links (
			merchant_offer_id, provider, destination_url, external_reference,
			disclosure_label, is_active, priority, program_id, commission_type,
			commission_rate_bps, commission_amount_minor, commission_currency
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, offerID,
		input.Provider, input.DestinationURL, input.ExternalReference, input.DisclosureLabel,
		input.IsActive, input.Priority, input.ProgramID, input.CommissionType,
		input.CommissionRateBPS, input.CommissionAmount, input.CommissionCurrency)
	return err
}

func upsertAffiliate(ctx context.Context, tx pgx.Tx, offerID string, input admin.AffiliateLinkInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO commerce.affiliate_links (
			merchant_offer_id, provider, destination_url, external_reference,
			disclosure_label, is_active, priority, program_id, commission_type,
			commission_rate_bps, commission_amount_minor, commission_currency
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
			destination_url=EXCLUDED.destination_url,
			external_reference=EXCLUDED.external_reference,
			disclosure_label=EXCLUDED.disclosure_label, is_active=EXCLUDED.is_active,
			priority=EXCLUDED.priority, program_id=EXCLUDED.program_id,
			commission_type=EXCLUDED.commission_type,
			commission_rate_bps=EXCLUDED.commission_rate_bps,
			commission_amount_minor=EXCLUDED.commission_amount_minor,
			commission_currency=EXCLUDED.commission_currency, updated_at=now()`, offerID,
		input.Provider, input.DestinationURL, input.ExternalReference, input.DisclosureLabel,
		input.IsActive, input.Priority, input.ProgramID, input.CommissionType,
		input.CommissionRateBPS, input.CommissionAmount, input.CommissionCurrency)
	return err
}

func (repository *Repository) ListAffiliateLinks(ctx context.Context, limit, offset int) (admin.Page[admin.AffiliateLink], error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT count(*) OVER(), links.id, links.merchant_offer_id, products.name,
			merchants.name, links.provider, links.destination_url, links.external_reference,
			links.disclosure_label, links.is_active, links.priority, links.program_id,
			links.commission_type, links.commission_rate_bps,
			links.commission_amount_minor, links.commission_currency, links.updated_at
		FROM commerce.affiliate_links AS links
		JOIN commerce.merchant_offers AS offers ON offers.id=links.merchant_offer_id
		JOIN catalog.products AS products ON products.id=offers.product_id
		JOIN commerce.merchants AS merchants ON merchants.id=offers.merchant_id
		ORDER BY links.updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return admin.Page[admin.AffiliateLink]{}, fmt.Errorf("list affiliate links: %w", err)
	}
	defer rows.Close()
	page := admin.Page[admin.AffiliateLink]{Items: make([]admin.AffiliateLink, 0)}
	for rows.Next() {
		var item admin.AffiliateLink
		if err := scanAffiliate(rows, &page.Total, &item); err != nil {
			return admin.Page[admin.AffiliateLink]{}, err
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (repository *Repository) UpdateAffiliateLink(ctx context.Context, actor identity.UserID, id string, input admin.AffiliateLinkInput) (admin.AffiliateLink, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return admin.AffiliateLink{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE commerce.affiliate_links SET provider=$2, destination_url=$3,
			external_reference=$4, disclosure_label=$5, is_active=$6, priority=$7,
			program_id=$8, commission_type=$9, commission_rate_bps=$10,
			commission_amount_minor=$11, commission_currency=$12, updated_at=now()
		WHERE id=$1`, id, input.Provider, input.DestinationURL, input.ExternalReference,
		input.DisclosureLabel, input.IsActive, input.Priority, input.ProgramID,
		input.CommissionType, input.CommissionRateBPS, input.CommissionAmount,
		input.CommissionCurrency)
	if err == nil && tag.RowsAffected() == 0 {
		err = ports.ErrNotFound
	}
	if err == nil {
		err = audit(ctx, tx, actor, "affiliate_link.update", "affiliate_link", id, map[string]string{"provider": input.Provider})
	}
	if err := finishMutation(ctx, tx, err); err != nil {
		return admin.AffiliateLink{}, err
	}
	return repository.getAffiliate(ctx, id)
}

func (repository *Repository) getAffiliate(ctx context.Context, id string) (admin.AffiliateLink, error) {
	var item admin.AffiliateLink
	err := scanAffiliate(repository.pool.QueryRow(ctx, `
		SELECT links.id, links.merchant_offer_id, products.name, merchants.name,
			links.provider, links.destination_url, links.external_reference,
			links.disclosure_label, links.is_active, links.priority, links.program_id,
			links.commission_type, links.commission_rate_bps,
			links.commission_amount_minor, links.commission_currency, links.updated_at
		FROM commerce.affiliate_links AS links
		JOIN commerce.merchant_offers AS offers ON offers.id=links.merchant_offer_id
		JOIN catalog.products AS products ON products.id=offers.product_id
		JOIN commerce.merchants AS merchants ON merchants.id=offers.merchant_id
		WHERE links.id=$1`, id), nil, &item)
	if err = notFound(err); err != nil {
		return admin.AffiliateLink{}, err
	}
	return item, nil
}

func scanAffiliate(row scanner, total *int64, item *admin.AffiliateLink) error {
	var external, program, currency sql.NullString
	var rate sql.NullInt32
	var amount sql.NullInt64
	targets := []interface{}{&item.ID, &item.OfferID, &item.ProductName, &item.MerchantName,
		&item.Provider, &item.DestinationURL, &external, &item.DisclosureLabel,
		&item.IsActive, &item.Priority, &program, &item.CommissionType, &rate,
		&amount, &currency, &item.UpdatedAt}
	if total != nil {
		targets = append([]interface{}{total}, targets...)
	}
	if err := row.Scan(targets...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrNotFound
		}
		return fmt.Errorf("scan affiliate link: %w", err)
	}
	if external.Valid {
		value := external.String
		item.ExternalReference = &value
	}
	if program.Valid {
		value := program.String
		item.ProgramID = &value
	}
	if rate.Valid {
		value := int(rate.Int32)
		item.CommissionRateBPS = &value
	}
	if amount.Valid {
		value := amount.Int64
		item.CommissionAmount = &value
	}
	if currency.Valid {
		value := currency.String
		item.CommissionCurrency = &value
	}
	return nil
}
