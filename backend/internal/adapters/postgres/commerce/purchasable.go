package commercepostgres

import "time"

// DefaultOfferMaximumAge is the freshness window applied when a deployment
// does not set OFFER_MAXIMUM_AGE. Every repository that decides whether an
// offer is servable must fall back to this same value, or a catalog filter and
// the vendor button it stands in for will disagree about which products have
// a live offer.
const DefaultOfferMaximumAge = 72 * time.Hour

// PurchasableOfferExists returns an SQL predicate that is true when the product
// identified by productIDExpr has at least one offer a vendor button may
// redirect through: an active, available, fresh, unexpired offer from an
// active merchant, carrying an active affiliate link inside its validity
// window.
//
// These are the conditions ListPurchasableByProducts and resolveDestination
// apply; a catalog query that says "only tools with a live vendor offer" has
// to mean exactly what the button means, so the conditions are written once
// here and the catalog listing embeds this fragment rather than restating
// them. maxAgeParam names the bound parameter holding the window in seconds,
// e.g. "$2"; productIDExpr is a column reference from the enclosing query,
// e.g. "products.id". Both are query text supplied by the calling repository,
// never request input.
func PurchasableOfferExists(productIDExpr, maxAgeParam string) string {
	return `EXISTS (
		SELECT 1
		FROM commerce.merchant_offers AS purchasable_offers
		JOIN commerce.merchants AS purchasable_merchants
			ON purchasable_merchants.id = purchasable_offers.merchant_id
		JOIN commerce.affiliate_links AS purchasable_links
			ON purchasable_links.merchant_offer_id = purchasable_offers.id
			AND purchasable_links.is_active = true
			AND (purchasable_links.valid_from IS NULL OR purchasable_links.valid_from <= now())
			AND (purchasable_links.valid_until IS NULL OR purchasable_links.valid_until > now())
		WHERE purchasable_offers.product_id = ` + productIDExpr + `
			AND purchasable_offers.is_active = true
			AND purchasable_offers.availability IN ('in_stock', 'backorder')
			AND purchasable_offers.last_checked_at >= now() - make_interval(secs => ` + maxAgeParam + `::double precision)
			AND (purchasable_offers.expires_at IS NULL OR purchasable_offers.expires_at > now())
			AND purchasable_merchants.status = 'active'
	)`
}
