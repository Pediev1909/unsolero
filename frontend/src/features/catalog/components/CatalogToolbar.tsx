import { SlidersHorizontal } from 'lucide-react'

import { Button } from '../../../components/ui/Button'
import { Select } from '../../../components/ui/Select'
import type { CatalogSort } from '../types'

interface CatalogToolbarProps {
  count: number
  sort: CatalogSort
  onOpenFilters: () => void
  onSort: (sort: CatalogSort) => void
}

export function CatalogToolbar({
  count,
  sort,
  onOpenFilters,
  onSort,
}: CatalogToolbarProps) {
  return (
    <div className="flex flex-col gap-4 border-y border-ink/15 py-4 sm:flex-row sm:items-center sm:justify-between">
      <p aria-live="polite" className="text-sm text-ink/70">
        {count} {count === 1 ? 'product' : 'products'}
      </p>
      <div className="flex items-end gap-3">
        <Button
          className="lg:hidden"
          onClick={onOpenFilters}
          size="sm"
          variant="secondary"
        >
          <SlidersHorizontal aria-hidden="true" size={16} />
          Filters
        </Button>
        <Select
          containerClassName="min-w-44 flex-1 sm:flex-none"
          label="Sort by"
          onChange={(event) => onSort(event.target.value as CatalogSort)}
          value={sort}
        >
          <option value="featured">Featured</option>
          <option value="name_asc">Name</option>
          <option value="price_asc">Price: low to high</option>
          <option value="price_desc">Price: high to low</option>
          <option value="quality_desc">Quality score</option>
          <option value="value_desc">Value score</option>
        </Select>
      </div>
    </div>
  )
}
