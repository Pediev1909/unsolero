import type { BuilderValues } from './schemas'

export const goalOptions = [
  {
    value: 'build_muscle',
    label: 'Build muscle',
    description: 'A balanced hypertrophy setup for progressive resistance.',
  },
  {
    value: 'strength',
    label: 'Get stronger',
    description: 'Prioritize load, stability and long-term progression.',
  },
  {
    value: 'general_fitness',
    label: 'General fitness',
    description: 'A versatile setup for strength, movement and conditioning.',
  },
  {
    value: 'weight_loss',
    label: 'Improve conditioning',
    description: 'Efficient equipment for consistent, repeatable training.',
  },
  {
    value: 'mobility',
    label: 'Move better',
    description: 'Low-impact tools for mobility, control and resilience.',
  },
] as const

export const experienceOptions = [
  {
    value: 'beginner',
    label: 'Beginner',
    description: 'I am building consistency and learning the fundamentals.',
  },
  {
    value: 'intermediate',
    label: 'Intermediate',
    description: 'I train regularly and know the movements I prefer.',
  },
  {
    value: 'advanced',
    label: 'Advanced',
    description: 'I need equipment that supports serious progression.',
  },
] as const

export const spacePresetOptions = [
  {
    value: 'small_apartment',
    label: 'Small apartment',
    description: 'A compact training zone with close neighbors.',
  },
  {
    value: 'spare_room',
    label: 'Spare room',
    description: 'A dedicated room with moderate floor space.',
  },
  {
    value: 'half_garage',
    label: 'Half garage',
    description: 'Room for larger equipment while sharing the space.',
  },
  {
    value: 'full_garage',
    label: 'Full garage',
    description: 'A dedicated footprint with generous clearance.',
  },
] as const

export const equipmentOptions = [
  { name: 'Pull-up bar', category_slug: 'pull-up-bars' },
  { name: 'Adjustable dumbbells', category_slug: 'adjustable-dumbbells' },
  { name: 'Training bench', category_slug: 'benches' },
  { name: 'Resistance bands', category_slug: 'resistance-bands' },
  { name: 'Kettlebell', category_slug: 'kettlebells' },
  { name: 'Barbell', category_slug: 'barbells' },
] as const

export const preferenceOptions = [
  { value: 'dumbbells', label: 'Dumbbells' },
  { value: 'barbell', label: 'Barbell training' },
  { value: 'kettlebell', label: 'Kettlebells' },
  { value: 'resistance_bands', label: 'Resistance bands' },
  { value: 'bodyweight', label: 'Bodyweight' },
  { value: 'cardio', label: 'Cardio' },
  { value: 'low_impact', label: 'Low impact' },
] as const

export const priorityOptions = [
  { value: 'budget', label: 'Best value' },
  { value: 'compact', label: 'Compact footprint' },
  { value: 'quality', label: 'Build quality' },
  { value: 'durability', label: 'Long-term durability' },
  { value: 'quiet', label: 'Quiet training' },
  { value: 'portability', label: 'Easy to move' },
] as const

export const steps = [
  {
    label: 'Goal',
    title: 'What do you want your training to achieve?',
    supporting:
      'We use this to decide which capabilities your setup must cover.',
  },
  {
    label: 'Experience',
    title: 'Where are you in your training journey?',
    supporting:
      'The right equipment should support your current skill without limiting your next step.',
  },
  {
    label: 'Space',
    title: 'Where will you train?',
    supporting:
      'Choose the closest match. We apply real product dimensions and apartment suitability.',
  },
  {
    label: 'Budget',
    title: 'What is your complete setup budget?',
    supporting:
      'We will stay inside this amount—not treat it as a target to spend.',
  },
  {
    label: 'Owned',
    title: 'What do you already own?',
    supporting:
      'We will avoid redundant equipment and account for useful compatibility.',
  },
  {
    label: 'Training',
    title: 'How do you like to train?',
    supporting:
      'Choose at least one. Your preferences refine the ranking after hard constraints.',
  },
  {
    label: 'Priorities',
    title: 'What matters most in the decision?',
    supporting:
      'Choose at least one. These priorities adjust transparent, configurable weights.',
  },
  {
    label: 'Context',
    title: 'Anything else we should know?',
    supporting:
      'Optional context is saved with your brief. It does not invent or override product facts.',
  },
] as const

export type GoalValue = Exclude<BuilderValues['goal'], ''>
export type ExperienceValue = Exclude<BuilderValues['experience'], ''>
export type SpacePresetValue = Exclude<BuilderValues['space_preset'], ''>
