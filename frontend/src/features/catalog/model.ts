import type { ProductDetail, ProductInsight, ProductSummary } from './schemas'

export function prominentSuitability(
  product: ProductSummary,
): ProductInsight[] {
  return [...product.suitability]
    .filter((item) => item.score >= 85)
    .sort((left, right) => right.score - left.score)
    .slice(0, 2)
}

export function suitabilityVariant(score: number) {
  if (score >= 90) return 'success' as const
  if (score >= 80) return 'accent' as const
  if (score >= 65) return 'neutral' as const
  return 'warning' as const
}

export function attributeLabel(key: string): string {
  const sentence = key.replaceAll('_', ' ')
  return sentence.charAt(0).toUpperCase() + sentence.slice(1)
}

export function attributeValue(
  attribute: ProductDetail['attributes'][number],
): string {
  if (attribute.numeric_value !== undefined) {
    return `${attribute.numeric_value}${attribute.unit ? ` ${attribute.unit}` : ''}`
  }
  if (attribute.text_value !== undefined)
    return attribute.text_value.replaceAll('_', ' ')
  return attribute.boolean_value ? 'Yes' : 'No'
}

export function formatMeasurement(grams: number): string {
  return `${new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 }).format(grams / 1000)} kg`
}
