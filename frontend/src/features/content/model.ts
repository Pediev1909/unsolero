import type { ContentType } from './schemas'

export function contentTypeLabel(type: ContentType) {
  switch (type) {
    case 'article':
      return 'Article'
    case 'guide':
      return 'Guide'
    case 'buying_guide':
      return 'Buying guide'
    case 'comparison':
      return 'Product comparison'
  }
}

export function formatEditorialDate(value: string) {
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'long' }).format(
    new Date(value),
  )
}

export function headingID(value: string) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}
