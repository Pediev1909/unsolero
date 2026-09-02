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
