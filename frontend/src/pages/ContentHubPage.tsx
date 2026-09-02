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
  section: 'articles' | 'guides' | 'comparisons' | 'stacks'
  eyebrow: string
  title: string
  description: string
  emptyTitle: string
  metaTitle: string
}

const guides: HubConfig = {
  metaTitle: 'Software Stack Guides | UNSOLERO',
  section: 'guides',
  eyebrow: 'Editorial guidance',
  title: 'Buy with a clearer brief.',
  description:
    'Practical guides built from structured software facts, real constraints, and explicit trade-offs—not review-volume rankings.',
  emptyTitle: 'No published guides yet',
}

const articles: HubConfig = {
  metaTitle: 'Software Stack Articles | UNSOLERO',
  section: 'articles',
  eyebrow: 'Planning library',
  title: 'Fewer tools, better chosen.',
  description:
    'Editorial field notes for planning and tightening a software stack before another subscription joins it.',
  emptyTitle: 'No published articles yet',
}

// Head-to-heads had no page listing them at all. Eight of them existed, were
// in the sitemap, and could be reached only from a product page or a search
// result -- the same gap the category index closed, in a different corner.
const comparisons: HubConfig = {
  metaTitle: 'Software Comparisons | UNSOLERO',
  section: 'comparisons',
  eyebrow: 'Head to head',
  title: 'Two tools, one decision.',
  description:
    'Direct comparisons of the tools people weigh against each other, with every price read from the vendor and the billing basis stated, so a monthly rate is never set against an annual one.',
  emptyTitle: 'No published comparisons yet',
}

// The builder's output as writing. /build is noindex and its result is a
// private page, so the site's one structural difference -- a whole stack with
// the rejections shown -- had no address a search result or a video could
// point at. Each stack here is that result for one kind of business.
const stacks: HubConfig = {
  metaTitle: 'Software stacks, priced for one kind of business | UNSOLERO',
  section: 'stacks',
  eyebrow: 'Complete setups',
  title: 'Software stacks',
  description:
    'A stack is a whole set of tools chosen for one kind of business and one budget: what to run, what it costs a month, and the products we left out and why. Every price is read from the vendor and dated.',
  emptyTitle: 'No published stacks yet',
}

export function GuidesPage() {
  return <ContentHub config={guides} />
}

export function StacksPage() {
  return <ContentHub config={stacks} />
}

export function ComparisonsPage() {
  return <ContentHub config={comparisons} />
}

export function ArticlesPage() {
  return <ContentHub config={articles} />
}

function ContentHub({ config }: { config: HubConfig }) {
  const content = useContent({ section: config.section, limit: 24 })
  usePageMetadata({
    title: config.metaTitle,
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
