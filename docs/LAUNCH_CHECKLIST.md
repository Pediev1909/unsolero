# Launch checklist

Everything that has to happen outside the codebase, in the order it has to
happen, with what it costs.

Written for one person, with no company and no budget. Nothing here assumes a
team, a legal entity, or paid tooling.

## The order matters

There is one hard dependency that shapes everything else:

> **Affiliate programs review a live site before approving you.** Most require a
> working website with real content in the category you want to promote. A bare
> domain, or a site with three placeholder pages, is rejected — and a rejection
> is harder to reverse than a first application.

So the sequence is: **site live → content published → apply to programs → earn**.
There is no shortcut that starts with the money.

## Total cost

| When | What | Cost |
| --- | --- | --- |
| Once | Domain, first year | about €12 |
| Monthly | VPS | €5–14, depending on size |
| Monthly | Cloudflare DNS, CDN, R2 storage | €0 |
| Monthly | SMTP (Resend or Brevo free tier) | €0 |
| Monthly | Alert webhook (Discord) | €0 |
| Once, optional | One accountant consultation | one session's fee |

**To start you need roughly €20–30.** That covers the domain and the first
month of hosting. Everything else is a free tier.

On VPS size: 4 GB is the floor and costs less, but ClamAV — required in
production and checked at startup — is the largest consumer of memory on the
box. 8 GB removes the pressure. Verify current pricing on the provider's page
before ordering; Hetzner raised prices twice in 2026 and published comparisons
are already out of date.

## Step 1 — Domain

**Where:** any registrar. Cloudflare Registrar sells at cost with no markup and
no first-year discount trick, which makes it the cheapest honest option if you
are using Cloudflare anyway. Namecheap and Porkbun are fine alternatives.

**What to buy:** one `.com` if available. Avoid `.local`, `.test`, `.example`,
`.internal` and `.invalid` — the application rejects them as reserved, and so
do certificate authorities.

**Why first:** TLS certificates are issued against a real DNS record, and
affiliate applications ask for the site address. Nothing downstream works
without it.

## Step 2 — Accounts

All free. Create them in one sitting.

| Service | For | Notes |
| --- | --- | --- |
| VPS provider | The server | Hetzner, Netcup or Contabo. Hetzner has no setup fee and bills hourly. |
| Cloudflare | DNS, CDN, TLS, R2 storage | Create **two** R2 buckets: one for media, one for backups. |
| Resend or Brevo | Account security email | Free tier is enough. You must verify your domain with them. |
| Discord | Operational alerts | A server you own; create one incoming webhook. |

## Step 3 — Deploy

Follow [DEPLOY_SINGLE_BOX.md](./DEPLOY_SINGLE_BOX.md). It has the commands, the
configuration, and a table of every startup error and what causes it.

Three things that will waste an evening if you miss them:

1. Generate the 32-byte secrets with `openssl rand -base64 32 | tr -d '='`. The
   padding must be stripped or startup rejects the value.
2. `ALERT_WEBHOOK_TOKEN` is mandatory and at least 32 characters, even though
   Discord ignores it.
3. The first start is slow because ClamAV downloads its signature database. The
   API waits for it deliberately.

## Step 4 — Catalog

This is the real work, and it does not finish. The catalog is what makes the
product worth using; everything else is plumbing around it.

**Start narrow.** Ten tools covered accurately beat sixty covered approximately.
Pick one goal from the policy — `client_services` is the most concrete — and
cover the tools that goal's roles require: CRM, project management, invoicing,
plus a cheaper and a more expensive option for each so the alternatives feature
has something to show.

**Per product you need:** name, category, entry paid tier price, what it does
(capabilities), what it integrates with, which tools it overlaps with, and
scores. Every fact needs a source recorded before the product can be published
— the system enforces this and will not let you skip it.

**Where the prices come from:** the vendor's own pricing page, with the date you
checked it. Not a comparison blog, not memory. Software pricing changes and a
wrong price produces a confidently wrong recommendation, which is the one
failure this product cannot afford.

Delete the fictional demo products before real traffic. They exist to prove the
pipeline works, not to be seen.

## Step 5 — Content

Three editorial pieces are already written and loaded as drafts:

- *How to choose business software without overbuying* — the method
- *Building a software stack for a small client services agency* — applied
- *How UNSOLERO ranks software* — the trust page

Review them, supply hero images, and publish. The third one matters more than
its traffic suggests: it is what a skeptical reader and an affiliate reviewer
both look for.

**What to add next:** one page per goal in the policy, and one per category you
have covered in the catalog. Write about method, not about vendors — vendor
claims go stale and belong in the catalog behind the evidence workflow.

## Step 6 — Affiliate programs

Only after the site is live with real content.

### Start with PartnerStack

[PartnerStack](https://partnerstack.com) is a network, not a single program. It
carries programs for 600+ B2B software companies, and one approval gets you
into a marketplace where you apply to individual programs. Several approve
automatically with no traffic minimum.

**It accepts individuals** — no company required. It does check that your site
is real: a dead link, a placeholder page, or a personal social profile in place
of a site is grounds for rejection. Make the LinkedIn profile you sign up with
match the name on the account.

### Then the direct programs

Typical 2026 rates run 20–40%, with the industry median around 22.5% of
first-year revenue. The headline rates worth knowing:

| Program | Rate | Notes |
| --- | --- | --- |
| Systeme.io | 60% lifetime | Highest headline rate in the category |
| Kit (ConvertKit) | 30–50% lifetime | Tiered; 30% base, rises with volume |
| Notion | 50% for 12 months | |
| Webflow | 50% of first-year revenue | Application-based |

Application-based programs — Webflow, Semrush, Jasper, beehiiv, Frase — do
approve small sites when the content angle is strong. Under 5,000 monthly
visitors is not automatically disqualifying if the site is clearly about their
category.

### Networks worth having

| Network | Payout threshold |
| --- | --- |
| Impact.com | about $10 |
| ShareASale | $50 |
| CJ Affiliate | $50 |

All three accept individuals and cost nothing to join.

**Verify every rate before you build a page around it.** Rates and terms change,
and the figures above are what was published at the time of writing, not a
contract.

## Step 7 — Tax

You do not need a company. In Bulgaria affiliate income can be declared as a
self-employed individual (самоосигуряващо се лице) registered with the NRA —
that is an individual registration, not a firm.

What the research shows, and what to confirm with an accountant rather than
with this document:

- Effective rate around 7.5% (25% statutory expense deduction, then 10% flat tax)
- Registration due within 7 days of starting activity
- VAT threshold about €51,130/year, and reverse-charged services to foreign
  platforms generally do not count toward it
- Social contributions are paid by you, and are the real recurring cost — not
  the tax

One consultation is enough to get this right and is cheap next to getting it
wrong.

## What to expect

The engine works and the site is real. What is missing is a catalog and an
audience, and both are earned slowly.

Be skeptical of your own first month. Affiliate approval takes days to weeks,
SEO takes months, and the first commissions of any size usually arrive well
after the effort that produced them. The compensating advantage of recurring
SaaS commissions is that they accumulate: a customer referred once keeps paying.
That only works if you are still there in a year.
