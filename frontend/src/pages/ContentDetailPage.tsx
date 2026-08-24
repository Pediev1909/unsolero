import { Scale } from 'lucide-react'
import { Link, Navigate, useLocation, useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { CatalogProductGrid } from '../features/catalog/components/CatalogProductGrid'
import { ProductComparison } from '../features/catalog/components/ProductComparison'
import { useCatalogActions } from '../features/catalog/useCatalogActions'
import { ContentBody } from '../features/content/components/ContentBody'
import { ContentGrid } from '../features/content/components/ContentGrid'
import {
  contentTypeLabel,
  formatEditorialDate,
  headingID,
} from '../features/content/model'
import { useContentEntry } from '../features/content/queries'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { EditorialHero } from '../features/content/components/PriceScale'

export function ContentDetailPage() {
  const { slug = '' } = useParams()
  const location = useLocation()
  const entry = useContentEntry(slug)
  const actions = useCatalogActions()

  usePageMetadata({
    title: entry.data?.seo.title ?? 'Editorial guide | UNSOLERO',
    description:
      entry.data?.seo.description ??
      'Constraint-aware software stack guidance from UNSOLERO Editorial.',
    type: entry.data ? 'article' : 'website',
    imagePath: entry.data?.hero_image.url,
    canonicalURL: entry.data?.seo.canonical_url,
    author: entry.data?.author.name,
    publishedAt: entry.data?.published_at,
    updatedAt: entry.data?.updated_at,
  })
  // The server emits schema.org/Article into the shell already, with the
  // author as a Person and both dates. Emitting a second one here duplicated
  // the page's own description of itself.

  if (entry.isPending) return <ContentState loading />
  if (entry.isError) {
    return <ContentState onRetry={() => void entry.refetch()} />
  }
  if (entry.data.path !== location.pathname) {
    return <Navigate replace to={entry.data.path} />
  }

  const item = entry.data
  const headings = item.content.filter(
    (block) => block.type === 'heading' && block.heading,
  )
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <article>
          <header className="border-b border-ink/15 py-10 sm:py-16 lg:py-20">
            <Container>
              <nav
                aria-label="Breadcrumb"
                className="flex flex-wrap items-center gap-2 text-xs text-ink/68"
              >
                <Link className="hover:text-ink" to="/">
                  Home
                </Link>
                <span aria-hidden="true">/</span>
                <Link
                  className="hover:text-ink"
                  to={item.type === 'article' ? '/articles' : '/guides'}
                >
                  {item.type === 'article' ? 'Articles' : 'Guides'}
                </Link>
                <span aria-hidden="true">/</span>
                <span aria-current="page" className="text-ink">
                  {contentTypeLabel(item.type)}
                </span>
              </nav>
              <div className="mt-10 grid gap-10 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-end lg:gap-20">
                <div>
                  <p className="eyebrow">{contentTypeLabel(item.type)}</p>
                  <Heading className="mt-5 max-w-5xl" level={1} size="display">
                    {item.title}
                  </Heading>
                  <p className="mt-7 max-w-3xl text-lg leading-8 text-ink/65 sm:text-xl sm:leading-9">
                    {item.description}
                  </p>
                </div>
                <div className="border-t border-ink/15 pt-5 text-xs leading-6 text-ink/70 lg:border-l lg:border-t-0 lg:pl-7 lg:pt-0">
                  <p className="font-semibold text-ink">
                    <Link
                      className="underline decoration-ink/25 underline-offset-4 hover:decoration-ink"
                      to={`/author/${item.author.slug}`}
                    >
                      {item.author.name}
                    </Link>
                  </p>
                  <p>Published {formatEditorialDate(item.published_at)}</p>
                  {item.updated_at.slice(0, 10) !==
                    item.published_at.slice(0, 10) && (
                    <p>Updated {formatEditorialDate(item.updated_at)}</p>
                  )}
                </div>
              </div>
            </Container>
          </header>

          {/* The comparison itself, where an abstract illustration used to
              be. A reader arriving here wants to know which one to pick, and
              the first screen should start answering rather than decorate.
              PriceScale returns null when a piece has fewer than two priced
              products or when they all cost the same, and the illustration
              stays as the hero for those. */}
          <Container className="py-7 sm:py-10">
            <EditorialHero
              image={item.hero_image}
              products={item.related_products}
            />
          </Container>

          <Container className="pb-16 pt-6 sm:pb-24 lg:pt-12">
            <div className="grid gap-12 lg:grid-cols-[14rem_minmax(0,45rem)] lg:justify-center lg:gap-16 xl:grid-cols-[16rem_minmax(0,45rem)_10rem]">
              <aside className="hidden lg:block">
                <div className="sticky top-28 border-t border-ink/15 pt-5">
                  <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
                    In this piece
                  </p>
                  <nav aria-label="Article sections" className="mt-4">
                    <ul className="space-y-3 text-sm leading-5 text-ink/70">
                      {headings.map((heading) => (
                        <li key={heading.heading}>
                          <a
                            className="hover:text-bronze-dark"
                            href={`#${headingID(heading.heading ?? '')}`}
                          >
                            {heading.heading}
                          </a>
                        </li>
                      ))}
                    </ul>
                  </nav>
                </div>
              </aside>
              <ContentBody blocks={item.content} />
              <div aria-hidden="true" className="hidden xl:block" />
            </div>
          </Container>
        </article>

        {item.related_categories.length > 0 && (
          <section className="border-y border-ink/15 bg-paper py-12 sm:py-16">
            <Container>
              <p className="eyebrow">Continue the research</p>
              <div className="mt-6 flex flex-wrap gap-3">
                {item.related_categories.map((category) => (
                  <Link
                    className="border border-ink/20 bg-surface px-4 py-3 text-sm font-semibold hover:border-ink"
                    key={category.id}
                    to={`/categories/${category.slug}`}
                  >
                    Explore {category.name}
                  </Link>
                ))}
              </div>
            </Container>
          </section>
        )}

        {item.related_products.length > 0 && (
          <section className="py-16 sm:py-24">
            <Container>
              <p className="eyebrow">Products referenced</p>
              <Heading className="mt-4" level={2} size="title">
                Inspect the underlying product facts.
              </Heading>
              <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/70">
                Product facts and reference prices come from the catalog. Open a
                product to review specifications and current demo offers.
              </p>
              <div className="mt-9">
                <CatalogProductGrid
                  comparedIDs={actions.comparedIDs}
                  onCompare={actions.compare}
                  onSave={actions.save}
                  products={item.related_products}
                  savedIDs={actions.savedIDs}
                  savePending={actions.savePending}
                />
              </div>
            </Container>
          </section>
        )}

        {item.related_content.length > 0 && (
          <section className="border-t border-ink/15 bg-surface py-16 sm:py-24">
            <Container>
              <p className="eyebrow">Related editorial</p>
              <Heading className="mt-4" level={2} size="title">
                Keep building the brief.
              </Heading>
              <div className="mt-9">
                <ContentGrid entries={item.related_content} />
              </div>
            </Container>
          </section>
        )}
      </main>
      {actions.comparedIDs.size > 0 && (
        <Button
          className="fixed bottom-4 right-4 z-30 shadow-overlay"
          onClick={() => actions.setComparisonOpen(true)}
        >
          <Scale aria-hidden="true" size={17} /> Compare{' '}
          {actions.comparedIDs.size}
        </Button>
      )}
      <ProductComparison
        error={actions.comparisonError}
        loading={actions.comparisonPending}
        onOpenChange={actions.setComparisonOpen}
        onRemove={actions.compare}
        open={actions.comparisonOpen}
        products={actions.compared}
      />
      <SiteFooter />
    </>
  )
}

function ContentState({
  loading,
  onRetry,
}: {
  loading?: boolean
  onRetry?: () => void
}) {
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-24">
          {loading ? (
            <LoadingState
              description="Loading the published article and its product references."
              title="Loading editorial content"
            />
          ) : (
            <ErrorState
              description="This page may be unpublished or the address may have changed."
              onRetry={onRetry}
              title="Editorial content unavailable"
            />
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
