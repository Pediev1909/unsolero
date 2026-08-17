import { PackageOpen } from 'lucide-react'
import type { ReactNode } from 'react'

import { StatePanel } from './StatePanel'

interface EmptyStateProps {
  title?: string
  description?: string
  action?: ReactNode
  compact?: boolean
}

export function EmptyState({
  title = 'Nothing here yet',
  description = 'Items will appear here when they are available.',
  action,
  compact,
}: EmptyStateProps) {
  return (
    <StatePanel
      action={action}
      compact={compact}
      description={description}
      icon={<PackageOpen aria-hidden="true" size={20} />}
      title={title}
    />
  )
}
