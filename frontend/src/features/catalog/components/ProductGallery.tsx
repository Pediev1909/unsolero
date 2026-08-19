import { useState } from 'react'

import { cn } from '../../../lib/styles/cn'
import type { ProductDetail } from '../schemas'

export function ProductGallery({ product }: { product: ProductDetail }) {
  const [selected, setSelected] = useState(0)
  const image = product.images[selected] ?? product.primary_image

  return (
    <div>
      <div
        className={cn(
          'overflow-hidden bg-paper',
          image ? 'aspect-[4/3]' : 'aspect-[16/6]',
        )}
      >
        {image ? (
          <img
            alt={image.alt_text}
            className="size-full object-cover"
            fetchPriority="high"
            height={image.height_px}
            loading="eager"
            src={image.url}
            width={image.width_px}
          />
        ) : (
          // Software has no product photograph, so the absence is the normal
          // case rather than a fault. Naming the vendor reads as a deliberate
          // mark; "Image unavailable" made the page look broken.
          <div className="grid size-full place-items-center px-6 text-center text-xl font-semibold tracking-[-0.01em] text-ink/70">
            {product.brand.name}
          </div>
        )}
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
