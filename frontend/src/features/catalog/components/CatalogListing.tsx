import { Scale } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { useLocation } from 'react-router-dom'

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
  title: string
  description: string
  categorySlug?: string
  brandSlug?: string
  noindex?: boolean
  afterCatalog?: ReactNode
}

export function CatalogListing({
  eyebrow = 'Software intelligence',
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
  const brands = useBrands()
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
        <section className="border-b border-ink/15 py-14 sm:py-20 lg:py-24">
          <Container>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-bronze-dark">
              {eyebrow}
            </p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              {title}
            </Heading>
            <p className="mt-6 max-w-2xl text-base leading-7 text-ink/70 sm:text-lg">
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
                count={products.data?.total ?? 0}
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
          // underneath it where its Compare button could not be clicked.
          style={{ bottom: 'calc(var(--bottom-bar-offset, 0px) + 1rem)' }}
        >
          <p className="text-sm">
            <strong>{actions.comparedIDs.size}</strong> of 4 selected
          </p>
          <Button
            className="border-canvas bg-canvas text-ink hover:border-bronze hover:bg-bronze"
            onClick={() => actions.setComparisonOpen(true)}
            size="sm"
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
