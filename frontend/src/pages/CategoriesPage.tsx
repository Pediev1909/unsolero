import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { groupCategories } from '../features/catalog/categoryGroups'
import { useCategories } from '../features/catalog/queries'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { useStructuredData } from '../lib/seo/useStructuredData'

const description =
  'Fifteen categories of business software, each holding the tools we have compared inside it, with every price read from the vendor and dated.'

export function CategoriesPage() {
  const categories = useCategories()
  usePageMetadata({
    title: 'Every software category we cover | UNSOLERO',
    description,
  })
  useStructuredData('category-collection', {
    '@context': 'https://schema.org',
    '@type': 'CollectionPage',
    name: 'Software categories',
    description,
    url: window.location.href,
  })

  const groups = groupCategories(categories.data ?? [])
  const total = groups.reduce((sum, group) => sum + group.categories.length, 0)

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-16 sm:py-24 lg:py-28">
          <Container>
            <p className="eyebrow">Browse by need</p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              What kind of tool are you after?
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/65">
              Pick the job you are trying to do. Each category holds the tools
              we have compared for it, priced from the vendor&rsquo;s own page
              and dated so you can see how fresh the figure is.
            </p>
          </Container>
        </section>

        <Container className="py-12 sm:py-18 lg:py-24">
          {categories.isPending && (
            <LoadingState
              description="Loading the categories in the catalog."
              title="Loading categories"
            />
          )}
          {categories.isError && (
            <ErrorState
              description="The category list could not be loaded."
              onRetry={() => void categories.refetch()}
              title="Categories unavailable"
            />
          )}

          {!categories.isPending && !categories.isError && (
            <div className="flex flex-col gap-14 sm:gap-18">
              {groups.map((group) => (
                <section aria-labelledby={`group-${group.key}`} key={group.key}>
                  <div className="flex items-baseline justify-between gap-4 border-b border-ink/15 pb-4">
                    <h2
                      className="font-display text-2xl font-medium tracking-[-0.03em] sm:text-3xl"
                      id={`group-${group.key}`}
                    >
                      {group.label}
                    </h2>
                    <p className="shrink-0 text-sm text-ink/55 tabular-nums">
                      {group.categories.length}{' '}
                      {group.categories.length === 1
                        ? 'category'
                        : 'categories'}
                    </p>
                  </div>

                  <ul className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {group.categories.map((category) => (
                      <li key={category.slug}>
                        {/* The whole card is the link. A card where only the
                            title is clickable is a card that a hurried or
                            imprecise reader misses. */}
                        <Link
                          className="group flex h-full flex-col rounded-sm border border-ink/15 bg-surface p-5 transition-colors hover:border-bronze focus-visible:border-bronze sm:p-6"
                          to={`/categories/${category.slug}`}
                        >
                          <span className="flex items-start justify-between gap-3">
                            <span className="font-display text-lg font-medium tracking-[-0.02em] group-hover:text-bronze">
                              {category.name}
                            </span>
                            <ArrowRight
                              aria-hidden="true"
                              className="mt-1 shrink-0 text-ink/35 transition-transform group-hover:translate-x-0.5 group-hover:text-bronze"
                              size={18}
                            />
                          </span>
                          {category.description && (
                            <span className="mt-2 text-sm leading-6 text-ink/65">
                              {category.description}
                            </span>
                          )}
                          {category.published_products !== undefined && (
                            <span className="mt-4 text-sm text-ink/55 tabular-nums">
                              {category.published_products}{' '}
                              {category.published_products === 1
                                ? 'tool compared'
                                : 'tools compared'}
                            </span>
                          )}
                        </Link>
                      </li>
                    ))}
                  </ul>
                </section>
              ))}

              <p className="border-t border-ink/15 pt-8 text-sm text-ink/60">
                {total} {total === 1 ? 'category' : 'categories'} in the
                catalog.{' '}
                <Link className="underline underline-offset-4" to="/products">
                  See every tool in one list
                </Link>{' '}
                or{' '}
                <Link className="underline underline-offset-4" to="/build">
                  answer a few questions
                </Link>{' '}
                and we will work out which ones you need.
              </p>
            </div>
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
