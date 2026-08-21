-- Sharpen the Shopify transaction-fee claim.
--
-- The published version said Shopify charges a platform fee unless you use
-- Shopify Payments, which is true but vague, and vague is how a comparison
-- site stops being useful. The figure was then read from Shopify's own pricing
-- page: 2% on the Basic plan for third-party payment providers. A number a
-- reader can act on, instead of a warning they have to go and price themselves.

UPDATE editorial.entries
SET content = jsonb_set(
        content,
        '{0}',
        '{"type":"paragraph","text":"Shopify Basic and BigCommerce Core both cost 29 USD a month. They are not the same 29 dollars. BigCommerce charges no platform transaction fee on top when you use a supported payment provider. Shopify Basic charges 2% on every order taken through a payment provider other than Shopify Payments, on top of the card rate you are already paying. On any real volume that line matters more than the subscription."}'::jsonb
    ),
    updated_at = now()
WHERE slug = 'shopify-vs-bigcommerce'
  AND content -> 0 ->> 'type' = 'paragraph';

UPDATE editorial.entries
SET content = jsonb_set(
        content,
        '{7}',
        '{"type":"paragraph","text":"Shopify''s equivalent is the payment processor. Use Shopify Payments and there is no platform fee. Use anything else on Basic and Shopify takes 2% of every order, forever, on top of whatever your processor charges. If you are committed to a processor Shopify does not own, add that two percent to the subscription before you compare the two platforms at all."}'::jsonb
    ),
    updated_at = now()
WHERE slug = 'shopify-vs-bigcommerce'
  AND content -> 7 ->> 'type' = 'paragraph';
