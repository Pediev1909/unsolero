import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { EmptyState } from '../components/ui/EmptyState'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { ComparisonData } from '../features/catalog/components/ComparisonData'
import { useProductsByIDs } from '../features/catalog/queries'
import { useComparisonSelection } from '../features/catalog/useProductCollections'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function ComparePage() {
  const selection = useComparisonSelection()
  const products = useProductsByIDs(selection.productIDs)
  const loading =
    selection.isPending ||
    (selection.productIDs.length > 0 && products.isPending)
  usePageMetadata({
    title: 'Compare Software | UNSOLERO',
    description: 'Compare structured software product facts side by side.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-14 sm:py-20">
          <p className="eyebrow">Decision workspace</p>
          <Heading className="mt-5" level={1} size="display">
            Compare tools.
          </Heading>
          <p className="mt-5 max-w-2xl text-ink/70">
            Two to four tools, side by side. Every price says whether it is
            monthly or annual, and every score says what it measures — because
            &ldquo;82 out of 100&rdquo; on its own decides nothing.
          </p>
          <div className="mt-10">
            {loading && <LoadingState title="Loading your comparison" />}
            {(selection.isError || products.isError) && (
              <ErrorState
                description="Your comparison could not be loaded."
                onRetry={() => {
                  void selection.refetch()
                  void products.refetch()
                }}
                title="Comparison unavailable"
              />
            )}
            {!loading &&
              !selection.isError &&
              selection.productIDs.length < 2 && (
                <>
                  <EmptyState
                    action={
                      <ButtonLink to="/products">Explore tools</ButtonLink>
                    }
                    description="Choose at least two products from the catalog. Your selection stays in this browser, or syncs to your account when signed in."
                    title={
                      selection.productIDs.length === 1
                        ? 'Choose one more product'
                        : 'No products selected'
                    }
                  />
                  {/* Someone who arrives here with nothing selected has an
                      empty table and no next move. The written head-to-heads
                      are the answer to the question they came with. */}
                  <p className="mt-8 text-center text-sm text-ink/65">
                    Or read a comparison we have already written:{' '}
                    <Link
                      className="underline underline-offset-4"
                      to="/comparisons"
                    >
                      Zapier against Make, Stripe against Paddle, and more
                    </Link>
                    .
                  </p>
                </>
              )}
            {products.data && selection.productIDs.length >= 2 && (
              <ComparisonData
                products={products.data.products}
                onRemove={(id) => selection.toggle(id)}
              />
            )}
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
