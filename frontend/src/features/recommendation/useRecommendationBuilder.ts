import { useEffect, useRef, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'

import { useCurrentUser } from '../auth/queries'
import { steps } from './options'
import {
  builderValuesSchema,
  recommendationInputSchema,
  recommendationPreviewInputSchema,
  recommendationResultSchema,
  type BuilderValues,
  type RecommendationInput,
  type RecommendationResult,
} from './schemas'
import {
  useGenerateRecommendation,
  useRecommendationDraft,
  useRecommendationPreview,
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
  // React Query clears isError the moment the next mutation starts, and this
  // one fires on every answer. While saving is genuinely failing that made the
  // warning strobe — gone on each click, back a moment later — which reads as a
  // glitch rather than as a problem. Latch it: it appears on the first failure
  // and stays until a save succeeds.
  const [saveHasFailed, setSaveHasFailed] = useState(false)
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
      persistDraft(toDraft(values, currentStep + 1), {
        onError: () => setSaveHasFailed(true),
        onSuccess: () => setSaveHasFailed(false),
      })
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
    if (currentStep < steps.length - 1) {
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

  // The live suggestion. It waits until the goal and the owner question are
  // answered, because those two are what make an input valid at all -- the
  // budget already carries a visible default the visitor is about to touch.
  // Nothing here is invented on their behalf.
  //
  // Debounced hard: this fires while somebody drags a slider, and a request
  // per frame would be rude to the server and useless to the reader.
  const [previewValues, setPreviewValues] = useState<BuilderValues | null>(null)
  const previewActive = currentStep >= 2 && !result
  useEffect(() => {
    if (!previewActive) return
    const timeout = window.setTimeout(() => setPreviewValues(values), 500)
    return () => window.clearTimeout(timeout)
  }, [previewActive, values])

  // Gated on previewActive rather than cleared when it turns off, so the
  // effect never sets state synchronously. Whatever the last preview held
  // stays put and is simply not used until the preview is live again.
  const previewInput =
    previewActive && previewValues ? toRecommendationInput(previewValues) : null
  const parsedPreview = previewInput
    ? recommendationPreviewInputSchema.safeParse(previewInput)
    : null
  const preview = useRecommendationPreview(
    parsedPreview?.success ? parsedPreview.data : null,
    previewActive,
  )

  return {
    form,
    values,
    currentStep,
    preview,
    stepError,
    result,
    account,
    draft,
    saveDraft,
    saveHasFailed,
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
      return values.experience
        ? null
        : 'Choose who will look after these tools.'
    case 2:
      return values.budget_minor >= 10_000
        ? null
        : 'Enter a monthly budget of at least $100.'
    case 4:
      return values.training_preferences.length > 0
        ? null
        : 'Choose at least one preference.'
    case 5:
      return values.priorities.length > 0
        ? null
        : 'Choose at least one priority.'
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
      currentStep: Math.min(
        steps.length - 1,
        Math.max(0, Math.floor(value.currentStep)),
      ),
      values: parsed.data,
    }
  } catch {
    return null
  }
}

export function parseStoredResult(value: unknown) {
  return recommendationResultSchema.safeParse(value)
}
