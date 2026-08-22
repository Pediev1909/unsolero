import { useCurrentUser } from '../auth/queries'
import { trackEvent } from '../analytics/tracking'
import {
  comparisonLimit,
  updateProductSelection,
  useLocalProductSelection,
} from './productSelections'
import {
  useComparison,
  useReplaceComparison,
  useToggleWishlist,
  useWishlist,
} from './queries'

const comparisonStorageKey = 'rigmark:comparison:v1'
const wishlistStorageKey = 'rigmark:wishlist:v1'

export function useWishlistSelection() {
  const account = useCurrentUser()
  const local = useLocalProductSelection(wishlistStorageKey)
  const server = useWishlist(Boolean(account.data))
  const mutation = useToggleWishlist()
  const authenticated = Boolean(account.data)
  const productIDs = !account.isSuccess
    ? []
    : authenticated
      ? (server.data?.product_ids ?? [])
      : local.productIDs

  function toggle(productID: string) {
    if (!account.isSuccess) return false
    const saved = productIDs.includes(productID)
    const persistence = authenticated ? 'account' : 'browser'
    if (authenticated) {
      mutation.mutate(
        { productID, saved },
        {
          onSuccess: () => {
            if (!saved)
              trackEvent('product_saved', 'catalog', {
                product_id: productID,
                persistence,
              })
          },
        },
      )
    } else {
      const updated = local.replace(
        updateProductSelection(productIDs, productID),
      )
      if (updated && !saved) {
        trackEvent('product_saved', 'catalog', {
          product_id: productID,
          persistence,
        })
      }
      return updated
    }
    return true
  }

  return {
    productIDs,
    // Exposed so a message about where something was saved can be accurate:
    // an account list follows you between devices, a browser list does not.
    authenticated,
    toggle,
    isPending:
      account.isPending ||
      (authenticated && (server.isPending || mutation.isPending)),
    isError:
      account.isError ||
      (authenticated && (server.isError || mutation.isError)),
    refetch: server.refetch,
  }
}

export function useComparisonSelection() {
  const account = useCurrentUser()
  const local = useLocalProductSelection(comparisonStorageKey, comparisonLimit)
  const server = useComparison(Boolean(account.data))
  const mutation = useReplaceComparison()
  const authenticated = Boolean(account.data)
  const productIDs = !account.isSuccess
    ? []
    : authenticated
      ? (server.data?.product_ids ?? [])
      : local.productIDs

  function replace(next: readonly string[]) {
    if (!account.isSuccess) return false
    const normalized = next.slice(0, comparisonLimit)
    const created = productIDs.length < 2 && normalized.length >= 2
    const persistence = authenticated ? 'account' : 'browser'
    if (authenticated) {
      mutation.mutate([...normalized], {
        onSuccess: () => {
          if (created)
            trackEvent('comparison_created', 'comparison', {
              product_count: normalized.length,
              persistence,
            })
        },
      })
    } else {
      const updated = local.replace(normalized)
      if (updated && created)
        trackEvent('comparison_created', 'comparison', {
          product_count: normalized.length,
          persistence,
        })
      return updated
    }
    return true
  }

  return {
    productIDs,
    replace,
    toggle(productID: string) {
      if (!account.isSuccess) return false
      const next = updateProductSelection(
        productIDs,
        productID,
        comparisonLimit,
      )
      if (
        next.length === productIDs.length &&
        !productIDs.includes(productID)
      ) {
        return false
      }
      return replace(next)
    },
    isPending:
      account.isPending ||
      (authenticated && (server.isPending || mutation.isPending)),
    isError:
      account.isError ||
      (authenticated && (server.isError || mutation.isError)),
    refetch: server.refetch,
  }
}
