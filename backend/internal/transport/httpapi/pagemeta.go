package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	content "rigmark/internal/modules/content/domain"
)

// Public HTML is served as a single-page shell, so without this every route
// returned the same title and description. Crawlers that do not execute
// JavaScript — every social preview scraper, and most non-Google crawlers —
// saw the home page for every product and article on the site. This injects
// per-route metadata into the shell before it is sent.
//
// Editorial routes additionally carry their text in the document, because a
// correct title above an empty body still reads as an empty page to anything
// that does not run the application. See renderEntryBody.

const (
	pageMetaStartMarker = "<!--PAGE_META_START-->"
	pageMetaEndMarker   = "<!--PAGE_META_END-->"
	// Must stay in step with frontend/index.html and frontend/src/main.tsx.
	rootMountPoint = `<div id="root"></div>`
)

// documentContentSecurityPolicy is the policy an HTML document needs, as
// opposed to the "default-src 'none'" the API applies to its JSON responses.
//
// This matters because serving the shell from the API moved the document out
// from behind nginx, which used to attach this header. Without it the browser
// blocks every script and stylesheet the page references and renders nothing:
// the API's policy is correct for JSON and fatal for HTML.
//
// It must stay in step with frontend/nginx-static-security.conf, which applies
// the same policy on the fallback path.
const documentContentSecurityPolicy = "default-src 'self'; script-src 'self' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline'; img-src 'self' https: data:; connect-src 'self'; " +
	"font-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; " +
	"frame-ancestors 'none'; upgrade-insecure-requests"

type pageMetadata struct {
	Title        string
	Description  string
	CanonicalURL string
	ImageURL     string
	// FallbackImageURL is the social card used when ImageURL is empty or is
	// an SVG, which the social platforms refuse to render. It is chosen per
	// page type (see socialImagePath); empty falls through to defaultSocialImage.
	FallbackImageURL string
	// StructuredData is emitted as JSON-LD. Empty means the shell's default
	// WebSite schema is left in place.
	StructuredData any
	// Indexable is false for account and admin routes, and for any URL
	// carrying a query string.
	Indexable bool
	// PrerenderedBody is the page's content as HTML, placed inside the root
	// element so clients that do not run JavaScript still receive it.
	PrerenderedBody string
}

// spaShellProvider fetches the built index.html from the web container and
// caches it briefly.
//
// The refresh interval is short on purpose. The shell references asset bundles
// by content hash, so deploying a new frontend deletes the files the cached
// copy points at: every second of cache after a deploy is a second of serving
// HTML whose script and stylesheet return 404, which is a blank page rather
// than a degraded one. Fetching is a request to nginx on the same Docker
// network, so a short interval costs effectively nothing.
type spaShellProvider struct {
	url     string
	client  *http.Client
	refresh time.Duration

	mutex     sync.RWMutex
	shell     string
	fetchedAt time.Time
}

func newSPAShellProvider(url string) *spaShellProvider {
	return &spaShellProvider{
		url:     url,
		client:  &http.Client{Timeout: 3 * time.Second},
		refresh: 10 * time.Second,
	}
}

// Shell returns the cached shell, refreshing it when stale. The second return
// value reports whether a shell is available at all; when it is not, the caller
// must fall back to serving the static file so a metadata problem never becomes
// an outage.
func (provider *spaShellProvider) Shell(ctx context.Context) (string, bool) {
	provider.mutex.RLock()
	shell, fetchedAt := provider.shell, provider.fetchedAt
	provider.mutex.RUnlock()

	if shell != "" && time.Since(fetchedAt) < provider.refresh {
		return shell, true
	}

	fetched, err := provider.fetch(ctx)
	if err != nil {
		// The previous copy is the best available answer while web is briefly
		// unreachable. It may point at assets that no longer exist, but the
		// alternative is serving no metadata at all, and the edge fallback
		// would serve the same stale file from disk anyway.
		return shell, shell != ""
	}
	provider.mutex.Lock()
	provider.shell, provider.fetchedAt = fetched, time.Now()
	provider.mutex.Unlock()
	return fetched, true
}

func (provider *spaShellProvider) fetch(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.url, nil)
	if err != nil {
		return "", err
	}
	response, err := provider.client.Do(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spa shell responded %d", response.StatusCode)
	}
	// The shell is a small document; the limit stops a misrouted response from
	// being buffered in full.
	body, err := io.ReadAll(io.LimitReader(response.Body, 512<<10))
	if err != nil {
		return "", err
	}
	shell := string(body)
	if !strings.Contains(shell, pageMetaStartMarker) || !strings.Contains(shell, pageMetaEndMarker) {
		return "", fmt.Errorf("spa shell has no page metadata markers")
	}
	return shell, nil
}

// renderShell replaces the shell's default metadata block with this route's.
// Every injected value is escaped: product names and descriptions are catalog
// data, and a stray quote would otherwise break out of an attribute.
func renderShell(shell string, meta pageMetadata) (string, bool) {
	start := strings.Index(shell, pageMetaStartMarker)
	end := strings.Index(shell, pageMetaEndMarker)
	if start < 0 || end < 0 || end < start {
		return "", false
	}

	var block strings.Builder
	block.WriteString(pageMetaStartMarker)
	writeMetaTag(&block, "title", meta.Title)
	writeNamedMeta(&block, "description", meta.Description)
	writeProperty(&block, "og:type", "website")
	writeProperty(&block, "og:site_name", "UNSOLERO")
	writeProperty(&block, "og:title", meta.Title)
	writeProperty(&block, "og:description", meta.Description)
	if meta.CanonicalURL != "" {
		block.WriteString("\n    <link rel=\"canonical\" href=\"" + html.EscapeString(meta.CanonicalURL) + "\" />")
		writeProperty(&block, "og:url", meta.CanonicalURL)
	}
	// A card is promised below as summary_large_image, so one has to exist.
	// Two things kept breaking that promise: products carry no images at all,
	// and every editorial hero is an SVG, which LinkedIn, X and Facebook all
	// refuse to render. Both cases fell through to no og:image and produced a
	// bare grey card on exactly the platforms the site is shared on.
	if image := meta.ImageURL; image != "" {
		writeProperty(&block, "og:image", image)
		writeProperty(&block, "og:image:width", "1200")
		writeProperty(&block, "og:image:height", "630")
		writeNamedMeta(&block, "twitter:image", image)
	}
	writeNamedMeta(&block, "twitter:card", "summary_large_image")
	writeNamedMeta(&block, "twitter:title", meta.Title)
	writeNamedMeta(&block, "twitter:description", meta.Description)
	if !meta.Indexable {
		writeNamedMeta(&block, "robots", "noindex, nofollow")
	}
	if meta.StructuredData != nil {
		if encoded, err := json.Marshal(meta.StructuredData); err == nil {
			// JSON-LD sits inside a script element. Escaping "<" as its
			// unicode form keeps a product name containing "</script>" from
			// closing the block and injecting markup.
			safe := strings.ReplaceAll(string(encoded), "<", `\u003c`)
			block.WriteString("\n    <script type=\"application/ld+json\">" + safe + "</script>")
		}
	}
	block.WriteString("\n    ")

	rendered := shell[:start] + block.String() + shell[end+len(pageMetaEndMarker):]

	if meta.PrerenderedBody != "" {
		// React clears the container on mount, so this is replaced the moment
		// the application starts. It exists for everything that never gets
		// that far. If the mount point is ever renamed the replacement simply
		// does not happen, which costs the prerender rather than the page.
		rendered = strings.Replace(rendered, rootMountPoint,
			`<div id="root">`+meta.PrerenderedBody+`</div>`, 1)
	}
	return rendered, true
}

func writeMetaTag(block *strings.Builder, tag, value string) {
	if value == "" {
		return
	}
	block.WriteString("\n    <" + tag + ">" + html.EscapeString(value) + "</" + tag + ">")
}

func writeNamedMeta(block *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	block.WriteString("\n    <meta name=\"" + name + "\" content=\"" + html.EscapeString(value) + "\" />")
}

func writeProperty(block *strings.Builder, property, value string) {
	if value == "" {
		return
	}
	block.WriteString("\n    <meta property=\"" + property + "\" content=\"" + html.EscapeString(value) + "\" />")
}

// truncateDescription keeps descriptions within the length search results
// actually display, cutting on a word boundary rather than mid-word.
//
// The limit counts characters rather than bytes. Slicing a UTF-8 string by
// byte offset splits any multi-byte character it lands inside, and product
// descriptions carry em-dashes and accented brand names routinely.
func truncateDescription(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	cut := string(runes[:limit])
	if index := strings.LastIndex(cut, " "); index > 0 && len([]rune(cut[:index])) > limit/2 {
		cut = cut[:index]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}

// productStructuredData emits schema.org/Product so search results can carry
// the price and vendor. It states only what the catalog holds with recorded
// provenance: no ratings, no review counts, no availability we have not
// observed. Structured data that overstates what is known is the fastest way
// to earn a manual penalty.
func productStructuredData(product catalog.Product, canonicalURL string) map[string]any {
	data := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "Product",
		"name":        product.Name,
		"description": truncateDescription(product.Description, 300),
		"category":    product.CategoryName,
	}
	if canonicalURL != "" {
		data["url"] = canonicalURL
	}
	if product.BrandName != "" {
		data["brand"] = map[string]any{"@type": "Brand", "name": product.BrandName}
	}
	if product.Price.AmountMinor > 0 && product.Price.Currency != "" {
		data["offers"] = map[string]any{
			"@type":         "Offer",
			"price":         fmt.Sprintf("%.2f", float64(product.Price.AmountMinor)/100),
			"priceCurrency": product.Price.Currency,
			"url":           canonicalURL,
		}
	}
	return data
}

// articleStructuredData emits schema.org/Article for editorial pages.
func (h *Handler) articleStructuredData(entry content.Entry, canonicalURL string) map[string]any {
	data := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "Article",
		"headline":    entry.Title,
		"description": truncateDescription(entry.Description, 300),
	}
	if canonicalURL != "" {
		data["url"] = canonicalURL
		data["mainEntityOfPage"] = canonicalURL
	}
	if entry.AuthorName != "" {
		// A Person, not an Organization. Search engines weigh attribution to a
		// named, identifiable human, and the url is what turns a name into an
		// entity they can look up rather than a string on a page.
		author := map[string]any{"@type": "Person", "name": entry.AuthorName}
		if entry.Author.Slug != "" {
			author["url"] = h.absolutePublicRoute("/author/" + entry.Author.Slug)
		}
		data["author"] = author
	}
	if !entry.PublishedAt.IsZero() {
		data["datePublished"] = entry.PublishedAt.UTC().Format(time.RFC3339)
	}
	if !entry.UpdatedAt.IsZero() {
		data["dateModified"] = entry.UpdatedAt.UTC().Format(time.RFC3339)
	}
	data["publisher"] = map[string]any{"@type": "Organization", "name": "UNSOLERO"}
	return data
}

// renderEntryBody turns an editorial entry into the HTML that ships inside the
// document, so the article text is present before any JavaScript runs.
//
// Until now a fetch of an article returned about 2.8 KB of shell: correct
// title, description and JSON-LD, and not one sentence of the article. Google
// executes JavaScript and eventually sees the page, but nothing else in the
// chain does — other crawlers, link previews, and the assistants the site
// publishes an llms.txt for, which was pointing them at pages that answered
// with an empty document.
//
// React clears the container when it mounts, so this content is what a reader
// sees for the moment before the application takes over, and is what a client
// that never runs the application sees permanently. Blocks are a closed set of
// structured types rather than stored markup, so every value here is escaped
// and no caller can inject elements.
func renderEntryBody(entry content.Entry) string {
	var body strings.Builder
	body.WriteString(`<article class="mx-auto max-w-reading px-4 py-12">`)
	body.WriteString(`<h1 class="font-editorial text-4xl">` + html.EscapeString(entry.Title) + `</h1>`)
	if entry.Description != "" {
		body.WriteString(`<p class="mt-4 text-body-lg text-ink/70">` +
			html.EscapeString(entry.Description) + `</p>`)
	}
	if entry.Author.Name != "" {
		body.WriteString(`<p class="mt-4 text-body-sm text-ink/70">By ` +
			html.EscapeString(entry.Author.Name))
		if !entry.PublishedAt.IsZero() {
			// A machine-readable date next to the visible one is what a
			// crawler uses to decide how current the page is.
			body.WriteString(` · <time datetime="` +
				html.EscapeString(entry.PublishedAt.Format(time.RFC3339)) + `">` +
				html.EscapeString(entry.PublishedAt.Format("2 January 2006")) + `</time>`)
		}
		body.WriteString(`</p>`)
	}
	for _, block := range entry.Content {
		writeEntryBlock(&body, block)
	}
	body.WriteString(`</article>`)
	return body.String()
}

// renderProductBody ships the product's heading, facts and its place in the
// catalog inside the document.
//
// A product route previously returned a correct title above an empty body. An
// audit of the live site reported 125 pages with no <h1> and 126 with no
// inbound internal link: both are the same defect seen twice, because a crawler
// that does not run the application sees neither the heading nor a single
// anchor. The category and brand links below are what connect a product page to
// the rest of the site without JavaScript.
func renderProductBody(product catalog.Product) string {
	var body strings.Builder
	body.WriteString(`<main class="mx-auto max-w-reading px-4 py-12">`)
	body.WriteString(`<h1 class="font-editorial text-4xl">` + html.EscapeString(product.Name) + `</h1>`)

	// Brand and category as links rather than text: this is the page's only
	// route back into the catalog for a client that never runs the bundle.
	var trail []string
	if product.BrandName != "" && product.BrandSlug != "" {
		trail = append(trail, `<a href="/brands/`+html.EscapeString(product.BrandSlug)+`">`+
			html.EscapeString(product.BrandName)+`</a>`)
	}
	if product.CategoryName != "" && product.CategorySlug != "" {
		trail = append(trail, `<a href="/categories/`+html.EscapeString(product.CategorySlug)+`">`+
			html.EscapeString(product.CategoryName)+`</a>`)
	}
	if len(trail) > 0 {
		body.WriteString(`<p class="mt-3 text-body-sm text-ink/70">` +
			strings.Join(trail, " · ") + `</p>`)
	}

	if product.Price.AmountMinor > 0 && product.Price.Currency != "" {
		body.WriteString(`<p class="mt-4 text-body-lg">` +
			html.EscapeString(formatMoney(product.Price)) +
			` <span class="text-ink/70">per month, entry paid tier</span></p>`)
	}
	if product.Description != "" {
		body.WriteString(`<p class="mt-5 text-body">` +
			html.EscapeString(product.Description) + `</p>`)
	}
	body.WriteString(`</main>`)
	return body.String()
}

// renderCatalogListingBody ships a heading and the products the listing holds,
// so a category or brand page carries real links instead of an empty container.
// This is what gives every product page an inbound internal link.
func renderCatalogListingBody(heading, description string, products []catalog.Product) string {
	var body strings.Builder
	body.WriteString(`<main class="mx-auto max-w-reading px-4 py-12">`)
	body.WriteString(`<h1 class="font-editorial text-4xl">` + html.EscapeString(heading) + `</h1>`)
	if description != "" {
		body.WriteString(`<p class="mt-4 text-body-lg text-ink/70">` +
			html.EscapeString(description) + `</p>`)
	}
	if len(products) > 0 {
		body.WriteString(`<ul class="mt-8 space-y-3 text-body">`)
		for _, product := range products {
			body.WriteString(`<li><a href="/products/` + html.EscapeString(product.Slug) + `">` +
				html.EscapeString(product.Name) + `</a>`)
			if product.BrandName != "" {
				body.WriteString(` <span class="text-ink/70">` +
					html.EscapeString(product.BrandName) + `</span>`)
			}
			body.WriteString(`</li>`)
		}
		body.WriteString(`</ul>`)
	}
	body.WriteString(`</main>`)
	return body.String()
}

func formatMoney(value catalog.Money) string {
	return fmt.Sprintf("%s %.2f", value.Currency, float64(value.AmountMinor)/100)
}

func writeEntryBlock(body *strings.Builder, block content.Block) {
	switch block.Type {
	case content.BlockHeading:
		if block.Heading != "" {
			body.WriteString(`<h2 class="mt-10 font-editorial text-2xl">` +
				html.EscapeString(block.Heading) + `</h2>`)
		}
	case content.BlockParagraph:
		if block.Text != "" {
			body.WriteString(`<p class="mt-5 text-body">` + html.EscapeString(block.Text) + `</p>`)
		}
	case content.BlockUnordered, content.BlockOrdered:
		if len(block.Items) == 0 {
			return
		}
		tag := "ul"
		if block.Type == content.BlockOrdered {
			tag = "ol"
		}
		body.WriteString(`<` + tag + ` class="mt-5 space-y-2 text-body">`)
		for _, item := range block.Items {
			body.WriteString(`<li>` + html.EscapeString(item) + `</li>`)
		}
		body.WriteString(`</` + tag + `>`)
	case content.BlockQuote, content.BlockCallout:
		if block.Text == "" {
			return
		}
		body.WriteString(`<blockquote class="mt-6 border-l-2 border-ink/20 pl-4 text-body">` +
			html.EscapeString(block.Text))
		if block.Attribution != "" {
			body.WriteString(`<footer class="mt-2 text-body-sm text-ink/70">— ` +
				html.EscapeString(block.Attribution) + `</footer>`)
		}
		body.WriteString(`</blockquote>`)
	case content.BlockCTA:
		if block.Promotion == "" || block.Label == "" {
			return
		}
		// source=promotion is not decoration: TrackPromotionClick rejects a
		// click whose source is anything else, and the handler defaults an
		// absent source to product_detail. Without it this link resolves to an
		// error for every reader who has JavaScript off, which is exactly the
		// reader this server-rendered body exists for.
		//
		// rel carries sponsored as well as nofollow. The body is served to
		// crawlers, and an undisclosed paid link in indexed HTML is the thing
		// search engines penalise a site for.
		href := "/api/affiliate/promotion/" + url.PathEscape(block.Promotion) + "?source=promotion"
		body.WriteString(`<aside class="mt-6 border border-ink/15 p-5">`)
		if block.Heading != "" {
			body.WriteString(`<h3 class="font-semibold">` + html.EscapeString(block.Heading) + `</h3>`)
		}
		body.WriteString(`<p class="mt-2 text-body">` + html.EscapeString(block.Text) + `</p>`)
		body.WriteString(`<p class="mt-4"><a class="underline" rel="nofollow noopener sponsored" target="_blank" href="` +
			html.EscapeString(href) + `">` + html.EscapeString(block.Label) + `</a></p>`)
		body.WriteString(`</aside>`)
	case content.BlockProsCons:
		if len(block.Pros) == 0 && len(block.Cons) == 0 {
			return
		}
		body.WriteString(`<section class="mt-8">`)
		if block.Heading != "" {
			body.WriteString(`<h2 class="font-editorial text-2xl">` + html.EscapeString(block.Heading) + `</h2>`)
		}
		if block.Text != "" {
			body.WriteString(`<p class="mt-3 text-body">` + html.EscapeString(block.Text) + `</p>`)
		}
		writeTitledList(body, "Pros", block.Pros)
		writeTitledList(body, "Cons", block.Cons)
		body.WriteString(`</section>`)
	case content.BlockFAQ:
		if len(block.Questions) == 0 {
			return
		}
		body.WriteString(`<section class="mt-8">`)
		if block.Heading != "" {
			body.WriteString(`<h2 class="font-editorial text-2xl">` + html.EscapeString(block.Heading) + `</h2>`)
		}
		for _, pair := range block.Questions {
			body.WriteString(`<h3 class="mt-5 font-semibold">` + html.EscapeString(pair.Question) + `</h3>`)
			body.WriteString(`<p class="mt-2 text-body">` + html.EscapeString(pair.Answer) + `</p>`)
		}
		body.WriteString(`</section>`)
	case content.BlockOffer:
		if block.Product == "" {
			return
		}
		// The link goes to the product page, never to the affiliate redirect.
		// This body is served to crawlers, and the product page is where the
		// live, disclosed, tracked control is resolved — from an offer that
		// may not exist by the time this HTML is read.
		href := "/products/" + url.PathEscape(block.Product)
		body.WriteString(`<aside class="mt-6 border border-ink/15 p-5">`)
		if block.Heading != "" {
			body.WriteString(`<h3 class="font-semibold">` + html.EscapeString(block.Heading) + `</h3>`)
		}
		if block.Text != "" {
			body.WriteString(`<p class="mt-2 text-body">` + html.EscapeString(block.Text) + `</p>`)
		}
		body.WriteString(`<p class="mt-4"><a class="underline" href="` + html.EscapeString(href) + `">` +
			html.EscapeString(defaultText(block.Label, "See the product page")) + `</a></p>`)
		body.WriteString(`<p class="mt-2 text-body-sm text-ink/70">Affiliate link on the product page. ` +
			`Commission never changes the ranking.</p>`)
		body.WriteString(`</aside>`)
	}
}

// writeTitledList is an h3 over a list, for the two halves of a pros-and-cons
// block. An empty side writes nothing, so the heading never stands alone.
func writeTitledList(body *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	body.WriteString(`<h3 class="mt-5 font-semibold">` + html.EscapeString(title) + `</h3>`)
	body.WriteString(`<ul class="mt-2 space-y-2 text-body">`)
	for _, item := range items {
		body.WriteString(`<li>` + html.EscapeString(item) + `</li>`)
	}
	body.WriteString(`</ul>`)
}

// renderAuthorBody ships the author's name, biography and published work inside
// the document, for the same reason the entries do: a page whose body only
// exists after JavaScript runs is an empty page to everything that does not run
// it, and this one exists to be read by exactly those readers.
func renderAuthorBody(author content.Author, entries []content.Summary) string {
	var body strings.Builder
	body.WriteString(`<main class="mx-auto max-w-reading px-4 py-12">`)
	body.WriteString(`<h1 class="font-editorial text-4xl">` + html.EscapeString(author.Name) + `</h1>`)
	if author.Bio != "" {
		body.WriteString(`<p class="mt-4 text-body">` + html.EscapeString(author.Bio) + `</p>`)
	}
	if len(entries) > 0 {
		body.WriteString(`<h2 class="mt-10 font-editorial text-2xl">Published work</h2><ul class="mt-5 space-y-3">`)
		for _, entry := range entries {
			body.WriteString(`<li><a href="` + html.EscapeString(entry.Path) + `">` +
				html.EscapeString(entry.Title) + `</a></li>`)
		}
		body.WriteString(`</ul>`)
	}
	body.WriteString(`</main>`)
	return body.String()
}

// Structured data that describes where a page sits and what it lists. Each
// builder returns one schema.org node without an @context; structuredDataGraph
// joins them, because a page gets one JSON-LD block and a product page has both
// a Product and a BreadcrumbList to declare.

// breadcrumb is one step of a trail: a name and the site-relative path it
// leads to. The last step is the page itself.
type breadcrumb struct {
	Name, Path string
}

// listedItem is one entry of an ItemList.
type listedItem struct {
	Name, Path string
}

// structuredDataGraph joins nodes under one @context. Nil nodes are skipped so
// a caller can pass an optional node — the FAQPage that exists only when the
// entry has a FAQ block — without a branch.
func structuredDataGraph(nodes ...map[string]any) map[string]any {
	graph := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		delete(node, "@context")
		graph = append(graph, node)
	}
	return map[string]any{"@context": "https://schema.org", "@graph": graph}
}

// breadcrumbStructuredData emits schema.org/BreadcrumbList for the trail from
// the home page to this one. Home is always the first step; callers pass the
// rest, ending with the current page.
func (h *Handler) breadcrumbStructuredData(trail ...breadcrumb) map[string]any {
	steps := append([]breadcrumb{{Name: "Home", Path: "/"}}, trail...)
	items := make([]map[string]any, 0, len(steps))
	for index, step := range steps {
		items = append(items, map[string]any{
			"@type": "ListItem", "position": index + 1,
			"name": step.Name, "item": h.publicRouteURL(step.Path),
		})
	}
	return map[string]any{"@type": "BreadcrumbList", "itemListElement": items}
}

// itemListStructuredData emits schema.org/ItemList for a listing page: the
// products under a category or vendor, the entries of an editorial hub.
func (h *Handler) itemListStructuredData(items []listedItem) map[string]any {
	elements := make([]map[string]any, 0, len(items))
	for index, item := range items {
		elements = append(elements, map[string]any{
			"@type": "ListItem", "position": index + 1,
			"name": item.Name, "url": h.publicRouteURL(item.Path),
		})
	}
	return map[string]any{
		"@type": "ItemList", "numberOfItems": len(elements), "itemListElement": elements,
	}
}

// faqStructuredData emits schema.org/FAQPage from an entry's FAQ blocks, or nil
// when the entry has none. The questions are the ones the editor wrote, so the
// markup never claims a FAQ the page does not show.
func faqStructuredData(blocks []content.Block) map[string]any {
	var questions []map[string]any
	for _, block := range blocks {
		if block.Type != content.BlockFAQ {
			continue
		}
		for _, pair := range block.Questions {
			questions = append(questions, map[string]any{
				"@type": "Question", "name": pair.Question,
				"acceptedAnswer": map[string]any{"@type": "Answer", "text": pair.Answer},
			})
		}
	}
	if len(questions) == 0 {
		return nil
	}
	return map[string]any{"@type": "FAQPage", "mainEntity": questions}
}

// publicRouteURL is the absolute URL for a site-relative path, falling back to
// the path itself when no site URL is configured, so structured data is never
// left with an empty value.
func (h *Handler) publicRouteURL(path string) string {
	return defaultText(h.absolutePublicRoute(path), path)
}
