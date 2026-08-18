package adminpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

const productMediaURLPrefix = "/api/media/products/"

func (repository *Repository) ScheduleMediaDeletion(ctx context.Context, productID catalog.ProductID, objectName string) error {
	_, err := repository.pool.Exec(ctx, `INSERT INTO admin.media_deletion_jobs (product_id,object_name)
		VALUES ($1,$2) ON CONFLICT (object_name) DO UPDATE SET status='pending',attempt_count=0,
		next_attempt_at=now(),last_error_code=NULL,updated_at=now(),completed_at=NULL`, productID, objectName)
	if err != nil {
		return fmt.Errorf("schedule media deletion: %w", err)
	}
	return nil
}

func (repository *Repository) ClaimMediaDeletions(ctx context.Context, limit int, now time.Time) ([]ports.MediaDeletion, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `UPDATE admin.media_deletion_jobs SET status='pending',next_attempt_at=$1,updated_at=$1,
		last_error_code='worker.lease_expired' WHERE status='processing'
		AND updated_at < ($1::timestamptz - interval '10 minutes')`, now); err != nil {
		return nil, fmt.Errorf("recover media deletion jobs: %w", err)
	}
	rows, err := tx.Query(ctx, `WITH claimed AS (
		SELECT id FROM admin.media_deletion_jobs WHERE status='pending' AND next_attempt_at<=$1
		ORDER BY next_attempt_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	) UPDATE admin.media_deletion_jobs jobs SET status='processing',attempt_count=attempt_count+1,updated_at=$1
	FROM claimed WHERE jobs.id=claimed.id
	RETURNING jobs.object_name,jobs.attempt_count,jobs.created_at`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("claim media deletion jobs: %w", err)
	}
	defer rows.Close()
	result := make([]ports.MediaDeletion, 0)
	for rows.Next() {
		var deletion ports.MediaDeletion
		if err = rows.Scan(&deletion.ObjectName, &deletion.AttemptCount, &deletion.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan media deletion job: %w", err)
		}
		result = append(result, deletion)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read media deletion jobs: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit media deletion claims: %w", err)
	}
	return result, nil
}

func (repository *Repository) CompleteMediaDeletion(ctx context.Context, objectName string, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE admin.media_deletion_jobs SET status='completed',completed_at=$2,
		updated_at=$2,last_error_code=NULL WHERE object_name=$1 AND status IN ('pending','processing')`, objectName, now)
	if err != nil {
		return fmt.Errorf("complete media deletion: %w", err)
	}
	return nil
}

func (repository *Repository) FailMediaDeletion(ctx context.Context, objectName, errorCode string, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE admin.media_deletion_jobs SET
		status=CASE WHEN attempt_count>=5 THEN 'dead' ELSE 'pending' END,
		next_attempt_at=CASE WHEN attempt_count>=5 THEN next_attempt_at ELSE $3 + make_interval(secs => LEAST(300,attempt_count*attempt_count*5)) END,
		last_error_code=$2,updated_at=$3 WHERE object_name=$1 AND status IN ('pending','processing')`, objectName, errorCode, now)
	if err != nil {
		return fmt.Errorf("fail media deletion: %w", err)
	}
	return nil
}

func (repository *Repository) BeginMediaReconciliation(ctx context.Context, mode string, batchSize int,
	objectCursor, referenceCursor string, now time.Time) (string, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `UPDATE admin.media_reconciliation_runs
		SET status='failed',error_code='reconciliation.lease_expired',completed_at=$1,updated_at=$1
		WHERE status='running' AND updated_at < ($1::timestamptz - interval '30 minutes')`, now); err != nil {
		return "", fmt.Errorf("recover stale media reconciliation: %w", err)
	}
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO admin.media_reconciliation_runs
		(mode,batch_size,object_cursor,reference_cursor,started_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$5) RETURNING id`, mode, batchSize, objectCursor, referenceCursor, now).Scan(&id)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return "", ports.ErrMediaReconciliationRunning
		}
		return "", fmt.Errorf("begin media reconciliation: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit media reconciliation start: %w", err)
	}
	return id, nil
}

func (repository *Repository) InspectMediaObject(ctx context.Context, productID catalog.ProductID, objectName string) (ports.MediaObjectState, error) {
	var state ports.MediaObjectState
	imageURL := productMediaURLPrefix + objectName
	if err := repository.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER (WHERE product_id=$2)
		FROM catalog.product_images WHERE url=$1`, imageURL, productID).Scan(&state.ReferenceCount, &state.MatchingProductCount); err != nil {
		return ports.MediaObjectState{}, fmt.Errorf("inspect media references: %w", err)
	}
	var updatedAt, nextAttempt time.Time
	err := repository.pool.QueryRow(ctx, `SELECT status,updated_at,next_attempt_at
		FROM admin.media_deletion_jobs WHERE object_name=$1`, objectName).Scan(&state.DeletionStatus, &updatedAt, &nextAttempt)
	if errors.Is(err, pgx.ErrNoRows) {
		return state, nil
	}
	if err != nil {
		return ports.MediaObjectState{}, fmt.Errorf("inspect media deletion state: %w", err)
	}
	state.DeletionUpdatedAt, state.DeletionNextAttempt = &updatedAt, &nextAttempt
	return state, nil
}

func (repository *Repository) ListMediaReferences(ctx context.Context, cursor string, limit int) ([]ports.MediaReference, string, error) {
	rows, err := repository.pool.Query(ctx, `WITH media_references AS (
		SELECT substring(url FROM char_length($1) + 1) AS object_name,
			array_agg(DISTINCT product_id::text ORDER BY product_id::text) AS product_ids
		FROM catalog.product_images
		WHERE url LIKE $1 || '%' AND substring(url FROM char_length($1) + 1) > $2
		GROUP BY object_name
	)
	SELECT object_name,product_ids FROM media_references ORDER BY object_name LIMIT $3`, productMediaURLPrefix, cursor, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("list media database references: %w", err)
	}
	defer rows.Close()
	result := make([]ports.MediaReference, 0, limit)
	next := ""
	for rows.Next() {
		var objectName string
		var productIDs []string
		if err = rows.Scan(&objectName, &productIDs); err != nil {
			return nil, "", fmt.Errorf("scan media database reference: %w", err)
		}
		if len(result) == limit {
			next = result[len(result)-1].ObjectName
			break
		}
		reference := ports.MediaReference{ObjectName: objectName, ProductIDs: make([]catalog.ProductID, len(productIDs))}
		for index, productID := range productIDs {
			reference.ProductIDs[index] = catalog.ProductID(productID)
		}
		result = append(result, reference)
	}
	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("read media database references: %w", err)
	}
	return result, next, nil
}

func (repository *Repository) ListStaleMediaDeletions(ctx context.Context, before time.Time, limit int) ([]ports.MediaDeletion, error) {
	rows, err := repository.pool.Query(ctx, `SELECT object_name,attempt_count,created_at
		FROM admin.media_deletion_jobs
		WHERE status='dead'
			OR (status='processing' AND updated_at < $1)
			OR (status='pending' AND next_attempt_at < $1)
			OR (status='completed' AND completed_at < $1)
		ORDER BY updated_at,id LIMIT $2`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale media deletions: %w", err)
	}
	defer rows.Close()
	result := make([]ports.MediaDeletion, 0)
	for rows.Next() {
		var deletion ports.MediaDeletion
		if err = rows.Scan(&deletion.ObjectName, &deletion.AttemptCount, &deletion.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan stale media deletion: %w", err)
		}
		result = append(result, deletion)
	}
	return result, rows.Err()
}

func (repository *Repository) RecordMediaReconciliationResult(ctx context.Context, runID string, result ports.MediaReconciliationResult) error {
	_, err := repository.pool.Exec(ctx, `INSERT INTO admin.media_reconciliation_results
		(run_id,classification,identifier_hash,safe_object_name,action,detail_code)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (run_id,classification,identifier_hash,detail_code) DO NOTHING`,
		runID, result.Classification, result.IdentifierHash, result.SafeObjectName, result.Action, result.DetailCode)
	if err != nil {
		return fmt.Errorf("record media reconciliation result: %w", err)
	}
	return nil
}

func (repository *Repository) FinishMediaReconciliation(ctx context.Context, run ports.MediaReconciliationRun, now time.Time) error {
	tag, err := repository.pool.Exec(ctx, `UPDATE admin.media_reconciliation_runs SET status='completed',
		next_object_cursor=$2,next_reference_cursor=$3,objects_inspected=$4,references_inspected=$5,
		discrepancy_count=$6,deletion_jobs_scheduled=$7,completed_at=$8,updated_at=$8
		WHERE id=$1 AND status='running'`, run.ID, run.NextObjectCursor, run.NextReferenceCursor,
		run.ObjectsInspected, run.ReferencesInspected, run.Discrepancies, run.DeletionJobsScheduled, now)
	if err != nil {
		return fmt.Errorf("finish media reconciliation: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("media reconciliation run is not active")
	}
	return nil
}

func (repository *Repository) FailMediaReconciliation(ctx context.Context, runID, errorCode string, now time.Time) error {
	_, err := repository.pool.Exec(ctx, `UPDATE admin.media_reconciliation_runs SET status='failed',
		error_code=$2,completed_at=$3,updated_at=$3 WHERE id=$1 AND status='running'`, runID, errorCode, now)
	if err != nil {
		return fmt.Errorf("fail media reconciliation: %w", err)
	}
	return nil
}
