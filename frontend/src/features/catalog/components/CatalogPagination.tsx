import { ArrowLeft, ArrowRight } from 'lucide-react'

import { Button } from '../../../components/ui/Button'

interface CatalogPaginationProps {
  page: number
  totalPages: number
  onPage: (page: number) => void
}

export function CatalogPagination({
  page,
  totalPages,
  onPage,
}: CatalogPaginationProps) {
  if (totalPages <= 1) return null
  return (
    <nav
      aria-label="Catalog pagination"
      className="mt-10 flex items-center justify-between border-t border-ink/15 pt-6"
    >
      <Button
        disabled={page <= 1}
        onClick={() => onPage(page - 1)}
        size="sm"
        variant="secondary"
      >
        <ArrowLeft aria-hidden="true" size={15} />
        Previous
      </Button>
      <p className="text-sm text-ink/60">
        Page <span className="font-semibold text-ink">{page}</span> of{' '}
        {totalPages}
      </p>
      <Button
        disabled={page >= totalPages}
        onClick={() => onPage(page + 1)}
        size="sm"
        variant="secondary"
      >
        Next
        <ArrowRight aria-hidden="true" size={15} />
      </Button>
    </nav>
  )
}
