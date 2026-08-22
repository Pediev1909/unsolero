import { Scale } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'

import { SiteFooter } from '../../../components/layout/SiteFooter'
import { SiteHeader } from '../../../components/layout/SiteHeader'
import { Button } from '../../../components/ui/Button'
import { Container } from '../../../components/ui/Container'
import { Drawer } from '../../../components/ui/Drawer'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Heading } from '../../../components/ui/Heading'
import { LoadingState } from '../../../components/ui/LoadingState'
import { usePageMetadata } from '../../../lib/seo/usePageMetadata'
import { useBrands, useCategories, useProducts } from '../queries'
import { useCatalogActions } from '../useCatalogActions'
import { useCatalogUrlState } from '../useCatalogUrlState'
import { CatalogFilters } from './CatalogFilters'
import { CatalogLoadingGrid } from './CatalogLoadingGrid'
import { CatalogPagination } from './CatalogPagination'
import { CatalogProductGrid } from './CatalogProductGrid'
import { CatalogToolbar } from './CatalogToolbar'
import { ProductComparison } from './ProductComparison'
import { catalogRobots } from './catalogSeo'

interface CatalogListingProps {
  eyebrow?: string
  /** Renders a trail above the heading. Omit it and none is shown. */
  breadcrumb?: { parentLabel: string; parentTo: string }
  title: string
  description: string
  categorySlug?: string
  brandSlug?: string
  noindex?: boolean
  afterCatalog?: ReactNode
}

export function CatalogListing({
  eyebrow = 'Software intelligence',
  breadcrumb,
  title,
  description,
  categorySlug,
  brandSlug,
  noindex = true,
  afterCatalog,
}: CatalogListingProps) {
  const [filterOpen, setFilterOpen] = useState(false)
  const { search } = useLocation()
  const urlState = useCatalogUrlState({
    category: categorySlug,
    brand: brandSlug,
  })
  const products = useProducts(urlState.query)
  const categories = useCategories()
  // Scoped to the category being viewed, so the filter never offers a brand
  // that would empty the page.
  const brands = useBrands(categorySlug)
  const actions = useCatalogActions()
  const displayedProductsAreAllDemo = Boolean(
    products.data &&
    products.data.products.length > 0 &&
    products.data.products.every((product) => product.is_demo),
  )

  usePageMetadata({
    title: `${title} | UNSOLERO`,
    description,
    // Filter/sort/pagination combinations are useful to people but duplicate the
    // canonical listing for crawlers. Only the clean landing URL is indexable.
    robots: catalogRobots(noindex, search),
  })

  const filters =
    categories.isPending || brands.isPending ? (
      <LoadingState compact title="Loading filters" />
    ) : categories.isError || brands.isError ? (
      <ErrorState
        compact
        description="Category and brand filters could not be loaded."
        onRetry={() => {
          void categories.refetch()
          void brands.refetch()
        }}
        title="Filters unavailable"
      />
    ) : (
      <CatalogFilters
        brands={brands.data}
        categories={categories.data}
        fixedBrand={Boolean(brandSlug)}
        fixedCategory={Boolean(categorySlug)}
        onApply={(values) => {
          urlState.applyFilters(values)
          setFilterOpen(false)
        }}
        onClear={() => {
          urlState.clearFilters()
          setFilterOpen(false)
        }}
        values={urlState.values}
      />
    )

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-10 sm:py-16 lg:py-24">
          <Container>
            {/* Product pages carried breadcrumbs and category and brand pages
                did not, so a visitor arriving on one from a search result had
                no way up. The crumbs are the smallest links on the page, so
                they carry a 24px hit box. */}
            {breadcrumb && (
              <nav
                aria-label="Breadcrumb"
                className="mb-5 flex flex-wrap items-center gap-x-2 text-xs text-ink/68"
              >
                <Link
                  className="flex min-h-6 items-center hover:text-ink"
                  to="/products"
                >
                  Software
                </Link>
                <span aria-hidden="true">/</span>
                <Link
                  className="flex min-h-6 items-center hover:text-ink"
                  to={breadcrumb.parentTo}
                >
                  {breadcrumb.parentLabel}
                </Link>
                <span aria-hidden="true">/</span>
                <span
                  aria-current="page"
                  className="flex min-h-6 items-center text-ink"
                >
                  {title}
                </span>
              </nav>
            )}
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-bronze-dark">
              {eyebrow}
            </p>
            <Heading
              className="mt-4 max-w-4xl sm:mt-5"
              level={1}
              size="display"
            >
              {title}
            </Heading>
            <p className="mt-5 max-w-2xl text-base leading-7 text-ink/70 sm:mt-6 sm:text-lg">
              {description}
            </p>
            {displayedProductsAreAllDemo && (
              <p className="mt-5 text-xs leading-5 text-ink/65">
                Products shown on this page are clearly marked fictional demo
                data. Prices are reference prices, not live quotes.
              </p>
            )}
          </Container>
        </section>

        <Container className="py-10 sm:py-14">
          <div className="grid gap-10 lg:grid-cols-[15rem_minmax(0,1fr)] xl:grid-cols-[17rem_minmax(0,1fr)]">
            <aside aria-label="Catalog filters" className="hidden lg:block">
              <div className="sticky top-28">
                <h2 className="mb-6 text-xs font-bold uppercase tracking-[0.14em]">
                  Refine results
                </h2>
                {filters}
              </div>
            </aside>

            <div id="catalog-results">
              <CatalogToolbar
                comparedCount={actions.comparedIDs.size}
                count={products.data?.total ?? 0}
                onOpenComparison={() => actions.setComparisonOpen(true)}
                onOpenFilters={() => setFilterOpen(true)}
                onSort={urlState.setSort}
                sort={urlState.sort}
              />

              <div className="mt-7">
                {products.isPending && <CatalogLoadingGrid />}
                {products.isError && (
                  <ErrorState
                    description="The software catalog could not be loaded."
                    onRetry={() => void products.refetch()}
                    title="Catalog unavailable"
                  />
                )}
                {products.data?.products.length === 0 && (
                  <EmptyState
                    action={
                      <Button
                        onClick={urlState.clearFilters}
                        variant="secondary"
                      >
                        Clear filters
                      </Button>
                    }
                    description="Try a broader search, another brand, or a wider price range."
                    title="No software matches these filters"
                  />
                )}
                {products.data && products.data.products.length > 0 && (
                  <>
                    <CatalogProductGrid
                      comparedIDs={actions.comparedIDs}
                      onCompare={actions.compare}
                      onSave={actions.save}
                      products={products.data.products}
                      savePending={actions.savePending}
                      savedIDs={actions.savedIDs}
                    />
                    <CatalogPagination
                      onPage={(page) => {
                        urlState.setPage(page)
                        document
                          .getElementById('catalog-results')
                          ?.scrollIntoView({ behavior: 'smooth' })
                      }}
                      page={products.data.page}
                      totalPages={products.data.total_pages}
                    />
                  </>
                )}
              </div>
            </div>
          </div>
        </Container>
        {afterCatalog}
      </main>

      {actions.comparedIDs.size > 0 && (
        <div
          className="fixed inset-x-0 z-30 mx-auto flex w-[min(calc(100%-2rem),32rem)] items-center justify-between gap-4 border border-ink/15 bg-ink px-4 py-3 text-canvas shadow-overlay"
          // Sits above the consent banner while that is showing, rather than
          // underneath it where its Compare button could not be clicked. The
          // safe-area term keeps it clear of a device's own bottom furniture;
          // the toolbar carries the same action in normal flow, so this bar
          // being obscured by anything is an inconvenience rather than a dead
          // end.
          style={{
            bottom:
              'calc(var(--bottom-bar-offset, 0px) + env(safe-area-inset-bottom, 0px) + 1rem)',
          }}
        >
          <p className="text-sm">
            <strong>{actions.comparedIDs.size}</strong> of 4 selected
          </p>
          <Button
            onClick={() => actions.setComparisonOpen(true)}
            size="sm"
            variant="inverse"
          >
            <Scale aria-hidden="true" size={16} /> Compare
          </Button>
        </div>
      )}

      <Drawer
        description="Narrow products by needs and reference price."
        onOpenChange={setFilterOpen}
        open={filterOpen}
        side="left"
        title="Filters"
      >
        {filters}
      </Drawer>
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
