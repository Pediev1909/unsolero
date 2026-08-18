// Command loadtest is a dependency-free HTTP load probe for reproducible local
// validation. It is intentionally a test utility, not part of the API runtime.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type stringFlags []string

func (values *stringFlags) String() string { return strings.Join(*values, ",") }
func (values *stringFlags) Set(value string) error {
	*values = append(*values, value)
	return nil
}

type options struct {
	scenario                 string
	url                      string
	method                   string
	body                     string
	concurrency              int
	duration                 time.Duration
	requests                 int64
	timeout                  time.Duration
	success                  statusMatcher
	headers                  http.Header
	setupURL                 string
	setupMethod              string
	setupBody                string
	setupExpected            statusMatcher
	transport                http.RoundTripper
	maxP95                   time.Duration
	maxErrorRate             float64
	allowSelfSignedLocalhost bool
}

type sample struct {
	duration time.Duration
	status   int
	err      bool
	bytes    int64
}

type report struct {
	Scenario       string  `json:"scenario"`
	Target         string  `json:"target"`
	Method         string  `json:"method"`
	Concurrency    int     `json:"concurrency"`
	ConfiguredTime string  `json:"configured_duration,omitempty"`
	RequestLimit   int64   `json:"request_limit,omitempty"`
	Elapsed        string  `json:"elapsed"`
	Requests       int     `json:"requests"`
	Successful     int     `json:"successful"`
	Failed         int     `json:"failed"`
	TransportError int     `json:"transport_errors"`
	ErrorRate      float64 `json:"error_rate"`
	RequestsSecond float64 `json:"requests_per_second"`
	ResponseBytes  int64   `json:"response_bytes"`
	P50            string  `json:"p50"`
	P95            string  `json:"p95"`
	P99            string  `json:"p99"`
	Maximum        string  `json:"maximum"`
	p95Duration    time.Duration
	Statuses       map[string]int `json:"statuses"`
}

type statusMatcher struct {
	exact  map[int]struct{}
	ranges [][2]int
}

func parseStatusMatcher(value string) (statusMatcher, error) {
	matcher := statusMatcher{exact: make(map[int]struct{})}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			low, lowErr := strconv.Atoi(bounds[0])
			high, highErr := strconv.Atoi(bounds[1])
			if lowErr != nil || highErr != nil || low < 100 || high > 599 || low > high {
				return statusMatcher{}, fmt.Errorf("invalid status range %q", part)
			}
			matcher.ranges = append(matcher.ranges, [2]int{low, high})
			continue
		}
		status, err := strconv.Atoi(part)
		if err != nil || status < 100 || status > 599 {
			return statusMatcher{}, fmt.Errorf("invalid status %q", part)
		}
		matcher.exact[status] = struct{}{}
	}
	if len(matcher.exact) == 0 && len(matcher.ranges) == 0 {
		return statusMatcher{}, errors.New("at least one expected status is required")
	}
	return matcher, nil
}

func (matcher statusMatcher) matches(status int) bool {
	if _, ok := matcher.exact[status]; ok {
		return true
	}
	for _, bounds := range matcher.ranges {
		if status >= bounds[0] && status <= bounds[1] {
			return true
		}
	}
	return false
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "loadtest:", err)
		os.Exit(1)
	}
}

func run(arguments []string, output, errorOutput io.Writer) error {
	flags := flag.NewFlagSet("loadtest", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	var headers stringFlags
	var bodyFile, setupBodyFile, successValue, setupSuccessValue string
	opts := options{}
	flags.StringVar(&opts.scenario, "scenario", "unnamed", "stable scenario label")
	flags.StringVar(&opts.url, "url", "", "absolute target URL")
	flags.StringVar(&opts.method, "method", http.MethodGet, "HTTP method")
	flags.StringVar(&bodyFile, "body-file", "", "request body file")
	flags.IntVar(&opts.concurrency, "concurrency", 8, "concurrent workers")
	flags.DurationVar(&opts.duration, "duration", 10*time.Second, "test duration (ignored with -requests)")
	flags.Int64Var(&opts.requests, "requests", 0, "exact request count; zero uses duration")
	flags.DurationVar(&opts.timeout, "timeout", 10*time.Second, "per-request timeout")
	flags.StringVar(&successValue, "success", "200-299", "comma-separated successful statuses/ranges")
	flags.Var(&headers, "header", "request header in Name: value form; repeatable")
	flags.StringVar(&opts.setupURL, "setup-url", "", "optional setup URL called once per worker")
	flags.StringVar(&opts.setupMethod, "setup-method", http.MethodPost, "setup HTTP method")
	flags.StringVar(&setupBodyFile, "setup-body-file", "", "setup request body file")
	flags.StringVar(&setupSuccessValue, "setup-success", "200-299", "successful setup statuses/ranges")
	flags.DurationVar(&opts.maxP95, "max-p95", 0, "optional local regression ceiling for p95 latency")
	flags.Float64Var(&opts.maxErrorRate, "max-error-rate", -1, "optional error-rate ceiling between 0 and 1")
	flags.BoolVar(&opts.allowSelfSignedLocalhost, "allow-self-signed-localhost", false, "trust only a localhost staging certificate")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if opts.url == "" || opts.concurrency < 1 || opts.concurrency > 10_000 || opts.timeout <= 0 ||
		(opts.requests <= 0 && opts.duration <= 0) || opts.requests < 0 || opts.maxP95 < 0 || opts.maxErrorRate > 1 {
		return errors.New("url, positive duration/request count, timeout, and concurrency between 1 and 10000 are required")
	}
	if opts.allowSelfSignedLocalhost {
		if !isLocalHTTPS(opts.url) || opts.setupURL != "" && !isLocalHTTPS(opts.setupURL) {
			return errors.New("self-signed certificate bypass is restricted to an HTTPS localhost target")
		}
		opts.transport = &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}} // #nosec G402 -- hostname-restricted local staging only
	}
	var err error
	opts.success, err = parseStatusMatcher(successValue)
	if err != nil {
		return fmt.Errorf("success statuses: %w", err)
	}
	opts.setupExpected, err = parseStatusMatcher(setupSuccessValue)
	if err != nil {
		return fmt.Errorf("setup statuses: %w", err)
	}
	opts.headers, err = parseHeaders(headers)
	if err != nil {
		return err
	}
	if bodyFile != "" {
		body, readErr := os.ReadFile(bodyFile)
		if readErr != nil {
			return fmt.Errorf("read body file: %w", readErr)
		}
		opts.body = string(body)
	}
	if setupBodyFile != "" {
		body, readErr := os.ReadFile(setupBodyFile)
		if readErr != nil {
			return fmt.Errorf("read setup body file: %w", readErr)
		}
		opts.setupBody = string(body)
	}

	result, err := execute(context.Background(), opts)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return err
	}
	return checkRegressionGates(opts, result)
}

func isLocalHTTPS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" {
		return false
	}
	return parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"
}

func checkRegressionGates(opts options, result report) error {
	if opts.maxErrorRate >= 0 && result.ErrorRate > opts.maxErrorRate {
		return fmt.Errorf("error rate %.6f exceeds local regression ceiling %.6f", result.ErrorRate, opts.maxErrorRate)
	}
	if opts.maxP95 > 0 && result.p95Duration > opts.maxP95 {
		return fmt.Errorf("p95 %s exceeds local regression ceiling %s", result.p95Duration, opts.maxP95)
	}
	return nil
}

func execute(parent context.Context, opts options) (report, error) {
	clients := make([]*http.Client, opts.concurrency)
	for index := range clients {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return report{}, fmt.Errorf("cookie jar: %w", err)
		}
		clients[index] = &http.Client{
			Timeout:   opts.timeout,
			Jar:       jar,
			Transport: opts.transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		if opts.setupURL != "" {
			status, err := send(parent, clients[index], opts.setupMethod, opts.setupURL, opts.setupBody, opts.headers, int64(index+1))
			if err != nil {
				return report{}, fmt.Errorf("worker %d setup failed: %w", index+1, err)
			}
			if !opts.setupExpected.matches(status.status) {
				return report{}, fmt.Errorf("worker %d setup returned HTTP %d", index+1, status.status)
			}
		}
	}

	ctx := parent
	cancel := func() {}
	if opts.requests == 0 {
		ctx, cancel = context.WithTimeout(parent, opts.duration)
	}
	defer cancel()
	results := make(chan sample, opts.concurrency*2)
	var sequence atomic.Int64
	started := time.Now()
	var workers sync.WaitGroup
	for worker := range clients {
		workers.Add(1)
		go func(client *http.Client) {
			defer workers.Done()
			for {
				current := sequence.Add(1)
				if opts.requests > 0 && current > opts.requests {
					return
				}
				if err := ctx.Err(); err != nil {
					return
				}
				startedRequest := time.Now()
				result, err := send(ctx, client, opts.method, opts.url, opts.body, opts.headers, current)
				// A duration-based run intentionally cancels requests still in flight
				// at its boundary. They were not completed samples and must not be
				// reported as transport failures. Request-count runs and every other
				// network error remain observable failures.
				if intentionalDurationBoundary(opts.requests, ctx.Err(), err) {
					return
				}
				result.duration = time.Since(startedRequest)
				result.err = err != nil
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			}
		}(clients[worker])
	}
	go func() {
		workers.Wait()
		close(results)
	}()
	samples := make([]sample, 0, 1024)
	for result := range results {
		samples = append(samples, result)
	}
	elapsed := time.Since(started)
	return buildReport(opts, samples, elapsed), nil
}

func intentionalDurationBoundary(requestLimit int64, runError, requestError error) bool {
	return requestLimit == 0 && errors.Is(runError, context.DeadlineExceeded) &&
		errors.Is(requestError, context.DeadlineExceeded)
}

func send(ctx context.Context, client *http.Client, method, rawURL, body string, headers http.Header, sequence int64) (sample, error) {
	resolvedURL, resolvedBody, err := applyTemplates(rawURL, body, sequence)
	if err != nil {
		return sample{}, err
	}
	request, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), resolvedURL, bytes.NewBufferString(resolvedBody))
	if err != nil {
		return sample{}, err
	}
	request.Header = headers.Clone()
	if resolvedBody != "" && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return sample{}, err
	}
	defer response.Body.Close()
	written, copyErr := io.Copy(io.Discard, io.LimitReader(response.Body, 8<<20))
	return sample{status: response.StatusCode, bytes: written}, copyErr
}

func applyTemplates(rawURL, body string, sequence int64) (string, string, error) {
	replacement := strconv.FormatInt(sequence, 10)
	rawURL = strings.ReplaceAll(rawURL, "{{SEQ}}", replacement)
	body = strings.ReplaceAll(body, "{{SEQ}}", replacement)
	if strings.Contains(rawURL, "{{UUID}}") || strings.Contains(body, "{{UUID}}") {
		identifier := make([]byte, 16)
		if _, err := rand.Read(identifier); err != nil {
			return "", "", fmt.Errorf("generate identifier: %w", err)
		}
		identifier[6] = (identifier[6] & 0x0f) | 0x40
		identifier[8] = (identifier[8] & 0x3f) | 0x80
		raw := hex.EncodeToString(identifier)
		value := raw[0:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:32]
		rawURL = strings.ReplaceAll(rawURL, "{{UUID}}", value)
		body = strings.ReplaceAll(body, "{{UUID}}", value)
	}
	return rawURL, body, nil
}

func parseHeaders(values []string) (http.Header, error) {
	headers := make(http.Header)
	for _, value := range values {
		name, item, ok := strings.Cut(value, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("invalid header %q", value)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(item))
	}
	return headers, nil
}

func buildReport(opts options, samples []sample, elapsed time.Duration) report {
	result := report{Scenario: opts.scenario, Target: opts.url, Method: strings.ToUpper(opts.method),
		Concurrency: opts.concurrency, RequestLimit: opts.requests, Elapsed: elapsed.String(),
		Requests: len(samples), Statuses: make(map[string]int)}
	if opts.requests == 0 {
		result.ConfiguredTime = opts.duration.String()
	}
	latencies := make([]time.Duration, 0, len(samples))
	for _, item := range samples {
		latencies = append(latencies, item.duration)
		result.ResponseBytes += item.bytes
		if item.err {
			result.TransportError++
			result.Failed++
			continue
		}
		result.Statuses[strconv.Itoa(item.status)]++
		if opts.success.matches(item.status) {
			result.Successful++
		} else {
			result.Failed++
		}
	}
	if result.Requests > 0 {
		result.ErrorRate = float64(result.Failed) / float64(result.Requests)
	}
	if elapsed > 0 {
		result.RequestsSecond = float64(result.Requests) / elapsed.Seconds()
	}
	sort.Slice(latencies, func(left, right int) bool { return latencies[left] < latencies[right] })
	result.P50 = percentile(latencies, 50).String()
	result.P95 = percentile(latencies, 95).String()
	result.p95Duration = percentile(latencies, 95)
	result.P99 = percentile(latencies, 99).String()
	result.Maximum = percentile(latencies, 100).String()
	return result
}

func percentile(sorted []time.Duration, percentage int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := (percentage*len(sorted) + 99) / 100
	if index < 1 {
		index = 1
	}
	return sorted[index-1]
}
