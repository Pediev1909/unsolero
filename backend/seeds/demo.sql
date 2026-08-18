INSERT INTO catalog.categories (name, slug, description, sort_order)
VALUES
    ('Adjustable Dumbbells', 'adjustable-dumbbells', 'Space-efficient selectable dumbbell systems.', 10),
    ('Benches', 'benches', 'Flat, incline, and folding strength benches.', 20),
    ('Power Racks', 'power-racks', 'Racks, cages, and stands for barbell training.', 30),
    ('Barbells', 'barbells', 'General-purpose and technique training bars.', 40),
    ('Weight Plates', 'weight-plates', 'Iron, bumper, and coated plate sets.', 50),
    ('Kettlebells', 'kettlebells', 'Fixed and adjustable kettlebells.', 60),
    ('Resistance Bands', 'resistance-bands', 'Compact elastic resistance systems.', 70),
    ('Cardio Machines', 'cardio-machines', 'Home-oriented conditioning equipment.', 80)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    is_active = true,
    updated_at = now();

INSERT INTO catalog.brands (name, slug, description, website_url, country_code)
VALUES
    ('Demo Northline', 'demo-northline', 'Fictional demonstration brand for compact strength equipment.', 'https://demo-northline.example.invalid', 'US'),
    ('Demo Forgefield', 'demo-forgefield', 'Fictional demonstration brand for traditional free weights.', 'https://demo-forgefield.example.invalid', 'US'),
    ('Demo QuietForm', 'demo-quietform', 'Fictional demonstration brand focused on quiet apartment training.', 'https://demo-quietform.example.invalid', 'CA'),
    ('Demo Oak & Iron', 'demo-oak-and-iron', 'Fictional demonstration brand for restrained home-gym equipment.', 'https://demo-oak-and-iron.example.invalid', 'GB'),
    ('Demo Civic Strength', 'demo-civic-strength', 'Fictional demonstration brand for accessible beginner equipment.', 'https://demo-civic-strength.example.invalid', 'US'),
    ('Demo Summit Yard', 'demo-summit-yard', 'Fictional demonstration brand for higher-capacity strength systems.', 'https://demo-summit-yard.example.invalid', 'DE'),
    ('Demo Pulseworks', 'demo-pulseworks', 'Fictional demonstration brand for home conditioning equipment.', 'https://demo-pulseworks.example.invalid', 'NL'),
    ('Demo Kinetic House', 'demo-kinetic-house', 'Fictional demonstration brand for portable training tools.', 'https://demo-kinetic-house.example.invalid', 'AU'),
    ('Demo Harbor Athletics', 'demo-harbor-athletics', 'Fictional demonstration brand for durable general fitness gear.', 'https://demo-harbor-athletics.example.invalid', 'SE'),
    ('Demo Range Lab', 'demo-range-lab', 'Fictional demonstration brand for advanced adjustable equipment.', 'https://demo-range-lab.example.invalid', 'JP')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    website_url = EXCLUDED.website_url,
    country_code = EXCLUDED.country_code,
    is_active = true,
    updated_at = now();

WITH product_seed (
    category_slug, brand_slug, name, slug, description,
    price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
    max_capacity_grams, material, warranty_months,
    quality_score, value_score, durability_score, beginner_score,
    advanced_score, apartment_score, noise_score, portability_score
) AS (
    VALUES
        ('adjustable-dumbbells', 'demo-northline', 'Demo Northline Nest 24 Dumbbell Pair', 'demo-northline-nest-24-pair', 'Fictional demo adjustable dumbbell pair with a compact nested plate design and 24 kg maximum per hand.', 32900, 'USD', 560, 340, 260, 50000, 48000, 'Powder-coated steel and nylon', 24, 84, 82, 82, 91, 72, 91, 88, 72),
        ('adjustable-dumbbells', 'demo-quietform', 'Demo QuietForm Dial 20 Dumbbell Pair', 'demo-quietform-dial-20-pair', 'Fictional demo adjustable pair designed with restrained plate movement for quieter apartment sessions.', 27900, 'USD', 540, 320, 250, 42000, 40000, 'Steel with thermoplastic coating', 24, 81, 85, 78, 93, 65, 95, 94, 76),
        ('adjustable-dumbbells', 'demo-civic-strength', 'Demo Civic Select 16 Dumbbell Pair', 'demo-civic-select-16-pair', 'Fictional demo entry-level selectable dumbbell pair intended for compact beginner setups.', 19900, 'USD', 500, 300, 240, 34000, 32000, 'Steel and reinforced polymer', 18, 75, 90, 74, 97, 48, 96, 91, 84),
        ('adjustable-dumbbells', 'demo-range-lab', 'Demo Range Lab ProBlock 32 Pair', 'demo-range-lab-problock-32-pair', 'Fictional demo high-range adjustable dumbbell pair for progressive home strength training.', 44900, 'USD', 620, 360, 280, 66000, 64000, 'Machined steel and nylon', 36, 91, 78, 90, 74, 94, 82, 85, 60),

        ('benches', 'demo-oak-and-iron', 'Demo Oak & Iron Foldaway Flat Bench', 'demo-oak-iron-foldaway-flat-bench', 'Fictional demo folding flat bench with a slim stored profile for shared living spaces.', 14900, 'USD', 1180, 430, 440, 18000, 250000, 'Steel frame with synthetic upholstery', 24, 80, 88, 81, 94, 70, 94, 92, 86),
        ('benches', 'demo-northline', 'Demo Northline Incline Five Bench', 'demo-northline-incline-five-bench', 'Fictional demo adjustable bench with five back positions and two seat positions.', 23900, 'USD', 1320, 550, 470, 28500, 320000, 'Powder-coated steel and vinyl', 36, 86, 83, 87, 88, 86, 76, 84, 54),
        ('benches', 'demo-quietform', 'Demo QuietForm Apartment Bench', 'demo-quietform-apartment-bench', 'Fictional demo compact adjustable bench with rubber contact points and vertical storage.', 21900, 'USD', 1240, 500, 460, 24000, 275000, 'Steel, dense foam, and rubber', 30, 83, 84, 82, 91, 79, 93, 94, 72),
        ('benches', 'demo-summit-yard', 'Demo Summit Yard Heavy Seven Bench', 'demo-summit-yard-heavy-seven-bench', 'Fictional demo high-capacity seven-position bench intended for advanced barbell setups.', 36900, 'USD', 1450, 620, 480, 41000, 450000, 'Heavy-gauge steel and vinyl', 60, 94, 74, 95, 69, 96, 54, 78, 31),

        ('power-racks', 'demo-northline', 'Demo Northline Slim Half Rack', 'demo-northline-slim-half-rack', 'Fictional demo half rack that balances barbell capability with a moderate residential footprint.', 69900, 'USD', 1200, 1250, 2150, 76000, 320000, '11-gauge powder-coated steel', 60, 90, 81, 91, 76, 92, 62, 76, 18),
        ('power-racks', 'demo-quietform', 'Demo QuietForm Wall Fold Rack', 'demo-quietform-wall-fold-rack', 'Fictional demo wall-mounted folding rack with lined contact points for compact rooms.', 74900, 'USD', 610, 1230, 2200, 68000, 300000, 'Powder-coated steel and rubber', 60, 88, 79, 89, 72, 91, 88, 88, 35),
        ('power-racks', 'demo-civic-strength', 'Demo Civic Compact Squat Stands', 'demo-civic-compact-squat-stands', 'Fictional demo independent squat stands for users who need a movable barbell support system.', 29900, 'USD', 650, 980, 1780, 34000, 220000, 'Powder-coated steel', 24, 76, 90, 78, 86, 70, 83, 82, 78),
        ('power-racks', 'demo-summit-yard', 'Demo Summit Yard Full Cage 90', 'demo-summit-yard-full-cage-90', 'Fictional demo full power cage with a high-capacity frame for dedicated training rooms.', 109900, 'USD', 1550, 1500, 2320, 118000, 450000, '11-gauge powder-coated steel', 84, 96, 76, 97, 65, 99, 35, 72, 8),

        ('barbells', 'demo-northline', 'Demo Northline Compact 15 Bar', 'demo-northline-compact-15-bar', 'Fictional demo 1.8 m barbell for smaller rooms and moderate plate loads.', 15900, 'USD', 1800, 50, 50, 15000, 180000, 'Alloy steel with hard chrome sleeves', 36, 82, 87, 84, 90, 74, 88, 82, 72),
        ('barbells', 'demo-forgefield', 'Demo Forgefield Standard 20 Bar', 'demo-forgefield-standard-20-bar', 'Fictional demo 2.2 m general-purpose barbell with dual training marks.', 22900, 'USD', 2200, 50, 50, 20000, 320000, 'Alloy steel with black oxide shaft', 60, 91, 84, 92, 82, 93, 58, 75, 42),
        ('barbells', 'demo-civic-strength', 'Demo Civic Technique 10 Bar', 'demo-civic-technique-10-bar', 'Fictional demo lightweight technique bar for learning barbell movement patterns.', 11900, 'USD', 1680, 50, 50, 10000, 70000, 'Aluminum and steel', 24, 74, 86, 70, 98, 38, 91, 88, 81),
        ('barbells', 'demo-quietform', 'Demo QuietForm Short 12 Bar', 'demo-quietform-short-12-bar', 'Fictional demo short barbell with coated sleeves for controlled compact-space training.', 17900, 'USD', 1600, 50, 50, 12000, 140000, 'Alloy steel with polymer sleeve guards', 36, 80, 82, 80, 91, 65, 94, 91, 79),

        ('weight-plates', 'demo-harbor-athletics', 'Demo Harbor Bumper 60 kg Set', 'demo-harbor-bumper-60-set', 'Fictional demo low-bounce bumper plate set totaling 60 kg.', 26900, 'USD', 450, 450, 260, 60000, NULL, 'Virgin rubber and stainless steel', 36, 87, 82, 88, 86, 88, 74, 89, 38),
        ('weight-plates', 'demo-forgefield', 'Demo Forgefield Iron 80 kg Set', 'demo-forgefield-iron-80-set', 'Fictional demo machined cast-iron plate set totaling 80 kg.', 22900, 'USD', 410, 410, 360, 80000, NULL, 'Machined cast iron', 60, 89, 91, 94, 78, 92, 44, 48, 22),
        ('weight-plates', 'demo-quietform', 'Demo QuietForm Urethane 40 kg Set', 'demo-quietform-urethane-40-set', 'Fictional demo compact urethane-coated plate set totaling 40 kg for quieter handling.', 23900, 'USD', 380, 380, 250, 40000, NULL, 'Urethane-coated iron', 48, 85, 76, 88, 91, 73, 93, 95, 54),
        ('weight-plates', 'demo-northline', 'Demo Northline Compact 50 kg Set', 'demo-northline-compact-50-set', 'Fictional demo thin-profile coated plate set totaling 50 kg.', 24900, 'USD', 400, 400, 240, 50000, NULL, 'Rubber-coated iron', 36, 84, 83, 86, 88, 80, 85, 90, 49),

        ('kettlebells', 'demo-range-lab', 'Demo Range Lab Adjustable 20 Kettlebell', 'demo-range-lab-adjustable-20-kettlebell', 'Fictional demo adjustable kettlebell spanning several common training weights up to 20 kg.', 15900, 'USD', 280, 220, 320, 20500, 20000, 'Cast iron and steel', 36, 87, 85, 88, 85, 88, 90, 87, 78),
        ('kettlebells', 'demo-civic-strength', 'Demo Civic Cast 12 Kettlebell', 'demo-civic-cast-12-kettlebell', 'Fictional demo 12 kg single-cast kettlebell suited to introductory strength work.', 5900, 'USD', 220, 170, 250, 12000, 12000, 'Powder-coated cast iron', 24, 77, 91, 85, 96, 48, 92, 86, 86),
        ('kettlebells', 'demo-harbor-athletics', 'Demo Harbor Cast 16 Kettlebell', 'demo-harbor-cast-16-kettlebell', 'Fictional demo 16 kg single-cast kettlebell with a smooth uncoated handle.', 7900, 'USD', 240, 185, 270, 16000, 16000, 'Powder-coated cast iron', 36, 84, 88, 91, 88, 72, 89, 84, 78),
        ('kettlebells', 'demo-summit-yard', 'Demo Summit Cast 24 Kettlebell', 'demo-summit-cast-24-kettlebell', 'Fictional demo 24 kg single-cast kettlebell intended for experienced trainees.', 10900, 'USD', 275, 205, 300, 24000, 24000, 'Powder-coated cast iron', 48, 90, 86, 95, 60, 94, 81, 80, 57),

        ('resistance-bands', 'demo-kinetic-house', 'Demo Kinetic House Starter Band Set', 'demo-kinetic-house-starter-band-set', 'Fictional demo five-band starter set with handles, anchor, and storage pouch.', 3900, 'USD', 250, 180, 90, 1800, 45000, 'Layered natural latex and nylon', 12, 75, 94, 72, 98, 54, 99, 98, 99),
        ('resistance-bands', 'demo-northline', 'Demo Northline Progressive Tube Set', 'demo-northline-progressive-tube-set', 'Fictional demo stackable tube resistance set with labeled resistance ranges.', 5900, 'USD', 300, 220, 100, 2600, 70000, 'Latex, nylon, and aluminum', 18, 82, 92, 80, 95, 72, 98, 97, 98),
        ('resistance-bands', 'demo-quietform', 'Demo QuietForm Fabric Loop Set', 'demo-quietform-fabric-loop-set', 'Fictional demo three-piece fabric loop set intended for silent lower-body training.', 2900, 'USD', 180, 140, 80, 900, 25000, 'Woven polyester and latex', 12, 76, 93, 79, 97, 48, 100, 100, 100),

        ('cardio-machines', 'demo-pulseworks', 'Demo Pulseworks Folding Rower', 'demo-pulseworks-folding-rower', 'Fictional demo magnetic rower with an upright storage position for home conditioning.', 64900, 'USD', 2050, 520, 710, 36000, 150000, 'Steel, aluminum, and nylon', 36, 84, 81, 83, 88, 78, 79, 86, 38),
        ('cardio-machines', 'demo-quietform', 'Demo QuietForm Walking Pad', 'demo-quietform-walking-pad', 'Fictional demo low-profile walking pad designed for moderate indoor walking sessions.', 49900, 'USD', 1420, 610, 130, 29000, 120000, 'Steel, ABS, and rubber belt', 24, 80, 84, 78, 94, 55, 91, 92, 62),
        ('cardio-machines', 'demo-pulseworks', 'Demo Pulseworks Compact Air Bike', 'demo-pulseworks-compact-air-bike', 'Fictional demo compact fan-resistance bike for higher-intensity conditioning.', 79900, 'USD', 1220, 610, 1280, 47000, 160000, 'Powder-coated steel and ABS', 48, 88, 79, 89, 76, 91, 48, 32, 28)
)
INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description,
    price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
    max_capacity_grams, material, warranty_months,
    quality_score, value_score, durability_score, beginner_score,
    advanced_score, apartment_score, noise_score, portability_score, status
)
SELECT
    categories.id,
    brands.id,
    product_seed.name,
    product_seed.slug,
    product_seed.description,
    product_seed.price_minor,
    product_seed.currency,
    product_seed.length_mm,
    product_seed.width_mm,
    product_seed.height_mm,
    product_seed.weight_grams,
    product_seed.max_capacity_grams,
    product_seed.material,
    product_seed.warranty_months,
    product_seed.quality_score,
    product_seed.value_score,
    product_seed.durability_score,
    product_seed.beginner_score,
    product_seed.advanced_score,
    product_seed.apartment_score,
    product_seed.noise_score,
    product_seed.portability_score,
    'published'
FROM product_seed
JOIN catalog.categories ON categories.slug = product_seed.category_slug
JOIN catalog.brands ON brands.slug = product_seed.brand_slug
ON CONFLICT (slug) DO UPDATE SET
    category_id = EXCLUDED.category_id,
    brand_id = EXCLUDED.brand_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_minor = EXCLUDED.price_minor,
    currency = EXCLUDED.currency,
    length_mm = EXCLUDED.length_mm,
    width_mm = EXCLUDED.width_mm,
    height_mm = EXCLUDED.height_mm,
    weight_grams = EXCLUDED.weight_grams,
    max_capacity_grams = EXCLUDED.max_capacity_grams,
    material = EXCLUDED.material,
    warranty_months = EXCLUDED.warranty_months,
    quality_score = EXCLUDED.quality_score,
    value_score = EXCLUDED.value_score,
    durability_score = EXCLUDED.durability_score,
    beginner_score = EXCLUDED.beginner_score,
    advanced_score = EXCLUDED.advanced_score,
    apartment_score = EXCLUDED.apartment_score,
    noise_score = EXCLUDED.noise_score,
    portability_score = EXCLUDED.portability_score,
    status = 'published',
    updated_at = now();

INSERT INTO catalog.product_images (
    product_id, url, alt_text, sort_order, is_primary, width_px, height_px
)
SELECT
    products.id,
    CASE categories.slug
        WHEN 'adjustable-dumbbells' THEN '/images/demo-adjustable-dumbbells.webp'
        WHEN 'benches' THEN '/images/demo-foldaway-bench.webp'
        WHEN 'power-racks' THEN '/images/demo-power-rack.webp'
        WHEN 'barbells' THEN '/images/demo-barbell.webp'
        WHEN 'weight-plates' THEN '/images/demo-weight-plates.webp'
        WHEN 'kettlebells' THEN '/images/demo-adjustable-kettlebell.webp'
        WHEN 'resistance-bands' THEN '/images/demo-resistance-bands.webp'
        WHEN 'cardio-machines' THEN '/images/demo-cardio-machine.webp'
    END,
    'Illustrative studio image for the fictional demo product ' || products.name,
    0,
    true,
    1000,
    750
FROM catalog.products AS products
JOIN catalog.categories AS categories ON categories.id = products.category_id
WHERE products.slug LIKE 'demo-%'
ON CONFLICT (product_id, url) DO UPDATE SET
    alt_text = EXCLUDED.alt_text,
    sort_order = EXCLUDED.sort_order,
    is_primary = EXCLUDED.is_primary,
    width_px = EXCLUDED.width_px,
    height_px = EXCLUDED.height_px;

WITH attribute_seed (product_slug, attribute_key, attribute_type, numeric_value, text_value, boolean_value, unit) AS (
    VALUES
        ('demo-northline-nest-24-pair', 'adjustment_steps', 'number', 12::numeric, NULL, NULL, 'count'),
        ('demo-quietform-dial-20-pair', 'adjustment_steps', 'number', 10::numeric, NULL, NULL, 'count'),
        ('demo-civic-select-16-pair', 'adjustment_steps', 'number', 8::numeric, NULL, NULL, 'count'),
        ('demo-range-lab-problock-32-pair', 'adjustment_steps', 'number', 16::numeric, NULL, NULL, 'count'),
        ('demo-oak-iron-foldaway-flat-bench', 'foldable', 'boolean', NULL, NULL, true, NULL),
        ('demo-northline-incline-five-bench', 'backrest_positions', 'number', 5::numeric, NULL, NULL, 'count'),
        ('demo-quietform-apartment-bench', 'vertical_storage', 'boolean', NULL, NULL, true, NULL),
        ('demo-summit-yard-heavy-seven-bench', 'backrest_positions', 'number', 7::numeric, NULL, NULL, 'count'),
        ('demo-northline-slim-half-rack', 'rack_hole_spacing', 'number', 50::numeric, NULL, NULL, 'mm'),
        ('demo-quietform-wall-fold-rack', 'foldable', 'boolean', NULL, NULL, true, NULL),
        ('demo-civic-compact-squat-stands', 'stand_type', 'text', NULL, 'independent', NULL, NULL),
        ('demo-summit-yard-full-cage-90', 'upright_size', 'text', NULL, '75x75 mm', NULL, NULL),
        ('demo-northline-compact-15-bar', 'sleeve_diameter', 'number', 50::numeric, NULL, NULL, 'mm'),
        ('demo-forgefield-standard-20-bar', 'sleeve_diameter', 'number', 50::numeric, NULL, NULL, 'mm'),
        ('demo-civic-technique-10-bar', 'sleeve_diameter', 'number', 50::numeric, NULL, NULL, 'mm'),
        ('demo-quietform-short-12-bar', 'sleeve_diameter', 'number', 50::numeric, NULL, NULL, 'mm'),
        ('demo-harbor-bumper-60-set', 'plate_type', 'text', NULL, 'bumper', NULL, NULL),
        ('demo-forgefield-iron-80-set', 'plate_type', 'text', NULL, 'cast_iron', NULL, NULL),
        ('demo-quietform-urethane-40-set', 'plate_type', 'text', NULL, 'urethane_coated', NULL, NULL),
        ('demo-northline-compact-50-set', 'plate_type', 'text', NULL, 'rubber_coated', NULL, NULL),
        ('demo-range-lab-adjustable-20-kettlebell', 'adjustment_steps', 'number', 6::numeric, NULL, NULL, 'count'),
        ('demo-kinetic-house-starter-band-set', 'resistance_levels', 'number', 5::numeric, NULL, NULL, 'count'),
        ('demo-northline-progressive-tube-set', 'stackable_resistance', 'boolean', NULL, NULL, true, NULL),
        ('demo-quietform-fabric-loop-set', 'resistance_levels', 'number', 3::numeric, NULL, NULL, 'count'),
        ('demo-pulseworks-folding-rower', 'foldable', 'boolean', NULL, NULL, true, NULL),
        ('demo-pulseworks-folding-rower', 'resistance_type', 'text', NULL, 'magnetic', NULL, NULL),
        ('demo-quietform-walking-pad', 'max_speed', 'number', 6::numeric, NULL, NULL, 'km/h'),
        ('demo-pulseworks-compact-air-bike', 'resistance_type', 'text', NULL, 'fan', NULL, NULL)
)
INSERT INTO catalog.product_attributes (
    product_id, attribute_key, attribute_type, numeric_value, text_value, boolean_value, unit
)
SELECT
    products.id,
    attribute_seed.attribute_key,
    attribute_seed.attribute_type,
    attribute_seed.numeric_value,
    attribute_seed.text_value,
    attribute_seed.boolean_value,
    attribute_seed.unit
FROM attribute_seed
JOIN catalog.products ON products.slug = attribute_seed.product_slug
ON CONFLICT (product_id, attribute_key) DO UPDATE SET
    attribute_type = EXCLUDED.attribute_type,
    numeric_value = EXCLUDED.numeric_value,
    text_value = EXCLUDED.text_value,
    boolean_value = EXCLUDED.boolean_value,
    unit = EXCLUDED.unit,
    updated_at = now();

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score)
VALUES
    ('Demo Iron Market', 'demo-iron-market', 'https://demo-iron-market.example.invalid', 'US', 86),
    ('Demo HomeFit Supply', 'demo-homefit-supply', 'https://demo-homefit-supply.example.invalid', 'US', 82),
    ('Demo Training Depot', 'demo-training-depot', 'https://demo-training-depot.example.invalid', 'US', 79)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    website_url = EXCLUDED.website_url,
    country_code = EXCLUDED.country_code,
    trust_score = EXCLUDED.trust_score,
    status = 'active',
    updated_at = now();

WITH merchant_terms (merchant_slug, price_percent, shipping_minor) AS (
    VALUES
        ('demo-iron-market', 98::bigint, 0::bigint),
        ('demo-homefit-supply', 95::bigint, 1299::bigint),
        ('demo-training-depot', 100::bigint, 699::bigint)
)
INSERT INTO commerce.merchant_offers (
    merchant_id,
    product_id,
    merchant_sku,
    product_url,
    price_minor,
    shipping_minor,
    currency,
    availability,
    condition,
    last_checked_at
)
SELECT
    merchants.id,
    products.id,
    upper(replace(merchants.slug, '-', '_')) || '__' || upper(replace(products.slug, '-', '_')),
    merchants.website_url || '/products/' || products.slug,
    (products.price_minor * merchant_terms.price_percent) / 100,
    merchant_terms.shipping_minor,
    products.currency,
    'in_stock',
    'new',
    now()
FROM catalog.products AS products
CROSS JOIN merchant_terms
JOIN commerce.merchants ON merchants.slug = merchant_terms.merchant_slug
WHERE products.slug LIKE 'demo-%'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    product_id = EXCLUDED.product_id,
    product_url = EXCLUDED.product_url,
    price_minor = EXCLUDED.price_minor,
    shipping_minor = EXCLUDED.shipping_minor,
    currency = EXCLUDED.currency,
    availability = EXCLUDED.availability,
    condition = EXCLUDED.condition,
    is_active = true,
    last_checked_at = EXCLUDED.last_checked_at,
    updated_at = now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id,
    provider,
    destination_url,
    external_reference,
    disclosure_label,
    is_active,
    priority,
    program_id,
    commission_type,
    commission_rate_bps
)
SELECT
    offers.id,
    'demo-affiliate-network',
    offers.product_url || '?ref=rigmark-demo',
    'DEMO-' || replace(offers.id::text, '-', ''),
    'Demo affiliate link',
    true,
    0,
    'rigmark-demo-program',
    'percentage',
    250
FROM commerce.merchant_offers AS offers
JOIN catalog.products AS products ON products.id = offers.product_id
WHERE products.slug LIKE 'demo-%'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url = EXCLUDED.destination_url,
    external_reference = EXCLUDED.external_reference,
    disclosure_label = EXCLUDED.disclosure_label,
    is_active = true,
    priority = EXCLUDED.priority,
    program_id = EXCLUDED.program_id,
    commission_type = EXCLUDED.commission_type,
    commission_rate_bps = EXCLUDED.commission_rate_bps,
    commission_amount_minor = NULL,
    commission_currency = NULL,
    updated_at = now();

INSERT INTO editorial.authors (name, slug, bio)
VALUES (
    'UNSOLERO Editorial',
    'rigmark-editorial',
    'UNSOLERO editors turn structured equipment facts into practical, constraint-aware home gym guidance.'
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    bio = EXCLUDED.bio,
    updated_at = now()
WHERE (editorial.authors.name, editorial.authors.bio)
    IS DISTINCT FROM (EXCLUDED.name, EXCLUDED.bio);

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'rigmark-editorial'
), editorial_seed (
    content_type, title, slug, description, hero_image_url, hero_image_alt,
    content, seo_title, seo_description, published_at
) AS (
    VALUES
        (
            'buying_guide',
            'The best adjustable dumbbells for different home gym constraints',
            'best-adjustable-dumbbells',
            'A practical framework for choosing fictional demo adjustable dumbbells by space, training range, noise, and budget.',
            '/images/demo-adjustable-dumbbells.webp',
            'A fictional adjustable dumbbell pair arranged in a restrained studio setting',
            $$[
              {"type":"paragraph","text":"Adjustable dumbbells solve a specific home-gym problem: they replace several fixed pairs with one compact system. The best choice is not automatically the model with the highest maximum weight. It is the one whose range, adjustment method, footprint, and handling match the way you train."},
              {"type":"heading","heading":"Begin with the constraint that cannot move"},
              {"type":"paragraph","text":"In a shared room, stored footprint and controlled plate movement may matter more than absolute capacity. In a dedicated garage, long-term loading range can reasonably take priority. Set the room and budget limits before comparing product scores."},
              {"type":"heading","heading":"What to compare"},
              {"type":"unordered_list","items":["Maximum load per hand and the size of each adjustment step","Length at lighter settings, because some systems remain full length","Noise and plate movement during controlled repetitions","Storage footprint and whether a stand is required","Warranty coverage and the materials used in adjustment components"]},
              {"type":"callout","heading":"How UNSOLERO evaluates the shortlist","text":"Our demo comparison uses structured specifications and suitability scores. It does not use customer-review claims, and affiliate commission never changes the ordering."},
              {"type":"heading","heading":"Choose for the next phase of training"},
              {"type":"paragraph","text":"A beginner may get more value from an accessible range and simple mechanism than unused maximum capacity. An experienced lifter should estimate whether the top weight will cover planned pressing, rowing, and lower-body work before committing."}
            ]$$::jsonb,
            'Best adjustable dumbbells by space and budget | UNSOLERO',
            'Compare adjustable dumbbell choices by training range, apartment fit, handling, and budget using structured demo product facts.',
            '2026-08-01T09:00:00Z'::timestamptz
        ),
        (
            'guide',
            'How to build a home gym for a small apartment',
            'best-home-gym-for-small-apartment',
            'A compact home gym planning guide built around storage, noise, movement coverage, and a staged equipment budget.',
            '/images/rigmark-home-gym-hero.webp',
            'Compact strength equipment arranged in a calm apartment training space',
            $$[
              {"type":"paragraph","text":"A useful apartment gym is not a smaller copy of a garage gym. It should cover the movements you actually train, return the room to normal quickly, and avoid equipment that creates more storage and noise problems than training value."},
              {"type":"heading","heading":"Measure the stored setup as well as the training area"},
              {"type":"paragraph","text":"Record the clear floor area available during a session, ceiling height, doorway width, and the exact storage location. A folding bench still needs a safe place to stand, and adjustable weights need enough clearance for loading and pickup."},
              {"type":"heading","heading":"Build movement coverage in layers"},
              {"type":"ordered_list","items":["Start with compact resistance that supports presses, rows, squats, hinges, and carries.","Add a stable bench only when supported pressing and rowing justify its stored footprint.","Use bands for warm-ups and accessory work rather than buying several single-purpose machines.","Delay racks, barbells, and plate storage until the room and training plan clearly support them."]},
              {"type":"heading","heading":"Treat noise as a product requirement"},
              {"type":"paragraph","text":"Floor protection, controlled repetitions, coated contact points, and a no-dropping rule usually matter more than marketing labels. UNSOLERO keeps noise suitability separate from quality so a strong garage product is not presented as an apartment fit."},
              {"type":"callout","heading":"A deliberately incomplete setup can be better","text":"Leaving budget unspent is useful when the remaining products would duplicate movement patterns or create a storage problem. Upgrade after your training reveals a real limitation."}
            ]$$::jsonb,
            'Small apartment home gym planning guide | UNSOLERO',
            'Plan a compact apartment home gym around real measurements, noise, storage, movement coverage, and a staged equipment budget.',
            '2026-08-04T09:00:00Z'::timestamptz
        ),
        (
            'comparison',
            'Demo Civic Select 16 vs Demo QuietForm Dial 20',
            'demo-civic-select-16-vs-demo-quietform-dial-20',
            'A transparent comparison of two fictional adjustable dumbbell systems for beginners and compact apartment setups.',
            '/images/demo-adjustable-dumbbells.webp',
            'Fictional adjustable dumbbells shown for a structured product comparison',
            $$[
              {"type":"paragraph","text":"These fictional demo products occupy a similar role but make different trade-offs. The Civic Select 16 emphasizes entry price, beginner suitability, and portability. The QuietForm Dial 20 adds loading range and a stronger quiet-use score at a higher reference price."},
              {"type":"heading","heading":"Where the Civic Select 16 fits"},
              {"type":"paragraph","text":"Its seeded reference price is lower, and its structured beginner and apartment scores make it the clearer starting point when budget and simple compact training dominate the brief."},
              {"type":"heading","heading":"Where the QuietForm Dial 20 fits"},
              {"type":"paragraph","text":"The additional load range gives more progression headroom, while the fictional product data indicates quieter handling. That advantage matters only if it justifies the higher cost for the user."},
              {"type":"callout","heading":"Decision rule","text":"Choose the Civic model when cost and beginner accessibility lead. Choose the QuietForm model when the extra range and quieter handling solve a specific constraint. Neither product is universally better."}
            ]$$::jsonb,
            'Civic Select 16 vs QuietForm Dial 20 | UNSOLERO',
            'Compare two fictional adjustable dumbbell systems using structured price, capacity, apartment, beginner, and noise suitability data.',
            '2026-08-07T09:00:00Z'::timestamptz
        ),
        (
            'article',
            'How to measure a room before buying home gym equipment',
            'how-to-measure-a-room-for-home-gym-equipment',
            'A measurement checklist that prevents clearance, storage, ceiling-height, and delivery problems in a home gym.',
            '/images/demo-foldaway-bench.webp',
            'A fictional folding training bench illustrating compact equipment planning',
            $$[
              {"type":"paragraph","text":"Product dimensions describe the equipment, not the space required to use it. Before building a shortlist, document the room as a training system: working area, circulation, storage, surfaces, access, and the people who share it."},
              {"type":"heading","heading":"Record six practical measurements"},
              {"type":"unordered_list","items":["Clear floor length and width after movable furniture is removed","Ceiling height at the exact training position","Door, stair, and hallway width along the delivery route","Distance to walls during the widest planned movement","Stored dimensions for anything described as folding or portable","Clearance around vents, radiators, outlets, and doors"]},
              {"type":"heading","heading":"Add movement clearance"},
              {"type":"paragraph","text":"A bar can fit inside a room and still be unusable near a wall. A bench can fit on paper and still block a door when positioned for training. Sketch the equipment at training size, then add the space required to approach, adjust, and safely leave it."},
              {"type":"callout","heading":"Keep the measurements with your brief","text":"Structured room constraints make product rejection explainable. They also prevent a later recommendation refinement from silently expanding beyond the available space."}
            ]$$::jsonb,
            'How to measure a room for home gym equipment | UNSOLERO',
            'Use this room-measurement checklist before comparing home gym equipment, including working clearance, storage size, and delivery access.',
            '2026-08-10T09:00:00Z'::timestamptz
        )
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description,
    published_at
)
SELECT
    author.id, editorial_seed.content_type, 'published', editorial_seed.title,
    editorial_seed.slug, editorial_seed.description, editorial_seed.hero_image_url,
    editorial_seed.hero_image_alt, editorial_seed.content, editorial_seed.seo_title,
    editorial_seed.seo_description, editorial_seed.published_at
FROM editorial_seed CROSS JOIN author
ON CONFLICT (slug) DO UPDATE SET
    author_id = EXCLUDED.author_id,
    content_type = EXCLUDED.content_type,
    status = 'published',
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    hero_image_url = EXCLUDED.hero_image_url,
    hero_image_alt = EXCLUDED.hero_image_alt,
    content = EXCLUDED.content,
    seo_title = EXCLUDED.seo_title,
    seo_description = EXCLUDED.seo_description,
    published_at = EXCLUDED.published_at,
    updated_at = now()
WHERE (
    editorial.entries.author_id,
    editorial.entries.content_type,
    editorial.entries.status,
    editorial.entries.title,
    editorial.entries.description,
    editorial.entries.hero_image_url,
    editorial.entries.hero_image_alt,
    editorial.entries.content,
    editorial.entries.seo_title,
    editorial.entries.seo_description,
    editorial.entries.published_at
) IS DISTINCT FROM (
    EXCLUDED.author_id,
    EXCLUDED.content_type,
    EXCLUDED.status,
    EXCLUDED.title,
    EXCLUDED.description,
    EXCLUDED.hero_image_url,
    EXCLUDED.hero_image_alt,
    EXCLUDED.content,
    EXCLUDED.seo_title,
    EXCLUDED.seo_description,
    EXCLUDED.published_at
);

DELETE FROM editorial.entry_products
WHERE entry_id IN (
    SELECT id FROM editorial.entries
    WHERE slug IN (
        'best-adjustable-dumbbells',
        'best-home-gym-for-small-apartment',
        'demo-civic-select-16-vs-demo-quietform-dial-20',
        'how-to-measure-a-room-for-home-gym-equipment'
    )
);

WITH relationships (entry_slug, product_slug, position) AS (
    VALUES
        ('best-adjustable-dumbbells', 'demo-civic-select-16-pair', 0),
        ('best-adjustable-dumbbells', 'demo-quietform-dial-20-pair', 1),
        ('best-adjustable-dumbbells', 'demo-northline-nest-24-pair', 2),
        ('best-home-gym-for-small-apartment', 'demo-civic-select-16-pair', 0),
        ('best-home-gym-for-small-apartment', 'demo-oak-iron-foldaway-flat-bench', 1),
        ('best-home-gym-for-small-apartment', 'demo-kinetic-house-starter-band-set', 2),
        ('demo-civic-select-16-vs-demo-quietform-dial-20', 'demo-civic-select-16-pair', 0),
        ('demo-civic-select-16-vs-demo-quietform-dial-20', 'demo-quietform-dial-20-pair', 1)
)
INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT entries.id, products.id, relationships.position
FROM relationships
JOIN editorial.entries AS entries ON entries.slug = relationships.entry_slug
JOIN catalog.products AS products ON products.slug = relationships.product_slug;

DELETE FROM editorial.entry_categories
WHERE entry_id IN (
    SELECT id FROM editorial.entries
    WHERE slug IN (
        'best-adjustable-dumbbells',
        'best-home-gym-for-small-apartment',
        'demo-civic-select-16-vs-demo-quietform-dial-20',
        'how-to-measure-a-room-for-home-gym-equipment'
    )
);

WITH relationships (entry_slug, category_slug, position) AS (
    VALUES
        ('best-adjustable-dumbbells', 'adjustable-dumbbells', 0),
        ('best-home-gym-for-small-apartment', 'adjustable-dumbbells', 0),
        ('best-home-gym-for-small-apartment', 'benches', 1),
        ('best-home-gym-for-small-apartment', 'resistance-bands', 2),
        ('demo-civic-select-16-vs-demo-quietform-dial-20', 'adjustable-dumbbells', 0),
        ('how-to-measure-a-room-for-home-gym-equipment', 'power-racks', 0),
        ('how-to-measure-a-room-for-home-gym-equipment', 'benches', 1),
        ('how-to-measure-a-room-for-home-gym-equipment', 'cardio-machines', 2)
)
INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, relationships.position
FROM relationships
JOIN editorial.entries AS entries ON entries.slug = relationships.entry_slug
JOIN catalog.categories AS categories ON categories.slug = relationships.category_slug;

DELETE FROM editorial.related_entries
WHERE entry_id IN (
    SELECT id FROM editorial.entries
    WHERE slug IN (
        'best-adjustable-dumbbells',
        'best-home-gym-for-small-apartment',
        'demo-civic-select-16-vs-demo-quietform-dial-20',
        'how-to-measure-a-room-for-home-gym-equipment'
    )
);

WITH relationships (entry_slug, related_slug, position) AS (
    VALUES
        ('best-adjustable-dumbbells', 'demo-civic-select-16-vs-demo-quietform-dial-20', 0),
        ('best-adjustable-dumbbells', 'best-home-gym-for-small-apartment', 1),
        ('best-home-gym-for-small-apartment', 'how-to-measure-a-room-for-home-gym-equipment', 0),
        ('best-home-gym-for-small-apartment', 'best-adjustable-dumbbells', 1),
        ('demo-civic-select-16-vs-demo-quietform-dial-20', 'best-adjustable-dumbbells', 0),
        ('how-to-measure-a-room-for-home-gym-equipment', 'best-home-gym-for-small-apartment', 0)
)
INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT entries.id, related.id, relationships.position
FROM relationships
JOIN editorial.entries AS entries ON entries.slug = relationships.entry_slug
JOIN editorial.entries AS related ON related.slug = relationships.related_slug;
-- Product governance is seeded last so fictional products are never made public
-- without explicit, auditable fictional provenance.
WITH source AS (
    INSERT INTO evidence.sources (
        external_key, source_type, title, publisher, is_fictional,
        review_status, reviewed_at, review_note
    ) VALUES (
        'unsolero-demo-fixture-v1', 'demo_fixture',
        'Fictional UNSOLERO development fixture', 'UNSOLERO development seed',
        true, 'verified', now(),
        'Fictional evidence for local development only; not a real-world source.'
    )
    ON CONFLICT (external_key) DO UPDATE SET
        title = EXCLUDED.title, publisher = EXCLUDED.publisher,
        is_fictional = true, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id
)
INSERT INTO evidence.observations (
    source_id, product_id, observed_at, confidence, notes
)
SELECT source.id, products.id, now(), 100,
       'Fictional demo observation; no real product claim is made.'
FROM source CROSS JOIN catalog.products AS products
WHERE products.slug LIKE 'demo-%'
  AND NOT EXISTS (
      SELECT 1 FROM evidence.observations AS existing
      WHERE existing.source_id = source.id AND existing.product_id = products.id
  );

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
    max_capacity_grams, material, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, 1, products.category_id, products.brand_id, products.name,
       products.slug, products.description, products.price_minor, products.currency,
       products.length_mm, products.width_mm, products.height_mm, products.weight_grams,
       products.max_capacity_grams, products.material, products.warranty_months,
       'published', now(), now(), now(),
       'Fictional development fixture approved for local demonstration only.'
FROM catalog.products AS products
WHERE products.slug LIKE 'demo-%'
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score,
    durability_score, beginner_score, advanced_score, apartment_score,
    noise_score, portability_score, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, facts.id, 1, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score, products.advanced_score,
       products.apartment_score, products.noise_score, products.portability_score,
       'published', now(), now(), now(),
       'Fictional development score fixture; not a real-world assessment.'
FROM catalog.products AS products
JOIN evidence.product_fact_revisions AS facts
  ON facts.product_id = products.id AND facts.version = 1
WHERE products.slug LIKE 'demo-%'
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (
    fact_revision_id, fact_key, observation_id, public_classification
)
SELECT facts.id, keys.fact_key, observations.id, 'editorial_assessment'
FROM evidence.product_fact_revisions AS facts
JOIN catalog.products AS products ON products.id = facts.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
JOIN evidence.sources AS sources
  ON sources.id = observations.source_id AND sources.external_key = 'unsolero-demo-fixture-v1'
CROSS JOIN (VALUES
    ('category'), ('brand'), ('name'), ('slug'), ('description'), ('price'),
    ('dimensions'), ('weight'), ('max_capacity'), ('material'), ('warranty')
) AS keys(fact_key)
WHERE products.slug LIKE 'demo-%' AND facts.version = 1
  AND (keys.fact_key <> 'max_capacity' OR products.max_capacity_grams IS NOT NULL)
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (
    score_revision_id, score_key, rationale, observation_id
)
SELECT scores.id, keys.score_key,
       'Fictional demo score supplied by the development fixture.', observations.id
FROM evidence.score_revisions AS scores
JOIN catalog.products AS products ON products.id = scores.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
JOIN evidence.sources AS sources
  ON sources.id = observations.source_id AND sources.external_key = 'unsolero-demo-fixture-v1'
CROSS JOIN (VALUES
    ('quality'), ('value'), ('durability'), ('beginner'),
    ('advanced'), ('apartment'), ('noise'), ('portability')
) AS keys(score_key)
WHERE products.slug LIKE 'demo-%' AND scores.version = 1
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores
  ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug LIKE 'demo-%'
  AND facts.version = 1
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';

-- Data-driven recommendation policy fixture. This is intentionally limited to
-- the fictional development catalog; new categories/products remain excluded
-- until explicitly configured and reviewed.
-- Preference and priority vocabulary for the fictional fitness policy. These
-- moved out of Go constants into policy data, and the fitness vertical is demo
-- data, so its vocabulary belongs here rather than in a migration. The draft
-- guard matches every other policy insert below: an activated policy is
-- immutable and must not be edited in place.
INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'fitness-v2', tags.tag_key, tags.label
FROM (VALUES
 ('dumbbells','Dumbbells'), ('barbell','Barbell'), ('kettlebell','Kettlebell'),
 ('resistance_bands','Resistance bands'), ('bodyweight','Bodyweight'),
 ('cardio','Cardio'), ('low_impact','Low impact')
) AS tags(tag_key, label)
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities
 (policy_version, priority_key, label, reason_code, reason_message, reason_dimension, reason_threshold, sort_order)
SELECT 'fitness-v2', priorities.priority_key, priorities.label, priorities.reason_code,
       priorities.reason_message, priorities.reason_dimension, priorities.reason_threshold, priorities.sort_order
FROM (VALUES
 ('budget','Budget','priority.value','Strong value for the available budget','value',85,0),
 ('compact','Compact','priority.compact','Uses your available space efficiently','space_match',85,1),
 ('quality','Quality','priority.quality','Strong structured quality score','quality',85,2),
 ('durability','Durability','priority.durability','Strong structured durability score','durability',85,3),
 ('quiet','Quiet','priority.quiet','Well suited to quieter training','noise',85,4),
 ('portability','Portability','priority.portable','Easy to move or store','portability',85,5)
) AS priorities(priority_key, label, reason_code, reason_message, reason_dimension, reason_threshold, sort_order)
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'fitness-v2', dimensions.priority_key, dimensions.dimension
FROM (VALUES
 ('budget','budget_match'), ('budget','value'), ('compact','space_match'),
 ('quality','quality'), ('durability','durability'), ('quiet','noise'),
 ('portability','portability')
) AS dimensions(priority_key, dimension)
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (policy_version,category_id,support_status)
SELECT 'fitness-v2', id, 'supported' FROM catalog.categories
WHERE slug IN ('adjustable-dumbbells','benches','power-racks','barbells','weight-plates','kettlebells','resistance-bands','cardio-machines')
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT (policy_version,category_id) DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups
SELECT 'fitness-v2', id, CASE slug
 WHEN 'adjustable-dumbbells' THEN 'dumbbell_system' WHEN 'benches' THEN 'bench'
 WHEN 'power-racks' THEN 'rack' WHEN 'barbells' THEN 'barbell'
 WHEN 'weight-plates' THEN 'weight_plates' WHEN 'kettlebells' THEN 'kettlebell_system'
 WHEN 'resistance-bands' THEN 'resistance_band_system' ELSE 'cardio_machine' END
FROM catalog.categories WHERE slug IN ('adjustable-dumbbells','benches','power-racks','barbells','weight-plates','kettlebells','resistance-bands','cardio-machines')
AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

WITH mappings(slug, capability) AS (VALUES
 ('adjustable-dumbbells','resistance_training'),('adjustable-dumbbells','strength_training'),('adjustable-dumbbells','hypertrophy'),
 ('benches','supported_training'),('power-racks','safe_barbell_training'),('power-racks','pull_up'),('power-racks','anchor_point'),
 ('barbells','barbell_training'),('barbells','strength_training'),('barbells','hypertrophy'),('weight-plates','weight_plates'),
 ('kettlebells','resistance_training'),('kettlebells','strength_training'),('kettlebells','hypertrophy'),('kettlebells','conditioning'),
 ('resistance-bands','resistance_training'),('resistance-bands','hypertrophy'),('resistance-bands','conditioning'),('resistance-bands','mobility'),
 ('cardio-machines','conditioning'))
INSERT INTO recommendation.category_policy_capabilities
SELECT 'fitness-v2', categories.id, mappings.capability FROM mappings
JOIN catalog.categories categories ON categories.slug=mappings.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policies
SELECT 'fitness-v2', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products WHERE slug LIKE 'demo-%' AND published_fact_revision_id IS NOT NULL AND published_score_revision_id IS NOT NULL
AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT (policy_version,product_id) DO NOTHING;

INSERT INTO recommendation.product_space_profiles (policy_version,product_id,footprint_length_mm,footprint_width_mm,footprint_height_mm)
SELECT 'fitness-v2', id, length_mm, width_mm, height_mm FROM catalog.products WHERE slug LIKE 'demo-%'
AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT (policy_version,product_id) DO NOTHING;

WITH mappings(slug, capability, relation_type) AS (VALUES
 ('barbells','weight_plates','requires'),('weight-plates','barbell_training','requires'),('power-racks','barbell_training','requires'),
 ('benches','resistance_training','compatible_with'),('benches','barbell_training','compatible_with'),
 ('resistance-bands','anchor_point','compatible_with'))
INSERT INTO recommendation.product_policy_capabilities
SELECT 'fitness-v2', products.id, mappings.capability, mappings.relation_type FROM mappings
JOIN catalog.categories categories ON categories.slug=mappings.slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%'
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

WITH goal_scores(category_slug,build_muscle,strength,general_fitness,weight_loss,mobility) AS (VALUES
 ('adjustable-dumbbells',100,88,92,76,50),('benches',92,84,76,52,54),('power-racks',86,100,62,42,35),
 ('barbells',94,100,70,50,35),('weight-plates',92,100,65,45,30),('kettlebells',84,86,96,92,65),
 ('resistance-bands',76,62,94,86,100),('cardio-machines',40,42,90,100,58)), expanded AS (
 SELECT category_slug, unnest(ARRAY['build_muscle','strength','general_fitness','weight_loss','mobility']) goal_key,
        unnest(ARRAY[build_muscle,strength,general_fitness,weight_loss,mobility]) score FROM goal_scores)
INSERT INTO recommendation.product_goal_support
SELECT 'fitness-v2', products.id, expanded.goal_key, expanded.score FROM expanded
JOIN catalog.categories categories ON categories.slug=expanded.category_slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%'
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT (policy_version,product_id,goal_key) DO NOTHING;

WITH tags(slug, preference) AS (VALUES
 ('adjustable-dumbbells','dumbbells'),('power-racks','barbell'),('power-racks','bodyweight'),
 ('barbells','barbell'),('weight-plates','barbell'),('kettlebells','kettlebell'),
 ('resistance-bands','resistance_bands'),('resistance-bands','bodyweight'),('cardio-machines','cardio'))
INSERT INTO recommendation.product_preference_tags
SELECT 'fitness-v2', products.id, tags.preference FROM tags
JOIN catalog.categories categories ON categories.slug=tags.slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%'
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags
SELECT 'fitness-v2', id, 'low_impact' FROM catalog.products
WHERE slug LIKE 'demo-%' AND noise_score >= 85
AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='fitness-v2' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

UPDATE recommendation.policy_versions SET workflow_status='active',activated_at=now()
WHERE version='fitness-v2' AND workflow_status='draft'
AND EXISTS (SELECT 1 FROM recommendation.category_policies WHERE policy_version='fitness-v2' AND support_status='supported')
AND EXISTS (SELECT 1 FROM recommendation.product_policies WHERE policy_version='fitness-v2');
