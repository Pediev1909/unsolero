import { useQuery } from '@tanstack/react-query'
import { z } from 'zod'

import { apiRequest } from '../../lib/api/client'
import { contentSummarySchema } from '../content/schemas'

/**
 * The editorial pieces a product appears in: the comparisons and guides that
 * weigh it against others.
 *
 * Lives with the catalog rather than with content because the product page is
 * what asks. The key sits under 'content' so an invalidation of everything
 * editorial takes this with it.
 */
export const relatedContentKeys = {
  byProduct: (slug: string) => ['content', 'by-product', slug] as const,
}

export function getProductEditorial(slug: string) {
  return apiRequest(
    `/content?product=${encodeURIComponent(slug)}`,
    { method: 'GET' },
    (value) => z.array(contentSummarySchema).parse(value),
  )
}

export function useProductEditorial(slug: string) {
  return useQuery({
    queryKey: relatedContentKeys.byProduct(slug),
    queryFn: () => getProductEditorial(slug),
    enabled: Boolean(slug),
  })
}
