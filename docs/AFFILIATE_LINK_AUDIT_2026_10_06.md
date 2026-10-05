# Affiliate link audit — 2026-10-06

Audit date uses the project owner's Europe/Sofia date. The first public API
snapshot reported `generated_at: 2026-10-05T18:28:28Z`; its raw timestamp is
preserved here separately from the local audit date.

## Scope and method

This audit covers all **15 existing product offers and three standalone
promotions** found in the production database before Tidio activation. The
unused referral codes in the Zoho dashboard inventory and the held
ActiveCampaign links are not live site destinations.

- Compare production rows with the exact destinations recorded in the seeds.
- Read `GET /api/catalog/offers` and per-product offers endpoints; these do not
  create affiliate clicks.
- Request only untagged public merchant pages to check availability. Do not
  request provider tracking URLs, register accounts, or make purchases.
- Check production promotion rows directly, since the public promotion
  endpoint records a click and page visibility alone does not prove a
  promotion is eligible.

All 15 existing offer links and all three promotion destinations matched their
repository seeds in the read-only production SQL inspection. All were marked
active. This verifies configuration and first-party eligibility; it does not
prove provider-side attribution, conversions, or payment. Those remain checks
against genuine visitor activity in the partner dashboards.

## Initial production state

`GET https://unsolero.com/api/catalog/offers` returned HTTP 200 with **13
offers**, each carrying a same-origin `purchase_path`. All 13 had
`last_checked_at: 2026-09-24T12:47:39Z` and `freshness_status: stale`.
ActiveCampaign was among them, superseding the old documentation that called
its seed unapplied.

Pipedrive Lite and SE Ranking Core each returned HTTP 200 with `[]` from
`/api/catalog/products/{slug}/offers`. Their affiliate rows were active and
correctly configured, but their last checks were August 28 and September 2,
respectively, outside the production offer window. An empty response does not
establish that the external referral link is broken.

The production promotion rows were all active, last checked at
`2026-09-24T12:47:39Z`, with no `valid_from` or `valid_until` bound. The
freshness guard also applies to promotions. Its implementation is
[`ResolvePromotionDestination`](../backend/internal/adapters/postgres/commerce/repository.go).
No timestamp was refreshed merely because a URL returned HTTP 200.

## Existing product destinations

The `Public page` result is for the untagged page linked in that column, not
for the affiliate URL. `Shown` means present in the initial production offers
response with a non-null `purchase_path`.

| Product | Exact configured affiliate URL | Public page result | Initial site state | Repository source |
| --- | --- | --- | --- | --- |
| Bigin Express | `https://go.zoho.com/SsgT` | [Pricing](https://www.bigin.com/pricing.html): 200 | Shown | [saas_offers.sql](../backend/seeds/saas_offers.sql) |
| Cal.com Teams | `https://refer.cal.com/unsolero-wtpd` | [Pricing](https://cal.com/pricing): 200 | Shown | [cal_com_affiliate.sql](../backend/seeds/cal_com_affiliate.sql) |
| Kit Creator | `https://partners.kit.com/ge4ub49bd32p-uw04r` | [Pricing](https://kit.com/pricing): 200 | Shown | [kit_affiliate.sql](../backend/seeds/kit_affiliate.sql) |
| MailerLite Comfort | `https://www.mailerlite.com/?linkId=lp_170762&sourceId=unsolero&tenantId=mailerlite` | [Pricing](https://www.mailerlite.com/pricing): 200 | Shown | [mailerlite_affiliate.sql](../backend/seeds/mailerlite_affiliate.sql) |
| monday.com Basic | `https://try.monday.com/o8rui6cwqodl-p7wfga` | [Pricing](https://monday.com/pricing): 200 | Shown | [monday_affiliate.sql](../backend/seeds/monday_affiliate.sql) |
| Pipedrive Lite | `https://aff.trypipedrive.com/8c0fqmk2j8mc` | [Programme landing](https://www.pipedrive.com/programLP): redirects to `/en/programlp`, 200; pricing page blocked local automated access, 403 | Not shown; stale observation | [pipedrive_affiliate.sql](../backend/seeds/pipedrive_affiliate.sql) |
| SE Ranking Core | `https://seranking.com/subscription.html?ga=5233991&source=link` | [Pricing](https://seranking.com/subscription.html): 200 | Not shown; stale observation | [se_ranking_affiliate.sql](../backend/seeds/se_ranking_affiliate.sql) |
| Teachable Starter | `https://partnerstack.teachable.com/y6u7cxavunjg` | [Pricing](https://teachable.com/pricing): redirects to `www.teachable.com/pricing`, 200 | Shown | [teachable_affiliate.sql](../backend/seeds/teachable_affiliate.sql) |
| Zoho Bookings Basic | `https://go.zoho.com/POSi` | [Pricing](https://www.zoho.com/bookings/pricing.html): 200 | Shown | [saas_catalog_three.sql](../backend/seeds/saas_catalog_three.sql) |
| Zoho Books Standard | `https://go.zoho.com/K0nf` | [Pricing](https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard): 200 | Shown | [saas_offers.sql](../backend/seeds/saas_offers.sql) |
| Zoho Campaigns Standard | `https://go.zoho.com/UCST` | [Pricing](https://www.zoho.com/campaigns/pricing.html): 200 | Shown | [saas_catalog_email2.sql](../backend/seeds/saas_catalog_email2.sql) |
| Zoho CRM Standard | `https://go.zoho.com/dNbV` | [Pricing](https://www.zoho.com/crm/zohocrm-pricing.html): 200 | Shown | [saas_offers.sql](../backend/seeds/saas_offers.sql) |
| Zoho Invoice | `https://go.zoho.com/dIhI` | [Pricing](https://www.zoho.com/invoice/pricing/): 200 | Shown | [saas_catalog_gaps.sql](../backend/seeds/saas_catalog_gaps.sql) |
| Zoho Projects Premium | `https://go.zoho.com/PCoS` | [Pricing](https://www.zoho.com/projects/pricing.html): redirects to `/projects/zohoprojects-pricing.html`, 200 | Shown | [saas_catalog_gaps.sql](../backend/seeds/saas_catalog_gaps.sql) |
| ActiveCampaign Starter | `https://try.activecampaign.com/ydwpszsmins9` | [Pricing](https://www.activecampaign.com/pricing): 200 | Shown | [activecampaign_affiliate.sql](../backend/seeds/activecampaign_affiliate.sql) |

## New partnership and release target

| Product | Exact supplied affiliate URL | Price evidence | Initial site state | Repository source |
| --- | --- | --- | --- | --- |
| Tidio Starter | `https://affiliate.tidio.com/s92aresrdhui` | [Official pricing](https://www.tidio.com/pricing/): USD 24.17/month billed annually, 100 billable conversations/month, reviewed 2026-10-06 | Not activated in the initial snapshot | [tidio_affiliate.sql](../backend/seeds/tidio_affiliate.sql) |

The target release activates Tidio and re-verifies the unchanged SE Ranking
USD 129 monthly offer using
[se_ranking_reverification_2026_10_06.sql](../backend/seeds/se_ranking_reverification_2026_10_06.sql).
It preserves the existing 13 visible offers, keeps Pipedrive unavailable until
its monthly USD price is confirmed, and leaves the three promotion rows intact.
This produces **15 eligible product offers out of 16 configured offers**.

Post-deployment production API and application-only redirect checks are
pending at the time of this audit document's initial write. The release check
must compare each UNSOLERO `302 Location` with the approved destination
without following it. Use the application's bot classification so diagnostic
requests cannot emit countable affiliate events; they may still create raw
diagnostic records. Never send those requests to the provider's tracking URL.

## Existing promotions

| Promotion slug | Exact configured affiliate URL | Untagged destination result | Repository source |
| --- | --- | --- | --- |
| `funnel-hacking-secrets-webinar` | `https://www.funnelhackingsecrets.com?cf_affiliate_id=4330879&affiliate_id=4330879` | `https://www.funnelhackingsecrets.com` redirects to `https://funnelhackingsecrets.com/`, then 403 | [clickfunnels_funnel_hacking_secrets.sql](../backend/seeds/clickfunnels_funnel_hacking_secrets.sql) |
| `funnel-hacking-secrets-order` | `https://www.funnelhackingsecrets.com/go?cf_affiliate_id=4330879&affiliate_id=4330879` | `https://www.funnelhackingsecrets.com/go` redirects to `https://funnelhackingsecrets.com/go`, then 403 | [clickfunnels_funnel_hacking_secrets.sql](../backend/seeds/clickfunnels_funnel_hacking_secrets.sql) |
| `activecampaign-mailchimp-switch` | `https://try.activecampaign.com/am4yesxqhxo9-c8qk4` | [Comparison page](https://www.activecampaign.com/compare/mailchimp): 200 | [activecampaign_mailchimp_switch.sql](../backend/seeds/activecampaign_mailchimp_switch.sql) |

UNSOLERO's `/offers/funnel-hacking-secrets` and
`/guides/mailchimp-alternatives` both returned HTTP 200. The guide's server HTML
contained the expected promotion CTA with `?source=promotion`. The
ClickFunnels page renders its buttons in React; its server HTML alone is not
evidence that a promotion row exists.

The ClickFunnels 403 responses are an **unresolved external check**, not proof
of a dead link: the provider may block automated requests. Keep the exact
approved tracking strings. Confirm the untagged destination in an ordinary
browser, or obtain a replacement link from the approved ClickFunnels
dashboard if the campaign is no longer available.

## Price and destination findings

**SE Ranking:** the official [pricing page](https://seranking.com/subscription.html)
still publishes Core at USD 129 monthly or USD 103.20 per month with annual
billing. The web reader exposed USD; the local HTTP response was localized to
EUR, so no currency conversion was inferred. The current catalog's USD 129
monthly basis can be re-verified without changing recommendation scoring or
the affiliate destination. Do not replay the original affiliate seed, which
contains the old annual-basis USD 103.20 figure.

**Pipedrive:** do not renew the stored USD 19.90 monthly assertion without
stronger evidence. The official [pricing page](https://www.pipedrive.com/en/pricing)
read through the web tool shows USD 14 per seat/month billed annually (USD 168
per year), while direct local HTTP and browser requests were blocked. The
monthly USD price could not be confirmed from the authoritative pricing toggle.
Secondary material and an `intercom.help/pipedrive` article were not accepted
as a replacement for a current official monthly price read. Keep the
price-review gap explicit instead of asserting a newly verified monthly price.

The Pipedrive referral still has the previously recorded generic programme
landing destination. If a pricing-specific referral is wanted, use the
programme's PartnerStack link builder if available, or request it from
`affiliates@pipedrive.com`; supply the existing link code `8c0fqmk2j8mc` and
request `https://www.pipedrive.com/en/pricing` as its destination. A plain
pricing URL is not a replacement for an owned referral URL.

## Release validation

Before deployment, both SQL files passed first-application and exact rerun
checks against a disposable PostgreSQL database with the current migrations.
IDs, row counts and timestamps remained stable on rerun. Catalog products,
published fact and score revisions, and recommendation policy snapshots were
unchanged. Twenty-two invalid-state cases failed and rolled back without
partial writes, including mismatched facts/prices, expired offers or links,
future validity, a conflicting selected destination and conflicting evidence.

The full Go test suite passed with PostgreSQL integration tests enabled, as
did `go vet` and `go build`. Frontend validation passed ESLint, all 303 tests,
TypeScript, the production build, formatting and bundle budgets. Tests that
require a separate Redis or S3 test service were not enabled for this data-only
change.

## Ongoing limitation

The existing 30-day eligibility window still requires real offer reviews;
there is no scheduled price-refresh job. The 13 initially visible offers and
three promotions share a September 24 check timestamp. Raising timestamps
without reviewing the associated facts would conceal stale evidence. The
public seven-day `stale` label correctly remains visible before the 30-day
eligibility cutoff.
