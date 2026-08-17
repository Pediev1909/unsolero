import { AlertCircle } from 'lucide-react'

import { Button } from './Button'
import { StatePanel } from './StatePanel'

interface ErrorStateProps {
  title?: string
  description?: string
  onRetry?: () => void
  compact?: boolean
}

export function ErrorState({
  title = 'Something went wrong',
  description = 'Please check your connection and try again.',
  onRetry,
  compact,
}: ErrorStateProps) {
  return (
    <StatePanel
      action={
        onRetry ? (
          <Button onClick={onRetry} variant="secondary">
            Try again
          </Button>
        ) : undefined
      }
      compact={compact}
      description={description}
      icon={<AlertCircle aria-hidden="true" size={20} />}
      live="assertive"
      title={title}
    />
  )
}
