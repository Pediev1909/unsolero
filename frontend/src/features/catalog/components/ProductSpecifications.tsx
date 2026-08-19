import { specifications } from '../specifications'
import type { ProductDetail } from '../schemas'

export function ProductSpecifications({ product }: { product: ProductDetail }) {
  const facts = specifications(product)
  if (facts.length === 0) return null

  return (
    <dl className="border-t border-ink/15">
      {facts.map(([label, value]) => (
        <SpecificationRow key={label} label={label} value={value} />
      ))}
    </dl>
  )
}

function SpecificationRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1 border-b border-ink/15 py-4 sm:grid-cols-[12rem_1fr] sm:gap-6">
      <dt className="text-xs font-bold uppercase tracking-[0.1em] text-ink/65">
        {label}
      </dt>
      <dd className="text-sm font-medium capitalize sm:text-right">{value}</dd>
    </div>
  )
}
