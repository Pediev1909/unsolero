import { attributeLabel, attributeValue, formatMeasurement } from './model'
import type { ProductDetail } from './schemas'

// The physical rows are inherited from the equipment catalog this site used to
// be. Software has no length, weight, material or warranty, and the API reports
// zero for all of them, so every product page printed "0 × 0 × 0 mm", "0 kg",
// an empty material and "Not specified" under a heading that promised facts.
// The comparison table was fixed the same way: a row that nothing can answer is
// left out rather than filled with zeroes.
export function specifications(
  product: ProductDetail,
): Array<readonly [string, string]> {
  const facts: Array<readonly [string, string]> = []

  const { length_mm, width_mm, height_mm } = product.dimensions
  if (length_mm > 0 || width_mm > 0 || height_mm > 0) {
    facts.push(['Dimensions', `${length_mm} × ${width_mm} × ${height_mm} mm`])
  }
  if (product.weight_grams > 0) {
    facts.push(['Weight', formatMeasurement(product.weight_grams)])
  }
  if (product.max_capacity_grams) {
    facts.push([
      'Maximum capacity',
      formatMeasurement(product.max_capacity_grams),
    ])
  }
  if (product.material.trim() !== '') {
    facts.push(['Material', product.material])
  }
  if (product.warranty_months > 0) {
    facts.push(['Warranty', `${product.warranty_months} months`])
  }

  for (const attribute of product.attributes) {
    facts.push([attributeLabel(attribute.key), attributeValue(attribute)])
  }
  return facts
}
