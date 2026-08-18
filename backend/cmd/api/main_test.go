package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"rigmark/internal/platform/config"
)

func TestServeHTTPDrainsInflightRequestOnCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		response.WriteHeader(http.StatusNoContent)
	})}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local sockets unavailable in this test environment: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serveHTTP(ctx, server, listener, time.Second) }()

	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			err = response.Body.Close()
		}
		requestDone <- err
	}()
	<-started
	cancel()
	close(release)
	if err := <-requestDone; err != nil {
		t.Fatalf("in-flight request failed: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("serveHTTP()=%v", err)
	}
}

func TestExternalRateLimiterFailsClosedWithoutAdapter(t *testing.T) {
	if _, err := selectRateLimiter(config.RateLimits{Provider: "external"}); err == nil {
		t.Fatal("external rate limiter unexpectedly configured")
	}
}
