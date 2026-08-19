import type { InputHTMLAttributes, ReactNode } from 'react'

import { cn } from '../../../lib/styles/cn'

interface ChoiceCardProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'type'
> {
  label: ReactNode
  description?: ReactNode
  type?: 'radio' | 'checkbox'
}

export function ChoiceCard({
  label,
  description,
  type = 'radio',
  className,
  ...props
}: ChoiceCardProps) {
  return (
    <label className={cn('group relative block cursor-pointer', className)}>
      <input {...props} className="peer sr-only" type={type} />
      <span className="flex min-h-28 flex-col border border-ink/15 bg-surface p-4 transition-colors group-hover:border-ink/40 peer-checked:border-ink peer-checked:bg-charcoal peer-checked:text-canvas peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-bronze sm:p-5">
        <span className="text-base font-semibold tracking-[-0.02em]">
          {label}
        </span>
        {/* Opacity rather than a fixed colour, so the description follows
            whatever the card is currently painted in. The previous rule tried
            to switch colour with peer-checked:, which only reaches siblings of
            the input — this span is nested, so the selected card kept dark text
            on its dark background and the description vanished. */}
        {description && (
          <span className="mt-2 text-sm leading-6 opacity-75">
            {description}
          </span>
        )}
      </span>
    </label>
  )
}
