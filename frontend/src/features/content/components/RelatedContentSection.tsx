import { Container } from '../../../components/ui/Container'
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

  // Nothing at all when there is nothing to show. Five of fifteen categories
  // have no editorial linked to them yet, and each was rendering a section
  // heading, a promise about "curated guidance" and then an empty box. An
  // empty state is for a list a reader emptied themselves by filtering; this
  // one only ever said that we have not written the piece.
  //
  // Errors are silent for the same reason: this sits below the whole catalog
  // listing, and an error panel there reports a broken page when the page is
  // fine.
  if (content.isError || content.data?.length === 0) return null

  return (
    <section className="border-t border-ink/15 bg-paper py-14 sm:py-20">
      <Container>
        <p className="eyebrow">Editorial guidance</p>
        <Heading className="mt-4" level={2} size="title">
          Research before you shortlist.
        </Heading>
        <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/70">
          Curated planning and buying guidance connected to this software
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
          {content.data && content.data.length > 0 && (
            <ContentGrid entries={content.data} />
          )}
        </div>
      </Container>
    </section>
  )
}
