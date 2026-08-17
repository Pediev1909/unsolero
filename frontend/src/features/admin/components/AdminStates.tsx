import type { ReactNode } from 'react'

import { Button } from '../../../components/ui/Button'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'

export function AdminQueryState({
  pending,
  error,
  empty,
  onRetry,
  children,
  emptyTitle = 'No data',
  emptyDescription = 'No records exist for this section yet.',
}: {
  pending: boolean
  error: boolean
  empty: boolean
  onRetry: () => void
  children: ReactNode
  emptyTitle?: string
  emptyDescription?: string
}) {
  if (pending)
    return (
      <LoadingState
        description="Loading current administrative data."
        title="Loading"
      />
    )
  if (error)
    return (
      <ErrorState
        description="The administrative data could not be loaded."
        onRetry={onRetry}
        title="Something went wrong"
      />
    )
  if (empty)
    return <EmptyState description={emptyDescription} title={emptyTitle} />
  return children
}

export function AdminTable({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto border border-ink/10 bg-surface">
      <table className="min-w-full border-collapse text-left text-sm">
        {children}
      </table>
    </div>
  )
}

export const adminTableHead =
  'border-b border-ink/10 bg-paper px-4 py-3 text-xs font-semibold tracking-[0.12em] text-ink/45 uppercase'
export const adminTableCell = 'border-b border-ink/8 px-4 py-4 align-top'

export function AdminPagination({
  page,
  totalPages,
  total,
  onPageChange,
}: {
  page: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
}) {
  return (
    <nav
      aria-label="Admin table pagination"
      className="mt-4 flex flex-wrap items-center justify-between gap-3"
    >
      <p className="text-xs text-ink/45">
        {total.toLocaleString()} {total === 1 ? 'record' : 'records'} · Page{' '}
        {page} of {Math.max(totalPages, 1)}
      </p>
      <div className="flex gap-2">
        <Button
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          size="sm"
          variant="secondary"
        >
          Previous
        </Button>
        <Button
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
          size="sm"
          variant="secondary"
        >
          Next
        </Button>
      </div>
    </nav>
  )
}

export function DataValue({ value }: { value: number }) {
  return value === 0 ? (
    <span className="text-sm font-medium text-ink/35">No data</span>
  ) : (
    <span>{value.toLocaleString()}</span>
  )
}
