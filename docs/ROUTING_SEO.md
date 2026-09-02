# Public routing and indexability

Status: local TLS edge semantics verified; server-rendered page content is not implemented  
Last reviewed: 2026-08-18

UNSOLERO uses a bounded edge route resolver rather than a general SSR system.
Nginx proxies browser navigation to `GET /api/v1/public-route` with the original
URI. The resolver validates the static manifest or performs a bounded database
lookup for product, category, brand, and editorial slugs. Only a known route is
internally redirected to the React shell.

```mermaid
flowchart LR
  B[Browser navigation] --> T[TLS Nginx]
  T --> R[Bounded route resolver]
  R -->|known| S[Internal SPA shell]
  R -->|unknown| N[HTTP 404 + noindex]
  R -->|malformed| M[HTTP 400 + noindex]
  R -->|trailing slash| C[HTTP 308 canonical path]
```

`/api/`, `/assets/`, `/images/`, `/sitemap.xml`, and `/robots.txt` have explicit
locations and cannot fall through to the application shell.

## Index policy

Indexable without a query string:

- `/`, `/products`, `/categories`, `/brands`, `/offers`;
- existing `/products/:slug`, `/categories/:slug`, `/brands/:slug`;
- `/guides`, `/articles`, `/comparisons`, `/stacks` and published matching
  editorial detail paths (`/guides/:slug`, `/articles/:slug`,
  `/compare/:slug`, `/stacks/:slug`).

Every indexable route now carries a server-rendered body inside `#root` as
well as its metadata, so a client that does not run JavaScript still receives
the page's text and internal links (see `prerender_static.go` for the static
routes and `pagemeta.go` for the dynamic ones).

Served with `X-Robots-Tag: noindex, nofollow`:

- authentication, onboarding/build, comparison, wishlist, setup, and account
  routes;
- `/links` (the social-bio landing page — a signpost, not a page),
  `/newsletter/confirm` and `/newsletter/unsubscribe` (token landings);
- every admin route;
- any otherwise-public route with query parameters;
- all resolver error responses.

Known indexable routes emit an HTTP canonical `Link` header. Trailing slash
variants redirect with 308. Encoded slash/backslash, control characters,
double slashes, and dot segments fail with 400. Unknown static/dynamic slugs
return a genuine 404 while retaining the existing React not-found experience.

The sitemap remains database-generated from published public records and
`robots.txt` points to the configured canonical origin. Private routes are not
added to the sitemap.

Phase 11 verified these semantics through handler tests and actual HTTPS edge
requests. The result fixes false-200 unknown routes but does not make client-
rendered catalog/editorial content equivalent to SSR or prerendering. Search
rendering, rich-result eligibility, and crawler behavior remain external SEO
validation work.
