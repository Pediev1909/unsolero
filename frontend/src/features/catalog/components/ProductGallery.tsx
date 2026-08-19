import { useState } from 'react'

import { cn } from '../../../lib/styles/cn'
import type { ProductDetail } from '../schemas'

export function ProductGallery({ product }: { product: ProductDetail }) {
  const [selected, setSelected] = useState(0)
  const image = product.images[selected] ?? product.primary_image

  return (
    <div>
      <div className="aspect-[4/3] overflow-hidden bg-paper">
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
          <div className="grid size-full place-items-center text-sm text-ink/65">
            Image unavailable
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
