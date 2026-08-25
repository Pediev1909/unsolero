-- Recommendation history must preserve the same physical/non-physical
-- distinction as the catalog. Software products deliberately have no
-- dimensions, so their immutable candidate snapshots store NULL rather than
-- inventing a physical footprint.

ALTER TABLE recommendation.candidate_snapshots
    ALTER COLUMN length_mm DROP NOT NULL,
    ALTER COLUMN width_mm DROP NOT NULL,
    ALTER COLUMN height_mm DROP NOT NULL,
    ADD CONSTRAINT candidate_snapshots_dimensions_complete_check CHECK (
        (length_mm IS NULL AND width_mm IS NULL AND height_mm IS NULL)
        OR (length_mm > 0 AND width_mm > 0 AND height_mm > 0)
    );

COMMENT ON COLUMN recommendation.candidate_snapshots.length_mm IS
    'Immutable product length at recommendation time; NULL for non-physical products.';
COMMENT ON COLUMN recommendation.candidate_snapshots.width_mm IS
    'Immutable product width at recommendation time; NULL for non-physical products.';
COMMENT ON COLUMN recommendation.candidate_snapshots.height_mm IS
    'Immutable product height at recommendation time; NULL for non-physical products.';
