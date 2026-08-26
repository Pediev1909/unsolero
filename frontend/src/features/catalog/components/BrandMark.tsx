import { useState } from 'react'

import { cn } from '../../../lib/styles/cn'

const sizes = {
  sm: 'h-7 w-7',
  md: 'h-10 w-10',
  lg: 'h-14 w-14',
} as const

interface BrandMarkProps {
  brandName: string
  brandSlug: string
  size?: keyof typeof sizes
  className?: string
  loading?: 'eager' | 'lazy'
}

/**
 * A vendor's own logo, next to their product's name.
 *
 * Every product in the catalog rendered as a name on a grey rectangle, which
 * told a reader nothing and made a page of six products look like a page of
 * placeholders. A recognisable mark is the fastest possible answer to "is this
 * the thing I was thinking of".
 *
 * These are the vendors' own icons, self-hosted rather than hotlinked so that
 * no visitor's browser has to call a third party to render this page. Using a
 * mark to identify the product it belongs to, in a comparison, is nominative
 * use; the marks stay unmodified, at one size, and the disclaimer sits in the
 * footer. Two of forty-six vendors publish nothing usable, and they fall back
 * to their initial rather than to an empty box.
 */
export function BrandMark({
  brandName,
  brandSlug,
  size = 'md',
  className,
  loading = 'lazy',
}: BrandMarkProps) {
  const [failed, setFailed] = useState(false)

  if (failed) {
    return (
      <span
        aria-hidden="true"
        className={cn(
          'flex shrink-0 items-center justify-center rounded-sm bg-paper font-display font-semibold text-ink/70',
          sizes[size],
          className,
        )}
      >
        {brandName.trim().charAt(0).toUpperCase()}
      </span>
    )
  }

  return (
    <img
      alt=""
      className={cn(
        'shrink-0 rounded-sm object-contain',
        sizes[size],
        className,
      )}
      // Decorative: the vendor's name is always printed beside it, so a screen
      // reader announcing the logo as well would just say everything twice.
      aria-hidden="true"
      height={128}
      loading={loading}
      onError={() => setFailed(true)}
      src={`/images/brands/${brandSlug}.png`}
      width={128}
    />
  )
}
