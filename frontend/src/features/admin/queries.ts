import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  adminApi,
  type AffiliateInput,
  type EvidenceObservationInput,
  type EvidenceRevisionInput,
  type EvidenceSourceInput,
  type OfferInput,
  type ProductInput,
  type CommerceProviderInput,
} from './api'

export const adminKeys = {
  root: ['admin'] as const,
  dashboard: ['admin', 'dashboard'] as const,
  analytics: ['admin', 'analytics'] as const,
  references: ['admin', 'references'] as const,
  recommendationPolicies: ['admin', 'recommendation-policies'] as const,
  evidenceSources: ['admin', 'evidence', 'sources'] as const,
  evidenceObservations: (productID: string) =>
    ['admin', 'evidence', 'observations', productID] as const,
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
  commerceProviders: ['admin', 'commerce', 'providers'] as const,
  commerceImports: (page: number) =>
    ['admin', 'commerce', 'imports', page] as const,
  commerceImportFailures: (id: string, page: number) =>
    ['admin', 'commerce', 'imports', id, 'failures', page] as const,
  verifiedConversions: (page: number) =>
    ['admin', 'commerce', 'conversions', page] as const,
  conversionImports: (page: number) =>
    ['admin', 'commerce', 'conversion-imports', page] as const,
  conversionReconciliations: (page: number) =>
    ['admin', 'commerce', 'reconciliations', page] as const,
  monetizationMetrics: ['admin', 'commerce', 'metrics'] as const,
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

export const useCommerceProviders = () =>
  useQuery({
    queryKey: adminKeys.commerceProviders,
    queryFn: adminApi.commerceProviders,
  })

export const useCommerceImports = (page = 1) =>
  useQuery({
    queryKey: adminKeys.commerceImports(page),
    queryFn: () => adminApi.commerceImports(page),
  })

export const useCommerceImportFailures = (id: string | undefined, page = 1) =>
  useQuery({
    queryKey: adminKeys.commerceImportFailures(id ?? '', page),
    queryFn: () => adminApi.commerceImportFailures(id ?? '', page),
    enabled: Boolean(id),
  })

export const useVerifiedConversions = (page = 1) =>
  useQuery({
    queryKey: adminKeys.verifiedConversions(page),
    queryFn: () => adminApi.verifiedConversions(page),
  })

export const useConversionImports = (page = 1) =>
  useQuery({
    queryKey: adminKeys.conversionImports(page),
    queryFn: () => adminApi.conversionImports(page),
  })

export const useConversionReconciliations = (page = 1) =>
  useQuery({
    queryKey: adminKeys.conversionReconciliations(page),
    queryFn: () => adminApi.conversionReconciliations(page),
  })

export const useMonetizationMetrics = () =>
  useQuery({
    queryKey: adminKeys.monetizationMetrics,
    queryFn: adminApi.monetizationMetrics,
  })

export const useRecommendationPolicies = () =>
  useQuery({
    queryKey: adminKeys.recommendationPolicies,
    queryFn: adminApi.recommendationPolicies,
  })

export function usePolicyTransition() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({
      version,
      action,
      note,
    }: {
      version: string
      action: 'submit' | 'approve' | 'reject' | 'activate' | 'deactivate'
      note: string
    }) => adminApi.transitionRecommendationPolicy(version, action, note),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.root }),
  })
}

export const useEvidenceSources = () =>
  useQuery({
    queryKey: adminKeys.evidenceSources,
    queryFn: adminApi.evidenceSources,
  })

export const useEvidenceObservations = (productID: string | undefined) =>
  useQuery({
    queryKey: adminKeys.evidenceObservations(productID ?? ''),
    queryFn: () => adminApi.evidenceObservations(productID ?? ''),
    enabled: Boolean(productID),
  })

// Every evidence write invalidates the whole admin tree. The governance detail,
// the source list and the observation list all move together — a new source can
// appear in a revision form, and a published revision changes the product row.
function useEvidenceMutation<Variables>(
  mutationFn: (variables: Variables) => Promise<unknown>,
) {
  const client = useQueryClient()
  return useMutation({
    mutationFn,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.root }),
  })
}

export const useCreateEvidenceSource = () =>
  useEvidenceMutation((input: EvidenceSourceInput) =>
    adminApi.createEvidenceSource(input),
  )

export const useReviewEvidenceSource = () =>
  useEvidenceMutation(
    ({ id, status, note }: { id: string; status: string; note: string }) =>
      adminApi.reviewEvidenceSource(id, status, note),
  )

export const useCreateEvidenceObservation = () =>
  useEvidenceMutation((input: EvidenceObservationInput) =>
    adminApi.createEvidenceObservation(input),
  )

export const useCreateEvidenceRevision = (productID: string) =>
  useEvidenceMutation((input: EvidenceRevisionInput) =>
    adminApi.createEvidenceRevision(productID, input),
  )

export const useEvidenceRevisionTransition = () =>
  useEvidenceMutation(
    ({
      revisionID,
      action,
      note,
    }: {
      revisionID: string
      action: 'submit' | 'approve' | 'reject' | 'publish'
      note: string
    }) => adminApi.transitionEvidenceRevision(revisionID, action, note),
  )

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

export function useCreateCommerceProvider() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (input: CommerceProviderInput) =>
      adminApi.createCommerceProvider(input),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.commerceProviders }),
  })
}

export function useCommerceProviderLifecycle() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      adminApi.setCommerceProviderLifecycle(id, status),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: adminKeys.commerceProviders }),
  })
}

export function useTriggerCommerceImport() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: adminApi.triggerCommerceImport,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}

export function useRetryCommerceImport() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: adminApi.retryCommerceImport,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}

export function useConversionProviderState() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      adminApi.setConversionProvider(id, enabled),
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}

export function useTriggerConversionImport() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: adminApi.triggerConversionImport,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}

export function useRetryConversionImport() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: adminApi.retryConversionImport,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}

export function useReconcileConversionImport() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: adminApi.reconcileConversionImport,
    onSuccess: () =>
      void client.invalidateQueries({ queryKey: ['admin', 'commerce'] }),
  })
}
