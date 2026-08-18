package main

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"rigmark/internal/platform/alerting"
)

type notifierStub struct{ calls atomic.Int32 }

func (stub *notifierStub) Notify(context.Context, alerting.Alert) error {
	stub.calls.Add(1)
	return nil
}
func (stub *notifierStub) Ready(context.Context) error { return nil }

func TestWorkerLoopCancelsInflightCycleGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- runWorkerLoop(ctx, time.Hour, time.Minute, 3, &notifierStub{}, slog.New(slog.NewTextHandler(io.Discard, nil)), func(cycleCtx context.Context) error {
			close(started)
			<-cycleCtx.Done()
			return cycleCtx.Err()
		})
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runWorkerLoop()=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
}

func TestWorkerLoopAlertsOnceAtFailureThreshold(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	notifier := &notifierStub{}
	var cycles atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- runWorkerLoop(ctx, time.Millisecond, time.Second, 2, notifier, slog.New(slog.NewTextHandler(io.Discard, nil)), func(context.Context) error {
			if cycles.Add(1) >= 3 {
				cancel()
			}
			return context.DeadlineExceeded
		})
	}()
	<-done
	if notifier.calls.Load() != 1 {
		t.Fatalf("alerts=%d", notifier.calls.Load())
	}
}
