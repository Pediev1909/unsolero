import { Check, Info, X, XCircle } from 'lucide-react'

import { cn } from '../../lib/styles/cn'
import { Button } from './Button'
import type { ToastRecord } from './toastTypes'

interface ToastProps {
  toast: ToastRecord
  onDismiss: (id: number) => void
}

const icons = {
  neutral: Info,
  success: Check,
  error: XCircle,
}

const accents = {
  neutral: 'border-ink/20',
  success: 'border-moss',
  error: 'border-ember',
}

export function Toast({ toast, onDismiss }: ToastProps) {
  const variant = toast.variant ?? 'neutral'
  const Icon = icons[variant]

  return (
    <div
      className={cn(
        'pointer-events-auto flex w-full gap-3 border-l-2 bg-surface p-4 text-ink shadow-overlay motion-safe:animate-[toast-enter_var(--duration-slow)_var(--ease-expressive)]',
        accents[variant],
      )}
      role={variant === 'error' ? 'alert' : 'status'}
    >
      <Icon aria-hidden="true" className="mt-0.5 shrink-0" size={18} />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold">{toast.title}</p>
        {toast.description && (
          <p className="mt-1 text-xs leading-5 text-ink/70">
            {toast.description}
          </p>
        )}
      </div>
      <Button
        aria-label="Dismiss notification"
        className="-mr-2 -mt-2 shrink-0"
        onClick={() => onDismiss(toast.id)}
        size="sm"
        variant="quiet"
      >
        <X aria-hidden="true" size={16} />
      </Button>
    </div>
  )
}
