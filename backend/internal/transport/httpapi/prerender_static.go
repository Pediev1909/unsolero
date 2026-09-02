package httpapi

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	catalogapp "rigmark/internal/modules/catalog/application"
	catalog "rigmark/internal/modules/catalog/domain"
	contentapp "rigmark/internal/modules/content/application"
	content "rigmark/internal/modules/content/domain"
)

// The static routes — the home page, the catalog and editorial indexes, the
// text pages — were the last ones served as an empty shell. A fetch of the
// home page returned 3.5 KB whose only text was the <title>, while every
// competitor's home page carries a readable body. The bodies below are what a
// client that never runs the application now receives for those routes.
//
// The copy for the text pages is duplicated from the React page it mirrors,
// because those pages are prose in a component rather than stored content.
// Each renderer names the file it must stay in step with.

// hubListingLimit is the largest page the content service will answer; List
// rejects anything above it as an invalid query, and a rejected query here
// would render a hub with a heading and no links.
const hubListingLimit = 24

// homeFeaturedCount matches frontend/src/components/home/FeaturedProductsSection.tsx.
const homeFeaturedCount = 8

// homeSteps must stay in step with frontend/src/components/home/MethodSection.tsx.
var homeSteps = []staticSection{
	{Heading: "Define your constraints", Paragraphs: []string{
		"Share what the business does, your team size, your monthly budget, and the tools you already run."}},
	{Heading: "Compare complete setups", Paragraphs: []string{
		"UNSOLERO weighs fit, quality, price, compatibility, and redundancy across the whole plan."}},
	{Heading: "Buy in the right order", Paragraphs: []string{
		"See what to buy now, what to skip, lower-cost options, and the upgrades worth considering later."}},
}

// staticSection is one heading with the paragraphs under it, and optionally a
// link that follows them. It is the shape every text page reduces to.
type staticSection struct {
	Heading    string
	Paragraphs []string
	LinkPath   string
	LinkText   string
}

// staticRoutePrerender returns the body and structured data for an indexable
// static route, or empty values for a route that has neither. Every lookup is
// non-fatal in the same way listingProducts is: a failed query costs the body a
// section, never the page.
func (h *Handler) staticRoutePrerender(ctx context.Context, path string) (string, any) {
	_, description := staticRouteMetadata(path)
	switch path {
	case "/":
		return h.renderHomeBody(ctx, description), nil
	case "/products":
		products := h.publishedCatalog(ctx)
		return renderProductIndexBody(description, products),
			structuredDataGraph(h.breadcrumbStructuredData(breadcrumb{"Products", "/products"}),
				h.itemListStructuredData(productListItems(products)))
	case "/categories":
		return renderCategoryIndexBody(description, h.activeCategories(ctx)), nil
	case "/brands":
		return renderBrandIndexBody(description, h.activeBrands(ctx)), nil
	case "/guides", "/articles", "/comparisons", "/stacks":
		section := strings.TrimPrefix(path, "/")
		entries := h.hubEntries(ctx, section)
		return renderHubBody(staticRouteHeading(path), description, entries),
			structuredDataGraph(h.breadcrumbStructuredData(editorialHub(section)),
				h.itemListStructuredData(entryListItems(entries)))
	case "/how-it-works":
		return renderHowItWorksBody(), nil
	case "/about":
		return renderAboutBody(), nil
	case "/affiliate-disclosure":
		return renderAffiliateDisclosureBody(), nil
	case "/offers":
		return renderOffersBody(description, h.purchasableProducts(ctx)), nil
	}
	return "", nil
}

// staticRouteHeading is the visible <h1>, which differs from the <title>: the
// title names the site for a search result, the heading speaks to a reader.
// Each must match the React page for the route (frontend/src/pages: HomePage
// via components/home/Hero, ProductsPage, CategoriesPage, BrandsPage and
// ContentHubPage, OffersPage).
func staticRouteHeading(path string) string {
	switch path {
	case "/":
		return "Build the right software stack."
	case "/products":
		return "Software, judged on what matters."
	case "/categories":
		return "What kind of tool are you after?"
	case "/brands":
		return "Looking for one company in particular?"
	case "/guides":
		return "Buy with a clearer brief."
	case "/articles":
		return "Fewer tools, better chosen."
	case "/comparisons":
		return "Two tools, one decision."
	case "/stacks":
		return "Software stacks"
	case "/offers":
		return "Live vendor offers"
	default:
		title, _ := staticRouteMetadata(path)
		return title
	}
}

// renderHomeBody: heading, promise, the three steps, the catalog's categories
// and a handful of featured tools. Nothing here is a count or a claim the
// catalog cannot back: the categories and products are the live ones.
func (h *Handler) renderHomeBody(ctx context.Context, description string) string {
	body := newStaticBody(staticRouteHeading("/"), description)
	body.writeLinkRow([]staticLink{
		{"/build", "Build my setup"}, {"/categories", "Explore categories"},
		{"/comparisons", "Comparisons"}, {"/guides", "Guides"}, {"/stacks", "Stacks"},
	})

	body.writeHeading("How it works")
	body.WriteString(`<ol class="mt-5 space-y-3 text-body">`)
	for _, step := range homeSteps {
		body.WriteString(`<li><strong>` + html.EscapeString(step.Heading) + `</strong> — ` +
			html.EscapeString(step.Paragraphs[0]) + `</li>`)
	}
	body.WriteString(`</ol>`)

	// Only categories with something in them. A home page that hands a
	// crawler twelve links to empty, non-indexable listings is spending its
	// strongest page on nothing.
	var categories []catalog.Category
	for _, category := range h.activeCategories(ctx) {
		if category.PublishedProducts > 0 {
			categories = append(categories, category)
		}
	}
	if len(categories) > 0 {
		body.writeHeading("Categories")
		body.WriteString(`<ul class="mt-5 space-y-2 text-body">`)
		for _, category := range categories {
			body.writeLinkItem("/categories/"+category.Slug, category.Name, "")
		}
		body.WriteString(`</ul>`)
	}

	if featured := h.featuredProducts(ctx); len(featured) > 0 {
		body.writeHeading("Featured tools")
		// Copy from FeaturedProductsSection.tsx.
		body.writeParagraph("Entry paid tiers, at the price recorded in our catalog. Open any product " +
			"to inspect its specifications, suitability, and currently available merchant offers.")
		writeProductList(body, featured)
	}
	return body.String()
}

// renderProductIndexBody lists every published product under its category,
// so the index is navigable rather than a wall of fifty names.
func renderProductIndexBody(description string, products []catalog.Product) string {
	body := newStaticBody(staticRouteHeading("/products"), description)
	for _, group := range groupByCategory(products) {
		body.WriteString(`<h2 class="mt-10 font-editorial text-2xl">`)
		if group.slug != "" {
			body.WriteString(`<a href="/categories/` + html.EscapeString(group.slug) + `">` +
				html.EscapeString(group.name) + `</a>`)
		} else {
			body.WriteString(html.EscapeString(group.name))
		}
		body.WriteString(`</h2>`)
		writeProductList(body, group.products)
	}
	return body.String()
}

func renderCategoryIndexBody(description string, categories []catalog.Category) string {
	body := newStaticBody(staticRouteHeading("/categories"), description)
	if len(categories) > 0 {
		body.WriteString(`<ul class="mt-8 space-y-3 text-body">`)
		for _, category := range categories {
			body.writeLinkItem("/categories/"+category.Slug, category.Name,
				countNote(category.PublishedProducts, "product"))
		}
		body.WriteString(`</ul>`)
	}
	return body.String()
}

// renderBrandIndexBody leaves out vendors with nothing published, for the same
// reason the React page and the sitemap do: their pages are an empty state.
func renderBrandIndexBody(description string, brands []catalog.Brand) string {
	body := newStaticBody(staticRouteHeading("/brands"), description)
	var listed []catalog.Brand
	for _, brand := range brands {
		if brand.PublishedProducts > 0 {
			listed = append(listed, brand)
		}
	}
	if len(listed) > 0 {
		body.WriteString(`<ul class="mt-8 space-y-3 text-body">`)
		for _, brand := range listed {
			body.writeLinkItem("/brands/"+brand.Slug, brand.Name, countNote(brand.PublishedProducts, "product"))
		}
		body.WriteString(`</ul>`)
	}
	return body.String()
}

// renderHubBody lists a section's entries with title, description and date:
// the same three things the React hub card shows.
func renderHubBody(heading, description string, entries []content.Summary) string {
	body := newStaticBody(heading, description)
	if len(entries) > 0 {
		body.WriteString(`<ul class="mt-8 space-y-6 text-body">`)
		for _, entry := range entries {
			body.WriteString(`<li><a href="` + html.EscapeString(entry.Path) + `">` +
				html.EscapeString(entry.Title) + `</a>`)
			if entry.Description != "" {
				body.WriteString(`<p class="mt-1 text-ink/70">` + html.EscapeString(entry.Description) + `</p>`)
			}
			if !entry.PublishedAt.IsZero() {
				body.WriteString(`<p class="mt-1 text-body-sm text-ink/70"><time datetime="` +
					html.EscapeString(entry.PublishedAt.Format(time.RFC3339)) + `">` +
					html.EscapeString(entry.PublishedAt.Format("2 January 2006")) + `</time></p>`)
			}
			body.WriteString(`</li>`)
		}
		body.WriteString(`</ul>`)
	}
	return body.String()
}

// renderOffersBody lists the products that have a live affiliate offer, each
// linking to its product page. The affiliate redirect itself is deliberately
// absent: a crawler handed /api/affiliate/click/... would follow it, and the
// product page is where the disclosed, tracked control belongs.
func renderOffersBody(description string, offers []offeredProduct) string {
	body := newStaticBody(staticRouteHeading("/offers"), description)
	if len(offers) > 0 {
		body.WriteString(`<ul class="mt-8 space-y-3 text-body">`)
		for _, offer := range offers {
			body.WriteString(`<li>`)
			writeProductLink(body, offer.Product)
			if offer.MerchantName != "" {
				body.WriteString(` <span class="text-ink/70">· via ` + html.EscapeString(offer.MerchantName) + `</span>`)
			}
			body.WriteString(`</li>`)
		}
		body.WriteString(`</ul>`)
	}
	body.writeParagraph("Every vendor link on a product page that earns a commission is labelled there. " +
		"Commission never changes the ranking.")
	return body.String()
}

// The three text pages. Each heading and paragraph is copied from the React
// page named in the comment and must be changed together with it; the server
// body is what a crawler reads and the React page is what a visitor reads, and
// a difference between them is the definition of cloaking.

// renderHowItWorksBody mirrors frontend/src/pages/HowItWorksPage.tsx.
func renderHowItWorksBody() string {
	body := newStaticBody("Start wherever you actually are.",
		"You do not need to know what you are looking for to use this site. Pick whichever of these three sounds like you.")
	body.writeSections([]staticSection{
		{Heading: "I know what kind of tool I need", Paragraphs: []string{
			"Pick the category — CRM, invoicing, help desk — and read the tools in it side by side. Every one shows its price, what that price includes, and where the figure came from."},
			LinkPath: "/categories", LinkText: "Browse the categories"},
		{Heading: "I have no idea where to start", Paragraphs: []string{
			"Answer a few plain questions about what your business does, what you already pay for, and what you can spend. We work out a whole set of tools that fit together, and tell you what to skip."},
			LinkPath: "/build", LinkText: "Build my setup"},
		{Heading: "I am stuck between two or three", Paragraphs: []string{
			"Put them next to each other. Price, what each tier actually gives you, and where they differ on the things that decide it."},
			LinkPath: "/compare", LinkText: "Open the comparison"},
	})
	body.writeHeading("Why you should believe any of this.")
	body.writeParagraph("Every comparison site says it is independent. Here is what that claim is worth here, in detail, so you can check it.")
	body.writeSections([]staticSection{
		{Heading: "Every price was read from the vendor, on a date we tell you", Paragraphs: []string{
			"Nobody here retypes a number from another comparison site. Each price is read from the vendor’s own pricing page, and the site records which page, what day it was read, and how confident that reading is. If a price cannot be verified, the product does not go in the catalog at all — which is why some well-known tools are missing."}},
		{Heading: "The billing basis is always stated, because it is where the trick lives", Paragraphs: []string{
			"One vendor quotes per month, the next quotes the same plan billed annually and shows a smaller number. They are not comparable, so we say which is which on every product. Where a vendor was running a promotion, we publish the standing rate, not the discount, so the comparison does not quietly go stale when the offer ends."}},
		{Heading: "The scores are opinions, and they are labelled as opinions", Paragraphs: []string{
			"Prices are facts. Whether a tool suits a beginner is a judgement. Those are kept apart: every score carries a written reason you can read, and none of them is presented as something the vendor said."}},
		{Heading: "We earn a commission, and it cannot move a ranking", Paragraphs: []string{
			"Some links here earn money if you buy through them, and every one of those is labelled where it appears. The ranking is produced by an engine that is never given the commission figure, so it has nothing to weigh even if it wanted to. A tool that pays us nothing can and does beat one that pays us well."}},
	})
	body.writeLinkRow([]staticLink{
		{"/articles/how-unsolero-ranks-software", "How UNSOLERO ranks software"},
		{"/affiliate-disclosure", "Affiliate disclosure"},
	})
	return body.String()
}

// renderAboutBody mirrors frontend/src/pages/AboutPage.tsx.
func renderAboutBody() string {
	body := newStaticBody("Who is behind this site.",
		"UNSOLERO is built and run by Andon Pediev. Not a team, not an agency — one person who writes the software, chooses which products are listed, and records where every fact came from.")
	body.writeSections([]staticSection{
		{Heading: "Why it exists", Paragraphs: []string{
			"Most software comparison sites rank by whoever pays the most and do not tell you. You read a list, you cannot see how it was ordered, and the order is for sale."},
			LinkPath: "/articles/how-unsolero-ranks-software", LinkText: "The full method"},
		{Heading: "Where the facts come from", Paragraphs: []string{
			"Prices and plan limits are read from each vendor’s own pricing and documentation pages. Every one of them records which page it came from and the date it was read, and that record is published next to the fact rather than kept internally."}},
		{Heading: "What this site does not claim", Paragraphs: []string{
			"Current scoring is built from documented specifications and pricing, not from months of daily use of every product. Where that changes, the page will say so: hands-on notes carry the date they were written and what was actually done, so you can tell a tested opinion from a documented fact."}},
		{Heading: "How it pays for itself", Paragraphs: []string{
			"Through affiliate commission, disclosed in full on the affiliate disclosure page. Commission cannot move a product up the list, and that is enforced by an automated test rather than promised in a sentence."},
			LinkPath: "/affiliate-disclosure", LinkText: "Affiliate disclosure"},
		{Heading: "Contact", Paragraphs: []string{
			"Corrections are welcome, particularly on prices and plan limits. If something here is wrong, write to hello@unsolero.com and say which page and which figure."}},
	})
	return body.String()
}

// renderAffiliateDisclosureBody mirrors frontend/src/pages/AffiliateDisclosurePage.tsx.
func renderAffiliateDisclosureBody() string {
	body := newStaticBody("How we make money.",
		"UNSOLERO may earn a commission when you subscribe to a product after following a link from this site. That costs you nothing extra — the price is the same as going direct.")
	body.writeSections([]staticSection{
		{Heading: "What commission does not do", Paragraphs: []string{
			"It does not affect which products are recommended, in what order, or why. Recommendations are produced by a deterministic engine that has no access to commercial data: commission rates and merchant relationships live in a separate part of the system and are never inputs to scoring."}},
		{Heading: "What we do not do", Paragraphs: []string{
			"We do not accept payment for placement or ranking. We do not hide products because they have no affiliate programme. We do not publish a product until its facts have a recorded source, which is why the catalog grows slowly."}},
		{Heading: "How to check", Paragraphs: []string{
			"Every recommendation shows its reasons and the facts behind them, and rejected products are shown with the reason they were rejected rather than hidden."},
			LinkPath: "/articles/how-unsolero-ranks-software", LinkText: "How we rank software"},
		{Heading: "Prices", Paragraphs: []string{
			"Prices are read from each vendor’s own pricing page and recorded with the date they were read. Software pricing changes often; confirm the current price with the vendor before subscribing."}},
	})
	return body.String()
}

// Data access. Each helper turns a failure into an empty list and a warning,
// so the document still carries its heading and metadata.

func (h *Handler) activeCategories(ctx context.Context) []catalog.Category {
	if h.catalog == nil {
		return nil
	}
	categories, err := h.catalog.ListCategories(ctx)
	if err != nil {
		h.logger.Warn("listing categories for prerendered body", "error", err)
		return nil
	}
	return categories
}

func (h *Handler) activeBrands(ctx context.Context) []catalog.Brand {
	if h.catalog == nil {
		return nil
	}
	brands, err := h.catalog.ListBrands(ctx)
	if err != nil {
		h.logger.Warn("listing brands for prerendered body", "error", err)
		return nil
	}
	return brands
}

// featuredProducts is the home page's preview, in the featured order the
// React section uses. Fixture rows carry the demo- slug prefix (see
// productSummaryDTO) and are development data, so they never reach the home
// page whatever the database holds.
func (h *Handler) featuredProducts(ctx context.Context) []catalog.Product {
	if h.catalog == nil {
		return nil
	}
	page, err := h.catalog.Search(ctx, catalogapp.Query{Sort: "featured", Page: 1, PageSize: homeFeaturedCount})
	if err != nil {
		h.logger.Warn("featured products for prerendered body", "error", err)
		return nil
	}
	products := make([]catalog.Product, 0, len(page.Products))
	for _, product := range page.Products {
		if !strings.HasPrefix(product.Slug, "demo-") {
			products = append(products, product)
		}
	}
	return products
}

// publishedCatalog is everyPublishedProduct (offers.go) with the failure mode
// this file uses everywhere: a warning and an empty list, so the document
// still carries its heading and metadata.
func (h *Handler) publishedCatalog(ctx context.Context) []catalog.Product {
	if h.catalog == nil {
		return nil
	}
	products, err := h.everyPublishedProduct(ctx)
	if err != nil {
		h.logger.Warn("listing catalog for prerendered body", "error", err)
		return nil
	}
	return products
}

func (h *Handler) hubEntries(ctx context.Context, section string) []content.Summary {
	if h.content == nil {
		return nil
	}
	entries, err := h.content.List(ctx, contentapp.ListQuery{Section: section, Limit: hubListingLimit})
	if err != nil {
		h.logger.Warn("listing entries for prerendered body", "error", err, "section", section)
		return nil
	}
	return entries
}

// offeredProduct is a published product together with the merchant behind its
// live offer. Only the merchant's name crosses over; the offer id, which is
// what the redirect path is built from, stays out of the static body.
type offeredProduct struct {
	Product      catalog.Product
	MerchantName string
}

// purchasableProducts asks commerce which published products have a servable
// affiliate offer right now, under the same conditions the redirect applies.
func (h *Handler) purchasableProducts(ctx context.Context) []offeredProduct {
	products := h.publishedCatalog(ctx)
	if len(products) == 0 || h.commerce == nil {
		return nil
	}
	ids := make([]catalog.ProductID, 0, len(products))
	for _, product := range products {
		ids = append(ids, product.ID)
	}
	offers, err := h.commerce.ListPurchasable(ctx, ids)
	if err != nil {
		h.logger.Warn("purchasable products for prerendered body", "error", err)
		return nil
	}
	var offered []offeredProduct
	for _, product := range products {
		if offer, found := offers[product.ID]; found {
			offered = append(offered, offeredProduct{Product: product, MerchantName: offer.MerchantName})
		}
	}
	return offered
}

// Rendering helpers shared by the bodies above.

type staticLink struct {
	Path, Text string
}

// staticBody is a <main> under construction: a heading, an optional intro, and
// whatever sections follow. String closes it.
type staticBody struct {
	strings.Builder
}

func newStaticBody(heading, description string) *staticBody {
	body := &staticBody{}
	body.WriteString(`<main class="mx-auto max-w-reading px-4 py-12">`)
	body.WriteString(`<h1 class="font-editorial text-4xl">` + html.EscapeString(heading) + `</h1>`)
	if description != "" {
		body.WriteString(`<p class="mt-4 text-body-lg text-ink/70">` + html.EscapeString(description) + `</p>`)
	}
	return body
}

func (body *staticBody) String() string {
	return body.Builder.String() + `</main>`
}

func (body *staticBody) writeHeading(text string) {
	body.WriteString(`<h2 class="mt-10 font-editorial text-2xl">` + html.EscapeString(text) + `</h2>`)
}

func (body *staticBody) writeParagraph(text string) {
	body.WriteString(`<p class="mt-5 text-body">` + html.EscapeString(text) + `</p>`)
}

func (body *staticBody) writeLinkRow(links []staticLink) {
	rendered := make([]string, 0, len(links))
	for _, link := range links {
		rendered = append(rendered, `<a class="underline" href="`+html.EscapeString(link.Path)+`">`+
			html.EscapeString(link.Text)+`</a>`)
	}
	body.WriteString(`<p class="mt-6 text-body">` + strings.Join(rendered, " · ") + `</p>`)
}

// writeLinkItem is one <li> holding a link and an optional note after it.
func (body *staticBody) writeLinkItem(path, text, note string) {
	body.WriteString(`<li><a href="` + html.EscapeString(path) + `">` + html.EscapeString(text) + `</a>`)
	if note != "" {
		body.WriteString(` <span class="text-ink/70">` + html.EscapeString(note) + `</span>`)
	}
	body.WriteString(`</li>`)
}

func (body *staticBody) writeSections(sections []staticSection) {
	for _, section := range sections {
		body.writeHeading(section.Heading)
		for _, paragraph := range section.Paragraphs {
			body.writeParagraph(paragraph)
		}
		if section.LinkPath != "" && section.LinkText != "" {
			body.writeLinkRow([]staticLink{{section.LinkPath, section.LinkText}})
		}
	}
}

// writeProductLink is a product's name as a link, its vendor, and its price
// with the caveat the product page states. The price only appears when the
// catalog holds one.
func writeProductLink(body *staticBody, product catalog.Product) {
	body.WriteString(`<a href="/products/` + html.EscapeString(product.Slug) + `">` +
		html.EscapeString(product.Name) + `</a>`)
	if product.BrandName != "" {
		body.WriteString(` <span class="text-ink/70">` + html.EscapeString(product.BrandName) + `</span>`)
	}
	if product.Price.AmountMinor > 0 && product.Price.Currency != "" {
		body.WriteString(` <span class="text-ink/70">— ` + html.EscapeString(formatMoney(product.Price)) +
			` per month, entry paid tier</span>`)
	}
}

func writeProductList(body *staticBody, products []catalog.Product) {
	if len(products) == 0 {
		return
	}
	body.WriteString(`<ul class="mt-5 space-y-3 text-body">`)
	for _, product := range products {
		body.WriteString(`<li>`)
		writeProductLink(body, product)
		body.WriteString(`</li>`)
	}
	body.WriteString(`</ul>`)
}

type categoryGroup struct {
	name, slug string
	products   []catalog.Product
}

// groupByCategory buckets products under their category, categories in name
// order and products in the order they arrived.
func groupByCategory(products []catalog.Product) []categoryGroup {
	index := map[string]int{}
	var groups []categoryGroup
	for _, product := range products {
		key := product.CategorySlug
		position, seen := index[key]
		if !seen {
			position = len(groups)
			index[key] = position
			groups = append(groups, categoryGroup{name: defaultText(product.CategoryName, "Other"), slug: key})
		}
		groups[position].products = append(groups[position].products, product)
	}
	sort.SliceStable(groups, func(left, right int) bool {
		return strings.ToLower(groups[left].name) < strings.ToLower(groups[right].name)
	})
	return groups
}

func countNote(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

// Structured data for the listings.

func productListItems(products []catalog.Product) []listedItem {
	items := make([]listedItem, 0, len(products))
	for _, product := range products {
		items = append(items, listedItem{Name: product.Name, Path: "/products/" + product.Slug})
	}
	return items
}

func entryListItems(entries []content.Summary) []listedItem {
	items := make([]listedItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, listedItem{Name: entry.Title, Path: entry.Path})
	}
	return items
}
