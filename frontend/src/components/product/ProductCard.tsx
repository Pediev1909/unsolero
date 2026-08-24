import { MoveUpRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { cn } from '../../lib/styles/cn'
import { BrandMonogram } from './BrandMonogram'
import { Badge } from '../ui/Badge'
import { PriceDisplay } from '../ui/PriceDisplay'

export interface ProductCardData {
  id: string
  name: string
  brand: string
  category: string
  priceMinor: number
  currency: string
  href?: string
  image?: {
    src: string
    alt: string
  }
  badge?: {
    label: string
    variant?: 'neutral' | 'accent' | 'success' | 'warning' | 'sponsored'
  }
}

interface ProductCardProps {
  product: ProductCardData
  priority?: boolean
  className?: string
}

export function ProductCard({
  product,
  priority = false,
  className,
}: ProductCardProps) {
  const content = (
    <>
      {/* Every product in a software catalog reaches this branch, so the
          media slot is never a photograph and a 4:3 box is 292px of empty
          paper on a phone before the name, the price or a button. The
          placeholder keeps a band deep enough to read as a deliberate mark
          and no deeper; 4:3 is kept for the day a vendor supplies artwork. */}
      <div
        className={cn(
          'relative overflow-hidden bg-paper',
          product.image ? 'aspect-[4/3]' : 'aspect-[16/6]',
        )}
      >
        {product.image ? (
          <img
            alt={product.image.alt}
            className="size-full object-cover transition-transform duration-280 ease-[var(--ease-standard)] group-hover:scale-[1.015] motion-reduce:transform-none"
            fetchPriority={priority ? 'high' : 'auto'}
            loading={priority ? 'eager' : 'lazy'}
            src={product.image.src}
          />
        ) : (
          // Most software has no product photograph, so the absence is the
          // normal case rather than a fault. A generic icon above the vendor
          // name announced the absence; a nameplate carrying the vendor's own
          // initials reads as a mark that was drawn on purpose. Replace with
          // the vendor's artwork once an affiliate programme supplies approved
          // brand assets.
          <BrandMonogram brand={product.brand} category={product.category} />
        )}
        {product.badge && (
          <Badge
            className="absolute left-3 top-3"
            variant={product.badge.variant}
          >
            {product.badge.label}
          </Badge>
        )}
      </div>
      <div className="grid gap-4 p-4 sm:p-5">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
              {product.brand} / {product.category}
            </p>
            <h3 className="mt-2 font-display text-xl font-medium leading-tight tracking-[-0.035em]">
              {product.name}
            </h3>
          </div>
          {product.href && (
            <MoveUpRight
              aria-hidden="true"
              className="mt-1 shrink-0 text-bronze transition-transform duration-180 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 motion-reduce:transform-none"
              size={18}
            />
          )}
        </div>
        <PriceDisplay
          amountMinor={product.priceMinor}
          currency={product.currency}
          size="sm"
        />
      </div>
    </>
  )

  return (
    <article
      className={cn(
        'group border-b border-r border-ink/15 bg-surface transition-colors duration-180',
        product.href && 'hover:bg-paper/55',
        className,
      )}
    >
      {product.href ? (
        <Link
          aria-label={`View details for ${product.name}`}
          className="block h-full"
          to={product.href}
        >
          {content}
        </Link>
      ) : (
        content
      )}
    </article>
  )
}
