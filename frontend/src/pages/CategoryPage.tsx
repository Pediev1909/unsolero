import { useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { LoadingState } from '../components/ui/LoadingState'
import { CatalogListing } from '../features/catalog/components/CatalogListing'
import { useCategory } from '../features/catalog/queries'
import { RelatedContentSection } from '../features/content/components/RelatedContentSection'
import { useStructuredData } from '../lib/seo/useStructuredData'

export function CategoryPage() {
  const { slug = '' } = useParams()
  const category = useCategory(slug)
  useStructuredData(
    'category-collection',
    category.data
      ? {
          '@context': 'https://schema.org',
          '@type': 'CollectionPage',
          name: category.data.name,
          description: category.data.description,
          url: window.location.href,
        }
      : null,
  )

  if (category.isPending) return <CollectionLoading label="category" />
  if (category.isError)
    return (
      <CollectionError label="category" retry={() => void category.refetch()} />
    )

  return (
    <CatalogListing
      categorySlug={category.data.slug}
      description={category.data.description}
      breadcrumb={{ parentLabel: 'Categories', parentTo: '/categories' }}
      eyebrow="Software category"
      noindex={false}
      title={category.data.name}
      afterCatalog={<RelatedContentSection categorySlug={category.data.slug} />}
    />
  )
}

function CollectionLoading({ label }: { label: string }) {
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-24">
          <LoadingState title={`Loading ${label}`} />
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}

function CollectionError({
  label,
  retry,
}: {
  label: string
  retry: () => void
}) {
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-24">
          <ErrorState
            description={`This ${label} could not be found or loaded.`}
            onRetry={retry}
            title={`${label[0]?.toUpperCase()}${label.slice(1)} unavailable`}
          />
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
