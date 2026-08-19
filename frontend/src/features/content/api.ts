import { z } from 'zod'

import { apiRequest } from '../../lib/api/client'
import {
  contentAuthorPageSchema,
  contentDetailSchema,
  contentSummarySchema,
} from './schemas'

export interface ContentQuery {
  section?: 'all' | 'articles' | 'guides' | 'comparisons'
  category?: string
  limit?: number
}

export function getContent(query: ContentQuery) {
  const values = new URLSearchParams()
  if (query.section) values.set('section', query.section)
  if (query.category) values.set('category', query.category)
  if (query.limit) values.set('limit', String(query.limit))
  const encoded = values.toString()
  return apiRequest(
    `/content${encoded ? `?${encoded}` : ''}`,
    { method: 'GET' },
    (value) => z.array(contentSummarySchema).parse(value),
  )
}

export function getContentEntry(slug: string) {
  return apiRequest(`/content/${slug}`, { method: 'GET' }, (value) =>
    contentDetailSchema.parse(value),
  )
}

export function getContentAuthor(slug: string) {
  return apiRequest(`/content/authors/${slug}`, { method: 'GET' }, (value) =>
    contentAuthorPageSchema.parse(value),
  )
}
