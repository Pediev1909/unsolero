package httpapi

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"time"

	content "rigmark/internal/modules/content/application"
	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

type contentSummaryResponse struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	HeroImage   imageResponse `json:"hero_image"`
	AuthorName  string        `json:"author_name"`
	PublishedAt time.Time     `json:"published_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	// Covered names the card after the products the piece compares, so a grid
	// of comparisons stops looking like one comparison repeated.
	Covered []coveredProductResponse `json:"covered"`
}

type coveredProductResponse struct {
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type contentBlockResponse struct {
	Type        string   `json:"type"`
	Heading     string   `json:"heading,omitempty"`
	Text        string   `json:"text,omitempty"`
	Items       []string `json:"items,omitempty"`
	Attribution string   `json:"attribution,omitempty"`
	// A cta block is only a heading and a paragraph without these two. The
	// first version of this struct omitted them, and the result was a call to
	// action that appeared in the prerendered body for crawlers and readers
	// without JavaScript, and rendered as text with no button for everybody
	// else — the entire audience that can actually click it.
	Promotion string `json:"promotion,omitempty"`
	Label     string `json:"label,omitempty"`
}

type contentAuthorResponse struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Bio       string  `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type contentSEOResponse struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	CanonicalURL string `json:"canonical_url"`
}

type contentDetailResponse struct {
	contentSummaryResponse
	Content           []contentBlockResponse   `json:"content"`
	Author            contentAuthorResponse    `json:"author"`
	RelatedProducts   []productSummaryResponse `json:"related_products"`
	RelatedCategories []categoryResponse       `json:"related_categories"`
	RelatedContent    []contentSummaryResponse `json:"related_content"`
	SEO               contentSEOResponse       `json:"seo"`
}

type contentAuthorPageResponse struct {
	Author  contentAuthorResponse    `json:"author"`
	Entries []contentSummaryResponse `json:"entries"`
}

// getAuthor serves the page behind a byline. Attribution to a named person is
// the signal a reader and a search engine both look for, and it is worth
// nothing if the name does not lead anywhere.
func (h *Handler) getAuthor(response http.ResponseWriter, request *http.Request) {
	author, entries, err := h.content.Author(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeContentError(response, err)
		return
	}
	result := contentAuthorPageResponse{
		Author: contentAuthorResponse{
			Name: author.Name, Slug: author.Slug, Bio: author.Bio, AvatarURL: author.AvatarURL,
		},
		Entries: make([]contentSummaryResponse, 0, len(entries)),
	}
	for _, entry := range entries {
		result.Entries = append(result.Entries, contentSummaryDTO(entry))
	}
	writeJSON(response, http.StatusOK, result, h.logger)
}

func (h *Handler) listContent(response http.ResponseWriter, request *http.Request) {
	limit := 12
	if raw := request.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			h.writeContentError(response, content.ErrInvalidQuery)
			return
		}
		limit = parsed
	}
	entries, err := h.content.List(request.Context(), content.ListQuery{
		Section:      request.URL.Query().Get("section"),
		CategorySlug: request.URL.Query().Get("category"),
		ProductSlug:  request.URL.Query().Get("product"),
		Limit:        limit,
	})
	if err != nil {
		h.writeContentError(response, err)
		return
	}
	result := make([]contentSummaryResponse, 0, len(entries))
	for _, entry := range entries {
		result = append(result, contentSummaryDTO(entry))
	}
	response.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=600")
	writeJSON(response, http.StatusOK, result, h.logger)
}

func (h *Handler) getContent(response http.ResponseWriter, request *http.Request) {
	entry, err := h.content.Get(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeContentError(response, err)
		return
	}
	detail := contentDetailDTO(entry)
	// The "Products referenced" grid on an alternatives or versus page is the
	// closest a reader gets to a decision anywhere on this site. The slice is
	// filled in place, so this has to run before the response is written and
	// cannot live in contentDetailDTO, which has neither the handler nor the
	// request.
	h.attachPurchasePaths(request.Context(), detail.RelatedProducts)
	response.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=600")
	writeJSON(response, http.StatusOK, detail, h.logger)
}

func contentSummaryDTO(entry domain.Summary) contentSummaryResponse {
	covered := make([]coveredProductResponse, 0, len(entry.Covered))
	for _, product := range entry.Covered {
		covered = append(covered, coveredProductResponse{
			Name: product.Name, PriceMinor: product.PriceMinor, Currency: product.Currency,
		})
	}
	return contentSummaryResponse{
		ID: entry.ID, Type: string(entry.Type), Title: entry.Title, Slug: entry.Slug,
		Path: entry.Path, Description: entry.Description,
		HeroImage:  imageResponse{URL: entry.HeroImageURL, AltText: entry.HeroImageAlt},
		AuthorName: entry.AuthorName, PublishedAt: entry.PublishedAt, UpdatedAt: entry.UpdatedAt,
		Covered: covered,
	}
}

func contentDetailDTO(entry domain.Entry) contentDetailResponse {
	blocks := make([]contentBlockResponse, 0, len(entry.Content))
	for _, block := range entry.Content {
		blocks = append(blocks, contentBlockResponse{
			Type: string(block.Type), Heading: block.Heading, Text: block.Text,
			Items: block.Items, Attribution: block.Attribution,
			Promotion: block.Promotion, Label: block.Label,
		})
	}
	products := make([]productSummaryResponse, 0, len(entry.RelatedProducts))
	for _, product := range entry.RelatedProducts {
		products = append(products, productSummaryDTO(product))
	}
	categories := make([]categoryResponse, 0, len(entry.RelatedCategories))
	for _, category := range entry.RelatedCategories {
		categories = append(categories, categoryResponse{
			ID: string(category.ID), Name: category.Name, Slug: category.Slug, Description: category.Description,
		})
	}
	related := make([]contentSummaryResponse, 0, len(entry.RelatedEntries))
	for _, item := range entry.RelatedEntries {
		related = append(related, contentSummaryDTO(item))
	}
	return contentDetailResponse{
		contentSummaryResponse: contentSummaryDTO(entry.Summary),
		Content:                blocks,
		Author:                 contentAuthorResponse{Name: entry.Author.Name, Slug: entry.Author.Slug, Bio: entry.Author.Bio, AvatarURL: entry.Author.AvatarURL},
		RelatedProducts:        products, RelatedCategories: categories, RelatedContent: related,
		SEO: contentSEOResponse{Title: entry.SEOTitle, Description: entry.SEODescription, CanonicalURL: entry.CanonicalURL},
	}
}

func (h *Handler) writeContentError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, content.ErrInvalidQuery):
		writeAPIError(response, http.StatusBadRequest, "invalid_content_query", "The content query is invalid.", nil, h.logger)
	case errors.Is(err, ports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "content_not_found", "The requested content could not be found.", nil, h.logger)
	default:
		h.logger.Error("editorial content request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "content_unavailable", "Editorial content is temporarily unavailable.", nil, h.logger)
	}
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location string `xml:"loc"`
	LastMod  string `xml:"lastmod,omitempty"`
}

func (h *Handler) sitemap(response http.ResponseWriter, request *http.Request) {
	entries, err := h.content.Sitemap(request.Context())
	if err != nil {
		h.logger.Error("generate sitemap", "error", err)
		http.Error(response, "Sitemap unavailable", http.StatusInternalServerError)
		return
	}
	result := sitemapURLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: make([]sitemapURL, 0, len(entries))}
	for _, entry := range entries {
		result.URLs = append(result.URLs, sitemapURL{
			Location: h.content.AbsoluteURL(entry.Path), LastMod: entry.ModifiedAt.UTC().Format("2006-01-02"),
		})
	}
	data, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		h.logger.Error("encode sitemap", "error", err)
		http.Error(response, "Sitemap unavailable", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	response.Header().Set("Cache-Control", "public, max-age=3600")
	response.WriteHeader(http.StatusOK)
	// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- encoding/xml escapes every dynamic sitemap value and the response is XML, not HTML.
	_, _ = response.Write(append([]byte(xml.Header), data...))
}

// llmsTxt describes the site for assistants that read it before crawling. The
// emerging convention is a short map of what a site is and where its useful
// pages are, in plain prose rather than markup.
//
// This is worth writing because software recommendations are increasingly
// asked of assistants rather than typed into a search box, and the sites those
// assistants cite become the answer. Being legible to them is the same
// distribution problem as ranking, one channel later.
func (h *Handler) llmsTxt(response http.ResponseWriter, _ *http.Request) {
	base := h.content.AbsoluteURL("")
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Header().Set("Cache-Control", "public, max-age=3600")
	// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- deliberate plain-text response; AbsoluteURL is configuration-validated.
	_, _ = response.Write([]byte(`# UNSOLERO

> An independent decision engine for business software. It turns a company's
> goal, budget and existing tools into an explainable stack recommendation.
> Commission never affects which products are recommended or in what order.

## What makes the data here citable

- Every product fact carries a recorded source and the date it was read.
- Prices are the entry paid tier at monthly billing, taken from the vendor's
  own pricing page. Promotional rates are excluded.
- Suitability scores are editorial judgements and are labelled as such,
  attributed separately from vendor facts.
- Recommendations are computed by a deterministic engine. The same inputs
  always produce the same output, and commercial data is not an input.

## Key pages

- [Software catalog](` + base + `/products): structured facts and prices per product
- [Guides](` + base + `/guides): how to plan a stack around real constraints
- [Articles](` + base + `/articles): editorial notes on choosing and combining tools
- [How we rank software](` + base + `/articles/how-unsolero-ranks-software): the method, in full
- [Affiliate disclosure](` + base + `/affiliate-disclosure): how the site earns money
- [Sitemap](` + base + `/sitemap.xml)

## Caveats worth repeating

Software pricing changes often. Each price on this site records when it was
read; confirm the current figure with the vendor before relying on it.
`))
}

func (h *Handler) robots(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Header().Set("Cache-Control", "public, max-age=3600")
	// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- deliberate plain-text robots response; AbsoluteURL is configuration-validated.
	_, _ = response.Write([]byte("User-agent: *\nAllow: /\nDisallow: /api/\nDisallow: /admin/\nDisallow: /account\nDisallow: /setups\nDisallow: /wishlist\nDisallow: /build\nDisallow: /login\nDisallow: /register\nDisallow: /design-system\nSitemap: " + h.content.AbsoluteURL("/sitemap.xml") + "\n"))
}
