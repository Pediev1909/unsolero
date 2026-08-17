import type { ProductSummary } from '../schemas'
import { CatalogProductCard } from './CatalogProductCard'

interface CatalogProductGridProps {
  products: ProductSummary[]
  comparedIDs: Set<string>
  savedIDs: Set<string>
  savePending?: boolean
  onCompare: (product: ProductSummary) => void
  onSave: (product: ProductSummary) => void
}

export function CatalogProductGrid({
  products,
  comparedIDs,
  savedIDs,
  savePending,
  onCompare,
  onSave,
}: CatalogProductGridProps) {
  return (
    <div className="grid grid-cols-1 border-l border-t border-ink/15 sm:grid-cols-2 xl:grid-cols-3">
      {products.map((product) => (
        <CatalogProductCard
          compared={comparedIDs.has(product.id)}
          key={product.id}
          onCompare={onCompare}
          onSave={onSave}
          product={product}
          saved={savedIDs.has(product.id)}
          savePending={savePending}
        />
      ))}
    </div>
  )
}
