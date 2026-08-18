package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSource combines process-local pool statistics with durable queue and
// checkpoint state. Its keys are compile-time bounded by allowedGauges.
type PostgresSource struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewPostgresSource(pool *pgxpool.Pool) *PostgresSource {
	return &PostgresSource{pool: pool, now: time.Now}
}

func RecordWorkerCheckpoint(ctx context.Context, pool *pgxpool.Pool, succeeded bool) error {
	if succeeded {
		_, err := pool.Exec(ctx, `INSERT INTO platform.operational_checkpoints
			(checkpoint_name,status,observed_at,last_success_at,failure_count,detail_code)
			VALUES ('worker','ok',now(),now(),0,NULL)
			ON CONFLICT (checkpoint_name) DO UPDATE SET status='ok',observed_at=now(),
			last_success_at=now(),failure_count=0,detail_code=NULL,updated_at=now()`)
		return err
	}
	_, err := pool.Exec(ctx, `INSERT INTO platform.operational_checkpoints
		(checkpoint_name,status,observed_at,failure_count,detail_code)
		VALUES ('worker','failed',now(),1,'worker.cycle_failed')
		ON CONFLICT (checkpoint_name) DO UPDATE SET status='failed',observed_at=now(),
		failure_count=platform.operational_checkpoints.failure_count+1,
		detail_code='worker.cycle_failed',updated_at=now()`)
	return err
}

func (source *PostgresSource) Collect(ctx context.Context) (map[string]float64, error) {
	statistics := source.pool.Stat()
	metrics := map[string]float64{
		"database_pool_acquired":               float64(statistics.AcquiredConns()),
		"database_pool_idle":                   float64(statistics.IdleConns()),
		"database_pool_total":                  float64(statistics.TotalConns()),
		"database_pool_max":                    float64(statistics.MaxConns()),
		"database_pool_wait_count":             float64(statistics.EmptyAcquireCount()),
		"database_pool_wait_seconds_total":     statistics.AcquireDuration().Seconds(),
		"database_pool_canceled_acquire_count": float64(statistics.CanceledAcquireCount()),
	}
	var active, backlog, successful, failed, retries, dead, leaseRecoveries int64
	var processingLatencySeconds float64
	var pendingMedia, mediaRetries, deadMedia, discrepancies int64
	err := source.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM commerce.offer_import_runs WHERE status='running') +
		(SELECT count(*) FROM commerce.conversion_import_runs WHERE status='running') +
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='processing'),
		(SELECT count(*) FROM commerce.offer_import_runs WHERE status IN ('queued','retry_wait')) +
		(SELECT count(*) FROM commerce.conversion_import_runs WHERE status IN ('queued','retry_wait')) +
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='pending'),
		(SELECT count(*) FROM commerce.offer_import_runs WHERE status='succeeded') +
		(SELECT count(*) FROM commerce.conversion_import_runs WHERE status='succeeded') +
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='completed'),
		(SELECT count(*) FROM commerce.offer_import_runs WHERE status IN ('failed','partial')) +
		(SELECT count(*) FROM commerce.conversion_import_runs WHERE status IN ('failed','partial')) +
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='dead'),
		(SELECT COALESCE(sum(GREATEST(attempt_count - 1, 0)),0) FROM commerce.offer_import_runs) +
		(SELECT COALESCE(sum(GREATEST(attempt_count - 1, 0)),0) FROM commerce.conversion_import_runs) +
		(SELECT COALESCE(sum(GREATEST(attempt_count - 1, 0)),0) FROM admin.media_deletion_jobs),
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='dead'),
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE last_error_code='worker.lease_expired'),
		COALESCE((SELECT avg(EXTRACT(epoch FROM (completed_at-started_at)))
			FROM (SELECT started_at,completed_at FROM commerce.offer_import_runs
				WHERE completed_at IS NOT NULL AND started_at IS NOT NULL
				UNION ALL SELECT started_at,completed_at FROM commerce.conversion_import_runs
				WHERE completed_at IS NOT NULL AND started_at IS NOT NULL) completed),0),
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='pending'),
		(SELECT COALESCE(sum(GREATEST(attempt_count - 1, 0)),0) FROM admin.media_deletion_jobs),
		(SELECT count(*) FROM admin.media_deletion_jobs WHERE status='dead'),
		COALESCE((SELECT discrepancy_count FROM admin.media_reconciliation_runs
			WHERE status='completed' ORDER BY completed_at DESC,id DESC LIMIT 1),0)`).Scan(
		&active, &backlog, &successful, &failed, &retries, &dead, &leaseRecoveries,
		&processingLatencySeconds, &pendingMedia, &mediaRetries, &deadMedia, &discrepancies)
	if err != nil {
		return metrics, fmt.Errorf("collect durable operational metrics: %w", err)
	}
	metrics["worker_active_jobs"] = float64(active)
	metrics["worker_backlog_depth"] = float64(backlog)
	metrics["worker_successful_jobs"] = float64(successful)
	metrics["worker_failed_jobs"] = float64(failed)
	metrics["worker_retry_count"] = float64(retries)
	metrics["worker_dead_jobs"] = float64(dead)
	metrics["worker_lease_recovery_count"] = float64(leaseRecoveries)
	metrics["worker_processing_latency_seconds"] = processingLatencySeconds
	metrics["media_pending_deletions"] = float64(pendingMedia)
	metrics["media_deletion_retry_count"] = float64(mediaRetries)
	metrics["media_dead_deletion_jobs"] = float64(deadMedia)
	metrics["media_reconciliation_discrepancies"] = float64(discrepancies)

	rows, checkpointErr := source.pool.Query(ctx, `SELECT checkpoint_name,status,last_success_at,failure_count
		FROM platform.operational_checkpoints WHERE checkpoint_name IN ('backup','restore_verification','worker')`)
	if checkpointErr != nil {
		return metrics, fmt.Errorf("collect operational checkpoints: %w", checkpointErr)
	}
	defer rows.Close()
	for rows.Next() {
		var name, status string
		var lastSuccess *time.Time
		var failureCount int64
		if err := rows.Scan(&name, &status, &lastSuccess, &failureCount); err != nil {
			return metrics, fmt.Errorf("scan operational checkpoint: %w", err)
		}
		switch name {
		case "backup":
			metrics["backup_failure_count"] = float64(failureCount)
			if lastSuccess != nil {
				metrics["backup_last_success_timestamp"] = float64(lastSuccess.Unix())
				metrics["backup_age_seconds"] = max(0, source.now().UTC().Sub(*lastSuccess).Seconds())
			}
		case "restore_verification":
			if status == "ok" {
				metrics["backup_restore_verified"] = 1
			}
			if status == "mismatch" {
				metrics["backup_migration_fingerprint_mismatch"] = 1
			}
		case "worker":
			metrics["worker_heartbeat_failure_count"] = float64(failureCount)
			if lastSuccess != nil {
				metrics["worker_last_success_timestamp"] = float64(lastSuccess.Unix())
				metrics["worker_heartbeat_age_seconds"] = max(0, source.now().UTC().Sub(*lastSuccess).Seconds())
			}
		}
	}
	return metrics, rows.Err()
}
