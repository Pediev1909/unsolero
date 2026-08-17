import { LoaderCircle } from 'lucide-react'

import { StatePanel } from './StatePanel'

interface LoadingStateProps {
  title?: string
  description?: string
  compact?: boolean
}

export function LoadingState({
  title = 'Loading',
  description = 'This should only take a moment.',
  compact,
}: LoadingStateProps) {
  return (
    <StatePanel
      compact={compact}
      description={description}
      icon={
        <LoaderCircle
          aria-hidden="true"
          className="animate-spin motion-reduce:animate-none"
          size={20}
        />
      }
      live="polite"
      title={title}
    />
  )
}
