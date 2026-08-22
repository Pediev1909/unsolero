import { useState } from 'react'
import { useToast } from '../../components/ui/useToast'
import { useProductsByIDs } from './queries'
import type { ProductSummary } from './schemas'
import {
  useComparisonSelection,
  useWishlistSelection,
} from './useProductCollections'

export function useCatalogActions() {
  const [comparisonOpen, setComparisonOpen] = useState(false)
  const { showToast } = useToast()
  const comparison = useComparisonSelection()
  const wishlist = useWishlistSelection()
  const comparedProducts = useProductsByIDs(comparison.productIDs)
  const compared = comparedProducts.data?.products ?? []
  const savedIDs = new Set(wishlist.productIDs)

  function compare(product: ProductSummary) {
    const isFull =
      !comparison.productIDs.includes(product.id) &&
      comparison.productIDs.length >= 4
    if (!comparison.toggle(product.id)) {
      showToast({
        title: isFull ? 'Comparison is full' : 'Comparison is unavailable',
        description: isFull
          ? 'You can compare up to four products at a time.'
          : 'Wait a moment, then try again.',
        variant: isFull ? 'neutral' : 'error',
      })
    }
  }

  function save(product: ProductSummary) {
    const wasSaved = wishlist.productIDs.includes(product.id)
    if (!wishlist.toggle(product.id)) {
      showToast({
        title: 'Saved software is unavailable',
        description: 'Wait a moment, then try again.',
        variant: wishlist.isError ? 'error' : 'neutral',
      })
      return
    }
    // Saving used to give no feedback at all beyond a bookmark icon changing
    // state, and no indication of where the thing had gone or whether it
    // would survive closing the tab. Both matter, and the honest answer
    // differs depending on whether they are signed in.
    if (!wasSaved) {
      showToast({
        title: `${product.name} saved`,
        description: wishlist.authenticated
          ? 'It is on your saved list, on any device you sign in from.'
          : 'It is on your saved list in this browser. Sign in to keep it.',
        variant: 'success',
      })
    }
  }

  return {
    compared,
    comparedIDs: new Set(comparison.productIDs),
    comparisonOpen,
    savedIDs,
    comparisonPending:
      comparison.isPending ||
      (comparison.productIDs.length > 0 && comparedProducts.isPending),
    comparisonError:
      comparison.isError ||
      (comparison.productIDs.length > 0 && comparedProducts.isError),
    savePending: wishlist.isPending,
    compare,
    save,
    setComparisonOpen,
  }
}
