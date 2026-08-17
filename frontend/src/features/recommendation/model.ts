import type {
  BuilderValues,
  RecommendationDraft,
  RecommendationInput,
} from './schemas'

export const spaceOptions = {
  small_apartment: {
    length_mm: 2400,
    width_mm: 1800,
    height_mm: 2400,
    apartment_living: true,
  },
  spare_room: {
    length_mm: 3500,
    width_mm: 3000,
    height_mm: 2400,
    apartment_living: false,
  },
  half_garage: {
    length_mm: 5000,
    width_mm: 3000,
    height_mm: 2600,
    apartment_living: false,
  },
  full_garage: {
    length_mm: 6000,
    width_mm: 5500,
    height_mm: 2800,
    apartment_living: false,
  },
} as const

export const defaultBuilderValues: BuilderValues = {
  goal: '',
  experience: '',
  space_preset: '',
  budget_minor: 70_000,
  existing_equipment: [],
  training_preferences: [],
  priorities: [],
  free_text: '',
}

export function toRecommendationInput(
  values: BuilderValues,
): RecommendationInput | null {
  if (!values.goal || !values.experience || !values.space_preset) return null
  return {
    goal: values.goal,
    experience: values.experience,
    budget_minor: values.budget_minor,
    currency: 'USD',
    available_space: spaceOptions[values.space_preset],
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
    available_space: values.space_preset
      ? spaceOptions[values.space_preset]
      : null,
    existing_equipment: values.existing_equipment,
    training_preferences: values.training_preferences,
    priorities: values.priorities,
    free_text: values.free_text,
  }
}

export function fromDraft(draft: RecommendationDraft): BuilderValues {
  const spacePreset = Object.entries(spaceOptions).find(
    ([, value]) =>
      draft.available_space?.length_mm === value.length_mm &&
      draft.available_space.width_mm === value.width_mm &&
      draft.available_space.height_mm === value.height_mm &&
      draft.available_space.apartment_living === value.apartment_living,
  )?.[0]
  return {
    goal: draft.goal ?? '',
    experience: draft.experience ?? '',
    space_preset: (spacePreset as BuilderValues['space_preset']) ?? '',
    budget_minor: draft.budget_minor ?? 70_000,
    existing_equipment: draft.existing_equipment,
    training_preferences: draft.training_preferences,
    priorities: draft.priorities,
    free_text: draft.free_text,
  }
}

export function fromRecommendationInput(
  input: RecommendationInput,
): BuilderValues {
  const spacePreset = Object.entries(spaceOptions).find(
    ([, value]) =>
      input.available_space.length_mm === value.length_mm &&
      input.available_space.width_mm === value.width_mm &&
      input.available_space.height_mm === value.height_mm &&
      input.available_space.apartment_living === value.apartment_living,
  )?.[0]
  return {
    goal: input.goal,
    experience: input.experience,
    space_preset: (spacePreset as BuilderValues['space_preset']) ?? '',
    budget_minor: input.budget_minor,
    existing_equipment: input.existing_equipment,
    training_preferences: input.training_preferences,
    priorities: input.priorities,
    free_text: input.free_text,
  }
}
