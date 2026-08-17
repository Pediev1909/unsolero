import { useMemo } from 'react'
import { z } from 'zod'

import {
  useLocalDocument,
  writeLocalDocument,
} from '../../lib/storage/localStorageStore'

export const comparisonLimit = 4
const selectionSchema = z.array(z.string().min(1))

export function updateProductSelection(
  current: readonly string[],
  productID: string,
  maximum = Number.POSITIVE_INFINITY,
): string[] {
  if (current.includes(productID)) {
    return current.filter((id) => id !== productID)
  }
  if (current.length >= maximum) return [...current]
  return [...current, productID]
}

export function normalizeProductSelection(
  productIDs: readonly string[],
  maximum = Number.POSITIVE_INFINITY,
) {
  return [...new Set(productIDs)].slice(0, maximum)
}

export function useLocalProductSelection(key: string, maximum?: number) {
  const raw = useLocalDocument(key)
  const productIDs = useMemo(() => {
    if (!raw) return []
    try {
      const parsed = selectionSchema.safeParse(JSON.parse(raw))
      return parsed.success
        ? normalizeProductSelection(parsed.data, maximum)
        : []
    } catch {
      return []
    }
  }, [maximum, raw])

  return {
    productIDs,
    replace(next: readonly string[]) {
      try {
        writeLocalDocument(key, normalizeProductSelection(next, maximum))
        return true
      } catch {
        return false
      }
    },
  }
}
