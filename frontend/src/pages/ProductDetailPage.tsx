import {
  AlertTriangle,
  Bookmark,
  Check,
  CheckCircle2,
  ExternalLink,
  Scale,
  Target,
} from 'lucide-react'
import { useEffect, useRef } from 'react'
import { Link, useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Badge } from '../components/ui/Badge'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { PriceDisplay } from '../components/ui/PriceDisplay'
import { Skeleton } from '../components/ui/Skeleton'
import { CatalogProductGrid } from '../features/catalog/components/CatalogProductGrid'
import { ProductComparison } from '../features/catalog/components/ProductComparison'
import { ProductGallery } from '../features/catalog/components/ProductGallery'
import { ProductOffers } from '../features/catalog/components/ProductOffers'
import { ProductSpecifications } from '../features/catalog/components/ProductSpecifications'
import { suitabilityVariant } from '../features/catalog/model'
import { useProduct } from '../features/catalog/queries'
import type { ProductInsight } from '../features/catalog/schemas'
import { useCatalogActions } from '../features/catalog/useCatalogActions'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { useStructuredData } from '../lib/seo/useStructuredData'
import { trackEvent } from '../features/analytics/tracking'

export function ProductDetailPage() {
  const { slug = '' } = useParams()
  const product = useProduct(slug)
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
      : 'Equipment details | UNSOLERO',
    description:
      product.data?.description ??
      'Structured equipment specifications, suitability, alternatives, and merchant offers.',
    imagePath:
      product.data?.primary_image?.url ?? '/images/unsolero-saas-hero.webp',
    robots:
      product.data?.is_demo === false ? 'index, follow' : 'noindex, follow',
    type: 'product',
  })
  useStructuredData(
    'product',
    product.data?.is_demo === false
      ? {
          '@context': 'https://schema.org',
          '@type': 'Product',
          name: product.data.name,
          description: product.data.description,
          image: product.data.images.map(
            (image) => new URL(image.url, window.location.origin).href,
          ),
          brand: { '@type': 'Brand', name: product.data.brand.name },
          category: product.data.category.name,
          url: window.location.href,
        }
      : null,
  )

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

  return (
    <PageFrame>
      <main id="main-content">
        <Container className="py-6 sm:py-8">
          <nav
            aria-label="Breadcrumb"
            className="flex flex-wrap items-center gap-2 text-xs text-ink/50"
          >
            <Link className="hover:text-ink" to="/products">
              Equipment
            </Link>
            <span aria-hidden="true">/</span>
            <Link
              className="hover:text-ink"
              to={`/categories/${item.category.slug}`}
            >
              {item.category.name}
            </Link>
            <span aria-hidden="true">/</span>
            <span aria-current="page" className="text-ink">
              {item.name}
            </span>
          </nav>
        </Container>

        <Container className="pb-16 sm:pb-24">
          <div className="grid gap-9 lg:grid-cols-[minmax(0,1.15fr)_minmax(22rem,0.85fr)] lg:gap-14">
            <ProductGallery product={item} />
            <div className="lg:pt-5">
              <div className="flex flex-wrap items-center gap-2">
                <Link
                  className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark hover:text-ink"
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
              <p className="mt-6 max-w-xl text-base leading-7 text-ink/65">
                {item.description}
              </p>
              <div className="mt-8 flex flex-wrap items-center justify-between gap-5 border-y border-ink/15 py-6">
                <div>
                  <p className="text-[0.625rem] font-bold uppercase tracking-[0.12em] text-ink/45">
                    Reference price
                  </p>
                  <PriceDisplay
                    amountMinor={item.price.amount_minor}
                    className="mt-2"
                    currency={item.price.currency}
                    size="lg"
                  />
                </div>
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

              <div className="mt-7">
                <h2 className="text-xs font-bold uppercase tracking-[0.14em]">
                  Suitability at a glance
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
                <p className="mt-4 text-xs leading-5 text-ink/45">
                  Scores are derived from structured catalog facts. They are not
                  customer ratings or reviews.
                </p>
              </div>
            </div>
          </div>
        </Container>

        <section className="border-y border-ink/15 bg-paper py-16 sm:py-24">
          <Container>
            <div className="grid gap-14 lg:grid-cols-2 lg:gap-20">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
                  Decision profile
                </p>
                <Heading className="mt-4" level={2} size="title">
                  Where it excels—and where it does not.
                </Heading>
                <div className="mt-9 space-y-8">
                  <InsightList
                    icon={<CheckCircle2 aria-hidden="true" size={19} />}
                    insights={item.strengths}
                    title="Strengths"
                    empty="No standout strength crossed the current evidence threshold."
                  />
                  <InsightList
                    icon={<AlertTriangle aria-hidden="true" size={19} />}
                    insights={item.weaknesses}
                    title="Trade-offs"
                    empty="No material weakness crossed the current evidence threshold."
                  />
                  <InsightList
                    icon={<Target aria-hidden="true" size={19} />}
                    insights={item.use_cases}
                    title="Best use cases"
                    empty="No use case crossed the current evidence threshold."
                  />
                </div>
              </div>
              <div>
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
            </div>
          </Container>
        </section>

        <section className="border-b border-ink/15 py-16 sm:py-24">
          <Container>
            <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
              Evidence record
            </p>
            <Heading className="mt-4" level={2} size="title">
              Where these product facts come from.
            </Heading>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/60">
              Published facts are tied to reviewed sources and immutable
              revisions. Commercial terms are kept outside this evidence and
              never affect recommendation scores.
            </p>
            <div className="mt-9 grid gap-px border border-ink/15 bg-ink/15 md:grid-cols-2">
              {item.evidence.map((evidence, index) => (
                <article
                  className="bg-canvas p-5 sm:p-6"
                  key={`${evidence.fact_key}-${evidence.source_title}-${index}`}
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="neutral">
                      {evidenceLabel(evidence.classification)}
                    </Badge>
                    {evidence.is_fictional && (
                      <Badge variant="warning">Fictional demo evidence</Badge>
                    )}
                  </div>
                  <h3 className="mt-4 text-sm font-semibold capitalize">
                    {evidence.fact_key.replaceAll('_', ' ')}
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-ink/60">
                    {evidence.source_title}
                  </p>
                  <p className="mt-3 text-xs text-ink/45">
                    Observed{' '}
                    {new Date(evidence.observed_at).toLocaleDateString()}
                    {' · '}Confidence {evidence.confidence}/100
                  </p>
                  {evidence.source_url && (
                    <a
                      className="mt-4 inline-flex min-h-11 items-center gap-2 text-xs font-semibold text-bronze-dark hover:text-ink"
                      href={evidence.source_url}
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
            <p className="mt-5 text-xs text-ink/40">
              Fact revision {item.fact_revision_id.slice(0, 8)} · Score revision{' '}
              {item.score_revision_id.slice(0, 8)}
            </p>
          </Container>
        </section>

        <section className="py-16 sm:py-24">
          <Container>
            <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
              Merchant comparison
            </p>
            <Heading className="mt-4" level={2} size="title">
              Available offers
            </Heading>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-ink/60">
              Compare the landed price and merchant status. The purchase button
              uses our tracked redirect; merchant destination URLs are never
              exposed by the API.
            </p>
            <div className="mt-9">
              <ProductOffers isDemo={item.is_demo} slug={item.slug} />
            </div>
          </Container>
        </section>

        <section className="border-t border-ink/15 py-16 sm:py-24">
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
                <p className="border-y border-ink/15 py-8 text-sm text-ink/60">
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
        <Container className="grid gap-10 py-16 lg:grid-cols-2">
          <Skeleton className="aspect-[4/3] w-full" />
          <div>
            <Skeleton className="h-4 w-24" />
            <Skeleton className="mt-6 h-20 w-4/5" />
            <Skeleton className="mt-7 h-24 w-full" />
            <Skeleton className="mt-8 h-24 w-full" />
          </div>
        </Container>
      </main>
    </PageFrame>
  )
}

function InsightList({
  icon,
  insights,
  title,
  empty,
}: {
  icon: React.ReactNode
  insights: ProductInsight[]
  title: string
  empty: string
}) {
  return (
    <div>
      <h3 className="flex items-center gap-2 text-sm font-semibold">
        {icon}
        {title}
      </h3>
      {insights.length ? (
        <ul className="mt-3 space-y-2 text-sm text-ink/65">
          {insights.map((insight) => (
            <li
              className="flex justify-between gap-4 border-b border-ink/10 pb-2"
              key={insight.key}
            >
              <span>{insight.label}</span>
              <span className="font-semibold text-ink">
                {insight.score}/100
              </span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-3 text-sm leading-6 text-ink/50">{empty}</p>
      )}
    </div>
  )
}
