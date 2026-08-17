import { cn } from '../../lib/styles/cn'

export type ButtonVariant = 'primary' | 'secondary' | 'quiet' | 'danger'
export type ButtonSize = 'sm' | 'md' | 'lg'

const variants: Record<ButtonVariant, string> = {
  primary:
    'border border-ink bg-ink text-canvas hover:border-bronze-dark hover:bg-bronze-dark',
  secondary:
    'border border-ink/30 bg-transparent text-ink hover:border-ink hover:bg-ink hover:text-canvas',
  quiet: 'border border-transparent bg-transparent text-ink hover:bg-paper',
  danger:
    'border border-ember bg-ember text-surface hover:border-ink hover:bg-ink',
}

const sizes: Record<ButtonSize, string> = {
  sm: 'min-h-10 px-3.5 text-xs',
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
