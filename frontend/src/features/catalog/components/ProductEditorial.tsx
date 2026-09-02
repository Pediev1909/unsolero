import { Container } from '../../../components/ui/Container'
import { Heading } from '../../../components/ui/Heading'
import { Skeleton } from '../../../components/ui/Skeleton'
import { cn } from '../../../lib/styles/cn'
import { ContentGrid } from '../../content/components/ContentGrid'
import { useProductEditorial } from '../relatedContent'
import { productSectionIDs, sectionAnchorClass } from './productSections'

/**
 * The comparisons and guides this product appears in.
 *
 * A reader on a product page is one step from a decision, and the pieces that
 * set this product against the alternatives are the most useful thing the site
 * has for that step. Every competitor links them from the product; ours were
 * reachable only from the content hub.
 *
 * Nothing at all when there is nothing to show, and nothing on error: this
 * sits below the price, the profile and the evidence, and a heading over an
 * empty box — or an error panel — would report a broken page when the page is
 * fine. While loading, the placeholders come without the heading, so a product
 * nobody has written about never flashes a title that then disappears.
 */
export function ProductEditorial({ slug }: { slug: string }) {
  const editorial = useProductEditorial(slug)

  if (editorial.isError || editorial.data?.length === 0) return null

  return (
    <section
      className={cn(
        'border-t border-ink/15 bg-paper py-16 sm:py-24',
        sectionAnchorClass,
      )}
      id={productSectionIDs.editorial}
    >
      <Container>
        {editorial.isPending ? (
          <div
            aria-label="Loading the comparisons this product appears in"
            className="grid gap-5 md:grid-cols-2 xl:grid-cols-3"
            role="status"
          >
            {[0, 1, 2].map((item) => (
              <Skeleton className="h-80" key={item} />
            ))}
          </div>
        ) : (
          <>
            <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
              Editorial
            </p>
            <Heading className="mt-4" level={2} size="title">
              Compared in
            </Heading>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/70">
              The guides and comparisons on this site that weigh this product
              against others.
            </p>
            <div className="mt-9">
              <ContentGrid entries={editorial.data} />
            </div>
          </>
        )}
      </Container>
    </section>
  )
}
