import { useState } from 'react'

import { cn } from '../../../lib/styles/cn'
import type { ProductDetail } from '../schemas'
import { BrandMark } from './BrandMark'

/**
 * What sits where a product photograph would be.
 *
 * Software has no photograph, so this slot has held two things in turn. First
 * a grey rectangle with the vendor's name typed into it. Then a price card —
 * entry price, billing basis, the day the vendor's page was read, a link to
 * that page — which was right about what matters and wrong about where it
 * sat: once the at-a-glance strip under the title carried the same facts, the
 * same number was printed twice on one screen. Two prices is one too many, and
 * the strip is the one a reader meets first, so it owns them now (see
 * ProductAtAGlance). The date, the source link and the "not a live quote"
 * caveat moved with the price rather than being dropped.
 *
 * Without images this is a small brand tile: the vendor's own mark at a size
 * it was drawn for, rather than a favicon stretched across a third of the
 * page. With images it is the gallery it always was.
 */
export function ProductGallery({ product }: { product: ProductDetail }) {
  const [selected, setSelected] = useState(0)
  const image = product.images[selected] ?? product.primary_image

  if (!image) return <BrandTile product={product} />

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

// Decorative: the vendor's name is the first line of text beside it.
function BrandTile({ product }: { product: ProductDetail }) {
  return (
    <div
      aria-hidden="true"
      className="flex size-20 items-center justify-center border border-ink/15 bg-surface sm:size-24"
    >
      <BrandMark
        brandName={product.brand.name}
        brandSlug={product.brand.slug}
        loading="eager"
        size="lg"
      />
    </div>
  )
}
