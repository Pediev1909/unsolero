import {
  forwardRef,
  useId,
  type ReactNode,
  type SelectHTMLAttributes,
} from 'react'
import { ChevronDown } from 'lucide-react'

import { cn } from '../../lib/styles/cn'
import { FieldFrame } from './FieldFrame'

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: ReactNode
  hint?: ReactNode
  error?: ReactNode
  containerClassName?: string
  children: ReactNode
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  function Select(
    {
      label,
      hint,
      error,
      containerClassName,
      className,
      id: suppliedId,
      required,
      children,
      'aria-describedby': describedBy,
      ...props
    },
    ref,
  ) {
    const generatedId = useId()
    const id = suppliedId ?? generatedId
    const descriptionId = error
      ? `${id}-error`
      : hint
        ? `${id}-hint`
        : undefined

    return (
      <FieldFrame
        className={containerClassName}
        error={error}
        hint={hint}
        id={id}
        label={label}
        required={required}
      >
        <div className="relative">
          <select
            {...props}
            aria-describedby={cn(describedBy, descriptionId) || undefined}
            aria-invalid={Boolean(error)}
            className={cn('control-base appearance-none pr-11', className)}
            id={id}
            ref={ref}
            required={required}
          >
            {children}
          </select>
          <ChevronDown
            aria-hidden="true"
            className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-ink/50"
            size={17}
          />
        </div>
      </FieldFrame>
    )
  },
)
