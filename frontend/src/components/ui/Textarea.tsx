import {
  forwardRef,
  useId,
  type ReactNode,
  type TextareaHTMLAttributes,
} from 'react'

import { cn } from '../../lib/styles/cn'
import { FieldFrame } from './FieldFrame'

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: ReactNode
  hint?: ReactNode
  error?: ReactNode
  containerClassName?: string
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  function Textarea(
    {
      label,
      hint,
      error,
      containerClassName,
      className,
      id: suppliedId,
      required,
      'aria-describedby': describedBy,
      rows = 5,
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
        <textarea
          {...props}
          aria-describedby={cn(describedBy, descriptionId) || undefined}
          aria-invalid={Boolean(error)}
          className={cn('control-base min-h-32 resize-y', className)}
          id={id}
          ref={ref}
          required={required}
          rows={rows}
        />
      </FieldFrame>
    )
  },
)
