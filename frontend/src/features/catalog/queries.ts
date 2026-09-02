import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  deleteWishlistProduct,
  getComparison,
  getBrand,
  getBrands,
  getCategories,
  getCategory,
  getLiveOffers,
  getOffers,
  getProduct,
  getProducts,
  getWishlist,
  saveWishlistProduct,
  replaceComparison,
} from './api'
import { productSelectionSchema } from './schemas'
import type { CatalogQuery } from './types'

export const catalogKeys = {
  all: ['catalog'] as const,
  products: (query: CatalogQuery) => ['catalog', 'products', query] as const,
  product: (slug: string) => ['catalog', 'product', slug] as const,
  categories: ['catalog', 'categories'] as const,
  category: (slug: string) => ['catalog', 'category', slug] as const,
  brands: ['catalog', 'brands'] as const,
  brand: (slug: string) => ['catalog', 'brand', slug] as const,
  offers: (slug: string) => ['catalog', 'offers', slug] as const,
  liveOffers: ['catalog', 'live-offers'] as const,
  wishlist: ['account', 'wishlist'] as const,
  comparison: ['account', 'comparison'] as const,
}

export function useProducts(query: CatalogQuery) {
  return useQuery({
    queryKey: catalogKeys.products(query),
    queryFn: () => getProducts(query),
    placeholderData: (previous) => previous,
  })
}

export function useProduct(slug: string) {
  return useQuery({
    queryKey: catalogKeys.product(slug),
    queryFn: () => getProduct(slug),
    enabled: Boolean(slug),
  })
}

export function useProductsByIDs(productIDs: string[]) {
  const query = { ids: productIDs, pageSize: Math.max(1, productIDs.length) }
  return useQuery({
    queryKey: catalogKeys.products(query),
    queryFn: () => getProducts(query),
    enabled: productIDs.length > 0,
  })
}

export function useCategories() {
  return useQuery({ queryKey: catalogKeys.categories, queryFn: getCategories })
}

export function useCategory(slug: string) {
  return useQuery({
    queryKey: catalogKeys.category(slug),
    queryFn: () => getCategory(slug),
    enabled: Boolean(slug),
  })
}

export function useBrands(categorySlug?: string) {
  return useQuery({
    queryKey: categorySlug
      ? [...catalogKeys.brands, categorySlug]
      : catalogKeys.brands,
    queryFn: () => getBrands(categorySlug),
  })
}

export function useBrand(slug: string) {
  return useQuery({
    queryKey: catalogKeys.brand(slug),
    queryFn: () => getBrand(slug),
    enabled: Boolean(slug),
  })
}

export function useOffers(slug: string) {
  return useQuery({
    queryKey: catalogKeys.offers(slug),
    queryFn: () => getOffers(slug),
    enabled: Boolean(slug),
  })
}

export function useLiveOffers() {
  return useQuery({ queryKey: catalogKeys.liveOffers, queryFn: getLiveOffers })
}

export function useWishlist(enabled: boolean) {
  return useQuery({
    queryKey: catalogKeys.wishlist,
    queryFn: getWishlist,
    enabled,
  })
}

export function useToggleWishlist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      productID,
      saved,
    }: {
      productID: string
      saved: boolean
    }) =>
      saved ? deleteWishlistProduct(productID) : saveWishlistProduct(productID),
    onMutate: async ({ productID, saved }) => {
      await queryClient.cancelQueries({ queryKey: catalogKeys.wishlist })
      const previous = queryClient.getQueryData(catalogKeys.wishlist)
      queryClient.setQueryData(catalogKeys.wishlist, (current: unknown) => {
        const parsed = productSelectionSchema.safeParse(current)
        const ids = parsed.success ? parsed.data.product_ids : []
        return {
          product_ids: saved
            ? ids.filter((id) => id !== productID)
            : [...ids, productID],
        }
      })
      return { previous }
    },
    onError: (_error, _variables, context) => {
      queryClient.setQueryData(catalogKeys.wishlist, context?.previous)
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: catalogKeys.wishlist }),
  })
}

export function useComparison(enabled: boolean) {
  return useQuery({
    queryKey: catalogKeys.comparison,
    queryFn: getComparison,
    enabled,
  })
}

export function useReplaceComparison() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: replaceComparison,
    onMutate: async (productIDs) => {
      await queryClient.cancelQueries({ queryKey: catalogKeys.comparison })
      const previous = queryClient.getQueryData(catalogKeys.comparison)
      queryClient.setQueryData(catalogKeys.comparison, {
        product_ids: productIDs,
      })
      return { previous }
    },
    onError: (_error, _variables, context) => {
      queryClient.setQueryData(catalogKeys.comparison, context?.previous)
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: catalogKeys.comparison }),
  })
}
