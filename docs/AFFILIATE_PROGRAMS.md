# Affiliate programme status

The single answer to "which programmes have accepted us, and which link is
live where". Everything else about affiliate links — the per-vendor detail, the
audit method, the application playbook — hangs off this page.

Last reconciled against the repository on **2026-08-29**.

Two things this page deliberately does not do. It does not follow an affiliate
URL: provider terms prohibit automated or artificial clicks, and a redirect
request can itself be recorded as one. And it does not claim revenue. A live
link proves UNSOLERO routes a genuine visitor correctly; only the partner
dashboard proves provider-side attribution, conversion, and payment.

## Summary

| | Count |
| --- | ---: |
| Brands in the SaaS catalog | 46 |
| Brands with an approved, active affiliate link | 8 |
| Live merchant offers behind those links | 14 |
| Approved and seeded, pending application | 1 (ActiveCampaign) |
| Standalone promotions deployed | 2 (ClickFunnels) |

Eight of forty-six is the number that matters. Thirty-eight brands are in the catalog
because the ranking has to be honest about the market, not because they pay.
That is the design — but it also means most of the catalog cannot earn, and
adding a programme is the cheapest growth lever available.

## Approved and live

Fourteen offers, all confirmed against the production API on 2026-08-29 —
`/api/catalog/products/{slug}/offers` returns a `purchase_path` only when the
offer carries an active affiliate link, so a non-null value is proof the link
is live rather than merely seeded.

Twelve of the fourteen are additionally asserted by
`backend/seeds/affiliate_offer_audit_2026_08_26.sql`, which fails its
transaction unless the product, price, currency, provider, and destination all
match exactly. Pipedrive and Teachable are not in it: that file is a dated
audit, and each of their seeds carries its own assertion block instead.

| Product | Price basis | Provider | Ownership evidence |
| --- | --- | --- | --- |
| Bigin Express | $9/user/month | Zoho | affiliate ID `PE2263909`, link code `SsgT` |
| Cal.com Teams | $12/user/month, yearly | Cal.com | referral code `unsolero-wtpd` |
| Kit Creator | $39/month at 1,000 subscribers | PartnerStack | copied from approved dashboard |
| MailerLite Comfort | $19/month at 1,000 subscribers | Trackdesk | source `unsolero`, link `lp_170762` |
| monday.com Basic | $9/seat/month, yearly | PartnerStack | copied from approved dashboard |
| Pipedrive Lite | $19.90/seat/month | PartnerStack | link code `8c0fqmk2j8mc` |
| SE Ranking Core | $103.20/month | SE Ranking | link-builder account `5233991` |
| Teachable Starter | $29/month, yearly | PartnerStack | link code `y6u7cxavunjg` |
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
| Pipedrive | `backend/seeds/pipedrive_affiliate.sql` | via PartnerStack; destination caveat below |
| SE Ranking | `backend/seeds/se_ranking_affiliate.sql` | |
| Teachable | `backend/seeds/teachable_affiliate.sql` | via PartnerStack, 30% for one year |
| the original twelve | `backend/seeds/affiliate_offer_audit_2026_08_26.sql` | the assertion that keeps them honest |

### Caveats on the two newest live links

Pipedrive and Teachable went live on 2026-08-28. Both carry open items that
this page tracked while they were still staged, and neither was closed by
deploying them.

**Pipedrive** — PartnerStack, default link created by the programme
2026-08-23. Seed: `backend/seeds/pipedrive_affiliate.sql`, attached to
`pipedrive-lite` at $19.90/seat/month.

Two caveats are recorded in the seed's own header and repeated here because
both affect money, and the first is still costing conversions today:

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
No open item; recorded here because its price review is worth keeping.

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

## Approved and seeded, not yet applied to the database

**ActiveCampaign** — PartnerStack, approved and read from the partner dashboard
2026-08-29. The only programme currently in this state: the seed is written and
`/api/catalog/products/activecampaign-starter/offers` returns `[]` in
production, so nothing earns until it is applied. Seed:
`backend/seeds/activecampaign_affiliate.sql`, attached to
`activecampaign-starter` at $15/month for 1,000 contacts, billed annually.

The dashboard publishes three default links. Only one is used, and only one
*could* be: `commerce.affiliate_links` is unique on
`(merchant_offer_id, provider)`, and `ResolveOfferDestination` returns a single
row per offer. A second link for the same product has nowhere to live and
nothing that would serve it.

| Dashboard link | Destination | Used |
| --- | --- | --- |
| Pricing Page | `try.activecampaign.com/ydwpszsmins9` | **yes** |
| Free Trial Page | `try.activecampaign.com/egv7yratfy4m-rvs4jt` | held in reserve |
| Affiliate Homepage | `try.activecampaign.com/4iq2pjt98jsg-b9q17i` | no |

Pricing wins on the same reasoning as Kit: a visitor who has just read the
price is looking for that plan. The trial link is the only real alternative,
since ActiveCampaign Starter has no free tier — but the pricing page carries
its own trial call to action, so pricing reaches the trial in one more click
while the trial page never shows the price. Keep the trial link for editorial
placements where no price has been shown.

As with Pipedrive and Teachable, the partner key is never exposed — PartnerStack
fills it at redirect time — so the external reference is the link code.

Every code on this page came from the dashboard's copy control, confirmed by
the account owner on 2026-08-29 — not read off the screen, which is the failure
[affiliate-links-zoho.md](./affiliate-links-zoho.md) exists to warn about.

**Commission: read, but deliberately not written to the columns.** The
programme's published terms are a 90-day cookie that restarts on each click,
and commission recurring for the first twelve months of the referred
subscription. The rate is where it stops being a single number: ActiveCampaign
tiers it — Silver 20%, Gold 25%, Platinum 30% — on referred MRR, so a new
partner starts at 20% and the 30% that most summaries quote is the top of a
ladder, not the entry rate. PartnerStack's own programme card states 30% flat,
which contradicts the tier table.

`commission_rate_bps` holds one integer, and there is no honest single value
here until the tier showing in our own dashboard is read. Writing 30% because a
directory says so would put a number in a money column that we have not seen
and are almost certainly not on. The columns stay NULL, and the seed header
says why. Read the tier from PartnerStack and fill them in the shape of
`mailerlite_affiliate.sql`.

### Two further ActiveCampaign links, activated and unassigned

Two of the seven links ActiveCampaign recommends were activated on 2026-08-29
and copied from the dashboard:

- `https://try.activecampaign.com/am4yesxqhxo9-c8qk4`
- `https://try.activecampaign.com/yb5i7jsind0c-txqy7s`

The pairing, confirmed by the account owner on 2026-08-29:

| Link | Landing page |
| --- | --- |
| `try.activecampaign.com/am4yesxqhxo9-c8qk4` | MailChimp Switch |
| `try.activecampaign.com/yb5i7jsind0c-txqy7s` | CRM |

**MailChimp Switch is built and seeded** as
`backend/seeds/activecampaign_mailchimp_switch.sql`. It goes in as a
*promotion*, not an offer: the offer slot on `activecampaign-starter` already
holds the pricing link, and more importantly the destination is a vendor-level
page for people leaving Mailchimp, tied to no tier and no price — attaching it
to a $15 plan would claim it sells something it never mentions. Its
unaffiliated equivalent, recorded so anyone can see where it goes without
following it, is `activecampaign.com/compare/mailchimp`, read 2026-08-29.

Rather than a standalone page nobody lands on, it appears inside the already
published `mailchimp-alternatives` guide, immediately after the line that tells
an automation-driven reader ActiveCampaign is their answer. The seed asserts
the article's shape before inserting and fails loudly if it has been re-edited.

This needed a new editorial block type, `cta` — see below.

**CRM is still held.** It needs ActiveCampaign to exist as a product in the
`crm` category first, with its own price read from the vendor page and its own
evidence source. Today ActiveCampaign is only in `email-marketing`, and an
offer cannot attach to a product that is not there.

### The `cta` editorial block

A block that can put a paid destination inside an article is the one place an
editor could publish an untracked link, a link to anywhere, or another
affiliate's code — and the article body is rendered into the served HTML, so it
would be indexed before anyone noticed.

So the block **names a promotion by slug and never a URL**. A slug can only
resolve to a row in `commerce.affiliate_promotions`, which already enforces
`https://`, an active merchant, a freshness window and a disclosure label. The
block chooses which approved destination to show; it cannot invent one.

Both renderers build the href themselves — React through `affiliateClickPath`,
and the Go prerenderer as
`/api/affiliate/promotion/{slug}?source=promotion`. That query parameter is
load-bearing rather than decorative: `TrackPromotionClick` rejects any other
source and the handler defaults an absent one to `product_detail`, so without
it every click from a reader without JavaScript resolves to an error. Both
carry `rel="nofollow noopener sponsored"`, and the block's own text says it
pays us, because a reader mid-article has seen no disclosure anywhere else.

The remaining five recommended links — Active Intelligence, Marketing
Automation, SMS Marketing, WhatsApp Messaging, Email Marketing Platform — were
deliberately not activated. The catalog has no SMS, WhatsApp or automation
category to hang them on, and Email Marketing Platform duplicates the pricing
destination already in use.

The price itself is not stale — read from the vendor page 2026-08-20 — but
ActiveCampaign is one of the twenty-five annual-basis products below. Unlike
Teachable there is no monthly figure to move to: the vendor publishes no
monthly rate for this plan at all.

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

## Standalone promotions, deployed

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

Everything is written, green and **deployed**: migration, repository, service,
handler, route, public page, tests. `https://unsolero.com/offers/funnel-hacking-secrets`
returned 200 with its own title and canonical on 2026-08-29, while an unknown
path on the same host returns 404 — so this is a real route, not an SPA
catch-all. The work reached `main` in `2c6dea8`.

One thing is **not** confirmed from outside: whether the two
`commerce.affiliate_promotions` rows exist in the production database. The only
endpoint that would prove it is the redirect itself, and requesting it would
record a click. Check it on the server instead:

```sql
SELECT slug, is_active, last_checked_at FROM commerce.affiliate_promotions;
```

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
