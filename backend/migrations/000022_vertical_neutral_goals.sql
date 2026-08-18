-- Opens the goal vocabulary the same way capabilities, preferences and
-- priorities were opened: the column keeps a strict normalized-code format,
-- and which goals actually exist is declared by the active recommendation
-- policy in recommendation.policy_goals.
--
-- The previous CHECK lists enumerated fitness goals, which would reject every
-- goal a non-fitness vertical defines. Format is still enforced, so a
-- malformed key is refused at write time; membership is enforced by the
-- engine, which rejects any goal the active policy does not declare.

ALTER TABLE planning.profiles
    DROP CONSTRAINT profiles_primary_goal_check,
    ADD CONSTRAINT profiles_primary_goal_check
        CHECK (primary_goal ~ '^[a-z][a-z0-9_]*$');

ALTER TABLE recommendation.recommendation_sessions
    DROP CONSTRAINT recommendation_sessions_primary_goal_check,
    ADD CONSTRAINT recommendation_sessions_primary_goal_check
        CHECK (primary_goal ~ '^[a-z][a-z0-9_]*$');

ALTER TABLE recommendation.drafts
    DROP CONSTRAINT drafts_primary_goal_check,
    ADD CONSTRAINT drafts_primary_goal_check
        CHECK (primary_goal IS NULL OR primary_goal ~ '^[a-z][a-z0-9_]*$');

COMMENT ON COLUMN planning.profiles.primary_goal IS
    'Goal key declared by the active recommendation policy for the deployment vertical.';
COMMENT ON COLUMN recommendation.recommendation_sessions.primary_goal IS
    'Goal key declared by the active recommendation policy for the deployment vertical.';
COMMENT ON COLUMN recommendation.drafts.primary_goal IS
    'Goal key declared by the active recommendation policy for the deployment vertical.';

-- experience_level stays a closed enumeration. Beginner, intermediate and
-- advanced describe the person rather than the product domain, and they carry
-- the same meaning for a software stack as for a gym, so opening them would
-- weaken validation without enabling anything.
