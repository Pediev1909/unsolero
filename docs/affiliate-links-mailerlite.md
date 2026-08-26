# MailerLite affiliate link

## Approved link

The default tracking link was copied from Andon Pediev's approved MailerLite
affiliate dashboard on 26 August 2026:

`https://www.mailerlite.com/?linkId=lp_170762&sourceId=unsolero&tenantId=mailerlite`

Ownership identifiers:

- source: `unsolero`
- tenant/program: `mailerlite`
- landing-page reference: `lp_170762`
- provider interface: Trackdesk

The repository stores the URL because an affiliate destination is a public
commerce value, not a credential. Do not alter, shorten, decode and rebuild, or
remove any of its tracking parameters. Only replace it with another complete
link copied from this same approved account.

## How UNSOLERO uses it

The browser never receives the raw destination in catalog API responses.
Product CTAs point to UNSOLERO's provider-neutral path:

`/api/affiliate/click/{offer-id}`

The backend checks that the merchant, offer, and affiliate link are active and
that the offer is fresh. It records the outbound navigation, then returns a
non-cacheable redirect to the exact MailerLite URL. Recommendation scoring has
no access to commission or affiliate-link data.

The public CTA and the affiliate disclosure explain that the link is an
affiliate link. MailerLite or its affiliate platform may place an attribution
cookie only after the visitor chooses to leave UNSOLERO; the privacy page
explains that handoff.

## Do we need another MailerLite link?

No. The approved default link is sufficient to launch tracking.

A pricing-page link is useful later because the product CTA discusses a
specific plan and price. It reduces the number of steps between the comparison
and the visitor's pricing decision, but it does not create a second affiliate
account or a different commission arrangement.

Create one only if the MailerLite dashboard makes the destination available:

1. Sign in to the MailerLite affiliate dashboard.
2. Open the pinned/default MailerLite offer.
3. Click **Advanced options** below the tracking link.
4. If **Target specific landing page** is available, enable it.
5. Select the official pricing landing page. If the dashboard instead exposes
   a **Deep link URL** field, enter `https://www.mailerlite.com/pricing` there.
6. Copy the newly generated complete tracking link.
7. Confirm that it still identifies source `unsolero` and tenant `mailerlite`.
8. Send the complete copied link to the maintainer. Do not send only the pricing
   URL and do not manually add parameters to it.

Trackdesk can expose different controls depending on whether MailerLite enabled
specific landing-page targeting or deep linking for the offer. If neither
control appears, the default link remains the correct link and no action is
required.

Optional Aff S1-S5 values are campaign-reporting labels, not new affiliate
links. Leave them empty until UNSOLERO has a documented reporting need; the
site already records the placement (`product_detail`, recommendation, and so
on) internally.

## Programme constraints recorded on 2026-08-26

- 30% on the initial paid subscription and 30% on recurring lifetime payments.
- Last-click attribution within a 45-day referral window.
- No self-referrals, automated traffic, hidden redirects, cookie stuffing, or
  forced tracking.
- A clear affiliate disclosure and a public privacy policy explaining affiliate
  links, cookies, and related processing are required.
- Conversions have a 30-day holding period. Settlement requires at least $100
  in open balance from at least two separate referrals.

MailerLite's own dashboard is authoritative for clicks and conversions. A local
redirect test can prove that UNSOLERO preserves the approved destination, but
it cannot prove a sale or payment. Never create a fake signup or self-purchase
to test conversion reporting.

## Activation and audit

Apply the idempotent data file:

`backend/seeds/mailerlite_affiliate.sql`

Then verify all of the following:

1. The MailerLite product API returns a provider-neutral `purchase_path`, not
   the raw affiliate URL.
2. A no-follow request to that path returns HTTP 302 with the complete approved
   MailerLite URL in `Location`.
3. The response is not cacheable.
4. The product page labels the CTA as an affiliate link.
5. The recommendation result and order are unchanged when commission metadata
   changes.
6. A later genuine visitor click appears in the MailerLite dashboard. Do not
   manufacture a click merely to make this number non-zero.
