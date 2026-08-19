package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
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
	if meta.ImageURL != "" {
		writeProperty(&block, "og:image", meta.ImageURL)
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
func articleStructuredData(entry content.Entry, canonicalURL string) map[string]any {
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
		data["author"] = map[string]any{"@type": "Organization", "name": entry.AuthorName}
	}
	if !entry.PublishedAt.IsZero() {
		data["datePublished"] = entry.PublishedAt.UTC().Format(time.RFC3339)
	}
	if !entry.UpdatedAt.IsZero() {
		data["dateModified"] = entry.UpdatedAt.UTC().Format(time.RFC3339)
	}
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
	}
}
