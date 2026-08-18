CREATE TABLE admin.media_reconciliation_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    mode text NOT NULL CHECK (mode IN ('dry_run', 'apply')),
    status text NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'failed')),
    batch_size smallint NOT NULL CHECK (batch_size BETWEEN 1 AND 500),
    object_cursor text NOT NULL DEFAULT '' CHECK (char_length(object_cursor) <= 1024),
    reference_cursor text NOT NULL DEFAULT '' CHECK (char_length(reference_cursor) <= 240),
    next_object_cursor text NOT NULL DEFAULT '' CHECK (char_length(next_object_cursor) <= 1024),
    next_reference_cursor text NOT NULL DEFAULT '' CHECK (char_length(next_reference_cursor) <= 240),
    objects_inspected integer NOT NULL DEFAULT 0 CHECK (objects_inspected >= 0),
    references_inspected integer NOT NULL DEFAULT 0 CHECK (references_inspected >= 0),
    discrepancy_count integer NOT NULL DEFAULT 0 CHECK (discrepancy_count >= 0),
    deletion_jobs_scheduled integer NOT NULL DEFAULT 0 CHECK (deletion_jobs_scheduled >= 0),
    error_code text CHECK (error_code IS NULL OR error_code ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX media_reconciliation_one_running_idx
    ON admin.media_reconciliation_runs ((true)) WHERE status = 'running';

CREATE INDEX media_reconciliation_runs_started_idx
    ON admin.media_reconciliation_runs (started_at DESC, id DESC);

CREATE TABLE admin.media_reconciliation_results (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id uuid NOT NULL REFERENCES admin.media_reconciliation_runs(id) ON DELETE CASCADE,
    classification text NOT NULL CHECK (classification IN (
        'orphan_object', 'missing_object', 'duplicate_reference',
        'ownership_mismatch', 'unexpected_namespace', 'stale_deletion', 'unclassified'
    )),
    identifier_hash text NOT NULL CHECK (identifier_hash ~ '^[a-f0-9]{64}$'),
    safe_object_name text CHECK (
        safe_object_name IS NULL OR (
            char_length(safe_object_name) BETWEEN 1 AND 240
            AND safe_object_name !~ '[\\/]'
        )
    ),
    action text NOT NULL CHECK (action IN ('none', 'deletion_scheduled', 'deletion_requeued')),
    detail_code text NOT NULL CHECK (detail_code ~ '^[a-z][a-z0-9_.-]{0,95}$'),
    detected_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (run_id, classification, identifier_hash, detail_code)
);

CREATE INDEX media_reconciliation_results_run_idx
    ON admin.media_reconciliation_results (run_id, classification, detected_at, id);

CREATE TABLE platform.operational_checkpoints (
    checkpoint_name text PRIMARY KEY CHECK (checkpoint_name IN ('backup', 'restore_verification')),
    status text NOT NULL CHECK (status IN ('ok', 'failed', 'mismatch', 'unknown')),
    observed_at timestamptz NOT NULL,
    last_success_at timestamptz,
    failure_count bigint NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    detail_code text CHECK (detail_code IS NULL OR detail_code ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    updated_at timestamptz NOT NULL DEFAULT now()
);
