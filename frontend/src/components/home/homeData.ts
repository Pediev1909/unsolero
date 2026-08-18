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
      name: 'Demo Pocket CRM',
      reason: 'Shared client records without a setup project.',
      priceMinor: 900,
    },
    {
      name: 'Demo Loop Projects',
      reason: 'Delivery tracking that connects to the client record.',
      priceMinor: 1900,
    },
    {
      name: 'Demo Ledgerly Books',
      reason: 'Invoices generated from what was actually agreed.',
      priceMinor: 1500,
    },
  ] satisfies SetupItem[],
  totalMinor: 4300,
  remainingMinor: 7700,
  rejected:
    'A help desk and an analytics tool were left out: at four people they duplicate work the inbox and the project tool already cover, and neither stops work or stops payment.',
  upgrade:
    'Add scheduling once client booking becomes the bottleneck rather than a convenience.',
}

export const comparisonProducts: ComparisonProduct[] = [
  {
    name: 'Demo Pocket CRM',
    shortName: 'Pocket CRM',
    priceMinor: 900,
    planLabel: 'Entry plan, monthly',
    easeScore: 90,
    integrationScore: 74,
    verdict: 'Best fit for this brief',
    recommended: true,
  },
  {
    name: 'Demo Northwind CRM',
    shortName: 'Northwind CRM',
    priceMinor: 2900,
    planLabel: 'Entry plan, monthly',
    easeScore: 78,
    integrationScore: 92,
    verdict: 'Connects to more, costs more',
  },
  {
    name: 'Demo Metricly Insights',
    shortName: 'Metricly Insights',
    priceMinor: 2600,
    planLabel: 'Entry plan, monthly',
    easeScore: 64,
    integrationScore: 80,
    verdict: 'Useful later, not at this size',
  },
]

export const featuredProducts: ProductCardData[] = [
  {
    id: 'saas-northwind-crm',
    href: '/products/saas-northwind-crm',
    name: 'Demo Northwind CRM',
    brand: 'Demo Northwind Software',
    category: 'CRM',
    priceMinor: 2900,
    currency: 'USD',
    image: {
      src: '/images/saas-crm.webp',
      alt: 'Illustrative interface image for a fictional CRM product',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'saas-loop-projects',
    href: '/products/saas-loop-projects',
    name: 'Demo Loop Projects',
    brand: 'Demo Looplane',
    category: 'Project management',
    priceMinor: 1900,
    currency: 'USD',
    image: {
      src: '/images/saas-project-management.webp',
      alt: 'Illustrative interface image for a fictional project management product',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'saas-ledgerly-books',
    href: '/products/saas-ledgerly-books',
    name: 'Demo Ledgerly Books',
    brand: 'Demo Ledgerly',
    category: 'Accounting and invoicing',
    priceMinor: 1500,
    currency: 'USD',
    image: {
      src: '/images/saas-invoicing.webp',
      alt: 'Illustrative interface image for a fictional invoicing product',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'saas-beacon-mail',
    href: '/products/saas-beacon-mail',
    name: 'Demo Beacon Mail',
    brand: 'Demo Beacon',
    category: 'Email marketing',
    priceMinor: 2400,
    currency: 'USD',
    image: {
      src: '/images/saas-email-marketing.webp',
      alt: 'Illustrative interface image for a fictional email marketing product',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
]
