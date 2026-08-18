package observability

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWorkerCheckpointIsDurableAndExported(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM platform.operational_checkpoints WHERE checkpoint_name='worker'`)
	})

	if err := RecordWorkerCheckpoint(ctx, pool, true); err != nil {
		t.Fatal(err)
	}
	if err := RecordWorkerCheckpoint(ctx, pool, false); err != nil {
		t.Fatal(err)
	}
	metrics, err := NewPostgresSource(pool).Collect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if metrics["worker_heartbeat_failure_count"] != 1 || metrics["worker_last_success_timestamp"] <= 0 ||
		metrics["worker_heartbeat_age_seconds"] < 0 {
		t.Fatalf("worker metrics=%v", metrics)
	}
}
