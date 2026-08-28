# Affiliate programme status

The single answer to "which programmes have accepted us, and which link is
live where". Everything else about affiliate links — the per-vendor detail, the
audit method, the application playbook — hangs off this page.

Last reconciled against the repository on **2026-08-28**.

Two things this page deliberately does not do. It does not follow an affiliate
URL: provider terms prohibit automated or artificial clicks, and a redirect
request can itself be recorded as one. And it does not claim revenue. A live
link proves UNSOLERO routes a genuine visitor correctly; only the partner
dashboard proves provider-side attribution, conversion, and payment.

## Summary

| | Count |
| --- | ---: |
| Brands in the SaaS catalog | 46 |
| Brands with an approved, active affiliate link | 6 |
| Live merchant offers behind those links | 12 |
| Approved and seeded, pending application | 2 (Pipedrive, Teachable) |
| Standalone promotions staged but not deployed | 2 (ClickFunnels) |

Six of forty-six is the number that matters. Forty brands are in the catalog
because the ranking has to be honest about the market, not because they pay.
That is the design — but it also means most of the catalog cannot earn, and
adding a programme is the cheapest growth lever available.

## Approved and live

Twelve offers. Each is asserted by
`backend/seeds/affiliate_offer_audit_2026_08_26.sql`, which fails its
transaction unless the product, price, currency, provider, and destination all
match exactly.

| Product | Price basis | Provider | Ownership evidence |
| --- | --- | --- | --- |
| Bigin Express | $9/user/month | Zoho | affiliate ID `PE2263909`, link code `SsgT` |
| Cal.com Teams | $12/user/month, yearly | Cal.com | referral code `unsolero-wtpd` |
| Kit Creator | $39/month at 1,000 subscribers | PartnerStack | copied from approved dashboard |
| MailerLite Comfort | $19/month at 1,000 subscribers | Trackdesk | source `unsolero`, link `lp_170762` |
| monday.com Basic | $9/seat/month, yearly | PartnerStack | copied from approved dashboard |
| SE Ranking Core | $103.20/month | SE Ranking | link-builder account `5233991` |
| Zoho Bookings Basic | $8/user/month | Zoho | `PE2263909`, code `POSi` |
| Zoho Books Standard | $20/organization/month | Zoho | `PE2263909`, code `K0nf` |
| Zoho Campaigns Standard | $5.25/month at 1,000 contacts, yearly | Zoho | `PE2263909`, code `UCST` |
| Zoho CRM Standard | $20/user/month | Zoho | `PE2263909`, code `dNbV` |
| Zoho Invoice | Free | Zoho | `PE2263909`, code `dIhI` |
| Zoho Projects Premium | $4/user/month, yearly | Zoho | `PE2263909`, code `PCoS` |

Note that PartnerStack appears twice as a *provider* while the network itself
sits outside this table. A programme reached through PartnerStack is approved
independently of the network account; a problem with the network account does
not retract Kit or monday.com.

### Where each programme lives in the repository

| Programme | Seed | Notes |
| --- | --- | --- |
| Zoho | `backend/seeds/saas_offers.sql`, `backend/seeds/zoho_books_price_update_2026_08_26.sql` | detail in [affiliate-links-zoho.md](./affiliate-links-zoho.md) |
| Cal.com | `backend/seeds/cal_com_affiliate.sql` | |
| Kit | `backend/seeds/kit_affiliate.sql` | via PartnerStack |
| MailerLite | `backend/seeds/mailerlite_affiliate.sql` | detail in [affiliate-links-mailerlite.md](./affiliate-links-mailerlite.md) |
| monday.com | `backend/seeds/monday_affiliate.sql` | via PartnerStack |
| SE Ranking | `backend/seeds/se_ranking_affiliate.sql` | |
| all twelve | `backend/seeds/affiliate_offer_audit_2026_08_26.sql` | the assertion that keeps them honest |

## Approved and seeded, not yet applied to the database

**Pipedrive** — PartnerStack, default link created by the programme
2026-08-23. Seed: `backend/seeds/pipedrive_affiliate.sql`, attached to
`pipedrive-lite` at $19.90/seat/month.

Two caveats are recorded in the seed's own header and repeated here because
both affect money:

- The destination is `www.pipedrive.com/programLP`, an affiliate programme
  landing page rather than the pricing page. Every other link here targets
  pricing deliberately. Pipedrive's dashboard offers no custom links and the
  default cannot be regenerated, so ask `affiliates@pipedrive.com` for a
  pricing destination before treating this as settled.
- The partner key is never shown — the destination template carries the literal
  `{partner_key}`, filled by PartnerStack at redirect time. The external
  reference is therefore the link code itself.

**Teachable** — PartnerStack, 30% for one year, link
`partnerstack.teachable.com/y6u7cxavunjg`, destination already set by the
programme to `teachable.com/pricing`. Seed:
`backend/seeds/teachable_affiliate.sql`, attached to `teachable-starter`.

Its price was re-read before the link was attached, recorded in
`backend/seeds/course_platform_price_review_2026_08_28.sql`:

- **Teachable Starter: $39/month monthly, $29/month billed annually**, plus the
  7.5% Starter transaction fee. The catalog's $29 is the correct annual figure
  and is not stale.
- **Teachable's own JSON-LD contradicts its own page**, declaring $59 monthly
  and $39 yearly — figures matching no control on the page. Anything reading
  the markup rather than the page gets numbers no customer pays.
- **Thinkific Basic could not be read in USD.** The page geolocates and served
  EUR only (€53 monthly, €39 annually). Its source is recorded at `pending`
  with confidence 40 so the gap stays visible.

Neither price was moved to its monthly basis, and the reasoning is in the seed
header: both are recorded annually, so within the category the comparison is
already like for like. Moving Teachable alone would put it a dollar from
Thinkific on screen while the real monthly gap is far wider — replacing a
correct comparison with a misleading one in order to satisfy a rule.

## The billing-basis defect

The rule is right and the catalog is inconsistent with it, but not one product
at a time. `catalog.products` has `price_minor` and `currency` and **no column
for the billing basis**. That basis exists only as prose in the description.

**Twenty-five published products state an annual or yearly basis in their
description while the engine compares their `price_minor` values as though all
were alike.** A product priced at its annual rate scores cheaper than a rival
priced monthly, and nothing in the data marks the difference. Among them:
Shopify Basic, BigCommerce Core, Framer, Webflow, Ecwid, Figma, Sketch,
Semrush, SE Ranking, Zapier, n8n, Make, monday.com, Cal.com, ActiveCampaign,
Microsoft Teams — several already carrying affiliate links.

This is the largest correctness issue currently open in the catalog. Fixing it
means a migration adding an explicit billing period, a backfill, a policy
revision in the shape of the Zoho Books correction, and surfacing the basis
wherever a price is shown. It is not started.

## Approved, staged, not yet deployed

**ClickFunnels — Funnel Hacking Secrets**, affiliate ID `4330879`, two
destinations copied from the approved dashboard on 2026-08-26.

These are *promotions*, not merchant offers. A free training and its order form
are not the ClickFunnels software subscription, and representing them as a
catalog product would let promotional economics reach recommendation scoring.
Migration `000026_affiliate_promotions.sql` gives them their own tables, their
own redirect path, and no product or recommendation columns at all.

| Slug | Type | Public page |
| --- | --- | --- |
| `funnel-hacking-secrets-webinar` | lead | funnelhackingsecrets.com |
| `funnel-hacking-secrets-order` | purchase | funnelhackingsecrets.com/go |

Everything is written and green — migration, repository, service, handler,
route, public page at `/offers/funnel-hacking-secrets`, tests. It is not
deployed: `https://unsolero.com/offers/funnel-hacking-secrets` returns 404 as of
2026-08-28 because the work is uncommitted on `fix/ci-and-faceless-guide`.

## Not applied, or unverified

The remaining forty catalog brands have no affiliate relationship. Of the ones
whose programmes were researched:

| Brand | Programme page | Status |
| --- | --- | --- |
| ClickUp | `clickup.com/partners/affiliates` | resolves; not applied |
| Pipedrive | via PartnerStack | **approved 2026-08-23** — see above |
| Teachable | via PartnerStack | **approved 2026-08-21, seeded** — see above |
| HubSpot | `hubspot.com/partners/affiliates` | resolves; pays through Impact |
| FreshBooks | `freshbooks.com/affiliates` | resolves; not applied |
| Teamwork | `teamwork.com/partners/` | resolves; not applied |
| Salesflare | — | **programme page withdrawn.** `/affiliates`, `/affiliate-program`, `/partner` and `/referral` all 404 as of 2026-08-28. Do not substitute a guessed address; ask their sales team. |

The application playbook — order, what reviewers check, payout thresholds, and
the Bulgarian tax position — is in
[LAUNCH_CHECKLIST.md](./LAUNCH_CHECKLIST.md), step 6.

## Adding a newly approved programme

What is needed per link, and why each part is needed:

1. **The exact destination URL, copied from the programme dashboard.** Not
   retyped, not shortened, not decoded and rebuilt. Every tracking parameter
   is load-bearing, and a transcribed character is how a link ends up crediting
   another affiliate — Zoho's portal font renders `I`, `l` and `1`
   identically, and that has already cost one link once.
2. **Which product it sells,** by catalog slug, so it attaches to a merchant
   offer rather than floating free. If the product is not in the catalog yet,
   it needs a catalog entry with priced evidence first.
3. **The list price and its billing basis** — monthly or annual — because
   UNSOLERO compares monthly-billing prices and mixing the two silently
   misranks the product.
4. **The provider name** (`zoho`, `partnerstack`, `trackdesk`, …) and the
   **account identifier** that proves the link is ours.
5. **Whether it is an offer or a promotion.** An offer sells a catalog product.
   A promotion is editorial — a webinar, a book, a bundle — and must go through
   the promotions path so it cannot enter scoring.

Given those, the change is: a seed file beside the ones above, an entry in the
audit assertion, and a row in this table. Nothing in the ranking engine
changes, because commission is not one of its inputs and never will be.

## Related

- [AFFILIATE_LINK_AUDIT_2026_08_26.md](./AFFILIATE_LINK_AUDIT_2026_08_26.md) — the audit method, and why no affiliate URL is ever followed
- [affiliate-links-zoho.md](./affiliate-links-zoho.md) — the 91-row Zoho reconciliation
- [affiliate-links-mailerlite.md](./affiliate-links-mailerlite.md) — the Trackdesk link and its parameters
- [MERCHANT_INTEGRATION.md](./MERCHANT_INTEGRATION.md) — how offers and links are modelled
- [LAUNCH_CHECKLIST.md](./LAUNCH_CHECKLIST.md) — applying to programmes not yet approved
- [../BUSINESS_MODEL.md](../BUSINESS_MODEL.md) — why commission is excluded from ranking
