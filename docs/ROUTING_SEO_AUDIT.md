# Routing, SEO, and HTTP semantics audit

Status: PARTIAL — genuine edge status/index policy is implemented; acquisition content remains client-rendered  
Last reviewed: 2026-08-18

## Current evidence

- React Router has explicit public, recommendation, account, admin, and client
  not-found routes.
- Public pages derive canonical, description, robots, social, and applicable
  structured metadata from validated application data. Ratings/reviews are
  absent unless real, and filter/query variants are noindex.
- The backend emits published-only sitemap and robots responses. API and missing
  catalog/content resources return genuine HTTP 404.
- Nginx sends browser navigation through a bounded API route resolver. Known
  static/dynamic routes serve the shell; unknown routes retain the React
  not-found experience with HTTP 404; malformed paths return 400; trailing
  slash variants return 308. API/assets/media/sitemap/robots are excluded.
- HTTP canonical `Link` and private/admin/auth/query `X-Robots-Tag` policies
  survive the internal SPA redirect and are tested at the TLS edge.

## Decision

Do not introduce a general SSR framework merely to change status codes; the
bounded resolver closes that blocker with much less operational complexity.
Server rendering or deterministic prerendering may still be justified for
indexable acquisition, editorial, category, brand, catalog, and product pages
if organic search/social previews are launch channels.

Crawler and social-preview reliability remains PARTIAL because metadata hooks
inside a client SPA do not guarantee every bot executes JavaScript. See
`ROUTING_SEO.md` for the implemented indexability contract.

## Required next validation

Any future public rendering layer must:

1. preserve the existing 200/404/400/308 edge contract;
2. preserve canonical/noindex/structured-data rules without fabricated facts;
3. keep account/admin/recommendation responses private and non-cacheable;
4. invalidate when published content/catalog revisions change;
5. test crawler HTML, social tags, hydration, sitemap/robots, redirects,
   canonical URLs, missing slugs, and mobile widths.
