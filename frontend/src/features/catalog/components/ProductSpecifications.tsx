import { attributeLabel, attributeValue, formatMeasurement } from '../model'
import type { ProductDetail } from '../schemas'

export function ProductSpecifications({ product }: { product: ProductDetail }) {
  const facts: Array<readonly [string, string]> = [
    [
      'Dimensions',
      `${product.dimensions.length_mm} × ${product.dimensions.width_mm} × ${product.dimensions.height_mm} mm`,
    ],
    ['Equipment weight', formatMeasurement(product.weight_grams)],
    [
      'Maximum capacity',
      product.max_capacity_grams
        ? formatMeasurement(product.max_capacity_grams)
        : 'Not applicable',
    ],
    ['Material', product.material],
    [
      'Warranty',
      product.warranty_months
        ? `${product.warranty_months} months`
        : 'Not specified',
    ],
  ]

  return (
    <dl className="border-t border-ink/15">
      {facts.map(([label, value]) => (
        <SpecificationRow key={label} label={label} value={value} />
      ))}
      {product.attributes.map((attribute) => (
        <SpecificationRow
          key={attribute.key}
          label={attributeLabel(attribute.key)}
          value={attributeValue(attribute)}
        />
      ))}
    </dl>
  )
}

function SpecificationRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1 border-b border-ink/15 py-4 sm:grid-cols-[12rem_1fr] sm:gap-6">
      <dt className="text-xs font-bold uppercase tracking-[0.1em] text-ink/45">
        {label}
      </dt>
      <dd className="text-sm font-medium capitalize sm:text-right">{value}</dd>
    </div>
  )
}
