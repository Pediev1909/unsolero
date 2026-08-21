import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { useBrands } from '../features/catalog/queries'
import type { Brand } from '../features/catalog/schemas'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { useStructuredData } from '../lib/seo/useStructuredData'

const description =
  'Every software vendor in the UNSOLERO catalog, listed alphabetically with the number of their products we have compared.'

/**
 * Groups vendors under their first letter.
 *
 * A vendor with nothing published is left out for the same reason the sitemap
 * leaves its page out: it is a promise rather than a page. A demo fixture is
 * left out because it is not a real company.
 */
function groupByLetter(brands: Brand[]) {
  const listed = brands.filter(
    (brand) =>
      !brand.slug.startsWith('demo-') &&
      (brand.published_products === undefined ||
        brand.published_products > 0),
  )
  const letters = new Map<string, Brand[]>()
  for (const brand of listed) {
    const first = brand.name.trim().charAt(0).toUpperCase()
    // A vendor whose name starts with a digit — n8n does not, but 1Password
    // would — belongs in one bucket rather than inventing a heading per digit.
    const key = /[A-Z]/.test(first) ? first : '#'
    const bucket = letters.get(key)
    if (bucket) bucket.push(brand)
    else letters.set(key, [brand])
  }
  return [...letters.entries()]
    .sort(([left], [right]) => {
      if (left === '#') return 1
      if (right === '#') return -1
      return left.localeCompare(right)
    })
    .map(([letter, items]) => ({
      letter,
      brands: items.sort((left, right) => left.name.localeCompare(right.name)),
    }))
}

export function BrandsPage() {
  const brands = useBrands()
  usePageMetadata({
    title: 'Every software vendor we cover | UNSOLERO',
    description,
  })
  useStructuredData('brand-collection', {
    '@context': 'https://schema.org',
    '@type': 'CollectionPage',
    name: 'Software vendors',
    description,
    url: window.location.href,
  })

  const groups = groupByLetter(brands.data ?? [])
  const total = groups.reduce((sum, group) => sum + group.brands.length, 0)

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-16 sm:py-24 lg:py-28">
          <Container>
            <p className="eyebrow">Browse by vendor</p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              Looking for one company in particular?
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/65">
              Every vendor whose products we have priced and scored. If you
              already know the name, start here.
            </p>
          </Container>
        </section>

        <Container className="py-12 sm:py-18 lg:py-24">
          {brands.isPending && (
            <LoadingState
              description="Loading the vendors in the catalog."
              title="Loading vendors"
            />
          )}
          {brands.isError && (
            <ErrorState
              description="The vendor list could not be loaded."
              onRetry={() => void brands.refetch()}
              title="Vendors unavailable"
            />
          )}

          {!brands.isPending && !brands.isError && (
            <div className="flex flex-col gap-12">
              {/* Jump links. With forty-odd vendors the page is longer than a
                  phone screen several times over, and hunting by scroll is
                  the thing an index is supposed to spare you. */}
              <nav aria-label="Jump to letter">
                <ul className="flex flex-wrap gap-2">
                  {groups.map((group) => (
                    <li key={group.letter}>
                      <a
                        className="flex h-10 min-w-10 items-center justify-center rounded-sm border border-ink/15 bg-surface px-2 text-sm font-semibold transition-colors hover:border-bronze hover:text-bronze"
                        href={`#letter-${group.letter}`}
                      >
                        {group.letter}
                      </a>
                    </li>
                  ))}
                </ul>
              </nav>

              {groups.map((group) => (
                <section
                  aria-labelledby={`letter-${group.letter}`}
                  key={group.letter}
                >
                  <h2
                    className="scroll-mt-24 border-b border-ink/15 pb-3 font-display text-2xl font-medium tracking-[-0.03em]"
                    id={`letter-${group.letter}`}
                  >
                    {group.letter}
                  </h2>
                  <ul className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {group.brands.map((brand) => (
                      <li key={brand.slug}>
                        <Link
                          className="group flex items-center justify-between gap-3 rounded-sm border border-ink/15 bg-surface px-4 py-3 transition-colors hover:border-bronze"
                          to={`/brands/${brand.slug}`}
                        >
                          <span className="font-medium group-hover:text-bronze">
                            {brand.name}
                          </span>
                          {brand.published_products !== undefined && (
                            <span className="shrink-0 text-sm text-ink/55 tabular-nums">
                              {brand.published_products}
                            </span>
                          )}
                        </Link>
                      </li>
                    ))}
                  </ul>
                </section>
              ))}

              <p className="border-t border-ink/15 pt-8 text-sm text-ink/60">
                {total} {total === 1 ? 'vendor' : 'vendors'} in the catalog. The
                number beside each is how many of its products we have
                compared.{' '}
                <Link className="underline underline-offset-4" to="/categories">
                  Browse by category instead
                </Link>
                .
              </p>
            </div>
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
