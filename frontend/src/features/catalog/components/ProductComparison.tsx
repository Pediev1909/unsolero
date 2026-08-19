import { Scale } from 'lucide-react'

import { Modal } from '../../../components/ui/Modal'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import type { ProductSummary } from '../schemas'
import { ComparisonData } from './ComparisonData'

interface ProductComparisonProps {
  products: ProductSummary[]
  open: boolean
  onOpenChange: (open: boolean) => void
  onRemove: (product: ProductSummary) => void
  loading?: boolean
  error?: boolean
}

export function ProductComparison({
  products,
  open,
  onOpenChange,
  onRemove,
  loading = false,
  error = false,
}: ProductComparisonProps) {
  return (
    <Modal
      description="Structured product facts only. Review scores appear only when verified review data exists."
      onOpenChange={onOpenChange}
      open={open}
      size="lg"
      title="Compare software"
    >
      {loading && <LoadingState compact title="Loading comparison" />}
      {error && !loading && (
        <ErrorState
          compact
          description="Your selected products could not be loaded."
          title="Comparison unavailable"
        />
      )}
      {!loading && !error && products.length >= 2 && (
        <ComparisonData
          onRemove={(id) => {
            const product = products.find((item) => item.id === id)
            if (product) onRemove(product)
          }}
          products={products}
        />
      )}
      {!loading && !error && products.length < 2 && (
        <div className="py-10 text-center text-sm text-ink/70">
          <Scale aria-hidden="true" className="mx-auto mb-3" size={24} />
          Select at least two products. You can compare up to four.
        </div>
      )}
      <div className="mt-5 flex justify-end">
        <ButtonLink
          onClick={() => onOpenChange(false)}
          size="sm"
          to="/compare"
          variant="secondary"
        >
          Open full comparison
        </ButtonLink>
      </div>
    </Modal>
  )
}
