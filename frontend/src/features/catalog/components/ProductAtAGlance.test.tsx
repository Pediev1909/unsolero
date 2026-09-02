import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Offer } from '../schemas'
import { ProductAtAGlance } from './ProductAtAGlance'
import { productDetailFixture } from './productDetailFixture'

// The strip reads offers through the same hook as the rest of the page. The
// point under test is what each answer of that hook makes the strip say, so
// the hook is replaced and steered per test.
const offersState = vi.hoisted(() => ({
  current: { isPending: false, isError: false, data: [] as Offer[] },
}))
vi.mock('../queries', () => ({
  useOffers: () => offersState.current,
}))

const liveOffer: Offer = {
  id: 'offer-1',
  merchant: {
    name: 'Mailchimp',
    slug: 'mailchimp',
    country_code: 'US',
    trust_score: 90,
  },
  price: { amount_minor: 2000, currency: 'USD' },
  shipping_minor: 0,
  landed_price_minor: 2000,
  availability: 'in_stock',
  condition: 'new',
  last_checked_at: '2026-08-26T12:00:00Z',
  observed_at: null,
  expires_at: null,
  freshness_status: 'fresh',
  purchase_path: '/api/affiliate/click/offer-1',
  disclosure_label: 'Affiliate link',
}

function renderStrip(product = productDetailFixture()) {
  return render(
    <MemoryRouter>
      <ProductAtAGlance product={product} />
    </MemoryRouter>,
  )
}

describe('ProductAtAGlance', () => {
  beforeEach(() => {
    offersState.current = { isPending: false, isError: false, data: [] }
  })

  it('answers price, billing basis, date read, best-for and evidence on one strip', () => {
    renderStrip()

    expect(screen.getByText('$20.00')).toBeInTheDocument()
    expect(screen.getByText('Flat rate, monthly billing')).toBeInTheDocument()
    // A monthly-only vendor has no second price to offer.
    expect(screen.queryByText(/billed yearly/)).toBeNull()
    // The same date source as the comparison table: the price fact's observation.
    expect(
      screen.getByText(
        /Read from the vendor on Aug 26, 2026 — not a live quote/,
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: /Open the page it came from/ }),
    ).toHaveAttribute('href', 'https://mailchimp.com/pricing/')
    expect(screen.getByText('Small-list newsletters')).toBeInTheDocument()
    // Three visible facts at 100, 70 and 90: the median is 90.
    expect(screen.getByText('3 facts')).toBeInTheDocument()
    expect(screen.getByText(/Median confidence 90\/100/)).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'See the record' }),
    ).toHaveAttribute('href', '#evidence')
    // There is no free-plan or trial field in the data, so the strip must
    // not claim one.
    expect(screen.queryByText(/free (plan|trial|version)/i)).toBeNull()
  })

  // Where the vendor sells both, the yearly rate follows the price as a
  // quieter line: offered, not pushed, and never mistaken for the compared
  // figure above it.
  it('offers the yearly rate as a second line when the vendor sells both', () => {
    renderStrip(
      productDetailFixture({
        billing: {
          period: 'monthly',
          unit: 'per_user',
          unit_note: null,
          annual_price_minor: 1500,
        },
      }),
    )
    expect(screen.getByText('$20.00')).toBeInTheDocument()
    expect(screen.getByText('Per user, monthly billing')).toBeInTheDocument()
    expect(screen.getByText('or $15.00/mo billed yearly')).toBeInTheDocument()
  })

  // An annual-only vendor's price already is the yearly rate. The basis says
  // so, and there is no second figure to offer.
  it('says billed yearly, with no second price, for an annual-only vendor', () => {
    renderStrip(
      productDetailFixture({
        billing: {
          period: 'annual',
          unit: 'flat',
          unit_note: null,
          annual_price_minor: null,
        },
      }),
    )
    expect(screen.getByText('Flat rate, billed yearly')).toBeInTheDocument()
    expect(screen.queryByText(/^or /)).toBeNull()
  })

  // A yearly rate that is not lower is a data error or a vendor with nothing
  // to offer for the commitment; either way it earns no line.
  it('drops the yearly line when the annual rate is not cheaper', () => {
    renderStrip(
      productDetailFixture({
        billing: {
          period: 'monthly',
          unit: 'flat',
          unit_note: null,
          annual_price_minor: 2000,
        },
      }),
    )
    expect(screen.queryByText(/billed yearly/)).toBeNull()
  })

  it('falls back to the server phrase when the response carries no billing object', () => {
    renderStrip(
      productDetailFixture({
        billing: undefined,
        key_specification: { label: 'Billing', value: 'Per month' },
      }),
    )
    expect(screen.getByText('Per month')).toBeInTheDocument()
  })

  it('sends the reader to the vendor with a disclosed affiliate link when there is a live offer', () => {
    offersState.current = {
      isPending: false,
      isError: false,
      data: [liveOffer],
    }
    renderStrip()

    const link = screen.getByRole('link', { name: /View at Mailchimp/ })
    expect(link.getAttribute('href')).toContain('/api/affiliate/click/offer-1')
    expect(link.getAttribute('href')).toContain('source=product_detail')
    expect(link).toHaveAttribute('rel', 'nofollow noopener sponsored')
    expect(screen.getByText(/Commission never changes/)).toBeInTheDocument()
    expect(screen.queryByText(/No affiliate offer/)).toBeNull()
  })

  // No invented link and no disabled button: a plain statement and the brand
  // page, which is where the vendor's other products are.
  it('says so, honestly, when there is no affiliate offer', () => {
    renderStrip()

    expect(
      screen.getByText('No affiliate offer for this product.'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'More from Mailchimp' }),
    ).toHaveAttribute('href', '/brands/mailchimp')
    expect(screen.queryByRole('link', { name: /View at/ })).toBeNull()
    expect(screen.queryByText(/Commission never/)).toBeNull()
  })

  it('treats an offers error like no offer rather than breaking the strip', () => {
    offersState.current = { isPending: false, isError: true, data: [] }
    renderStrip()
    expect(screen.getByText('$20.00')).toBeInTheDocument()
    expect(
      screen.getByText('No affiliate offer for this product.'),
    ).toBeInTheDocument()
  })

  it('leaves out the cells it has no answer for', () => {
    renderStrip(productDetailFixture({ use_cases: [], evidence: [] }))

    expect(screen.queryByText('Best for')).toBeNull()
    expect(screen.queryByText('Evidence')).toBeNull()
    expect(
      screen.getByText(/The day it was read is not recorded/),
    ).toBeInTheDocument()
    expect(screen.getByText('Entry price')).toBeInTheDocument()
    expect(screen.getByText('Where to get it')).toBeInTheDocument()
  })
})
