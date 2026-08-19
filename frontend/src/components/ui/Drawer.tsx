import { X } from 'lucide-react'
import { useEffect, useId, useRef, type ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'
import { Button } from './Button'

interface DrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  side?: 'left' | 'right'
  children: ReactNode
  footer?: ReactNode
}

export function Drawer({
  open,
  onOpenChange,
  title,
  description,
  side = 'right',
  children,
  footer,
}: DrawerProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleId = useId()
  const descriptionId = useId()

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (open && !dialog.open) {
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open])

  return (
    <dialog
      aria-describedby={description ? descriptionId : undefined}
      aria-labelledby={titleId}
      className={cn(
        'fixed inset-y-0 m-0 h-dvh max-h-none w-[min(90vw,26rem)] max-w-none border-0 bg-surface p-0 text-ink shadow-overlay',
        side === 'right'
          ? 'ml-auto animate-[drawer-enter_var(--duration-slow)_var(--ease-expressive)]'
          : 'mr-auto animate-[drawer-left-enter_var(--duration-slow)_var(--ease-expressive)]',
      )}
      onCancel={(event) => {
        event.preventDefault()
        onOpenChange(false)
      }}
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onOpenChange(false)
        }
      }}
      onClose={() => {
        if (open) onOpenChange(false)
      }}
      ref={dialogRef}
    >
      <div className="flex min-h-full flex-col">
        <div className="flex items-start justify-between gap-6 border-b border-ink/10 px-5 py-5 sm:px-7">
          <div>
            <h2
              className="font-display text-2xl font-medium tracking-[-0.04em]"
              id={titleId}
            >
              {title}
            </h2>
            {description && (
              <p
                className="mt-2 text-sm leading-6 text-ink/70"
                id={descriptionId}
              >
                {description}
              </p>
            )}
          </div>
          <Button
            aria-label="Close drawer"
            className="-mr-2 -mt-2 shrink-0"
            onClick={() => onOpenChange(false)}
            size="sm"
            variant="quiet"
          >
            <X aria-hidden="true" size={18} />
          </Button>
        </div>
        <div className="flex-1 px-5 py-6 sm:px-7">{children}</div>
        {footer && (
          <div className="border-t border-ink/10 px-5 py-5 sm:px-7">
            {footer}
          </div>
        )}
      </div>
    </dialog>
  )
}
