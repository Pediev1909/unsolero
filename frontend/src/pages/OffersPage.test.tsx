import { screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { OffersPage } from './OffersPage'

function product(overrides: {
  id: string
  name: string
  slug: string
  category: { name: string; slug: string }
  merchant: string
  offer: string
  period?: 'monthly' | 'annual'
}) {
  return {
    id: overrides.id,
    name: overrides.name,
    slug: overrides.slug,
    brand: { name: 'Vendor', slug: 'vendor' },
    category: overrides.category,
    price: { amount_minor: 1300, currency: 'USD' },
    primary_image: null,
    key_specification: { label: 'Billing', value: 'Per month' },
    billing: {
      period: overrides.period ?? ('monthly' as const),
      unit: 'flat' as const,
      unit_note: null,
      annual_price_minor: null,
    },
    suitability: [],
    scores: {
      quality: 0,
      value: 0,
      durability: 0,
      beginner: 0,
      advanced: 0,
      apartment: 0,
      noise: 0,
      portability: 0,
    },
    is_demo: false,
    purchase_path: `/api/affiliate/click/${overrides.offer}`,
    disclosure_label: 'Affiliate link',
    merchant_name: overrides.merchant,
  }
}

const crm = product({
  id: 'p-1',
  name: 'HubSpot Starter',
  slug: 'hubspot-starter',
  category: { name: 'CRM', slug: 'crm' },
  merchant: 'HubSpot',
  offer: 'offer-1',
  period: 'annual',
})
const email = product({
  id: 'p-2',
  name: 'Mailchimp Standard',
  slug: 'mailchimp-standard',
  category: { name: 'Email marketing', slug: 'email-marketing' },
  merchant: 'Mailchimp',
  offer: 'offer-2',
})

const liveItems = [
  {
    product: crm,
    offer: {
      price: { amount_minor: 2000, currency: 'USD' },
      merchant_name: 'HubSpot',
      last_checked_at: '2026-08-20T12:00:00Z',
      freshness_status: 'stale' as const,
    },
  },
  {
    product: email,
    offer: {
      price: { amount_minor: 1300, currency: 'USD' },
      merchant_name: 'Mailchimp',
      last_checked_at: '2026-09-01T12:00:00Z',
      freshness_status: 'fresh' as const,
    },
  },
]

// What the mocked query returns, set per test. The factory below reads it
// lazily, on render, so each test can swap the state without re-mocking.
let liveOffersState: Record<string, unknown> = {}

vi.mock('../features/catalog/queries', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../features/catalog/queries')>()
  return { ...actual, useLiveOffers: () => liveOffersState }
})

function success(items: typeof liveItems) {
  return {
    data: { items, generated_at: '2026-09-02T08:00:00Z' },
    isPending: false,
    isError: false,
    isSuccess: true,
    refetch: vi.fn(),
  }
}

describe('OffersPage', () => {
  beforeEach(() => {
    liveOffersState = success(liveItems)
  })

  it('groups the live offers by category with a dated price and a tracked vendor exit', () => {
    renderWithProviders(<OffersPage />, { route: '/offers' })

    expect(document.title).toBe('Live vendor offers and trials | UNSOLERO')
    expect(
      screen.getByRole('heading', { level: 1, name: 'Live vendor offers' }),
    ).toBeInTheDocument()

    const crmGroup = screen.getByRole('region', { name: 'CRM' })
    const emailGroup = screen.getByRole('region', { name: 'Email marketing' })
    expect(
      within(crmGroup).getByRole('link', { name: 'HubSpot Starter' }),
    ).toHaveAttribute('href', '/products/hubspot-starter')
    expect(
      within(emailGroup).getByRole('link', { name: 'Mailchimp Standard' }),
    ).toHaveAttribute('href', '/products/mailchimp-standard')

    // The vendor exit is the tracked redirect, never a vendor URL, and is
    // marked as the paid relationship it is.
    const exit = within(emailGroup).getByRole('link', {
      name: /view at mailchimp/i,
    })
    expect(exit.getAttribute('href')).toMatch(
      /^\/api\/affiliate\/click\/offer-2\?source=product_detail/,
    )
    expect(exit).toHaveAttribute('rel', 'nofollow noopener sponsored')

    // The basis travels with every price: a yearly-only vendor's per-month
    // figure says so, and the phrase comes from the structured object rather
    // than the older string beside it.
    expect(
      within(crmGroup).getByText('Flat rate, billed yearly'),
    ).toBeInTheDocument()
    expect(
      within(emailGroup).getByText('Flat rate, monthly billing'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Per month')).toBeNull()

    // A fresh price is "checked"; a stale one says so in words.
    expect(
      within(emailGroup).getByText(/checked september 1, 2026/i),
    ).toBeInTheDocument()
    expect(
      within(crmGroup).getByText(/price last read august 20, 2026/i),
    ).toBeInTheDocument()
    expect(
      within(crmGroup).getByText(/not re-verified since/i),
    ).toBeInTheDocument()

    // Disclosure, and where the ranking rule is explained, both reachable.
    expect(
      screen.getAllByRole('link', { name: /affiliate disclosure/i })[0],
    ).toHaveAttribute('href', '/affiliate-disclosure')
    expect(
      screen.getByRole('link', { name: /how we rank software/i }),
    ).toHaveAttribute('href', '/articles/how-unsolero-ranks-software')
    expect(
      screen.getByText(/2 products with a working vendor link/i),
    ).toBeInTheDocument()
  })

  it('says honestly when nothing is live', () => {
    liveOffersState = success([])
    renderWithProviders(<OffersPage />, { route: '/offers' })

    expect(
      screen.getByRole('heading', { name: 'No live offers right now' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('region', { name: 'CRM' })).toBeNull()
  })

  it('shows an error rather than an empty list when the lookup fails', () => {
    liveOffersState = {
      data: undefined,
      isPending: false,
      isError: true,
      isSuccess: false,
      refetch: vi.fn(),
    }
    renderWithProviders(<OffersPage />, { route: '/offers' })

    expect(
      screen.getByRole('heading', { name: 'Offers unavailable' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/no live offers right now/i)).toBeNull()
  })
})
