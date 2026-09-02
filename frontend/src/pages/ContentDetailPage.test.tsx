import { screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { ToastProvider } from '../components/ui/ToastProvider'
import type { ProductSummary } from '../features/catalog/schemas'
import type { ContentDetail } from '../features/content/schemas'
import { renderWithProviders } from '../test/renderWithProviders'
import { ContentDetailPage } from './ContentDetailPage'

function product(overrides: Partial<ProductSummary>): ProductSummary {
  return {
    id: 'p',
    name: 'Tool',
    slug: 'tool',
    brand: { name: 'Vendor', slug: 'vendor' },
    category: { name: 'CRM', slug: 'crm' },
    price: { amount_minor: 2000, currency: 'USD' },
    primary_image: null,
    key_specification: { label: 'Billing', value: 'Per month' },
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
    ...overrides,
  }
}

const zohoCRM = product({
  id: 'zoho-crm',
  name: 'Zoho CRM Standard',
  slug: 'zoho-crm-standard',
  brand: { name: 'Zoho', slug: 'zoho' },
  purchase_path: '/api/affiliate/click/06a56f58-2d93-4e37-b965-be55e672def4',
  merchant_name: 'Zoho',
})
const hubspot = product({
  id: 'hubspot',
  name: 'HubSpot Starter Customer Platform',
  slug: 'hubspot-starter-customer-platform',
  brand: { name: 'HubSpot', slug: 'hubspot' },
})
const bigin = product({
  id: 'bigin',
  name: 'Bigin Express',
  slug: 'bigin-express',
  brand: { name: 'Zoho', slug: 'zoho' },
  price: { amount_minor: 900, currency: 'USD' },
  purchase_path: '/api/affiliate/click/07d4b7a5-b471-4c95-b379-5a70b31bbd0e',
  merchant_name: 'Zoho',
})

const entry: ContentDetail = {
  id: '3f0e2c8a-6d1b-4c5e-9a7f-2b4d6e8f0a1c',
  type: 'comparison',
  title: 'Zoho CRM vs HubSpot: the same price, two different bets',
  slug: 'zoho-crm-vs-hubspot',
  path: '/compare/zoho-crm-vs-hubspot',
  description:
    'Both cost 20 USD per user per month. Which one costs more later.',
  hero_image: {
    url: '/images/crm.svg',
    alt_text: 'Two CRMs',
    is_primary: true,
  },
  author_name: 'Andon Pediev',
  published_at: '2026-08-21T13:50:04Z',
  updated_at: '2026-08-24T13:08:20Z',
  covered: [],
  content: [
    { type: 'paragraph', text: 'Both cost 20 USD per user per month.' },
    { type: 'heading', heading: 'The prices' },
    {
      type: 'faq',
      heading: 'Questions people ask',
      questions: [
        {
          question: 'Is Zoho CRM cheaper than HubSpot?',
          answer: 'Not on the entry tier.',
        },
      ],
    },
  ],
  author: {
    name: 'Andon Pediev',
    slug: 'andon-pediev',
    bio: 'Builds and runs UNSOLERO.',
    avatar_url: null,
  },
  related_products: [zohoCRM, hubspot, bigin],
  related_categories: [],
  related_content: [],
  seo: {
    title: 'Zoho CRM vs HubSpot | UNSOLERO',
    description: 'Both cost 20 USD per user per month.',
    canonical_url: 'https://unsolero.com/compare/zoho-crm-vs-hubspot',
  },
}

vi.mock('../features/content/queries', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../features/content/queries')>()
  return {
    ...actual,
    useContentEntry: () => ({
      data: entry,
      isPending: false,
      isError: false,
      refetch: vi.fn(),
    }),
  }
})

// The strip's vendor control fetches each product's offer. The page test is
// about where the control appears, so every slug resolves to one live offer;
// the strip's own rule — no control without a purchase_path on the summary —
// is what decides whether it is drawn.
vi.mock('../features/catalog/queries', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../features/catalog/queries')>()
  return {
    ...actual,
    useOffers: (slug: string) => ({
      isPending: false,
      isError: false,
      data: [
        {
          id: `offer-${slug}`,
          merchant: {
            name: 'Zoho',
            slug: 'zoho',
            country_code: 'IN',
            trust_score: 80,
          },
          price: { amount_minor: 2000, currency: 'USD' },
          shipping_minor: 0,
          landed_price_minor: 2000,
          availability: 'in_stock',
          condition: 'new',
          last_checked_at: '2026-08-26T10:48:13Z',
          observed_at: null,
          expires_at: null,
          freshness_status: 'fresh',
          purchase_path: `/api/affiliate/click/${slug}`,
          disclosure_label: 'Affiliate link',
        },
      ],
    }),
  }
})

// The side-by-side fetches three product records. Its rendering is covered by
// ComparisonTable's own tests; here the point is that the page asks for it,
// for these products, read only.
vi.mock('../features/catalog/components/ComparisonData', () => ({
  ComparisonData: ({
    products,
    readOnly,
  }: {
    products: ProductSummary[]
    readOnly?: boolean
  }) => (
    <div
      data-products={products.map((item) => item.slug).join(',')}
      data-readonly={String(Boolean(readOnly))}
      data-testid="comparison-data"
    />
  ),
}))

function renderPage() {
  return renderWithProviders(
    <ToastProvider>
      <ContentDetailPage />
    </ToastProvider>,
    { route: entry.path },
  )
}

describe('ContentDetailPage', () => {
  it('puts the products the piece covers at a glance, with a vendor control only where there is a live offer', () => {
    renderPage()

    const strip = screen
      .getByRole('heading', { level: 2, name: 'At a glance' })
      .closest('section') as HTMLElement
    expect(strip).not.toBeNull()

    for (const item of [zohoCRM, hubspot, bigin]) {
      expect(
        within(strip).getByRole('link', { name: item.name }),
      ).toHaveAttribute('href', `/products/${item.slug}`)
    }
    // Two of the three carry a purchase_path. HubSpot has none, and gets
    // nothing rather than a disabled button.
    expect(
      within(strip).getAllByRole('link', { name: /View at Zoho/ }),
    ).toHaveLength(2)
    expect(within(strip).getByText(/affiliate links/i)).toBeInTheDocument()
  })

  it('draws the catalog comparison table read only for a versus piece, and drops the price scale', () => {
    renderPage()

    expect(
      screen.getByRole('heading', { level: 2, name: 'Side by side' }),
    ).toBeInTheDocument()
    const table = screen.getByTestId('comparison-data')
    expect(table).toHaveAttribute(
      'data-products',
      'zoho-crm-standard,hubspot-starter-customer-platform,bigin-express',
    )
    expect(table).toHaveAttribute('data-readonly', 'true')
    expect(
      screen.queryByText('What each one costs, entry paid tier'),
    ).toBeNull()
  })

  it('names the author with a role and the two links a sceptical reader needs', () => {
    renderPage()

    expect(screen.getByRole('link', { name: 'Andon Pediev' })).toHaveAttribute(
      'href',
      '/author/andon-pediev',
    )
    expect(screen.getByText('Founder, UNSOLERO')).toBeInTheDocument()
    expect(screen.getByText(/Published August 21, 2026/)).toBeInTheDocument()
    expect(screen.getByText(/Updated August 24, 2026/)).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'How we rank software' }),
    ).toHaveAttribute('href', '/articles/how-unsolero-ranks-software')
    expect(
      screen.getAllByRole('link', { name: 'Affiliate disclosure' })[0],
    ).toHaveAttribute('href', '/affiliate-disclosure')
  })

  it('lists the FAQ heading in the table of contents, pointing at its anchor', () => {
    renderPage()

    const toc = screen.getByRole('navigation', { name: 'Article sections' })
    expect(
      within(toc).getByRole('link', { name: 'The prices' }),
    ).toHaveAttribute('href', '#the-prices')
    expect(
      within(toc).getByRole('link', { name: 'Questions people ask' }),
    ).toHaveAttribute('href', '#questions-people-ask')
    expect(document.getElementById('questions-people-ask')).not.toBeNull()
  })
})
