import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  adminApi,
  type AffiliateInput,
  type OfferInput,
  type ProductInput,
} from './api'

export const adminKeys = {
  root: ['admin'] as const,
  dashboard: ['admin', 'dashboard'] as const,
  analytics: ['admin', 'analytics'] as const,
  references: ['admin', 'references'] as const,
  products: (search: string, page: number) =>
    ['admin', 'products', search, page] as const,
  product: (id: string) => ['admin', 'product', id] as const,
  governedProducts: (page: number) =>
    ['admin', 'evidence', 'products', page] as const,
  productGovernance: (id: string) =>
    ['admin', 'evidence', 'product', id] as const,
  categories: ['admin', 'categories'] as const,
  brands: ['admin', 'brands'] as const,
  merchants: ['admin', 'merchants'] as const,
  offers: (page: number) => ['admin', 'offers', page] as const,
  affiliate: (page: number) => ['admin', 'affiliate-links', page] as const,
  recommendations: (page: number) =>
    ['admin', 'recommendations', page] as const,
  recommendation: (id: string) => ['admin', 'recommendation', id] as const,
  users: (page: number) => ['admin', 'users', page] as const,
  events: (name: string, page: number) =>
    ['admin', 'events', name, page] as const,
}

export const useAdminDashboard = () =>
  useQuery({ queryKey: adminKeys.dashboard, queryFn: adminApi.dashboard })
export const useAdminAnalytics = () =>
  useQuery({ queryKey: adminKeys.analytics, queryFn: adminApi.analytics })
export const useAdminReferences = () =>
  useQuery({ queryKey: adminKeys.references, queryFn: adminApi.references })
export const useAdminProducts = (search = '', page = 1) =>
  useQuery({
    queryKey: adminKeys.products(search, page),
    queryFn: () => adminApi.products(search, page),
  })
export const useAdminProduct = (id: string | undefined) =>
  useQuery({
    queryKey: adminKeys.product(id ?? ''),
    queryFn: () => adminApi.product(id ?? ''),
    enabled: Boolean(id),
  })
export const useGovernedProducts = (page = 1) =>
  useQuery({
    queryKey: adminKeys.governedProducts(page),
    queryFn: () => adminApi.governedProducts(page),
  })
export const useProductGovernance = (id: string | undefined) =>
  useQuery({
    queryKey: adminKeys.productGovernance(id ?? ''),
    queryFn: () => adminApi.productGovernance(id ?? ''),
    enabled: Boolean(id),
  })
export const useAdminCategories = () =>
  useQuery({ queryKey: adminKeys.categories, queryFn: adminApi.categories })
export const useAdminBrands = () =>
  useQuery({ queryKey: adminKeys.brands, queryFn: adminApi.brands })
export const useAdminMerchants = () =>
  useQuery({ queryKey: adminKeys.merchants, queryFn: adminApi.merchants })
export const useAdminOffers = (page = 1) =>
  useQuery({
    queryKey: adminKeys.offers(page),
    queryFn: () => adminApi.offers(page),
  })
export const useAdminAffiliateLinks = (page = 1) =>
  useQuery({
    queryKey: adminKeys.affiliate(page),
    queryFn: () => adminApi.affiliateLinks(page),
  })
export const useAdminRecommendations = (page = 1) =>
  useQuery({
    queryKey: adminKeys.recommendations(page),
    queryFn: () => adminApi.recommendations(page),
  })
export const useAdminRecommendation = (id: string | undefined) =>
  useQuery({
    queryKey: adminKeys.recommendation(id ?? ''),
    queryFn: () => adminApi.recommendation(id ?? ''),
    enabled: Boolean(id),
  })
export const useAdminUsers = (page = 1) =>
  useQuery({
    queryKey: adminKeys.users(page),
    queryFn: () => adminApi.users(page),
  })
export const useAdminEvents = (name = '', page = 1) =>
  useQuery({
    queryKey: adminKeys.events(name, page),
    queryFn: () => adminApi.events(name, page),
  })

export function useProductMutation(id?: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (input: ProductInput) =>
      id ? adminApi.updateProduct(id, input) : adminApi.createProduct(input),
    onSuccess: (product) => {
      client.setQueryData(adminKeys.product(product.id), product)
      void client.invalidateQueries({ queryKey: adminKeys.root })
    },
  })
}

export function useProductStatusMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      adminApi.setProductStatus(id, status),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.root }),
  })
}

export function useOfferMutation(id?: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (input: OfferInput) =>
      id ? adminApi.updateOffer(id, input) : adminApi.createOffer(input),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.root }),
  })
}

export function useAffiliateMutation(id: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (input: AffiliateInput) => adminApi.updateAffiliate(id, input),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.root }),
  })
}
