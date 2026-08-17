import { useQuery } from '@tanstack/react-query'

import { getContent, getContentEntry, type ContentQuery } from './api'

export const contentKeys = {
  all: ['content'] as const,
  list: (query: ContentQuery) => ['content', 'list', query] as const,
  entry: (slug: string) => ['content', 'entry', slug] as const,
}

export function useContent(query: ContentQuery) {
  return useQuery({
    queryKey: contentKeys.list(query),
    queryFn: () => getContent(query),
  })
}

export function useContentEntry(slug: string) {
  return useQuery({
    queryKey: contentKeys.entry(slug),
    queryFn: () => getContentEntry(slug),
    enabled: Boolean(slug),
  })
}
