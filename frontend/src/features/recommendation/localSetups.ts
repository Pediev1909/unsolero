import { useMemo } from 'react'
import { z } from 'zod'

import {
  useLocalDocument,
  writeLocalDocument,
} from '../../lib/storage/localStorageStore'
import {
  recommendationResultSchema,
  type RecommendationResult,
} from './schemas'
import { trackEvent } from '../analytics/tracking'

const storageKey = 'rigmark:saved-setups:v1'
const localSetupSchema = z.object({
  id: z.string(),
  name: z.string().min(1).max(120),
  created_at: z.string(),
  updated_at: z.string(),
  result: recommendationResultSchema,
})
const localSetupsSchema = z.array(localSetupSchema)
export type LocalSetup = z.infer<typeof localSetupSchema>

export function useLocalSetups() {
  const raw = useLocalDocument(storageKey)
  const setups = useMemo(() => parseLocalSetups(raw), [raw])
  function commit(next: LocalSetup[]) {
    writeLocalDocument(storageKey, next)
  }
  return {
    setups,
    get(id: string) {
      return setups.find((setup) => setup.id === id)
    },
    save(result: RecommendationResult, name = 'Personalized Home Gym') {
      const now = new Date().toISOString()
      const setupName = uniqueName(setups, name)
      const setup: LocalSetup = {
        id: crypto.randomUUID(),
        name: setupName,
        created_at: now,
        updated_at: now,
        result: { ...result, saved: false, setup_name: setupName },
      }
      commit([setup, ...setups])
      trackEvent('setup_saved', 'recommendation', {
        setup_id: setup.id,
        persistence: 'browser',
      })
      return setup
    },
    rename(id: string, name: string) {
      const trimmed = name.trim()
      if (!trimmed || trimmed.length > 120)
        throw new Error(
          'Setup names must contain between 1 and 120 characters.',
        )
      commit(
        setups.map((setup) =>
          setup.id === id
            ? {
                ...setup,
                name: trimmed,
                updated_at: new Date().toISOString(),
                result: { ...setup.result, setup_name: trimmed },
              }
            : setup,
        ),
      )
    },
    remove(id: string) {
      commit(setups.filter((setup) => setup.id !== id))
    },
  }
}

function parseLocalSetups(raw: string): LocalSetup[] {
  if (!raw) return []
  try {
    const parsed = localSetupsSchema.safeParse(JSON.parse(raw))
    return parsed.success ? parsed.data : []
  } catch {
    return []
  }
}

function uniqueName(setups: LocalSetup[], base: string) {
  const names = new Set(setups.map((setup) => setup.name.toLocaleLowerCase()))
  if (!names.has(base.toLocaleLowerCase())) return base
  let suffix = 2
  while (names.has(`${base} ${suffix}`.toLocaleLowerCase())) suffix += 1
  return `${base} ${suffix}`
}
