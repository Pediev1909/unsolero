-- Removes the fictional SaaS fixture from a database.
--
-- saas_demo.sql inserts ten invented products at status 'published', which is
-- the same status the public catalog queries filter on. A database that has
-- ever had the fixture applied therefore serves invented products and invented
-- prices to real visitors. This undoes that.
--
-- The fixture is identified by its brands' reserved example.invalid websites
-- rather than by a slug list, so a fixture brand added later is still caught.
--
-- Two passes, because they have different guarantees:
--
--   1. Unpublishing always succeeds. catalog.products carries no policy trigger
--      and no inbound restriction on UPDATE, so this pass cannot fail and is
--      what actually takes the fixture off the public site.
--   2. Deleting is best-effort. recommendation.product_policies is guarded by
--      reject_immutable_policy_child_change, which rejects DELETE while the
--      owning policy version is active, or retired with recommendations against
--      it. A product held by such a policy keeps its rows and stays at 'draft'.
--      That is the designed behaviour: policy history is immutable on purpose.
--
-- Safe to run repeatedly.

DO $purge$
DECLARE
    unpublished integer;
    deletable   integer;
    withheld    integer;
BEGIN
    CREATE TEMP TABLE fixture_products ON COMMIT DROP AS
    SELECT products.id
    FROM catalog.products products
    JOIN catalog.brands brands ON brands.id = products.brand_id
    WHERE brands.website_url LIKE 'https://example.invalid/%';

    CREATE TEMP TABLE fixture_brands ON COMMIT DROP AS
    SELECT id FROM catalog.brands WHERE website_url LIKE 'https://example.invalid/%';

    -- Pass 1: leave the public catalog. Unconditional.
    UPDATE catalog.products
       SET status = 'draft', updated_at = now()
     WHERE id IN (SELECT id FROM fixture_products)
       AND status <> 'draft';
    GET DIAGNOSTICS unpublished = ROW_COUNT;

    -- Pass 2: only products no immutable policy is holding.
    CREATE TEMP TABLE removable ON COMMIT DROP AS
    SELECT id FROM fixture_products
    EXCEPT
    SELECT policies.product_id
    FROM recommendation.product_policies policies
    JOIN recommendation.policy_versions versions ON versions.version = policies.policy_version
    WHERE versions.workflow_status = 'active'
       OR (versions.workflow_status = 'retired' AND EXISTS (
            SELECT 1 FROM recommendation.recommendations
             WHERE policy_version = versions.version));

    SELECT count(*) INTO deletable FROM removable;
    SELECT count(*) INTO withheld FROM fixture_products
     WHERE id NOT IN (SELECT id FROM removable);

    -- Policy children first: each restricts the row above it.
    DELETE FROM recommendation.product_redundancy_groups WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.product_preference_tags   WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.product_goal_support      WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.product_policy_capabilities WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.product_space_profiles    WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.product_policies          WHERE product_id IN (SELECT id FROM removable);

    -- Offer dependants, then the offers, then the remaining product restrictions.
    DELETE FROM commerce.affiliate_clicks WHERE merchant_offer_id IN (
        SELECT id FROM commerce.merchant_offers WHERE product_id IN (SELECT id FROM removable));
    DELETE FROM commerce.provider_offer_mappings WHERE merchant_offer_id IN (
        SELECT id FROM commerce.merchant_offers WHERE product_id IN (SELECT id FROM removable));
    DELETE FROM commerce.price_observations WHERE merchant_offer_id IN (
        SELECT id FROM commerce.merchant_offers WHERE product_id IN (SELECT id FROM removable));
    DELETE FROM commerce.availability_observations WHERE merchant_offer_id IN (
        SELECT id FROM commerce.merchant_offers WHERE product_id IN (SELECT id FROM removable));
    DELETE FROM commerce.affiliate_clicks WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM commerce.merchant_offers  WHERE product_id IN (SELECT id FROM removable);

    DELETE FROM recommendation.candidate_snapshots  WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM recommendation.recommendation_items WHERE product_id IN (SELECT id FROM removable)
                                                       OR alternative_for_product_id IN (SELECT id FROM removable);
    DELETE FROM planning.setup_items    WHERE product_id IN (SELECT id FROM removable);
    DELETE FROM editorial.entry_products WHERE product_id IN (SELECT id FROM removable);

    -- Products carry their own evidence, images, attributes and saved lists by
    -- cascade. Brands restrict on evidence.product_fact_revisions.brand_id,
    -- which that cascade has already cleared by this point.
    DELETE FROM catalog.products WHERE id IN (SELECT id FROM removable);
    DELETE FROM catalog.brands
     WHERE id IN (SELECT id FROM fixture_brands)
       AND NOT EXISTS (SELECT 1 FROM catalog.products WHERE brand_id = catalog.brands.id);

    RAISE NOTICE 'fixture purge: % unpublished, % deleted, % held by an immutable policy',
        unpublished, deletable, withheld;
END
$purge$;
