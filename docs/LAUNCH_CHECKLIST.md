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

### Apply in this order

**1. PartnerStack first.** It is a network, not a single programme, and it
carries ClickUp among others. One approval opens a marketplace you then apply
into, and several of its programmes approve automatically.

- Partner signup: https://partnerstack.com/partners-and-publishers
- Programme directory: https://market.partnerstack.com

The network application asks for eight to ten details and takes about five
minutes. It accepts individuals. What it checks is that your site is real: a
dead link, a placeholder page, or a social profile in place of a site is
grounds for rejection. Sign up with an email and name that match the LinkedIn
profile you give it.

**2. Then the direct programmes.** Every product in the catalog has one, which
is why those six were chosen.

| Product | Where to apply | Published terms |
| --- | --- | --- |
| ClickUp | via PartnerStack, or https://clickup.com/partners/affiliates | up to $25 per free workspace signup; 25% recurring at higher tiers; 30-day cookie |
| HubSpot | https://www.hubspot.com/partners/affiliates | 30% recurring monthly for up to one year; paid through Impact; $10 minimum balance |
| Zoho | https://www.zoho.com/affiliate/ | 15% recurring for the first 12 months; 90-day cookie; reviewed in about 3 business days |
| FreshBooks | https://www.freshbooks.com/affiliates | up to $10 per free trial, up to $200 per paying customer; paid on the 20th after a 30-day lock |
| Teamwork | https://www.teamwork.com/partners/ | 15% recurring for the lifetime of the customer |
| Salesflare | https://www.salesflare.com/affiliates | 30% recurring |

Rates and terms are what each programme published when this was written. Verify
on the programme page before building a page around a number.

### What the applications look at

Zoho states its requirement plainly, and it is representative: a functional
site with relevant business or technology content. The site now has a catalog,
six content pages, a privacy policy and an affiliate disclosure, which is the
state these reviews are looking for.

Apply to two or three first rather than all six at once. If something in the
application is wrong — a mismatched name, a missing disclosure — you would
rather learn it on the second rejection than the sixth.

### Networks worth having

| Network | Payout threshold |
| --- | --- |
| Impact.com | about $10 |
| ShareASale | $50 |
| CJ Affiliate | $50 |

HubSpot pays through Impact, so that account is needed regardless. All three
accept individuals and cost nothing to join.

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
