import { useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { LoadingState } from '../components/ui/LoadingState'
import { CatalogListing } from '../features/catalog/components/CatalogListing'
import { useBrand } from '../features/catalog/queries'

export function BrandPage() {
  const { slug = '' } = useParams()
  const brand = useBrand(slug)

  if (brand.isPending) return <BrandState loading />
  if (brand.isError) return <BrandState retry={() => void brand.refetch()} />

  return (
    <CatalogListing
      brandSlug={brand.data.slug}
      description={brand.data.description}
      eyebrow="Equipment brand"
      noindex={false}
      title={brand.data.name}
    />
  )
}

function BrandState({
  loading,
  retry,
}: {
  loading?: boolean
  retry?: () => void
}) {
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-24">
          {loading ? (
            <LoadingState title="Loading brand" />
          ) : (
            <ErrorState
              description="This brand could not be found or loaded."
              onRetry={retry}
              title="Brand unavailable"
            />
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
