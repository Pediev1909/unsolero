import { X } from 'lucide-react'
import { useEffect, useId, useRef, type ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'
import { Button } from './Button'

interface ModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: ReactNode
  footer?: ReactNode
  size?: 'sm' | 'md' | 'lg'
}

const sizes = {
  sm: 'max-w-md',
  md: 'max-w-xl',
  lg: 'max-w-3xl',
}

export function Modal({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
  size = 'md',
}: ModalProps) {
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
        'm-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] overflow-auto rounded-sm border border-ink/15 bg-surface p-0 text-ink shadow-overlay',
        sizes[size],
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
              className="mt-2 max-w-lg text-sm leading-6 text-ink/70"
              id={descriptionId}
            >
              {description}
            </p>
          )}
        </div>
        <Button
          aria-label="Close dialog"
          className="-mr-2 -mt-2 shrink-0"
          onClick={() => onOpenChange(false)}
          size="sm"
          variant="quiet"
        >
          <X aria-hidden="true" size={18} />
        </Button>
      </div>
      <div className="px-5 py-6 sm:px-7">{children}</div>
      {footer && (
        <div className="flex flex-col-reverse gap-3 border-t border-ink/10 px-5 py-5 sm:flex-row sm:justify-end sm:px-7">
          {footer}
        </div>
      )}
    </dialog>
  )
}
