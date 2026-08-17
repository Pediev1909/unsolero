import { apiRequest } from '../../lib/api/client'
import {
  draftSchema,
  recommendationResultSchema,
  setupListSchema,
  type RecommendationInput,
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

export function getSetups() {
  return apiRequest('/account/setups', { method: 'GET' }, (value) =>
    setupListSchema.parse(value),
  )
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
