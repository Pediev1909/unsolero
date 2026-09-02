package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	content "rigmark/internal/modules/content/application"
	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

type contentStub struct{}

func (contentStub) List(context.Context, content.ListQuery) ([]domain.Summary, error) {
	return []domain.Summary{}, nil
}
func (contentStub) Get(context.Context, string) (domain.Entry, error) {
	return domain.Entry{}, nil
}
func (contentStub) Author(context.Context, string) (domain.Author, []domain.Summary, error) {
	return domain.Author{}, nil, nil
}

func (contentStub) Sitemap(context.Context) ([]domain.SitemapEntry, error) {
	return []domain.SitemapEntry{{Path: "/guides/example", ModifiedAt: time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)}}, nil
}
func (contentStub) AbsoluteURL(path string) string {
	return "https://rigmark.example" + path
}

func TestSitemapAndRobotsExposePublicEditorialDiscovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig, logger, PublicServices{Content: contentStub{}})

	sitemapRequest := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	sitemapResponse := httptest.NewRecorder()
	router.ServeHTTP(sitemapResponse, sitemapRequest)
	if sitemapResponse.Code != http.StatusOK || !strings.Contains(sitemapResponse.Body.String(), "https://rigmark.example/guides/example") {
		t.Fatalf("sitemap response = %d %q", sitemapResponse.Code, sitemapResponse.Body.String())
	}
	if sitemapResponse.Header().Get("Content-Type") != "application/xml; charset=utf-8" {
		t.Fatalf("sitemap content type = %q", sitemapResponse.Header().Get("Content-Type"))
	}

	robotsRequest := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	robotsResponse := httptest.NewRecorder()
	router.ServeHTTP(robotsResponse, robotsRequest)
	robots := robotsResponse.Body.String()
	if robotsResponse.Code != http.StatusOK || !strings.Contains(robots, "Sitemap: https://rigmark.example/sitemap.xml") ||
		!strings.Contains(robots, "Disallow: /admin/") {
		t.Fatalf("robots response = %d %q", robotsResponse.Code, robots)
	}
}

// A cta block that reaches the client without its promotion and label is a
// heading and a paragraph. That shipped once: the link rendered in the
// prerendered body for crawlers and readers without JavaScript, and rendered
// as text with no button for everyone who could actually click it.
func TestContentDetailCarriesCTAPromotionAndLabel(t *testing.T) {
	entry := domain.Entry{Content: []domain.Block{{
		Type:      domain.BlockCTA,
		Heading:   "If automation is why you are leaving",
		Text:      "Their own comparison is the honest place to start.",
		Label:     "See ActiveCampaign against Mailchimp",
		Promotion: "activecampaign-mailchimp-switch",
	}}}

	encoded, err := json.Marshal(contentDetailDTO(entry))
	if err != nil {
		t.Fatalf("marshal content detail: %v", err)
	}
	body := string(encoded)
	for _, want := range []string{
		`"promotion":"activecampaign-mailchimp-switch"`,
		`"label":"See ActiveCampaign against Mailchimp"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("content response is missing %s:\n%s", want, body)
		}
	}
}

// contentRepositoryStub sits under the real application service, so a handler
// test exercises the validation the service does rather than a stub's idea of
// it. It records the filter it was asked for.
type contentRepositoryStub struct {
	filter ports.Filter
}

func (stub *contentRepositoryStub) ListPublished(_ context.Context, filter ports.Filter) ([]domain.Summary, error) {
	stub.filter = filter
	return []domain.Summary{}, nil
}
func (*contentRepositoryStub) GetAuthorBySlug(context.Context, string) (domain.Author, error) {
	return domain.Author{}, ports.ErrNotFound
}
func (*contentRepositoryStub) GetPublishedBySlug(context.Context, string) (domain.Entry, error) {
	return domain.Entry{}, ports.ErrNotFound
}
func (*contentRepositoryStub) ListSitemapEntries(context.Context) ([]domain.SitemapEntry, error) {
	return nil, nil
}

// The product filter is how a product page finds the comparisons it appears
// in. The slug arrives from the URL, so a well-formed one reaches the
// repository as given and anything else is refused with a 400 before SQL.
func TestListContentFiltersByProductSlugAndRejectsGarbage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repository := &contentRepositoryStub{}
	service, err := content.NewService(repository, nil, "https://rigmark.example")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig, logger, PublicServices{Content: service})

	request := httptest.NewRequest(http.MethodGet, "/api/content?product=mailchimp-standard&limit=6", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("product filter response = %d %q", response.Code, response.Body.String())
	}
	if repository.filter.ProductSlug != "mailchimp-standard" || repository.filter.Limit != 6 {
		t.Fatalf("repository filter = %#v", repository.filter)
	}
	if body := strings.TrimSpace(response.Body.String()); body != "[]" {
		t.Fatalf("empty listing body = %q, want []", body)
	}

	for _, raw := range []string{"Mailchimp%20Standard", "mailchimp_standard", "x%27%20OR%201%3D1", "-leading"} {
		repository.filter = ports.Filter{}
		request := httptest.NewRequest(http.MethodGet, "/api/content?product="+raw, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("product=%s response = %d %q, want 400", raw, response.Code, response.Body.String())
		}
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error.Code != "invalid_content_query" {
			t.Fatalf("product=%s error body = %q (%v)", raw, response.Body.String(), err)
		}
		if repository.filter.ProductSlug != "" {
			t.Fatalf("product=%s reached the repository as %q", raw, repository.filter.ProductSlug)
		}
	}
}

// Every field of every block type has to survive the trip to JSON. Two block
// types shipped without theirs — cta once, then offer, faq and pros_cons on
// 2026-09-02 — and both times the data was stored correctly, the crawler body
// printed it, and the reader running the application got a heading with
// nothing under it. This test fails the build rather than the page.
func TestContentDetailCarriesEveryBlockField(t *testing.T) {
	entry := domain.Entry{
		Content: []domain.Block{
			{Type: domain.BlockCTA, Heading: "H", Text: "T", Promotion: "a-promo", Label: "Go"},
			{Type: domain.BlockOffer, Heading: "Where to get them", Text: "T", Product: "kit-creator", Label: "View at Kit"},
			{Type: domain.BlockProsCons, Heading: "P", Pros: []string{"Cheap"}, Cons: []string{"Thin"}},
			{Type: domain.BlockFAQ, Heading: "Questions people ask", Questions: []domain.QuestionAnswer{{Question: "Is it cheaper?", Answer: "At 1,000 contacts, yes."}}},
			{Type: domain.BlockParagraph, Text: "Plain."},
		},
	}

	encoded, err := json.Marshal(contentDetailDTO(entry).Content)
	if err != nil {
		t.Fatalf("marshal blocks: %v", err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode blocks: %v", err)
	}
	if len(decoded) != 5 {
		t.Fatalf("blocks = %d, want 5", len(decoded))
	}

	if decoded[0]["promotion"] != "a-promo" || decoded[0]["label"] != "Go" {
		t.Errorf("cta lost its destination: %v", decoded[0])
	}
	if decoded[1]["product"] != "kit-creator" || decoded[1]["label"] != "View at Kit" {
		t.Errorf("offer lost the product it points at: %v", decoded[1])
	}
	pros, _ := decoded[2]["pros"].([]any)
	cons, _ := decoded[2]["cons"].([]any)
	if len(pros) != 1 || len(cons) != 1 {
		t.Errorf("pros_cons lost a side: %v", decoded[2])
	}
	questions, _ := decoded[3]["questions"].([]any)
	if len(questions) != 1 {
		t.Fatalf("faq lost its questions: %v", decoded[3])
	}
	first, _ := questions[0].(map[string]any)
	if first["question"] != "Is it cheaper?" || first["answer"] != "At 1,000 contacts, yes." {
		t.Errorf("faq question round-tripped wrong: %v", first)
	}

	// A paragraph must not gain empty collections just because the fields exist.
	for _, key := range []string{"pros", "cons", "questions", "product", "promotion", "label"} {
		if _, present := decoded[4][key]; present {
			t.Errorf("paragraph carries %q: %v", key, decoded[4])
		}
	}
}
