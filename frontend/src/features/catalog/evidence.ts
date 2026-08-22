import type { ProductDetail } from './schemas'

// Evidence cards were labelled with the raw fact key, so a reader weighing this
// product's sourcing was shown "Score.Durability", "Warranty" and "Slug". Two of
// those are inherited from the equipment catalog and one is an internal
// identifier. The section is the strongest thing on the page — it says every
// number here is tied to a reviewed source — and it was undermined by its own
// labels.

const factLabels: Record<string, string> = {
  'score.quality': 'Product quality',
  'score.value': 'Value for money',
  'score.durability': 'Vendor stability',
  'score.beginner': 'Ease of adoption',
  'score.advanced': 'Depth for power users',
  'score.portability': 'Data portability',
  'score.apartment': 'Apartment suitability',
  'score.noise': 'Quiet operation',
  brand: 'Vendor',
  category: 'Category',
  description: 'What it does',
  name: 'Product name',
  price: 'Price',
  warranty: 'Warranty',
  material: 'Material',
  weight: 'Weight',
  dimensions: 'Dimensions',
}

export function evidenceFactLabel(key: string): string {
  return factLabels[key] ?? key.replaceAll('_', ' ').replaceAll('.', ' · ')
}

// A fact key that describes nothing this product has. Evidence for a score of
// zero on a dimension the product is not measured on reads as a gap in the
// product rather than as a dimension that does not apply to it.
export function evidenceApplies(key: string, product: ProductDetail): boolean {
  switch (key) {
    // The URL this page is served at is not a claim anyone needs sourced.
    case 'slug':
      return false
    case 'score.apartment':
      return product.scores.apartment > 0
    case 'score.noise':
      return product.scores.noise > 0
    case 'score.portability':
      return product.scores.portability > 0
    case 'warranty':
      return product.warranty_months > 0
    case 'material':
      return product.material.trim() !== ''
    case 'weight':
      return product.weight_grams > 0
    case 'dimensions':
      return (
        product.dimensions.length_mm > 0 ||
        product.dimensions.width_mm > 0 ||
        product.dimensions.height_mm > 0
      )
    default:
      return true
  }
}

export function visibleEvidence(product: ProductDetail) {
  return product.evidence.filter((entry) =>
    evidenceApplies(entry.fact_key, product),
  )
}
