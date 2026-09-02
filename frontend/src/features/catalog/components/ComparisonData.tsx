import { useQueries } from '@tanstack/react-query'

import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import { getProduct } from '../api'
import { catalogKeys } from '../queries'
import type { ProductSummary } from '../schemas'
import { ComparisonTable } from './ComparisonTable'

export function ComparisonData({
  products,
  onRemove,
  readOnly,
}: {
  products: ProductSummary[]
  onRemove?: (productID: string) => void
  readOnly?: boolean
}) {
  const details = useQueries({
    queries: products.map((product) => ({
      queryKey: catalogKeys.product(product.slug),
      queryFn: () => getProduct(product.slug),
    })),
  })
  if (details.some((query) => query.isPending))
    return <LoadingState compact title="Loading comparison facts" />
  if (details.some((query) => query.isError))
    return (
      <ErrorState
        compact
        description="One or more product records could not be loaded."
        title="Comparison unavailable"
      />
    )
  return (
    <ComparisonTable
      onRemove={onRemove}
      products={details.flatMap((query) => (query.data ? [query.data] : []))}
      readOnly={readOnly}
    />
  )
}
