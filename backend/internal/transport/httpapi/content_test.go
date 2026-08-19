package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rigmark/internal/modules/content/domain"
)

type contentStub struct{}

func (contentStub) List(context.Context, string, string, int) ([]domain.Summary, error) {
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
