import type {
  BuilderValues,
  RecommendationDraft,
  RecommendationInput,
} from './schemas'

export const defaultBuilderValues: BuilderValues = {
  goal: '',
  experience: '',
  budget_minor: 12_000,
  existing_equipment: [],
  training_preferences: [],
  priorities: [],
  free_text: '',
}

export function toRecommendationInput(
  values: BuilderValues,
): RecommendationInput | null {
  if (!values.goal || !values.experience) return null
  return {
    goal: values.goal,
    experience: values.experience,
    budget_minor: values.budget_minor,
    currency: 'USD',
    existing_equipment: values.existing_equipment,
    training_preferences: values.training_preferences,
    priorities: values.priorities,
    free_text: values.free_text.trim(),
  }
}

export function toDraft(values: BuilderValues, currentStep: number) {
  return {
    current_step: currentStep,
    goal: values.goal || null,
    experience: values.experience || null,
    budget_minor: values.budget_minor,
    currency: 'USD' as const,
    existing_equipment: values.existing_equipment,
    training_preferences: values.training_preferences,
    priorities: values.priorities,
    free_text: values.free_text,
  }
}

export function fromDraft(draft: RecommendationDraft): BuilderValues {
  return {
    goal: draft.goal ?? '',
    experience: draft.experience ?? '',
    budget_minor: draft.budget_minor ?? 12_000,
    existing_equipment: draft.existing_equipment,
    training_preferences: draft.training_preferences,
    priorities: draft.priorities,
    free_text: draft.free_text,
  }
}

export function fromRecommendationInput(
  input: RecommendationInput,
): BuilderValues {
  return {
    goal: input.goal,
    experience: input.experience,
    budget_minor: input.budget_minor,
    existing_equipment: input.existing_equipment,
    training_preferences: input.training_preferences,
    priorities: input.priorities,
    free_text: input.free_text,
  }
}
