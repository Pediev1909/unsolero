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
      <Link
        aria-label={`View ${product.name}`}
        className="relative block aspect-[4/3] overflow-hidden bg-paper"
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
          <div className="grid size-full place-items-center text-sm text-ink/65">
            Image unavailable
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
            <Link
              className="text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/65 hover:text-bronze-dark"
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

        <div className="mt-4 flex min-h-7 flex-wrap gap-2">
          {suitability.map((item) => (
            <Badge key={item.key} variant="success">
              {item.label} {item.score}
            </Badge>
          ))}
        </div>

        <Link
          className="mt-4 text-xs font-semibold text-ink/70 hover:text-bronze-dark"
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
