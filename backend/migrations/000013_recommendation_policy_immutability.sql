-- Preserve every policy-derived input needed to reproduce a completed run.
ALTER TABLE recommendation.session_existing_equipment
    ADD COLUMN capabilities text[] NOT NULL DEFAULT '{}',
    ADD COLUMN redundancy_groups text[] NOT NULL DEFAULT '{}';

-- Once a policy is active, its behavior is immutable. Changes require a new
-- policy version and the review/activation workflow. Retired policy data stays
-- immutable so historical recommendations remain reproducible.
CREATE FUNCTION recommendation.reject_immutable_policy_child_change()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    target_policy_version text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        target_policy_version := OLD.policy_version;
    ELSE
        target_policy_version := NEW.policy_version;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM recommendation.policy_versions
        WHERE version = target_policy_version
          AND (
              workflow_status = 'active'
              OR (workflow_status = 'retired' AND EXISTS (
                  SELECT 1 FROM recommendation.recommendations
                  WHERE policy_version = target_policy_version
              ))
          )
    ) THEN
        RAISE EXCEPTION 'recommendation policy % is immutable in its current state', target_policy_version
            USING ERRCODE = '55000';
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$function$;

DO $block$
DECLARE
    table_name text;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[
        'policy_capabilities',
        'policy_goals',
        'policy_setup_roles',
        'policy_setup_role_capabilities',
        'category_policies',
        'category_policy_capabilities',
        'policy_redundancy_groups',
        'category_redundancy_groups',
        'product_policies',
        'product_policy_capabilities',
        'product_goal_support',
        'product_preference_tags',
        'product_redundancy_groups',
        'product_space_profiles'
    ] LOOP
        EXECUTE format(
            'CREATE TRIGGER reject_immutable_policy_child_change BEFORE INSERT OR UPDATE OR DELETE ON recommendation.%I FOR EACH ROW EXECUTE FUNCTION recommendation.reject_immutable_policy_child_change()',
            table_name
        );
    END LOOP;
END;
$block$;

CREATE FUNCTION recommendation.reject_immutable_policy_definition_change()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    IF OLD.workflow_status IN ('active', 'retired') AND (
        NEW.version,
        NEW.vertical_key,
        NEW.goal_match_weight,
        NEW.budget_match_weight,
        NEW.space_match_weight,
        NEW.experience_match_weight,
        NEW.preference_match_weight,
        NEW.quality_weight,
        NEW.value_weight,
        NEW.durability_weight,
        NEW.compatibility_weight,
        NEW.portability_weight,
        NEW.noise_weight,
        NEW.priority_boost_percent,
        NEW.maximum_setup_items,
        NEW.candidates_per_slot,
        NEW.optional_slot_bonus
    ) IS DISTINCT FROM (
        OLD.version,
        OLD.vertical_key,
        OLD.goal_match_weight,
        OLD.budget_match_weight,
        OLD.space_match_weight,
        OLD.experience_match_weight,
        OLD.preference_match_weight,
        OLD.quality_weight,
        OLD.value_weight,
        OLD.durability_weight,
        OLD.compatibility_weight,
        OLD.portability_weight,
        OLD.noise_weight,
        OLD.priority_boost_percent,
        OLD.maximum_setup_items,
        OLD.candidates_per_slot,
        OLD.optional_slot_bonus
    ) THEN
        RAISE EXCEPTION 'recommendation policy % definition is immutable', OLD.version
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER reject_immutable_policy_definition_change
BEFORE UPDATE ON recommendation.policy_versions
FOR EACH ROW EXECUTE FUNCTION recommendation.reject_immutable_policy_definition_change();

COMMENT ON FUNCTION recommendation.reject_immutable_policy_child_change() IS
    'Requires a new reviewed policy version instead of mutating active or historically referenced recommendation behavior.';
