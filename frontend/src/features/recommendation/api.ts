import { apiRequest } from '../../lib/api/client'
import {
  draftSchema,
  recommendationResultSchema,
  setupListSchema,
  type RecommendationInput,
  type SetupList,
} from './schemas'

export function getDraft() {
  return apiRequest('/recommendations/draft', { method: 'GET' }, (value) =>
    value === undefined ? null : draftSchema.parse(value),
  )
}

export function saveDraft(input: unknown) {
  return apiRequest(
    '/recommendations/draft',
    { method: 'PUT', body: input },
    (value) => draftSchema.parse(value),
  )
}

export function generateRecommendation(input: RecommendationInput) {
  return apiRequest(
    '/recommendations/generate',
    { method: 'POST', body: input },
    (value) => recommendationResultSchema.parse(value),
  )
}

export async function getSetups() {
  const setups: SetupList['setups'] = []
  let page = 1
  while (true) {
    const result = await apiRequest(
      `/account/setups?page=${page}&page_size=100`,
      { method: 'GET' },
      (value) => setupListSchema.parse(value),
    )
    setups.push(...result.setups)
    if (page >= result.total_pages || page >= 10_000) break
    page += 1
  }
  return { setups }
}

export function getSetup(setupID: string) {
  return apiRequest(
    `/account/setups/${encodeURIComponent(setupID)}`,
    { method: 'GET' },
    (value) => recommendationResultSchema.parse(value),
  )
}

export function renameSetup(setupID: string, name: string) {
  return apiRequest(
    `/account/setups/${encodeURIComponent(setupID)}`,
    { method: 'PATCH', body: { name } },
    () => undefined,
  )
}

export function deleteSetup(setupID: string) {
  return apiRequest(
    `/account/setups/${encodeURIComponent(setupID)}`,
    { method: 'DELETE' },
    () => undefined,
  )
}
