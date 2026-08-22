import type { ProductCardData } from '../product'

export interface SetupItem {
  name: string
  reason: string
  priceMinor: number
}

export interface ComparisonProduct {
  name: string
  shortName: string
  priceMinor: number
  /** Which tier the price refers to. Software is tiered, so a bare price is
   *  ambiguous without saying what it buys. */
  planLabel: string
  easeScore: number
  /** How well the tool connects to the rest of a stack. This is the dimension
   *  that separates a coherent stack from a pile of subscriptions. */
  integrationScore: number
  verdict: string
  recommended?: boolean
}

export const exampleSetup = {
  profile: {
    goal: 'Run a client services business',
    teamSize: '4 people',
    experience: 'No dedicated admin',
    budgetMinor: 12000,
    owned: 'Team chat',
  },
  items: [
    {
      name: 'HubSpot Starter Customer Platform',
      reason:
        'Shared client records without a setup project, and a free tier to start on.',
      priceMinor: 2000,
    },
    {
      name: 'ClickUp Unlimited',
      reason: 'Delivery tracking that connects to the client record.',
      priceMinor: 1000,
    },
    {
      name: 'Zoho Books Standard',
      reason: 'Invoices generated from what was actually agreed.',
      priceMinor: 1200,
    },
  ] satisfies SetupItem[],
  totalMinor: 4200,
  remainingMinor: 7800,
  rejected:
    'A help desk and an analytics tool were left out: at four people they duplicate work the inbox and the project tool already cover, and neither stops work or stops payment.',
  upgrade:
    'Add scheduling once client booking becomes the bottleneck rather than a convenience.',
}

export const comparisonProducts: ComparisonProduct[] = [
  {
    name: 'HubSpot Starter Customer Platform',
    shortName: 'HubSpot Starter',
    priceMinor: 2000,
    planLabel: 'Entry plan, monthly',
    easeScore: 90,
    integrationScore: 74,
    verdict: 'Best fit for this brief',
    recommended: true,
  },
  {
    name: 'Salesflare Growth',
    shortName: 'Salesflare',
    priceMinor: 3900,
    planLabel: 'Entry plan, monthly',
    easeScore: 78,
    integrationScore: 92,
    verdict: 'Connects to more, costs more',
  },
  {
    name: 'ClickUp Unlimited',
    shortName: 'ClickUp',
    priceMinor: 1000,
    planLabel: 'Entry plan, monthly',
    easeScore: 64,
    integrationScore: 80,
    verdict: 'Useful later, not at this size',
  },
]

export const featuredProducts: ProductCardData[] = [
  {
    id: 'clickup-unlimited',
    href: '/products/clickup-unlimited',
    name: 'ClickUp Unlimited',
    brand: 'ClickUp',
    category: 'Project management',
    priceMinor: 1000,
    currency: 'USD',
  },
  {
    id: 'zoho-books-standard',
    href: '/products/zoho-books-standard',
    name: 'Zoho Books Standard',
    brand: 'Zoho',
    category: 'Accounting and invoicing',
    priceMinor: 1200,
    currency: 'USD',
  },
  {
    id: 'hubspot-starter-customer-platform',
    href: '/products/hubspot-starter-customer-platform',
    name: 'HubSpot Starter Customer Platform',
    brand: 'HubSpot',
    category: 'CRM',
    priceMinor: 2000,
    currency: 'USD',
  },
  {
    id: 'salesflare-growth',
    href: '/products/salesflare-growth',
    name: 'Salesflare Growth',
    brand: 'Salesflare',
    category: 'CRM',
    priceMinor: 3900,
    currency: 'USD',
  },
]
