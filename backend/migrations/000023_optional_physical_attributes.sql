-- Physical attributes become optional so a non-physical vertical can hold a
-- catalog at all. A software product has no footprint, weight or material, and
-- the previous NOT NULL columns forced inventing values that would then feed
-- the recommendation engine as if they were facts.
--
-- The guarantee for physical goods is preserved rather than dropped: a
-- category declares whether its products occupy space, and a trigger enforces
-- that a physical product still carries complete physical facts.

ALTER TABLE catalog.categories
    ADD COLUMN is_physical boolean NOT NULL DEFAULT true;

COMMENT ON COLUMN catalog.categories.is_physical IS
    'True when products in this category occupy physical space and must carry dimensions, weight and material.';

ALTER TABLE catalog.products
    ALTER COLUMN length_mm DROP NOT NULL,
    ALTER COLUMN width_mm DROP NOT NULL,
    ALTER COLUMN height_mm DROP NOT NULL,
    ALTER COLUMN weight_grams DROP NOT NULL,
    ALTER COLUMN material DROP NOT NULL;

-- The pre-existing positive-value CHECKs still apply whenever a value is
-- present, because a CHECK expression evaluating to NULL is not a violation.

CREATE FUNCTION catalog.enforce_physical_product_attributes()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    category_is_physical boolean;
BEGIN
    SELECT is_physical INTO category_is_physical
    FROM catalog.categories WHERE id = NEW.category_id;

    IF category_is_physical THEN
        IF NEW.length_mm IS NULL OR NEW.width_mm IS NULL OR NEW.height_mm IS NULL
           OR NEW.weight_grams IS NULL OR NEW.material IS NULL THEN
            RAISE EXCEPTION 'product % is in a physical category and requires dimensions, weight and material', NEW.slug
                USING ERRCODE = '23514';
        END IF;
    ELSE
        IF NEW.length_mm IS NOT NULL OR NEW.width_mm IS NOT NULL OR NEW.height_mm IS NOT NULL
           OR NEW.weight_grams IS NOT NULL OR NEW.material IS NOT NULL
           OR NEW.max_capacity_grams IS NOT NULL THEN
            RAISE EXCEPTION 'product % is in a non-physical category and must not carry physical attributes', NEW.slug
                USING ERRCODE = '23514';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER enforce_physical_product_attributes
BEFORE INSERT OR UPDATE ON catalog.products
FOR EACH ROW EXECUTE FUNCTION catalog.enforce_physical_product_attributes();

-- products_recommendation_scores_idx orders by apartment_score, which only
-- means anything for a physical vertical. It is replaced by an index on the
-- dimensions that rank candidates in every vertical.
DROP INDEX IF EXISTS catalog.products_recommendation_scores_idx;

CREATE INDEX products_recommendation_scores_idx ON catalog.products (
    beginner_score DESC,
    value_score DESC,
    quality_score DESC
) WHERE status = 'published';
