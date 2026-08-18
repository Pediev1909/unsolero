ALTER TABLE platform.operational_checkpoints
    DROP CONSTRAINT operational_checkpoints_checkpoint_name_check;

ALTER TABLE platform.operational_checkpoints
    ADD CONSTRAINT operational_checkpoints_checkpoint_name_check
    CHECK (checkpoint_name IN ('backup', 'restore_verification', 'worker'));
