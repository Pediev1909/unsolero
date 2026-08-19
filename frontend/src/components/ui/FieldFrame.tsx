import type { ReactNode } from 'react'

interface FieldFrameProps {
  id: string
  label: ReactNode
  hint?: ReactNode
  error?: ReactNode
  required?: boolean
  className?: string
  children: ReactNode
}

export function FieldFrame({
  id,
  label,
  hint,
  error,
  required,
  className,
  children,
}: FieldFrameProps) {
  return (
    <div className={className}>
      <label
        className="mb-2 flex items-baseline justify-between gap-3 text-label font-bold uppercase tracking-[0.12em]"
        htmlFor={id}
      >
        <span>
          {label}
          {required && <span aria-hidden="true"> *</span>}
        </span>
      </label>
      {children}
      {error ? (
        <p className="mt-2 text-xs leading-5 text-ember" id={`${id}-error`}>
          {error}
        </p>
      ) : hint ? (
        <p className="mt-2 text-xs leading-5 text-ink/70" id={`${id}-hint`}>
          {hint}
        </p>
      ) : null}
    </div>
  )
}
