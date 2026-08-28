package commercepostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

type Repository struct {
	pool        *pgxpool.Pool
	maxOfferAge time.Duration
}

func New(pool *pgxpool.Pool, maximumAge ...time.Duration) *Repository {
	maxOfferAge := 72 * time.Hour
	if len(maximumAge) > 0 && maximumAge[0] > 0 {
		maxOfferAge = maximumAge[0]
	}
	return &Repository{pool: pool, maxOfferAge: maxOfferAge}
}

func (repository *Repository) ListActive(ctx context.Context) ([]domain.Merchant, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, name, slug, website_url, country_code, trust_score, status
		FROM commerce.merchants
		WHERE status = 'active'
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list active merchants: %w", err)
	}
	defer rows.Close()

	merchants := make([]domain.Merchant, 0)
	for rows.Next() {
		var merchant domain.Merchant
		if err := rows.Scan(
			&merchant.ID,
			&merchant.Name,
			&merchant.Slug,
			&merchant.WebsiteURL,
			&merchant.CountryCode,
			&merchant.TrustScore,
			&merchant.Status,
		); err != nil {
			return nil, fmt.Errorf("scan merchant: %w", err)
		}
		merchants = append(merchants, merchant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read merchants: %w", err)
	}
	return merchants, nil
}

func (repository *Repository) ListAvailableByProduct(
	ctx context.Context,
	productID catalog.ProductID,
	currency string,
) ([]domain.Offer, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT
			offers.id,
			offers.product_id,
			offers.merchant_sku,
			offers.product_url,
			offers.price_minor,
			offers.shipping_minor,
			offers.currency,
			offers.availability,
			offers.condition,
			offers.last_checked_at,
			offers.provider_observed_at,
			offers.imported_at,
			offers.expires_at,
			offers.is_active,
			merchants.id,
			merchants.name,
			merchants.slug,
			merchants.website_url,
			merchants.country_code,
			merchants.trust_score,
			merchants.status,
			affiliate_links.id,
			affiliate_links.provider,
			affiliate_links.destination_url,
			affiliate_links.external_reference,
			affiliate_links.disclosure_label,
			affiliate_links.is_active,
			affiliate_links.priority,
			affiliate_links.program_id,
			affiliate_links.commission_type,
			affiliate_links.commission_rate_bps,
			affiliate_links.commission_amount_minor,
			affiliate_links.commission_currency
		FROM commerce.merchant_offers AS offers
		JOIN commerce.merchants AS merchants ON merchants.id = offers.merchant_id
		LEFT JOIN commerce.affiliate_links AS affiliate_links
			ON affiliate_links.merchant_offer_id = offers.id
			AND affiliate_links.is_active = true
			AND (affiliate_links.valid_from IS NULL OR affiliate_links.valid_from <= now())
			AND (affiliate_links.valid_until IS NULL OR affiliate_links.valid_until > now())
		WHERE offers.product_id = $1
			AND offers.currency = $2
			AND offers.is_active = true
			AND offers.availability IN ('in_stock', 'backorder')
			AND offers.last_checked_at >= now() - make_interval(secs => $3::double precision)
			AND (offers.expires_at IS NULL OR offers.expires_at > now())
			AND merchants.status = 'active'
		ORDER BY offers.price_minor + offers.shipping_minor, merchants.trust_score DESC,
			affiliate_links.priority DESC, affiliate_links.provider`,
		productID,
		currency,
		int64(repository.maxOfferAge.Seconds()),
	)
	if err != nil {
		return nil, fmt.Errorf("list available product offers: %w", err)
	}
	defer rows.Close()

	offers := make([]domain.Offer, 0)
	offerIndexes := make(map[string]int)
	for rows.Next() {
		var offer domain.Offer
		var providerObservedAt sql.NullTime
		var importedAt sql.NullTime
		var expiresAt sql.NullTime
		var affiliateID sql.NullString
		var provider sql.NullString
		var destinationURL sql.NullString
		var externalReference sql.NullString
		var disclosureLabel sql.NullString
		var affiliateActive sql.NullBool
		var priority sql.NullInt16
		var programID sql.NullString
		var commissionType sql.NullString
		var commissionRate sql.NullInt32
		var commissionAmount sql.NullInt64
		var commissionCurrency sql.NullString
		if err := rows.Scan(
			&offer.ID,
			&offer.ProductID,
			&offer.MerchantSKU,
			&offer.ProductURL,
			&offer.Price.AmountMinor,
			&offer.ShippingMinor,
			&offer.Price.Currency,
			&offer.Availability,
			&offer.Condition,
			&offer.LastCheckedAt,
			&providerObservedAt,
			&importedAt,
			&expiresAt,
			&offer.IsActive,
			&offer.Merchant.ID,
			&offer.Merchant.Name,
			&offer.Merchant.Slug,
			&offer.Merchant.WebsiteURL,
			&offer.Merchant.CountryCode,
			&offer.Merchant.TrustScore,
			&offer.Merchant.Status,
			&affiliateID,
			&provider,
			&destinationURL,
			&externalReference,
			&disclosureLabel,
			&affiliateActive,
			&priority,
			&programID,
			&commissionType,
			&commissionRate,
			&commissionAmount,
			&commissionCurrency,
		); err != nil {
			return nil, fmt.Errorf("scan product offer: %w", err)
		}
		if providerObservedAt.Valid {
			offer.ProviderObservedAt = &providerObservedAt.Time
		}
		if importedAt.Valid {
			offer.ImportedAt = &importedAt.Time
		}
		if expiresAt.Valid {
			offer.ExpiresAt = &expiresAt.Time
		}

		index, exists := offerIndexes[string(offer.ID)]
		if !exists {
			if err := offer.Validate(); err != nil {
				return nil, fmt.Errorf("validate persisted offer %q: %w", offer.ID, err)
			}
			offers = append(offers, offer)
			index = len(offers) - 1
			offerIndexes[string(offer.ID)] = index
		}

		if affiliateID.Valid {
			link := domain.AffiliateLink{
				ID:              domain.AffiliateLinkID(affiliateID.String),
				Provider:        provider.String,
				DestinationURL:  destinationURL.String,
				DisclosureLabel: disclosureLabel.String,
				IsActive:        affiliateActive.Bool,
				Priority:        priority.Int16,
				Commission:      domain.CommissionMetadata{Type: commissionType.String},
			}
			if externalReference.Valid {
				link.ExternalReference = &externalReference.String
			}
			if programID.Valid {
				link.ProgramID = &programID.String
			}
			if commissionRate.Valid {
				value := int(commissionRate.Int32)
				link.Commission.RateBPS = &value
			}
			if commissionAmount.Valid {
				link.Commission.AmountMinor = &commissionAmount.Int64
			}
			if commissionCurrency.Valid {
				link.Commission.Currency = &commissionCurrency.String
			}
			offers[index].AffiliateLinks = append(offers[index].AffiliateLinks, link)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read product offers: %w", err)
	}

	return offers, nil
}

func (repository *Repository) ResolveOfferDestination(ctx context.Context, click domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error) {
	return repository.resolveDestination(ctx, string(click.OfferID), click, true)
}

func (repository *Repository) ResolveLegacyDestination(ctx context.Context, click domain.AffiliateClick) (domain.ResolvedAffiliateDestination, error) {
	return repository.resolveDestination(ctx, string(click.LinkID), click, false)
}

func (repository *Repository) ResolvePromotionDestination(ctx context.Context, click domain.AffiliateClick) (domain.ResolvedPromotionDestination, error) {
	var destination domain.ResolvedPromotionDestination
	err := repository.pool.QueryRow(ctx, `
		SELECT promotions.id, promotions.slug, promotions.destination_url
		FROM commerce.affiliate_promotions promotions
		JOIN commerce.merchants merchants ON merchants.id=promotions.merchant_id
		WHERE promotions.slug=$1 AND promotions.is_active=true
			AND (promotions.valid_from IS NULL OR promotions.valid_from<=now())
			AND (promotions.valid_until IS NULL OR promotions.valid_until>now())
			AND promotions.last_checked_at >= now() - make_interval(secs => $2::double precision)
			AND merchants.status='active'
		LIMIT 1`, click.PromotionSlug, int64(repository.maxOfferAge.Seconds())).Scan(
		&destination.PromotionID, &destination.PromotionSlug, &destination.DestinationURL)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ResolvedPromotionDestination{}, ports.ErrAffiliateDestinationNotFound
	}
	if err != nil {
		return domain.ResolvedPromotionDestination{}, fmt.Errorf("resolve affiliate promotion: %w", err)
	}
	return destination, nil
}

func (repository *Repository) resolveDestination(
	ctx context.Context,
	targetID string,
	click domain.AffiliateClick,
	byOffer bool,
) (domain.ResolvedAffiliateDestination, error) {
	targetPredicate := "offers.id = $1"
	if !byOffer {
		targetPredicate = "affiliate_links.id = $1"
	}
	query := fmt.Sprintf(`
		SELECT affiliate_links.id, offers.id, offers.product_id,
			CASE WHEN $3::uuid IS NULL THEN NULL ELSE items.id END,
			affiliate_links.destination_url
		FROM commerce.affiliate_links affiliate_links
		JOIN commerce.merchant_offers offers ON offers.id=affiliate_links.merchant_offer_id
		JOIN commerce.merchants merchants ON merchants.id=offers.merchant_id
		LEFT JOIN recommendation.recommendation_items items
			ON items.recommendation_id=$3::uuid AND items.product_id=offers.product_id
		LEFT JOIN recommendation.recommendations recommendations ON recommendations.id=$3::uuid
		LEFT JOIN recommendation.recommendation_sessions sessions
			ON sessions.id=recommendations.session_id AND sessions.user_id=$2::uuid
		WHERE %s AND affiliate_links.is_active=true
			AND (affiliate_links.valid_from IS NULL OR affiliate_links.valid_from<=now())
			AND (affiliate_links.valid_until IS NULL OR affiliate_links.valid_until>now())
			AND offers.is_active=true AND offers.availability IN ('in_stock','backorder')
			AND offers.last_checked_at >= now() - make_interval(secs => $4::double precision)
			AND (offers.expires_at IS NULL OR offers.expires_at > now())
			AND merchants.status='active'
			AND ($3::uuid IS NULL OR (sessions.user_id IS NOT NULL AND items.id IS NOT NULL))
		ORDER BY affiliate_links.priority DESC, affiliate_links.provider
		LIMIT 1`, targetPredicate)

	var userID any
	if click.UserID != nil {
		userID = *click.UserID
	}
	var destination domain.ResolvedAffiliateDestination
	err := repository.pool.QueryRow(ctx, query, targetID, userID, click.RecommendationID,
		int64(repository.maxOfferAge.Seconds())).Scan(&destination.AffiliateLinkID,
		&destination.OfferID, &destination.ProductID, &destination.RecommendationItem,
		&destination.DestinationURL)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ResolvedAffiliateDestination{}, ports.ErrAffiliateDestinationNotFound
	}
	if err != nil {
		return domain.ResolvedAffiliateDestination{}, fmt.Errorf("resolve affiliate destination: %w", err)
	}
	return destination, nil
}

func (repository *Repository) RecordClick(ctx context.Context, destination domain.ResolvedAffiliateDestination, click domain.AffiliateClick) error {
	retentionExpires := click.RetentionExpires
	if retentionExpires.IsZero() {
		retentionExpires = time.Now().UTC().Add(397 * 24 * time.Hour)
	}
	var userID any
	if click.UserID != nil {
		userID = *click.UserID
	}
	_, err := repository.pool.Exec(ctx, `WITH recorded AS (
		INSERT INTO commerce.affiliate_clicks (
			affiliate_link_id, merchant_offer_id, product_id, user_id, anonymous_id,
			session_id, source, campaign, referrer, request_id, traffic_source,
			traffic_medium, referrer_host, recommendation_id, recommendation_item_id,
			classification, is_countable, user_agent_hash, idempotency_key,
			retention_expires_at)
		VALUES ($1,$2,$3,$4::uuid,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::uuid,
			$15::uuid,$16,$17,$18,$19,$20)
		ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
		RETURNING id, merchant_offer_id, product_id
	), analytics_event AS (
		INSERT INTO analytics.events (
			event_name, schema_version, user_id, anonymous_id, session_id, request_id,
			surface, properties, traffic_source, traffic_medium, campaign, referrer_host,
			consent_state, classification, is_reportable, retention_expires_at, occurred_at)
		SELECT 'affiliate_clicked', 2, $4::uuid, $5, $6, $10, $7,
			jsonb_strip_nulls(jsonb_build_object('offer_id', recorded.merchant_offer_id,
			'product_id', recorded.product_id, 'source', $7, 'campaign', $8)),
			$11,$12,$8,$13,'essential',$16,$17,$20,now() FROM recorded WHERE $17
	)
	SELECT count(*) FROM recorded`, destination.AffiliateLinkID, destination.OfferID,
		destination.ProductID, userID, click.AnonymousID, click.SessionID, click.Source,
		click.Campaign, click.Referrer, click.RequestID, click.TrafficSource,
		click.TrafficMedium, click.ReferrerHost, click.RecommendationID,
		destination.RecommendationItem, click.Classification, click.IsCountable,
		click.UserAgentHash, click.IdempotencyKey, retentionExpires)
	if err != nil {
		return fmt.Errorf("record affiliate click: %w", err)
	}
	return nil
}

func (repository *Repository) RecordPromotionClick(ctx context.Context, destination domain.ResolvedPromotionDestination, click domain.AffiliateClick) error {
	retentionExpires := click.RetentionExpires
	if retentionExpires.IsZero() {
		retentionExpires = time.Now().UTC().Add(397 * 24 * time.Hour)
	}
	var userID any
	if click.UserID != nil {
		userID = *click.UserID
	}
	_, err := repository.pool.Exec(ctx, `
		INSERT INTO commerce.affiliate_promotion_clicks (
			promotion_id, user_id, anonymous_id, session_id, source, campaign,
			referrer, request_id, traffic_source, traffic_medium, referrer_host,
			classification, is_countable, user_agent_hash, idempotency_key,
			retention_expires_at)
		VALUES ($1,$2::uuid,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING`,
		destination.PromotionID, userID, click.AnonymousID, click.SessionID, click.Source,
		click.Campaign, click.Referrer, click.RequestID, click.TrafficSource,
		click.TrafficMedium, click.ReferrerHost, click.Classification, click.IsCountable,
		click.UserAgentHash, click.IdempotencyKey, retentionExpires)
	if err != nil {
		return fmt.Errorf("record affiliate promotion click: %w", err)
	}
	return nil
}

var (
	_ ports.MerchantRepository          = (*Repository)(nil)
	_ ports.OfferRepository             = (*Repository)(nil)
	_ ports.AffiliateRedirectRepository = (*Repository)(nil)
)
