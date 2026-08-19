import {
  forwardRef,
  useId,
  type InputHTMLAttributes,
  type ReactNode,
} from 'react'

import { cn } from '../../lib/styles/cn'
import { FieldFrame } from './FieldFrame'

export interface InputProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'size'
> {
  label: ReactNode
  hint?: ReactNode
  error?: ReactNode
  leadingIcon?: ReactNode
  containerClassName?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  {
    label,
    hint,
    error,
    leadingIcon,
    containerClassName,
    className,
    id: suppliedId,
    required,
    'aria-describedby': describedBy,
    ...props
  },
  ref,
) {
  const generatedId = useId()
  const id = suppliedId ?? generatedId
  const descriptionId = error ? `${id}-error` : hint ? `${id}-hint` : undefined

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
        {leadingIcon && (
          <span
            aria-hidden="true"
            className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-ink/65"
          >
            {leadingIcon}
          </span>
        )}
        <input
          {...props}
          aria-describedby={cn(describedBy, descriptionId) || undefined}
          aria-invalid={Boolean(error)}
          className={cn(
            'control-base',
            Boolean(leadingIcon) && 'pl-11',
            className,
          )}
          id={id}
          ref={ref}
          required={required}
        />
      </div>
    </FieldFrame>
  )
})
