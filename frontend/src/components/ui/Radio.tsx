import {
  forwardRef,
  useId,
  type InputHTMLAttributes,
  type ReactNode,
} from 'react'

import { cn } from '../../lib/styles/cn'

export interface RadioProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'type'
> {
  label: ReactNode
  description?: ReactNode
}

export const Radio = forwardRef<HTMLInputElement, RadioProps>(function Radio(
  {
    label,
    description,
    className,
    id: suppliedId,
    'aria-describedby': describedBy,
    ...props
  },
  ref,
) {
  const generatedId = useId()
  const id = suppliedId ?? generatedId
  const descriptionId = description ? `${id}-description` : undefined

  return (
    <div className="flex items-start gap-3">
      <input
        {...props}
        aria-describedby={cn(describedBy, descriptionId) || undefined}
        className={cn(
          'mt-0.5 size-5 shrink-0 cursor-pointer accent-bronze disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
        id={id}
        ref={ref}
        type="radio"
      />
      <div>
        <label className="cursor-pointer text-sm font-semibold" htmlFor={id}>
          {label}
        </label>
        {description && (
          <p className="mt-1 text-xs leading-5 text-ink/55" id={descriptionId}>
            {description}
          </p>
        )}
      </div>
    </div>
  )
})
