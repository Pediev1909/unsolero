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
      priceMinor: 2000,
    },
  ] satisfies SetupItem[],
  totalMinor: 5000,
  remainingMinor: 7000,
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
