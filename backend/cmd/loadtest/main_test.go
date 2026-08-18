package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func testResponse(request *http.Request, status int, headers http.Header) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: headers,
		Body: io.NopCloser(strings.NewReader("response")), Request: request}
}

func TestExecuteCountsStatusesAndDoesNotFollowRedirects(t *testing.T) {
	var requests atomic.Int64
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Path == "/redirect" {
			return testResponse(request, http.StatusFound, http.Header{"Location": []string{"/destination"}}), nil
		}
		return testResponse(request, http.StatusTeapot, make(http.Header)), nil
	})

	matcher, _ := parseStatusMatcher("302")
	setupMatcher, _ := parseStatusMatcher("200-299")
	report, err := execute(context.Background(), options{scenario: "redirect", url: "https://example.test/redirect",
		method: http.MethodGet, concurrency: 4, requests: 20, timeout: time.Second,
		success: matcher, setupExpected: setupMatcher, headers: make(http.Header), transport: transport})
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if report.Requests != 20 || report.Successful != 20 || report.Failed != 0 || report.Statuses["302"] != 20 {
		t.Fatalf("report = %#v", report)
	}
	if requests.Load() != 20 {
		t.Fatalf("server requests = %d, want 20", requests.Load())
	}
}

func TestSetupPreservesCookiesPerWorker(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/setup" {
			headers := make(http.Header)
			headers.Add("Set-Cookie", "session=valid; Path=/; HttpOnly")
			return testResponse(request, http.StatusNoContent, headers), nil
		}
		cookie, err := request.Cookie("session")
		if err != nil || cookie.Value != "valid" {
			return testResponse(request, http.StatusUnauthorized, make(http.Header)), nil
		}
		return testResponse(request, http.StatusOK, make(http.Header)), nil
	})

	matcher, _ := parseStatusMatcher("200")
	setupMatcher, _ := parseStatusMatcher("204")
	report, err := execute(context.Background(), options{scenario: "authenticated", url: "https://example.test/target",
		method: http.MethodGet, concurrency: 3, requests: 12, timeout: time.Second, success: matcher,
		setupURL: "https://example.test/setup", setupMethod: http.MethodPost, setupExpected: setupMatcher,
		headers: make(http.Header), transport: transport})
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if report.Successful != 12 {
		t.Fatalf("report = %#v", report)
	}
}

func TestTemplatesProduceUniqueReservedValues(t *testing.T) {
	url, body, err := applyTemplates("https://example.test/{{SEQ}}/{{UUID}}", `{"email":"phase8-{{SEQ}}@example.invalid","id":"{{UUID}}"}`, 42)
	if err != nil {
		t.Fatalf("applyTemplates() error = %v", err)
	}
	if !strings.Contains(url, "/42/") || !strings.Contains(body, "phase8-42@example.invalid") || strings.Contains(url+body, "{{") {
		t.Fatalf("unresolved templates: url=%q body=%q", url, body)
	}
}

func TestStatusMatcherRejectsInvalidRanges(t *testing.T) {
	for _, value := range []string{"", "99", "600", "300-200", "abc"} {
		if _, err := parseStatusMatcher(value); err == nil {
			t.Fatalf("parseStatusMatcher(%q) unexpectedly succeeded", value)
		}
	}
}

func TestRegressionGateRejectsExcessiveErrorsAndLatency(t *testing.T) {
	result := report{ErrorRate: 0.5, p95Duration: 2 * time.Second}
	if err := checkRegressionGates(options{maxErrorRate: 0, maxP95: time.Second}, result); err == nil {
		t.Fatal("regression gate accepted excessive error rate and latency")
	}
	if err := checkRegressionGates(options{maxErrorRate: 0.5, maxP95: 2 * time.Second}, result); err != nil {
		t.Fatalf("regression gate rejected boundary: %v", err)
	}
}

func TestIntentionalDurationBoundaryOnlyIgnoresItsOwnDeadline(t *testing.T) {
	if !intentionalDurationBoundary(0, context.DeadlineExceeded, context.DeadlineExceeded) {
		t.Fatal("duration deadline was not recognized")
	}
	if intentionalDurationBoundary(10, context.DeadlineExceeded, context.DeadlineExceeded) ||
		intentionalDurationBoundary(0, nil, context.DeadlineExceeded) ||
		intentionalDurationBoundary(0, context.DeadlineExceeded, errors.New("connection reset")) {
		t.Fatal("non-boundary failure was incorrectly ignored")
	}
}
