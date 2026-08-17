import type { ReactNode } from 'react'
import { Link, type LinkProps } from 'react-router-dom'

import {
  buttonStyles,
  type ButtonSize,
  type ButtonVariant,
} from './buttonStyles'

interface ButtonLinkProps extends LinkProps {
  variant?: ButtonVariant
  size?: ButtonSize
  fullWidth?: boolean
  children: ReactNode
}

export function ButtonLink({
  variant,
  size,
  fullWidth,
  className,
  children,
  ...props
}: ButtonLinkProps) {
  return (
    <Link
      {...props}
      className={buttonStyles({ variant, size, fullWidth, className })}
    >
      {children}
    </Link>
  )
}
