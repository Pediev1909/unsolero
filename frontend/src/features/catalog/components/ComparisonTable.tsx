import type { ReactNode } from 'react'
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
  // Rows that describe a physical object are meaningless for software, and
  // printing "0 cm × 0 cm × 0 cm" and an empty Materials cell made a working
  // comparison look broken. Each such row declares when it has something to
  // say; a row no selected product can answer is left out entirely rather than
  // filled with zeroes.
  const rows: {
    label: string
    render: (product: ProductDetail) => ReactNode
    applies?: (products: ProductDetail[]) => boolean
  }[] = [
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
      applies: (all: ProductDetail[]) =>
        all.some(
          (p) =>
            p.dimensions.length_mm > 0 ||
            p.dimensions.width_mm > 0 ||
            p.dimensions.height_mm > 0,
        ),
    },
    {
      label: 'Weight',
      render: (p: ProductDetail) => kilograms(p.weight_grams),
      applies: (all: ProductDetail[]) => all.some((p) => p.weight_grams > 0),
    },
    {
      label: 'Capacity',
      render: (p: ProductDetail) =>
        p.max_capacity_grams
          ? kilograms(p.max_capacity_grams)
          : 'Not applicable',
      applies: (all: ProductDetail[]) =>
        all.some((p) => (p.max_capacity_grams ?? 0) > 0),
    },
    {
      label: 'Materials',
      render: (p: ProductDetail) => p.material,
      applies: (all: ProductDetail[]) => all.some((p) => p.material !== ''),
    },
    {
      label: 'Warranty',
      render: (p: ProductDetail) => warranty(p.warranty_months),
      applies: (all: ProductDetail[]) => all.some((p) => p.warranty_months > 0),
    },
    {
      label: 'Portability',
      render: (p: ProductDetail) => score(p.scores.portability),
      applies: (all: ProductDetail[]) =>
        all.some((p) => p.scores.portability > 0),
    },
    {
      label: 'Apartment suitability',
      render: (p: ProductDetail) => score(p.scores.apartment),
      applies: (all: ProductDetail[]) => all.some((p) => p.scores.apartment > 0),
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
  ].filter((row) => !row.applies || row.applies(products))

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
            <th className="sticky left-0 z-20 w-36 border-b border-r border-ink/15 bg-paper p-4 text-xs uppercase tracking-[0.12em] text-ink/68">
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
                    <p className="text-[0.625rem] uppercase tracking-[0.12em] text-ink/65">
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
