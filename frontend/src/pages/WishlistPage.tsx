import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { EmptyState } from '../components/ui/EmptyState'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { CatalogProductCard } from '../features/catalog/components/CatalogProductCard'
import { MerchantAction } from '../features/catalog/components/MerchantAction'
import { useProductsByIDs } from '../features/catalog/queries'
import {
  useComparisonSelection,
  useWishlistSelection,
} from '../features/catalog/useProductCollections'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function WishlistPage() {
  const wishlist = useWishlistSelection()
  const comparison = useComparisonSelection()
  const products = useProductsByIDs(wishlist.productIDs)
  const loading =
    wishlist.isPending || (wishlist.productIDs.length > 0 && products.isPending)
  usePageMetadata({
    title: 'Saved Equipment | UNSOLERO',
    description: 'Review saved software and current offers.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-14 sm:py-20">
          <p className="eyebrow">Your shortlist</p>
          <Heading className="mt-5" level={1} size="display">
            Saved equipment.
          </Heading>
          <p className="mt-5 max-w-2xl text-ink/70">
            Current catalog prices and available merchant offers are loaded when
            you open this page. Guest saves stay on this device.
          </p>
          <div className="mt-10">
            {loading && <LoadingState title="Loading saved equipment" />}
            {(wishlist.isError || products.isError) && (
              <ErrorState
                description="Your saved equipment could not be loaded."
                onRetry={() => {
                  void wishlist.refetch()
                  void products.refetch()
                }}
                title="Saved equipment unavailable"
              />
            )}
            {!loading &&
              !wishlist.isError &&
              wishlist.productIDs.length === 0 && (
                <EmptyState
                  action={
                    <ButtonLink to="/products">Explore equipment</ButtonLink>
                  }
                  description="Save products from the catalog to keep a focused shortlist."
                  title="Nothing saved yet"
                />
              )}
            {products.data && products.data.products.length > 0 && (
              <div className="grid gap-px bg-ink/15 sm:grid-cols-2 xl:grid-cols-3">
                {products.data.products.map((product) => (
                  <div className="flex flex-col bg-surface" key={product.id}>
                    <CatalogProductCard
                      compared={comparison.productIDs.includes(product.id)}
                      onCompare={() => comparison.toggle(product.id)}
                      onSave={() => wishlist.toggle(product.id)}
                      product={product}
                      saved
                      savePending={wishlist.isPending}
                    />
                    <div className="px-4 pb-5 sm:px-5">
                      <MerchantAction
                        className="mt-0"
                        slug={product.slug}
                        source="wishlist"
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
