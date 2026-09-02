-- The compared price gets a billing basis.
--
-- catalog.products.price_minor has always been "the price", and the SaaS
-- catalog filled it from vendor pages that quote two figures for one plan:
-- the monthly-billing price and the per-month equivalent of an annual
-- contract. Twenty-five of the fifty-three published software products stored
-- the annual-contract figure while the rest stored the monthly one, and the
-- recommendation engine compared them as alike: budget eligibility, value and
-- the cheaper/premium alternatives all rested on a figure whose basis nothing
-- recorded.
--
-- The site rule, set by the Zoho Books correction of 2026-08-26, is that
-- price_minor is the monthly-billing list price wherever the vendor sells
-- monthly. These columns record the basis so that rule can be checked, and
-- carry the annual figure separately so it can be shown without being
-- compared. The recommendation engine keeps reading price_minor and needs no
-- change. Rows are corrected by a dated data seed, not here: every existing
-- row takes the defaults, which state what the catalog already assumed of it.

ALTER TABLE catalog.products
    ADD COLUMN billing_period text NOT NULL DEFAULT 'monthly'
        CHECK (billing_period IN ('monthly', 'annual', 'free', 'usage')),
    ADD COLUMN pricing_unit text NOT NULL DEFAULT 'flat'
        CHECK (pricing_unit IN ('flat', 'per_user', 'per_contacts', 'per_transaction', 'usage')),
    ADD COLUMN pricing_unit_note text
        CHECK (pricing_unit_note IS NULL OR char_length(btrim(pricing_unit_note)) BETWEEN 1 AND 120),
    ADD COLUMN annual_price_minor bigint
        CHECK (annual_price_minor IS NULL OR annual_price_minor >= 0),
    -- The annual figure exists beside a monthly price, never instead of one:
    -- when the vendor sells only annually, price_minor already is that figure.
    ADD CONSTRAINT products_annual_price_requires_monthly_billing
        CHECK (annual_price_minor IS NULL OR billing_period = 'monthly'),
    ADD CONSTRAINT products_free_plan_costs_nothing
        CHECK (billing_period <> 'free' OR price_minor = 0);

COMMENT ON COLUMN catalog.products.price_minor IS
    'The compared price, in minor units: the monthly-billing list price wherever the vendor sells monthly. See billing_period for the basis when it does not.';
COMMENT ON COLUMN catalog.products.billing_period IS
    'Basis of price_minor. monthly: the vendor sells monthly and this is the monthly-billing price. annual: the vendor offers no monthly billing and this is the per-month equivalent of an annual contract. free: the entry tier is free and price_minor is 0. usage: the price is usage-based (payments, automation tasks) and price_minor is the entry figure or 0, with pricing_unit_note explaining.';
COMMENT ON COLUMN catalog.products.pricing_unit IS
    'What one unit of price_minor buys: flat (per account or organization), per_user, per_contacts (a contact-count tier), per_transaction, or usage.';
COMMENT ON COLUMN catalog.products.pricing_unit_note IS
    'Short qualifier for the unit, shown beside the price: "at 1,000 contacts", "per seat, minimum 3 seats", "2.9% + 30 cents per transaction".';
COMMENT ON COLUMN catalog.products.annual_price_minor IS
    'Per-month equivalent when billed annually, in minor units. Present only when the vendor offers both monthly and annual billing; displayed, never compared.';

-- A fact revision restates the billing basis as one fact or not at all. A
-- revision with a NULL period does not touch billing and publication leaves
-- the product''s current basis in place; a revision with a period replaces all
-- four columns, so an annual option the vendor has withdrawn can be cleared.
ALTER TABLE evidence.product_fact_revisions
    ADD COLUMN billing_period text
        CHECK (billing_period IS NULL OR billing_period IN ('monthly', 'annual', 'free', 'usage')),
    ADD COLUMN pricing_unit text
        CHECK (pricing_unit IS NULL OR pricing_unit IN ('flat', 'per_user', 'per_contacts', 'per_transaction', 'usage')),
    ADD COLUMN pricing_unit_note text
        CHECK (pricing_unit_note IS NULL OR char_length(btrim(pricing_unit_note)) BETWEEN 1 AND 120),
    ADD COLUMN annual_price_minor bigint
        CHECK (annual_price_minor IS NULL OR annual_price_minor >= 0),
    ADD CONSTRAINT fact_revisions_billing_basis_is_whole
        CHECK ((billing_period IS NULL) = (pricing_unit IS NULL)),
    ADD CONSTRAINT fact_revisions_billing_detail_requires_basis
        CHECK (billing_period IS NOT NULL OR (pricing_unit_note IS NULL AND annual_price_minor IS NULL)),
    ADD CONSTRAINT fact_revisions_annual_price_requires_monthly_billing
        CHECK (annual_price_minor IS NULL OR billing_period = 'monthly'),
    ADD CONSTRAINT fact_revisions_free_plan_costs_nothing
        CHECK (billing_period <> 'free' OR price_minor = 0);

COMMENT ON COLUMN evidence.product_fact_revisions.billing_period IS
    'Revised basis of price_minor; NULL leaves the product''s basis untouched at publication. Values as catalog.products.billing_period.';
COMMENT ON COLUMN evidence.product_fact_revisions.pricing_unit IS
    'Revised pricing unit; present exactly when billing_period is. Values as catalog.products.pricing_unit.';
COMMENT ON COLUMN evidence.product_fact_revisions.pricing_unit_note IS
    'Revised unit qualifier; requires billing_period. See catalog.products.pricing_unit_note.';
COMMENT ON COLUMN evidence.product_fact_revisions.annual_price_minor IS
    'Revised per-month annual-billing equivalent; requires billing_period = monthly. See catalog.products.annual_price_minor.';
