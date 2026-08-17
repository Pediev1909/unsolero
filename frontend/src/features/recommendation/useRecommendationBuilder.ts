import { useEffect, useRef, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'

import { useCurrentUser } from '../auth/queries'
import {
  builderValuesSchema,
  recommendationInputSchema,
  recommendationResultSchema,
  type BuilderValues,
  type RecommendationInput,
  type RecommendationResult,
} from './schemas'
import {
  useGenerateRecommendation,
  useRecommendationDraft,
  useSaveRecommendationDraft,
} from './queries'
import {
  defaultBuilderValues,
  fromDraft,
  fromRecommendationInput,
  toDraft,
  toRecommendationInput,
} from './model'

const storageKey = 'rigmark:recommendation-builder:v1'

export function useRecommendationBuilder(initialInput?: RecommendationInput) {
  const account = useCurrentUser()
  const draft = useRecommendationDraft(Boolean(account.data))
  const saveDraft = useSaveRecommendationDraft()
  const persistDraft = saveDraft.mutate
  const generate = useGenerateRecommendation()
  const restored = initialInput ? null : readLocalProgress()
  const form = useForm<BuilderValues>({
    defaultValues: initialInput
      ? fromRecommendationInput(initialInput)
      : (restored?.values ?? defaultBuilderValues),
  })
  const [currentStep, setCurrentStep] = useState(restored?.currentStep ?? 0)
  const [stepError, setStepError] = useState<string | null>(null)
  const [result, setResult] = useState<RecommendationResult | null>(null)
  const appliedServerDraft = useRef(Boolean(initialInput))
  const values = useWatch({ control: form.control }) as BuilderValues

  useEffect(() => {
    if (!account.data || !draft.isSuccess || appliedServerDraft.current) return
    appliedServerDraft.current = true
    if (draft.data) {
      const savedDraft = draft.data
      const timeout = window.setTimeout(() => {
        form.reset(fromDraft(savedDraft))
        setCurrentStep(savedDraft.current_step - 1)
      }, 0)
      return () => window.clearTimeout(timeout)
    }
  }, [account.data, draft.data, draft.isSuccess, form])

  useEffect(() => {
    window.sessionStorage.setItem(
      storageKey,
      JSON.stringify({ currentStep, values }),
    )
    if (!account.data || !draft.isSuccess || generate.isPending || result)
      return
    const timeout = window.setTimeout(() => {
      persistDraft(toDraft(values, currentStep + 1))
    }, 650)
    return () => window.clearTimeout(timeout)
  }, [
    account.data,
    currentStep,
    draft.isSuccess,
    generate.isPending,
    persistDraft,
    result,
    values,
  ])

  async function next() {
    const error = validateStep(currentStep, values)
    if (error) {
      setStepError(error)
      return
    }
    setStepError(null)
    if (currentStep < 7) {
      setCurrentStep((step) => step + 1)
      window.scrollTo({ top: 0, behavior: 'smooth' })
      return
    }
    const input = toRecommendationInput(values)
    const parsed = recommendationInputSchema.safeParse(input)
    if (!parsed.success) {
      setStepError('Review your answers before building the setup.')
      return
    }
    try {
      const generated = await generate.mutateAsync(parsed.data)
      setResult(generated)
      if (generated.saved) window.sessionStorage.removeItem(storageKey)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } catch {
      // The mutation exposes the server error to the page without discarding answers.
    }
  }

  function back() {
    setStepError(null)
    setCurrentStep((step) => Math.max(0, step - 1))
  }

  function edit(step: number) {
    setResult(null)
    setCurrentStep(step)
    setStepError(null)
  }

  return {
    form,
    values,
    currentStep,
    stepError,
    result,
    account,
    draft,
    saveDraft,
    generate,
    next,
    back,
    edit,
  }
}

function validateStep(step: number, values: BuilderValues): string | null {
  switch (step) {
    case 0:
      return values.goal ? null : 'Choose your primary goal.'
    case 1:
      return values.experience ? null : 'Choose your experience level.'
    case 2:
      return values.space_preset
        ? null
        : 'Choose the space that best matches your training area.'
    case 3:
      return values.budget_minor >= 10_000
        ? null
        : 'Enter a budget of at least $100.'
    case 5:
      return values.training_preferences.length > 0
        ? null
        : 'Choose at least one training preference.'
    case 6:
      return values.priorities.length > 0
        ? null
        : 'Choose at least one priority.'
    case 7:
      return values.free_text.length <= 1000
        ? null
        : 'Keep the description under 1,000 characters.'
    default:
      return null
  }
}

function readLocalProgress(): {
  currentStep: number
  values: BuilderValues
} | null {
  try {
    const raw = window.sessionStorage.getItem(storageKey)
    if (!raw) return null
    const value = JSON.parse(raw) as { currentStep?: unknown; values?: unknown }
    const parsed = builderValuesSchema.safeParse(value.values)
    if (!parsed.success || typeof value.currentStep !== 'number') return null
    return {
      currentStep: Math.min(7, Math.max(0, Math.floor(value.currentStep))),
      values: parsed.data,
    }
  } catch {
    return null
  }
}

export function parseStoredResult(value: unknown) {
  return recommendationResultSchema.safeParse(value)
}
