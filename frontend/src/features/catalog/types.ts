export interface CatalogQuery {
  ids?: string[]
  q?: string
  category?: string
  brand?: string
  minPriceMinor?: number
  maxPriceMinor?: number
  /** Only products with a live vendor offer; sent as `has_offer=true`. */
  hasOffer?: boolean
  sort?: CatalogSort
  page?: number
  pageSize?: number
}

export type CatalogSort =
  | 'featured'
  | 'name_asc'
  | 'price_asc'
  | 'price_desc'
  | 'quality_desc'
  | 'value_desc'
