import { Container } from '../../../components/ui/Container'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Heading } from '../../../components/ui/Heading'
import { Skeleton } from '../../../components/ui/Skeleton'
import { useContent } from '../queries'
import { ContentGrid } from './ContentGrid'

export function RelatedContentSection({
  categorySlug,
}: {
  categorySlug: string
}) {
  const content = useContent({
    section: 'all',
    category: categorySlug,
    limit: 3,
  })
  return (
    <section className="border-t border-ink/15 bg-paper py-14 sm:py-20">
      <Container>
        <p className="eyebrow">Editorial guidance</p>
        <Heading className="mt-4" level={2} size="title">
          Research before you shortlist.
        </Heading>
        <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/60">
          Curated planning and buying guidance connected to this equipment
          category.
        </p>
        <div className="mt-9">
          {content.isPending && (
            <div
              aria-label="Loading related editorial content"
              className="grid gap-5 md:grid-cols-3"
              role="status"
            >
              {[0, 1, 2].map((item) => (
                <Skeleton className="h-80" key={item} />
              ))}
            </div>
          )}
          {content.isError && (
            <ErrorState
              compact
              description="Related editorial guidance could not be loaded."
              onRetry={() => void content.refetch()}
              title="Guidance unavailable"
            />
          )}
          {content.data?.length === 0 && (
            <EmptyState
              compact
              description="No reviewed editorial piece is connected to this category yet."
              title="No related guidance"
            />
          )}
          {content.data && content.data.length > 0 && (
            <ContentGrid entries={content.data} />
          )}
        </div>
      </Container>
    </section>
  )
}
