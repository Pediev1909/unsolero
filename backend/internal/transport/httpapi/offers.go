package httpapi

import (
	"context"
	"net/http"
	"sort"
	"time"

	catalog "rigmark/internal/modules/catalog/application"
	"rigmark/internal/modules/catalog/domain"
	commercedomain "rigmark/internal/modules/commerce/domain"
)

// liveOffersResponse is every product in the catalog that a reader can act on
// right now: an active, in-stock, fresh, unexpired offer behind a live
// affiliate link, under exactly the conditions the redirect will re-check.
//
// It exists because every affiliate site has a deals page and this one did
// not. A video whose answer is "the offers" had nowhere to point, and a
// social bio needed one URL that stays true as offers come and go.
type liveOffersResponse struct {
	Items []liveOfferItemResponse `json:"items"`
	// GeneratedAt is when this list was assembled. The page prints it so the
	// reader knows the list, not just each price, has a date.
	GeneratedAt string `json:"generated_at"`
}

type liveOfferItemResponse struct {
	// Product is the same summary the catalog grid draws, vendor button
	// fields included, so the offers page and the catalog cannot disagree
	// about a product.
	Product productSummaryResponse `json:"product"`
	Offer   liveOfferResponse      `json:"offer"`
}

// liveOfferResponse is the part of an offer a listing row prints. It carries
// no merchant URL, affiliate destination or provider reference: the only way
// out is the product's tracked purchase_path.
type liveOfferResponse struct {
	Price           moneyResponse `json:"price"`
	MerchantName    string        `json:"merchant_name"`
	LastCheckedAt   string        `json:"last_checked_at"`
	FreshnessStatus string        `json:"freshness_status"`
}

// listLiveOffers answers GET /api/catalog/offers.
//
// It walks the published catalog and asks commerce once for the whole set,
// the same way a grid page does, then keeps only the products commerce
// vouched for. Unlike attachPurchasePaths this fails closed when commerce is
// down: there the buttons are supplementary to a catalog that still stands,
// here the offers are the page, and an empty list would read as "nothing is
// on offer" when the truth is "we could not check".
func (h *Handler) listLiveOffers(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	products, err := h.everyPublishedProduct(ctx)
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	ids := make([]domain.ProductID, 0, len(products))
	for _, product := range products {
		ids = append(ids, product.ID)
	}
	offers, err := h.commerce.ListPurchasable(ctx, ids)
	if err != nil {
		h.logger.Error("list live offers", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "offers_unavailable", "Vendor offers are temporarily unavailable.", nil, h.logger)
		return
	}

	items := make([]liveOfferItemResponse, 0, len(offers))
	for _, product := range products {
		offer, found := offers[product.ID]
		if !found {
			continue
		}
		items = append(items, liveOfferItemDTO(product, offer))
	}
	// Grouped by category for the reader, in the catalog's own order within
	// each group. Stable, so two products in one category keep the ranking
	// the catalog gave them.
	sort.SliceStable(items, func(left, right int) bool {
		return items[left].Product.Category.Name < items[right].Product.Category.Name
	})

	response.Header().Set("Cache-Control", "public, max-age=30")
	writeJSON(response, http.StatusOK, liveOffersResponse{
		Items:       items,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}, h.logger)
}

// everyPublishedProduct reads the whole published catalog, a page at a time,
// and fails if any page does: a partial list here is a lie about what is on
// offer. (prerender_static.go has a tolerant sibling for a prerendered body,
// where a shorter index is better than none.)
//
// The catalog refuses a page larger than MaximumPageSize rather than
// clamping it, and the catalog is already larger than that. Two requests for
// fifty-odd products is cheaper than a new repository method whose one caller
// would be this listing.
func (h *Handler) everyPublishedProduct(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	for page := 1; ; page++ {
		result, err := h.catalog.Search(ctx, catalog.Query{Page: page, PageSize: catalog.MaximumPageSize})
		if err != nil {
			return nil, err
		}
		products = append(products, result.Products...)
		if page >= result.TotalPages || len(result.Products) == 0 {
			return products, nil
		}
	}
}

func liveOfferItemDTO(product domain.Product, offer commercedomain.PurchasableOffer) liveOfferItemResponse {
	summary := productSummaryDTO(product)
	path := "/api/affiliate/click/" + string(offer.OfferID)
	label := offer.DisclosureLabel
	merchant := offer.MerchantName
	summary.PurchasePath = &path
	summary.DisclosureLabel = &label
	summary.MerchantName = &merchant
	return liveOfferItemResponse{
		Product: summary,
		Offer: liveOfferResponse{
			Price:           moneyResponse{AmountMinor: offer.Price.AmountMinor, Currency: offer.Price.Currency},
			MerchantName:    offer.MerchantName,
			LastCheckedAt:   offer.LastCheckedAt.UTC().Format("2006-01-02T15:04:05Z"),
			FreshnessStatus: offerFreshness(offer.LastCheckedAt),
		},
	}
}
