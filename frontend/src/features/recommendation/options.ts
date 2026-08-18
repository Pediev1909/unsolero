import type { BuilderValues } from './schemas'

// Every value below must exist in the active recommendation policy. The engine
// validates membership rather than shape, so an option that drifts out of the
// policy is rejected at submission instead of quietly scoring as no match.

export const goalOptions = [
  {
    value: 'client_services',
    label: 'Run a client services business',
    description: 'Win work, deliver it, and get paid without losing track.',
  },
  {
    value: 'sell_products_online',
    label: 'Sell products online',
    description: 'A storefront, payments, and a way to reach buyers again.',
  },
  {
    value: 'creator_business',
    label: 'Run a creator business',
    description: 'An audience you own, and something worth selling them.',
  },
  {
    value: 'software_product',
    label: 'Run a software product',
    description: 'Measure what users do, support them, and ship changes.',
  },
  {
    value: 'solo_consulting',
    label: 'Work solo as a consultant',
    description: 'Few clients, tight admin, no time for tool maintenance.',
  },
] as const

export const experienceOptions = [
  {
    value: 'beginner',
    label: 'No dedicated admin',
    description: 'Nobody is paid to configure or maintain tools.',
  },
  {
    value: 'intermediate',
    label: 'Someone owns the tooling',
    description: 'One person keeps the stack tidy alongside other work.',
  },
  {
    value: 'advanced',
    label: 'Comfortable with technical setup',
    description: 'APIs, automations, and migrations are not a barrier.',
  },
] as const

export const existingToolOptions = [
  { name: 'Team chat', category_slug: 'team-communication' },
  { name: 'CRM', category_slug: 'crm' },
  { name: 'Project management', category_slug: 'project-management' },
  { name: 'Invoicing or accounting', category_slug: 'accounting-invoicing' },
  { name: 'Email marketing', category_slug: 'email-marketing' },
  { name: 'Website builder', category_slug: 'website-builder' },
] as const

export const preferenceOptions = [
  { value: 'all_in_one', label: 'One suite over many tools' },
  { value: 'best_of_breed', label: 'The strongest tool per job' },
  { value: 'open_source', label: 'Open source' },
  { value: 'no_code', label: 'No-code' },
  { value: 'api_first', label: 'API-first' },
  { value: 'privacy_focused', label: 'Privacy-focused vendors' },
  { value: 'eu_hosted', label: 'EU-hosted data' },
] as const

export const priorityOptions = [
  { value: 'budget', label: 'Best value' },
  { value: 'ease_of_use', label: 'Easy to adopt' },
  { value: 'integrations', label: 'Connects to my stack' },
  { value: 'reliability', label: 'Reliability' },
  { value: 'vendor_stability', label: 'Vendor stability' },
  { value: 'data_portability', label: 'Data portability' },
] as const

// Seven steps. A software stack has no room to measure, so the physical space
// question that the equipment vertical asks does not exist here.
export const steps = [
  {
    label: 'Goal',
    title: 'What does your business do?',
    supporting:
      'We use this to decide which jobs your stack has to cover.',
  },
  {
    label: 'Team',
    title: 'Who will look after these tools?',
    supporting:
      'A tool nobody is paid to maintain gets abandoned. We weigh setup cost accordingly.',
  },
  {
    label: 'Budget',
    title: 'What can you spend per month?',
    supporting:
      'We will stay inside this amount—not treat it as a target to spend.',
  },
  {
    label: 'Current stack',
    title: 'What do you already run?',
    supporting:
      'We avoid tools that duplicate these, and favour ones that connect to them.',
  },
  {
    label: 'Preferences',
    title: 'How do you prefer to buy software?',
    supporting:
      'Choose at least one. Preferences refine the ranking after hard constraints.',
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
