import { LoaderCircle } from 'lucide-react'
import type { ButtonHTMLAttributes, ReactNode } from 'react'

import {
  buttonStyles,
  type ButtonSize,
  type ButtonVariant,
} from './buttonStyles'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  fullWidth?: boolean
  loading?: boolean
  loadingLabel?: string
  children: ReactNode
}

export function Button({
  variant,
  size,
  fullWidth,
  loading = false,
  loadingLabel = 'Working…',
  className,
  children,
  disabled,
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      aria-busy={loading || undefined}
      className={buttonStyles({ variant, size, fullWidth, className })}
      disabled={disabled || loading}
      type={type}
    >
      {loading && (
        <LoaderCircle
          aria-hidden="true"
          className="animate-spin motion-reduce:animate-none"
          size={16}
        />
      )}
      {loading ? loadingLabel : children}
    </button>
  )
}
