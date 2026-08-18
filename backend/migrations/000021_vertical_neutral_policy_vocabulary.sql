-- Moves the preference and priority vocabularies out of Go constants and into
-- policy data, so a new vertical ships as a policy version rather than as an
-- engine change. Capabilities and goals were already data-driven; these two
-- were the remaining closed enumerations.
--
-- spatial_constraints records whether the vertical has physical products. A
-- non-spatial vertical (for example a software stack) has no room or footprint
-- to reason about, and the engine skips space eligibility and space scoring
-- rather than demanding invented measurements.

ALTER TABLE recommendation.policy_versions
    ADD COLUMN spatial_constraints boolean NOT NULL DEFAULT true;

COMMENT ON COLUMN recommendation.policy_versions.spatial_constraints IS
    'True when products occupy physical space. False disables space eligibility and space scoring for the vertical.';

CREATE TABLE recommendation.policy_preference_tags (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    tag_key text NOT NULL CHECK (tag_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    PRIMARY KEY (policy_version, tag_key)
);

COMMENT ON TABLE recommendation.policy_preference_tags IS
    'Preference vocabulary a policy version accepts. A tag outside this set is rejected by the engine.';

-- reason_dimension and dimension are constrained to the engine's canonical
-- scored dimensions. A priority pointing at an unknown dimension would boost
-- nothing and silently read as a scoring bug, so it is rejected at write time.
CREATE TABLE recommendation.policy_priorities (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    priority_key text NOT NULL CHECK (priority_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    reason_code text NOT NULL CHECK (reason_code ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$'),
    reason_message text NOT NULL CHECK (char_length(btrim(reason_message)) BETWEEN 1 AND 200),
    reason_dimension text NOT NULL CHECK (reason_dimension IN (
        'goal_match','budget_match','space_match','experience_match','preference_match',
        'quality','value','durability','compatibility','portability','noise')),
    reason_threshold smallint NOT NULL CHECK (reason_threshold BETWEEN 0 AND 100),
    sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    PRIMARY KEY (policy_version, priority_key)
);

COMMENT ON TABLE recommendation.policy_priorities IS
    'User-selectable priorities and the explanation each produces. Ordering is policy-defined so scoring stays independent of user selection order.';

CREATE TABLE recommendation.policy_priority_dimensions (
    policy_version text NOT NULL,
    priority_key text NOT NULL,
    dimension text NOT NULL CHECK (dimension IN (
        'goal_match','budget_match','space_match','experience_match','preference_match',
        'quality','value','durability','compatibility','portability','noise')),
    PRIMARY KEY (policy_version, priority_key, dimension),
    FOREIGN KEY (policy_version, priority_key)
        REFERENCES recommendation.policy_priorities(policy_version, priority_key) ON DELETE RESTRICT
);

COMMENT ON TABLE recommendation.policy_priority_dimensions IS
    'Scored dimensions each priority boosts. A priority may boost more than one dimension.';

-- The new tables are part of a policy definition and must inherit the same
-- immutability guarantee as every other policy child table; without this an
-- active policy could be altered underneath published recommendations.
DO $block$
DECLARE
    table_name text;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[
        'policy_preference_tags',
        'policy_priorities',
        'policy_priority_dimensions'
    ] LOOP
        EXECUTE format(
            'CREATE TRIGGER reject_immutable_policy_child_change BEFORE INSERT OR UPDATE OR DELETE ON recommendation.%I FOR EACH ROW EXECUTE FUNCTION recommendation.reject_immutable_policy_child_change()',
            table_name
        );
    END LOOP;
END;
$block$;

-- No existing policy is modified here. recommendation.policy_versions marks an
-- active policy immutable and forbids the active -> draft transition, so the
-- fitness-v2 demo policy cannot be backfilled in place, and rewriting it as a
-- new version would only preserve fictional demo data that the vertical pivot
-- retires. The SaaS policy ships as its own policy version with its complete
-- vocabulary; fitness-v2 remains valid but declares no preferences or
-- priorities, so those inputs are simply unavailable while it is selected.
