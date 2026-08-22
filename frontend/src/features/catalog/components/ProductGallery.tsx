import { CalendarCheck, ExternalLink, Receipt } from 'lucide-react'
import { useState } from 'react'

import { cn } from '../../../lib/styles/cn'
import { priceCheckedOn, priceSource } from '../comparisonRows'
import type { ProductDetail } from '../schemas'
import { BrandMark } from './BrandMark'

/**
 * What sits where a product photograph would be.
 *
 * Software has no photograph, so this slot held a grey rectangle with the
 * vendor's name typed into it — the single most prominent area of the page,
 * saying nothing. A logo blown up to fill it would have been no better: a
 * 128-pixel favicon stretched across a third of the page looks like a mistake.
 *
 * So it holds the thing this site knows and no competitor prints: what the
 * price is, on what basis it is billed, and the day somebody opened the
 * vendor's page and read it — with a link to that page. A reader who wants to
 * check us can do it in one click from the top of the page.
 */
export function ProductGallery({ product }: { product: ProductDetail }) {
  const [selected, setSelected] = useState(0)
  const image = product.images[selected] ?? product.primary_image

  if (!image) return <PriceCard product={product} />

  return (
    <div>
      <div className="aspect-[4/3] overflow-hidden bg-paper">
        <img
          alt={image.alt_text}
          className="size-full object-cover"
          fetchPriority="high"
          height={image.height_px}
          loading="eager"
          src={image.url}
          width={image.width_px}
        />
      </div>
      {product.images.length > 1 && (
        <div
          aria-label="Product images"
          className="mt-3 grid grid-cols-4 gap-3"
          role="group"
        >
          {product.images.map((item, index) => (
            <button
              aria-label={`View image ${index + 1} of ${product.name}`}
              aria-pressed={selected === index}
              className={cn(
                'aspect-[4/3] overflow-hidden border bg-paper',
                selected === index
                  ? 'border-ink'
                  : 'border-transparent hover:border-ink/30',
              )}
              key={`${item.url}-${index}`}
              onClick={() => setSelected(index)}
              type="button"
            >
              <img
                alt=""
                className="size-full object-cover"
                height={item.height_px}
                loading="lazy"
                src={item.url}
                width={item.width_px}
              />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function PriceCard({ product }: { product: ProductDetail }) {
  const checked = priceCheckedOn(product)
  const source = priceSource(product)
  const free = product.price.amount_minor === 0

  return (
    <aside className="border border-ink/15 bg-surface">
      <div className="flex items-center gap-3 border-b border-ink/12 bg-paper px-5 py-4">
        <BrandMark
          brandName={product.brand.name}
          brandSlug={product.brand.slug}
          size="lg"
        />
        <div className="min-w-0">
          <p className="text-[0.625rem] font-bold tracking-[0.13em] text-ink/65 uppercase">
            {product.brand.name}
          </p>
          <p className="mt-0.5 truncate font-display text-lg font-medium tracking-[-0.02em]">
            {product.name}
          </p>
        </div>
      </div>

      <dl className="divide-y divide-ink/10">
        <Row
          icon={<Receipt aria-hidden="true" size={16} />}
          label="Entry price"
        >
          <span className="font-display text-3xl font-medium tracking-[-0.04em]">
            {free
              ? 'No monthly fee'
              : new Intl.NumberFormat('en-US', {
                  style: 'currency',
                  currency: product.price.currency,
                }).format(product.price.amount_minor / 100)}
          </span>
          <span className="mt-1 block text-sm text-ink/70">
            {product.key_specification.value || 'Billing basis not recorded'}
          </span>
        </Row>

        <Row
          icon={<CalendarCheck aria-hidden="true" size={16} />}
          label="Read from the vendor on"
        >
          <span className="text-base font-semibold">
            {checked ?? 'Not recorded'}
          </span>
          {source && (
            <a
              className="mt-1.5 inline-flex items-center gap-1 text-sm text-bronze underline-offset-4 hover:underline"
              href={source}
              rel="nofollow noopener noreferrer"
              target="_blank"
            >
              Open the page it came from
              <ExternalLink aria-hidden="true" size={13} />
            </a>
          )}
        </Row>
      </dl>

      <p className="border-t border-ink/12 bg-paper px-5 py-3 text-xs leading-5 text-ink/70">
        Not a live quote. Vendors change prices without notice, and this one is
        as good as its date.
      </p>
    </aside>
  )
}

function Row({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="px-5 py-4">
      <dt className="flex items-center gap-2 text-[0.625rem] font-bold tracking-[0.13em] text-ink/65 uppercase">
        <span className="text-bronze">{icon}</span>
        {label}
      </dt>
      <dd className="mt-2">{children}</dd>
    </div>
  )
}
