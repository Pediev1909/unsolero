import { cn } from '../../lib/styles/cn'

export type ButtonVariant =
  'primary' | 'secondary' | 'quiet' | 'danger' | 'inverse'
export type ButtonSize = 'sm' | 'md' | 'lg'

const variants: Record<ButtonVariant, string> = {
  primary:
    'border border-ink bg-ink text-canvas hover:border-bronze-dark hover:bg-bronze-dark',
  secondary:
    'border border-ink/30 bg-transparent text-ink hover:border-ink hover:bg-ink hover:text-canvas',
  quiet: 'border border-transparent bg-transparent text-ink hover:bg-paper',
  danger:
    'border border-ember bg-ember text-surface hover:border-ink hover:bg-ink',
  // For a button sitting on an ink surface. A variant rather than a className
  // override, because cn() concatenates: passing bg-canvas alongside the
  // variant's bg-ink left both on the element and let the stylesheet's own
  // order pick. On the compare bar bg-ink won — a black button on a black
  // bar, 1.00:1 contrast, invisible.
  inverse:
    'border border-canvas bg-canvas text-ink hover:border-bronze-soft hover:bg-bronze-soft',
}

const sizes: Record<ButtonSize, string> = {
  // 44px, not 40. The small size carries the consent banner's two buttons, the
  // catalog card's compare and save controls and the mobile menu toggle, which
  // are among the most-tapped things on the site.
  sm: 'min-h-11 px-3.5 text-xs',
  md: 'min-h-12 px-5 text-sm',
  lg: 'min-h-14 px-6 text-sm',
}

export function buttonStyles({
  variant = 'primary',
  size = 'md',
  fullWidth = false,
  className,
}: {
  variant?: ButtonVariant
  size?: ButtonSize
  fullWidth?: boolean
  className?: string
} = {}): string {
  return cn(
    'inline-flex items-center justify-center gap-2 rounded-xs font-semibold tracking-[-0.01em] transition-[background-color,border-color,color,transform] duration-180 ease-[var(--ease-standard)] hover:-translate-y-px disabled:pointer-events-none disabled:opacity-50 motion-reduce:transform-none',
    variants[variant],
    sizes[size],
    fullWidth && 'w-full',
    className,
  )
}
