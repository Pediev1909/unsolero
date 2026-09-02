import { Bookmark, Check, ExternalLink, Scale } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { Link, useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Badge } from '../components/ui/Badge'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { Skeleton } from '../components/ui/Skeleton'
import { CatalogProductGrid } from '../features/catalog/components/CatalogProductGrid'
import { ProductAtAGlance } from '../features/catalog/components/ProductAtAGlance'
import { ProductComparison } from '../features/catalog/components/ProductComparison'
import { ProductEditorial } from '../features/catalog/components/ProductEditorial'
import { ProductGallery } from '../features/catalog/components/ProductGallery'
import { ProductOffers } from '../features/catalog/components/ProductOffers'
import { ProductProsCons } from '../features/catalog/components/ProductProsCons'
import { ProductSectionNav } from '../features/catalog/components/ProductSectionNav'
import { ProductSpecifications } from '../features/catalog/components/ProductSpecifications'
import {
  latestObservation,
  shortRevision,
} from '../features/catalog/components/productRecord'
import {
  productSectionIDs,
  productSections,
  sectionAnchorClass,
} from '../features/catalog/components/productSections'
import {
  evidenceFactLabel,
  visibleEvidence,
} from '../features/catalog/evidence'
import { specifications } from '../features/catalog/specifications'
import { suitabilityVariant } from '../features/catalog/model'
import { useOffers, useProduct } from '../features/catalog/queries'
import { useProductEditorial } from '../features/catalog/relatedContent'
import { useCatalogActions } from '../features/catalog/useCatalogActions'
import { cn } from '../lib/styles/cn'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { trackEvent } from '../features/analytics/tracking'

export function ProductDetailPage() {
  const { slug = '' } = useParams()
  const product = useProduct(slug)
  // Read here as well as inside the sections that draw them, so the jump row
  // can leave out a section that is going to render nothing. TanStack serves
  // both reads from one request.
  const offers = useOffers(slug)
  const editorial = useProductEditorial(slug)
  const actions = useCatalogActions()
  const viewedProductID = useRef<string | null>(null)

  useEffect(() => {
    if (!product.data || viewedProductID.current === product.data.id) return
    viewedProductID.current = product.data.id
    trackEvent('product_viewed', 'product_detail', {
      product_id: product.data.id,
    })
  }, [product.data])

  usePageMetadata({
    title: product.data
      ? `${product.data.name} by ${product.data.brand.name} | UNSOLERO`
      : 'Product details | UNSOLERO',
    description:
      product.data?.description ??
      'Structured product facts, suitability, alternatives, and current offers.',
    imagePath:
      product.data?.primary_image?.url ?? '/images/unsolero-saas-hero.svg',
    robots:
      product.data?.is_demo === false ? 'index, follow' : 'noindex, follow',
    type: 'product',
  })
  // The server already emits schema.org/Product into the shell, with the
  // offer attached and no invented ratings. A second block here described the
  // same product differently and left crawlers to pick one. The server's is
  // what they see without running any JavaScript, so it wins.

  if (product.isPending) return <ProductLoading />
  if (product.isError) {
    return (
      <PageFrame>
        <Container className="py-24">
          <ErrorState
            description="This product could not be found or loaded."
            onRetry={() => void product.refetch()}
            title="Product unavailable"
          />
        </Container>
      </PageFrame>
    )
  }

  const item = product.data
  const saved = actions.savedIDs.has(item.id)
  const hasImages = item.images.length > 0 || item.primary_image !== null
  const evidence = visibleEvidence(item)
  const observed = latestObservation(item)
  const sections = productSections({
    evidence: evidence.length > 0,
    offers: (offers.data?.length ?? 0) > 0,
    editorial: (editorial.data?.length ?? 0) > 0,
  })

  return (
    <PageFrame>
      <main id="main-content">
        <Container className="py-6 sm:py-8">
          {/* "Equipment" was the first crumb until now: a label left behind by
              the fitness catalog this site was before it sold software, sitting
              on every product page. The crumbs are also the smallest links on
              the page, so they carry a 24px hit box. */}
          <nav
            aria-label="Breadcrumb"
            className="flex flex-wrap items-center gap-x-2 text-xs text-ink/68"
          >
            <Link
              className="flex min-h-6 items-center hover:text-ink"
              to="/products"
            >
              Software
            </Link>
            <span aria-hidden="true">/</span>
            <Link
              className="flex min-h-6 items-center hover:text-ink"
              to={`/categories/${item.category.slug}`}
            >
              {item.category.name}
            </Link>
            <span aria-hidden="true">/</span>
            <span
              aria-current="page"
              className="flex min-h-6 items-center text-ink"
            >
              {item.name}
            </span>
          </nav>
        </Container>

        <Container className="pb-10 sm:pb-14">
          {/* With images the gallery takes the left column and the words the
              right. Without — which is every software product — that grid
              left a 1.15fr column holding a small brand tile and a wall of
              white beside the title, so the tile sits above the words instead.
              items-start, so a column does not stretch to the other's height. */}
          <div
            className={cn(
              hasImages &&
                'grid items-start gap-9 lg:grid-cols-[minmax(0,1.15fr)_minmax(22rem,0.85fr)] lg:gap-14',
            )}
          >
            <ProductGallery product={item} />
            <div className={hasImages ? 'lg:pt-5' : 'mt-6'}>
              <div className="flex flex-wrap items-center gap-2">
                <Link
                  className="inline-flex min-h-6 items-center text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark hover:text-ink"
                  to={`/brands/${item.brand.slug}`}
                >
                  {item.brand.name}
                </Link>
                {item.is_demo && (
                  <Badge variant="neutral">Fictional demo product</Badge>
                )}
              </div>
              <Heading className="mt-5" level={1} size="display">
                {item.name}
              </Heading>
              <p className="mt-5 max-w-xl text-base leading-7 text-ink/70">
                {item.description}
              </p>
            </div>
          </div>

          {/* How much, for whom, how well sourced, where: answered before the
              first scroll, once. The strip owns the price; nothing below
              prints it again. */}
          <ProductAtAGlance product={item} />

          <div className="mt-8 flex flex-wrap items-center justify-between gap-4 border-y border-ink/15 py-5">
            <p className="max-w-xs text-sm leading-6 text-ink/70">
              Keep it on a shortlist you can come back to.
            </p>
            <Button
              aria-pressed={saved}
              loading={actions.savePending}
              onClick={() => actions.save(item)}
              variant="secondary"
            >
              {saved ? (
                <Check aria-hidden="true" size={17} />
              ) : (
                <Bookmark aria-hidden="true" size={17} />
              )}
              {saved ? 'Saved' : 'Save product'}
            </Button>
          </div>

          <div className="mt-6">
            <h2 className="text-xs font-bold uppercase tracking-[0.14em]">
              Suitability
            </h2>
            <div className="mt-4 flex flex-wrap gap-2">
              {item.suitability.map((insight) => (
                <Badge
                  key={insight.key}
                  variant={suitabilityVariant(insight.score)}
                >
                  {insight.label} {insight.score}
                </Badge>
              ))}
            </div>
            <p className="mt-4 text-xs leading-5 text-ink/65">
              Scores are derived from structured catalog facts. They are not
              customer ratings or reviews.
            </p>
          </div>
        </Container>

        <ProductSectionNav sections={sections} />

        <section
          className={cn(
            'border-b border-ink/15 bg-paper py-16 sm:py-24',
            sectionAnchorClass,
          )}
          id={productSectionIDs.profile}
        >
          <Container>
            <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
              Decision profile
            </p>
            <Heading className="mt-4" level={2} size="title">
              Where it is strong, and what it costs you.
            </Heading>
            <div className="mt-9">
              <ProductProsCons
                strengths={item.strengths}
                useCases={item.use_cases}
                weaknesses={item.weaknesses}
              />
            </div>
            {/* A heading that promises facts and then lists nothing is worse
                than no heading. Software answers none of the physical rows,
                so on those products the whole block is absent. */}
            {specifications(item).length > 0 && (
              <div className="mt-14 border-t border-ink/15 pt-10">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
                  Product facts
                </p>
                <Heading className="mt-4" level={2} size="title">
                  Specifications
                </Heading>
                <div className="mt-9">
                  <ProductSpecifications product={item} />
                </div>
              </div>
            )}
          </Container>
        </section>

        {evidence.length > 0 && (
          <section
            className={cn(
              'border-b border-ink/15 py-16 sm:py-24',
              sectionAnchorClass,
            )}
            id={productSectionIDs.evidence}
          >
            <Container>
              <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
                Evidence record
              </p>
              <Heading className="mt-4" level={2} size="title">
                Where these product facts come from.
              </Heading>
              <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/70">
                Published facts are tied to reviewed sources and immutable
                revisions. Commercial terms are kept outside this evidence and
                never affect recommendation scores.
              </p>
              <div className="mt-9 grid gap-px border border-ink/15 bg-ink/15 md:grid-cols-2">
                {evidence.map((entry, index) => (
                  <article
                    className="bg-canvas p-5 sm:p-6"
                    key={`${entry.fact_key}-${entry.source_title}-${index}`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="neutral">
                        {evidenceLabel(entry.classification)}
                      </Badge>
                      {entry.is_fictional && (
                        <Badge variant="warning">Fictional demo evidence</Badge>
                      )}
                    </div>
                    <h3 className="mt-4 text-sm font-semibold">
                      {evidenceFactLabel(entry.fact_key)}
                    </h3>
                    <p className="mt-2 text-sm leading-6 text-ink/70">
                      {entry.source_title}
                    </p>
                    <p className="mt-3 text-xs text-ink/65">
                      Observed{' '}
                      {new Date(entry.observed_at).toLocaleDateString()}
                      {' · '}Confidence {entry.confidence}/100
                    </p>
                    {entry.source_url && (
                      <a
                        className="mt-4 inline-flex min-h-11 items-center gap-2 text-xs font-semibold text-bronze-dark hover:text-ink"
                        href={entry.source_url}
                        rel="noreferrer"
                        target="_blank"
                      >
                        Inspect source
                        <ExternalLink aria-hidden="true" size={14} />
                      </a>
                    )}
                  </article>
                ))}
              </div>
              {/* One line that names the record this page is: which fact and
                  score revisions it was rendered from, and the most recent day
                  any of it was observed. Not a price history — the API carries
                  one dated observation per fact, from the published revision
                  only; see productRecord.ts. */}
              <p className="mt-5 text-xs text-ink/65">
                Record: fact revision {shortRevision(item.fact_revision_id)} ·
                score revision {shortRevision(item.score_revision_id)}
                {observed && <> · observed {observed}</>}
              </p>
            </Container>
          </section>
        )}

        {/* Renders nothing when the vendor has no programme, which is most of
            them. The heading used to live here and the list inside the
            component, so a section title sat above an empty state on 45 of 53
            product pages. */}
        <ProductOffers slug={item.slug} />

        <ProductEditorial slug={item.slug} />

        <section
          className={cn(
            'border-t border-ink/15 py-16 sm:py-24',
            sectionAnchorClass,
          )}
          id={productSectionIDs.alternatives}
        >
          <Container>
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
                  Same category
                </p>
                <Heading className="mt-4" level={2} size="title">
                  Alternatives worth considering
                </Heading>
              </div>
              <Link
                className="text-sm font-semibold text-bronze-dark hover:text-ink"
                to={`/categories/${item.category.slug}`}
              >
                View all {item.category.name.toLowerCase()}
              </Link>
            </div>
            <div className="mt-9">
              {item.alternatives.length > 0 ? (
                <CatalogProductGrid
                  comparedIDs={actions.comparedIDs}
                  onCompare={actions.compare}
                  onSave={actions.save}
                  products={item.alternatives}
                  savedIDs={actions.savedIDs}
                  savePending={actions.savePending}
                />
              ) : (
                <p className="border-y border-ink/15 py-8 text-sm text-ink/70">
                  No comparable alternatives are available in this category.
                </p>
              )}
            </div>
          </Container>
        </section>
      </main>

      {actions.comparedIDs.size > 0 && (
        <Button
          className="fixed bottom-4 right-4 z-30 shadow-overlay"
          onClick={() => actions.setComparisonOpen(true)}
        >
          <Scale aria-hidden="true" size={17} /> Compare{' '}
          {actions.comparedIDs.size}
        </Button>
      )}
      <ProductComparison
        error={actions.comparisonError}
        loading={actions.comparisonPending}
        onOpenChange={actions.setComparisonOpen}
        onRemove={actions.compare}
        open={actions.comparisonOpen}
        products={actions.compared}
      />
    </PageFrame>
  )
}

function evidenceLabel(classification: string) {
  switch (classification) {
    case 'verified_fact':
      return 'Verified fact'
    case 'manufacturer_claim':
      return 'Manufacturer claim'
    case 'merchant_observation':
      return 'Merchant observation'
    default:
      return 'Editorial assessment'
  }
}

function PageFrame({ children }: { children: React.ReactNode }) {
  return (
    <>
      <SiteHeader position="sticky" />
      {children}
      <SiteFooter />
    </>
  )
}

function ProductLoading() {
  return (
    <PageFrame>
      <main id="main-content">
        <Container className="py-16">
          <Skeleton className="size-20 sm:size-24" />
          <Skeleton className="mt-6 h-4 w-24" />
          <Skeleton className="mt-6 h-20 w-4/5" />
          <Skeleton className="mt-7 h-24 w-full max-w-xl" />
          <Skeleton className="mt-8 h-40 w-full" />
        </Container>
      </main>
    </PageFrame>
  )
}
