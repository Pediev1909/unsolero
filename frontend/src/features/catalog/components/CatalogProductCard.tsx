import { Bookmark, Check, Scale } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { cn } from '../../../lib/styles/cn'
import { prominentSuitability } from '../model'
import type { ProductSummary } from '../schemas'

interface CatalogProductCardProps {
  product: ProductSummary
  compared: boolean
  saved: boolean
  savePending?: boolean
  onCompare: (product: ProductSummary) => void
  onSave: (product: ProductSummary) => void
}

export function CatalogProductCard({
  product,
  compared,
  saved,
  savePending = false,
  onCompare,
  onSave,
}: CatalogProductCardProps) {
  const suitability = prominentSuitability(product)

  return (
    <article className="group flex h-full flex-col border-b border-r border-ink/15 bg-surface">
        {/* Every product in a software catalog reaches this branch, so the
            media slot is never a photograph and a 4:3 box is 292px of empty
            paper on a phone before the name, the price or a button. The
            placeholder keeps a band deep enough to read as a deliberate mark
            and no deeper; 4:3 is kept for the day a vendor supplies artwork. */}
      <Link
        aria-label={`View ${product.name}`}
        className={cn(
          'relative block overflow-hidden bg-paper',
          product.primary_image ? 'aspect-[4/3]' : 'aspect-[16/5] sm:aspect-[16/6]',
        )}
        to={`/products/${product.slug}`}
      >
        {product.primary_image ? (
          <img
            alt={product.primary_image.alt_text}
            className="size-full object-cover transition-transform duration-280 group-hover:scale-[1.015] motion-reduce:transform-none"
            height={product.primary_image.height_px}
            loading="lazy"
            src={product.primary_image.url}
            width={product.primary_image.width_px}
          />
        ) : (
          // Software has no product photograph, so the absence is the normal
          // case rather than a fault. Naming the vendor reads as a deliberate
          // mark and still says whose product this is; "Image unavailable"
          // made every card look broken.
          <div className="grid size-full place-items-center px-4 text-center text-sm font-semibold text-ink/70">
            {product.brand.name}
          </div>
        )}
        {product.is_demo && (
          <Badge className="absolute left-3 top-3" variant="neutral">
            Demo product
          </Badge>
        )}
      </Link>

      <div className="flex flex-1 flex-col p-4 sm:p-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            {/* Small caps set at 10px left an 11px-tall tap target. The hit
                box is raised to 24px and pulled back up by the same amount it
                added, so the card's spacing is unchanged. */}
            <Link
              className="-mt-1.5 inline-flex min-h-6 items-center text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/65 hover:text-bronze-dark"
              to={`/brands/${product.brand.slug}`}
            >
              {product.brand.name}
            </Link>
            <h2 className="mt-2 font-display text-xl font-medium leading-tight tracking-[-0.035em]">
              <Link
                className="hover:text-bronze-dark"
                to={`/products/${product.slug}`}
              >
                {product.name}
              </Link>
            </h2>
          </div>
          <Button
            aria-label={
              saved
                ? `Remove ${product.name} from saved products`
                : `Save ${product.name}`
            }
            aria-pressed={saved}
            className={cn('shrink-0', saved && 'text-bronze-dark')}
            disabled={savePending}
            onClick={() => onSave(product)}
            size="sm"
            variant="quiet"
          >
            {saved ? (
              <Check aria-hidden="true" size={18} />
            ) : (
              <Bookmark aria-hidden="true" size={18} />
            )}
          </Button>
        </div>

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
      </div>
    </article>
  )
}
