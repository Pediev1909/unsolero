import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { EmptyState } from '../components/ui/EmptyState'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { ContentGrid } from '../features/content/components/ContentGrid'
import { useContent } from '../features/content/queries'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { useStructuredData } from '../lib/seo/useStructuredData'

interface HubConfig {
  section: 'articles' | 'guides'
  eyebrow: string
  title: string
  description: string
  emptyTitle: string
}

const guides: HubConfig = {
  section: 'guides',
  eyebrow: 'Editorial guidance',
  title: 'Buy with a clearer brief.',
  description:
    'Practical guides built from structured equipment facts, real constraints, and explicit trade-offs—not review-volume rankings.',
  emptyTitle: 'No published guides yet',
}

const articles: HubConfig = {
  section: 'articles',
  eyebrow: 'Planning library',
  title: 'Make the room work first.',
  description:
    'Editorial field notes for measuring, planning, and improving a home gym before another product enters the room.',
  emptyTitle: 'No published articles yet',
}

export function GuidesPage() {
  return <ContentHub config={guides} />
}

export function ArticlesPage() {
  return <ContentHub config={articles} />
}

function ContentHub({ config }: { config: HubConfig }) {
  const content = useContent({ section: config.section, limit: 24 })
  usePageMetadata({
    title: `${config.section === 'guides' ? 'Home Gym Guides' : 'Home Gym Articles'} | UNSOLERO`,
    description: config.description,
  })
  useStructuredData('content-collection', {
    '@context': 'https://schema.org',
    '@type': 'CollectionPage',
    name: config.title,
    description: config.description,
    url: window.location.href,
  })

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-16 sm:py-24 lg:py-28">
          <Container>
            <p className="eyebrow">{config.eyebrow}</p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              {config.title}
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/65">
              {config.description}
            </p>
          </Container>
        </section>
        <Container className="py-12 sm:py-18 lg:py-24">
          {content.isPending && (
            <LoadingState
              description="Loading published editorial work."
              title={`Loading ${config.section}`}
            />
          )}
          {content.isError && (
            <ErrorState
              description="The editorial library could not be loaded."
              onRetry={() => void content.refetch()}
              title="Editorial content unavailable"
            />
          )}
          {content.data?.length === 0 && (
            <EmptyState
              description="Only reviewed, published editorial work appears here."
              title={config.emptyTitle}
            />
          )}
          {content.data && content.data.length > 0 && (
            <ContentGrid entries={content.data} />
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
