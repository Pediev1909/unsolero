import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  generateRecommendation,
  deleteSetup,
  getDraft,
  getSetup,
  getSetups,
  saveDraft,
  renameSetup,
} from './api'
import type { RecommendationInput } from './schemas'
import { completeOnboarding, trackEvent } from '../analytics/tracking'

export const recommendationKeys = {
  draft: ['recommendations', 'draft'] as const,
  setups: ['account', 'setups'] as const,
  setup: (id: string) => ['account', 'setups', id] as const,
}

export function useRecommendationDraft(enabled: boolean) {
  return useQuery({
    queryKey: recommendationKeys.draft,
    queryFn: getDraft,
    enabled,
    staleTime: Infinity,
  })
}

export function useSaveRecommendationDraft() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: saveDraft,
    onSuccess: (draft) =>
      queryClient.setQueryData(recommendationKeys.draft, draft),
  })
}

export function useGenerateRecommendation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: RecommendationInput) => generateRecommendation(input),
    onSuccess: (result) => {
      completeOnboarding(result.status)
      trackEvent('recommendation_generated', 'recommendation', {
        status: result.status,
        persistence: result.saved ? 'account' : 'browser',
      })
      if (result.saved) {
        void queryClient.invalidateQueries({
          queryKey: recommendationKeys.setups,
        })
        queryClient.setQueryData(recommendationKeys.draft, null)
        if (result.setup_id) {
          trackEvent('setup_saved', 'recommendation', {
            setup_id: result.setup_id,
            persistence: 'account',
          })
        }
      }
    },
  })
}

export function useSetups(enabled = true) {
  return useQuery({
    queryKey: recommendationKeys.setups,
    queryFn: getSetups,
    enabled,
  })
}

export function useSetup(setupID: string, enabled = true) {
  return useQuery({
    queryKey: recommendationKeys.setup(setupID),
    queryFn: () => getSetup(setupID),
    enabled: Boolean(setupID) && enabled,
  })
}

export function useRenameSetup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ setupID, name }: { setupID: string; name: string }) =>
      renameSetup(setupID, name),
    onSuccess: (_data, { setupID }) => {
      void queryClient.invalidateQueries({
        queryKey: recommendationKeys.setups,
      })
      void queryClient.invalidateQueries({
        queryKey: recommendationKeys.setup(setupID),
      })
    },
  })
}

export function useDeleteSetup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteSetup,
    onSuccess: (_data, setupID) => {
      queryClient.removeQueries({ queryKey: recommendationKeys.setup(setupID) })
      void queryClient.invalidateQueries({
        queryKey: recommendationKeys.setups,
      })
    },
  })
}
