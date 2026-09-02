import { Bookmark, BookmarkCheck, ExternalLink, Scale } from 'lucide-react'
import { Link } from 'react-router-dom'

import {
  affiliateClickPath,
  type AffiliateSource,
} from '../../analytics/tracking'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { cn } from '../../../lib/styles/cn'
import { prominentSuitability } from '../model'
import { BrandMark } from './BrandMark'
import type { ProductSummary } from '../schemas'

interface CatalogProductCardProps {
  product: ProductSummary
  compared: boolean
  saved: boolean
  savePending?: boolean
  source?: AffiliateSource
  onCompare: (product: ProductSummary) => void
  onSave: (product: ProductSummary) => void
}

export function CatalogProductCard({
  product,
  compared,
  saved,
  savePending = false,
  source = 'product_detail',
  onCompare,
  onSave,
}: CatalogProductCardProps) {
  const suitability = prominentSuitability(product)
  const purchasePath = product.purchase_path

  return (
    <article className="group flex h-full flex-col border-b border-r border-ink/15 bg-surface">
      {/* The media slot used to be a grey band with the vendor's name typed
          into it, which is what a placeholder looks like rather than what a
          product looks like. A page of six of them read as a page that had
          failed to load. The vendor's own mark sits in the card body now,
          where it does the job the band was pretending to do: tell you at a
          glance whose product this is. */}

      <div className="flex flex-1 flex-col p-4 sm:p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-3">
            <BrandMark
              brandName={product.brand.name}
              brandSlug={product.brand.slug}
              size="md"
            />
            <div className="min-w-0">
              {/* Small caps set at 10px left an 11px-tall tap target. The hit
                  box is raised to 24px and pulled back up by the same amount
                  it added, so the card's spacing is unchanged. */}
              <Link
                className="-mt-1 inline-flex min-h-6 items-center text-[0.625rem] font-bold tracking-[0.13em] text-ink/65 uppercase hover:text-bronze-dark"
                to={`/brands/${product.brand.slug}`}
              >
                {product.brand.name}
              </Link>
              <h2 className="mt-1 font-display text-xl leading-tight font-medium tracking-[-0.035em]">
                <Link
                  className="hover:text-bronze-dark"
                  to={`/products/${product.slug}`}
                >
                  {product.name}
                </Link>
              </h2>
            </div>
          </div>
          {/* A toggle, so the state is in aria-pressed and the word changes
              with it. The product name rides in the accessible name because a
              grid holds twelve of these and "Save, Save, Save" tells a screen
              reader nothing about which one. */}
          <Button
            aria-label={`${saved ? 'Saved' : 'Save'} ${product.name}`}
            aria-pressed={saved}
            className={cn('shrink-0', saved && 'text-bronze-dark')}
            disabled={savePending}
            onClick={() => onSave(product)}
            size="sm"
            variant="quiet"
          >
            {saved ? (
              <BookmarkCheck aria-hidden="true" size={18} />
            ) : (
              <Bookmark aria-hidden="true" size={18} />
            )}
            {saved ? 'Saved' : 'Save'}
          </Button>
        </div>

        {product.is_demo && (
          <Badge className="mt-3 self-start" variant="neutral">
            Demo product
          </Badge>
        )}

        <div className="mt-5 flex items-end justify-between gap-4 border-y border-ink/10 py-4">
          <PriceDisplay
            amountMinor={product.price.amount_minor}
            currency={product.price.currency}
            size="md"
          />
          <div className="text-right">
            <p className="text-[0.5625rem] font-bold uppercase tracking-[0.12em] text-ink/65">
              {product.key_specification.label}
            </p>
            <p className="mt-1 text-xs font-semibold">
              {product.key_specification.value}
            </p>
          </div>
        </div>

        {/* Reserved height keeps the cards in a grid row aligned, but a card
            whose scores did not clear the threshold then carries an empty
            band. On a phone the cards are stacked full width, so nothing is
            being aligned to and the band is just a hole. */}
        {suitability.length > 0 && (
          <div className="mt-4 flex min-h-7 flex-wrap gap-2">
            {suitability.map((item) => (
              <Badge key={item.key} variant="success">
                {item.label} {item.score}
              </Badge>
            ))}
          </div>
        )}

        <Link
          className="mt-3 flex min-h-6 items-center text-xs font-semibold text-ink/70 hover:text-bronze-dark"
          to={`/categories/${product.category.slug}`}
        >
          {product.category.name}
        </Link>

        <div className="mt-auto grid grid-cols-2 gap-2 pt-6">
          <Button
            aria-pressed={compared}
            onClick={() => onCompare(product)}
            size="sm"
            variant={compared ? 'primary' : 'secondary'}
          >
            <Scale aria-hidden="true" size={15} />
            {compared ? 'Compared' : 'Compare'}
          </Button>
          <ButtonLink
            size="sm"
            to={`/products/${product.slug}`}
            variant="secondary"
          >
            View details
          </ButtonLink>
        </div>

        {/* The vendor exit, on the card rather than two clicks past it.
            This one component draws the catalog, the category and brand
            pages, and the "Products referenced" grid under every alternatives
            and versus article — the pages written for people about to choose,
            which until now ended in an internal link.

            It sits below Compare and View details, not above them, because
            the card's job is still to help someone decide and this is what
            they do after deciding. Cards with no live offer show nothing
            here: a disabled button would imply we are withholding something,
            and an empty slot reads as what it is. */}
        {purchasePath && (
          <div className="mt-3 border-t border-ink/10 pt-3">
            <a
              className="inline-flex min-h-11 w-full items-center justify-center gap-2 bg-charcoal px-4 text-sm font-semibold text-canvas hover:bg-ink"
              href={affiliateClickPath(purchasePath, source)}
              rel="nofollow noopener sponsored"
              target="_blank"
            >
              View at {product.merchant_name ?? 'vendor'}
              <ExternalLink aria-hidden="true" size={14} />
            </a>
            <p className="mt-2 text-[0.6875rem] leading-4 text-ink/65">
              {product.disclosure_label ?? 'Affiliate link'}. Commission never
              changes the ranking.
            </p>
          </div>
        )}
      </div>
    </article>
  )
}
