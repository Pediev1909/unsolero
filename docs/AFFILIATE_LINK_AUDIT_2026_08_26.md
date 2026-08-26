# Affiliate link audit — 2026-08-26

## Result

Twelve offers have current first-party pricing evidence, an exact destination
copied from an approved programme dashboard, and an active public UNSOLERO
redirect. Zoho Books has been corrected to its reviewed $20 monthly-billing
price and reactivated.

No affiliate URL was followed during this audit. Provider terms prohibit
automated or artificial clicks, and a redirect request can itself be recorded
as a click. The audit instead checks each exact owned destination in the
database, requests only the provider-neutral UNSOLERO URL without following its
302, and verifies that `Location` is byte-for-byte equal to the approved link.

This proves that UNSOLERO routes genuine visitors correctly. Only the partner
dashboard can prove provider-side click attribution, conversion, and payment.
Those should be confirmed using later genuine visitor activity, never a fake
signup, self-referral, or test purchase.

## Active and verified

| Product | Recorded basis | Provider | Ownership evidence |
|---|---:|---|---|
| Bigin Express | $9/user/month | Zoho | Approved affiliate ID `PE2263909`, portal link code `SsgT` |
| Cal.com Teams | $12/user/month, yearly | Cal.com | Referral dashboard code `unsolero-wtpd` |
| Kit Creator | $39/month at 1,000 subscribers | PartnerStack | Link copied from the approved dashboard |
| MailerLite Comfort | $19/month at 1,000 subscribers | Trackdesk/MailerLite | Source `unsolero`, link `lp_170762` |
| monday.com Basic | $9/seat/month, yearly | PartnerStack | Link copied from the approved dashboard |
| SE Ranking Core | $103.20/month | SE Ranking | Official link-builder account `5233991` |
| Zoho Bookings Basic | $8/user/month | Zoho | Approved affiliate ID `PE2263909`, portal link code `POSi` |
| Zoho Books Standard | $20/organization/month | Zoho | Approved affiliate ID `PE2263909`, portal link code `K0nf` |
| Zoho Campaigns Standard | $5.25/month at 1,000 contacts, yearly | Zoho | Approved affiliate ID `PE2263909`, portal link code `UCST` |
| Zoho CRM Standard | $20/user/month | Zoho | Approved affiliate ID `PE2263909`, portal link code `dNbV` |
| Zoho Invoice | Free | Zoho | Approved affiliate ID `PE2263909`, portal link code `dIhI` |
| Zoho Projects Premium | $4/user/month, yearly | Zoho | Approved affiliate ID `PE2263909`, portal link code `PCoS` |

First-party sources used for the current check:

- MailerLite pricing and affiliate programme terms
- Cal.com pricing
- Kit pricing and billing documentation
- monday.com pricing documentation
- SE Ranking pricing
- Zoho live pricing pages and the JSON pricing data those pages use
- approved partner-dashboard evidence already recorded beside each link seed

## Corrected: Zoho Books Standard

Zoho's current US pricing page shows $20 monthly or $15/month billed annually.
UNSOLERO compares monthly-billing prices, so the product, offer, and public
recommendation policy now use $20.

The correction includes:

1. a new dated manufacturer observation;
2. product fact revision 2 with the monthly billing basis stated clearly;
3. an affiliate-independent value-score review;
4. published fact and score revision 2;
5. a $20 monthly merchant offer;
6. immutable recommendation policy `saas-v11`;
7. another no-follow UNSOLERO redirect audit.

This separation prevents affiliate availability from overriding product truth.

## Reproducible production operation

The idempotent, assertion-backed audit operation is:

`backend/seeds/affiliate_offer_audit_2026_08_26.sql`

It fails the transaction unless all twelve approved links, providers, products,
currencies, and verified prices match exactly.
