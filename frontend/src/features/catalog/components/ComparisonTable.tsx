import { X } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import type { ProductDetail } from '../schemas'

interface ComparisonTableProps {
  products: ProductDetail[]
  onRemove: (productID: string) => void
}

export function ComparisonTable({ products, onRemove }: ComparisonTableProps) {
  const rows = [
    {
      label: 'Price',
      render: (p: ProductDetail) => (
        <PriceDisplay
          amountMinor={p.price.amount_minor}
          currency={p.price.currency}
          size="sm"
        />
      ),
    },
    {
      label: 'Dimensions',
      render: (p: ProductDetail) =>
        `${millimeters(p.dimensions.length_mm)} × ${millimeters(p.dimensions.width_mm)} × ${millimeters(p.dimensions.height_mm)}`,
    },
    {
      label: 'Weight',
      render: (p: ProductDetail) => kilograms(p.weight_grams),
    },
    {
      label: 'Capacity',
      render: (p: ProductDetail) =>
        p.max_capacity_grams
          ? kilograms(p.max_capacity_grams)
          : 'Not applicable',
    },
    { label: 'Materials', render: (p: ProductDetail) => p.material },
    {
      label: 'Warranty',
      render: (p: ProductDetail) => warranty(p.warranty_months),
    },
    {
      label: 'Portability',
      render: (p: ProductDetail) => score(p.scores.portability),
    },
    {
      label: 'Apartment suitability',
      render: (p: ProductDetail) => score(p.scores.apartment),
    },
    {
      label: 'Beginner suitability',
      render: (p: ProductDetail) => score(p.scores.beginner),
    },
    {
      label: 'Advanced suitability',
      render: (p: ProductDetail) => score(p.scores.advanced),
    },
    { label: 'Value', render: (p: ProductDetail) => score(p.scores.value) },
    { label: 'Quality', render: (p: ProductDetail) => score(p.scores.quality) },
  ]

  return (
    <div
      className="overflow-x-auto border border-ink/15"
      role="region"
      aria-label="Product comparison"
      tabIndex={0}
    >
      <table className="w-full min-w-[44rem] border-collapse text-left text-sm">
        <caption className="sr-only">
          Structured product facts for the selected tools
        </caption>
        <thead>
          <tr>
            <th className="sticky left-0 z-20 w-36 border-b border-r border-ink/15 bg-paper p-4 text-xs uppercase tracking-[0.12em] text-ink/50">
              Specification
            </th>
            {products.map((product) => (
              <th
                className="min-w-44 border-b border-r border-ink/15 bg-surface p-4 align-top last:border-r-0"
                key={product.id}
                scope="col"
              >
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-[0.625rem] uppercase tracking-[0.12em] text-ink/45">
                      {product.brand.name}
                    </p>
                    <Link
                      className="mt-2 block font-display text-xl font-medium leading-tight hover:text-bronze-dark"
                      to={`/products/${product.slug}`}
                    >
                      {product.name}
                    </Link>
                  </div>
                  <Button
                    aria-label={`Remove ${product.name} from comparison`}
                    onClick={() => onRemove(product.id)}
                    size="sm"
                    variant="quiet"
                  >
                    <X aria-hidden="true" size={15} />
                  </Button>
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.label}>
              <th
                className="sticky left-0 z-10 border-b border-r border-ink/15 bg-paper p-4 font-semibold"
                scope="row"
              >
                {row.label}
              </th>
              {products.map((product) => (
                <td
                  className="border-b border-r border-ink/10 bg-surface p-4 align-top text-ink/70 last:border-r-0"
                  key={product.id}
                >
                  {row.render(product)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function millimeters(value: number) {
  return `${(value / 10).toLocaleString()} cm`
}
function kilograms(value: number) {
  return `${(value / 1000).toLocaleString(undefined, { maximumFractionDigits: 1 })} kg`
}
function score(value: number) {
  return `${value}/100`
}
function warranty(months: number) {
  if (months === 0) return 'No stated warranty'
  if (months % 12 === 0)
    return `${months / 12} ${months === 12 ? 'year' : 'years'}`
  return `${months} months`
}
