import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import type { Billing, ProductSummary } from '../schemas'
import { SeatCostCalculator } from './SeatCostCalculator'

const perUser: Billing = {
  period: 'monthly',
  unit: 'per_user',
  unit_note: null,
  annual_price_minor: null,
}
const flatRate: Billing = { ...perUser, unit: 'flat' }

function product(
  name: string,
  amountMinor: number,
  billing: Billing | undefined,
): ProductSummary {
  const slug = name.toLowerCase().replace(/\s+/g, '-')
  return {
    id: slug,
    name,
    slug,
    brand: { name: 'Vendor', slug: 'vendor' },
    category: { name: 'CRM', slug: 'crm' },
    price: { amount_minor: amountMinor, currency: 'USD' },
    primary_image: null,
    key_specification: { label: 'Billing', value: 'Per month' },
    billing,
    suitability: [],
    scores: {
      quality: 70,
      value: 70,
      durability: 70,
      beginner: 70,
      advanced: 70,
      apartment: 70,
      noise: 70,
      portability: 70,
    },
    is_demo: false,
  }
}

const seatCRM = product('Seat CRM', 2000, perUser)
const liteCRM = product('Lite CRM', 1000, perUser)
const flatSuite = product('Flat Suite', 900, flatRate)

function renderCalculator(products: ProductSummary[]) {
  return render(
    <MemoryRouter>
      <SeatCostCalculator products={products} />
    </MemoryRouter>,
  )
}

describe('SeatCostCalculator', () => {
  it('renders nothing for a category with no per-seat prices', () => {
    const { container } = renderCalculator([flatSuite])
    expect(container).toBeEmptyDOMElement()
  })

  // The basis comes from the API's structured unit. Before that existed, the
  // phrase "Per month" on every product kept the calculator dormant; a
  // response without the object still does, rather than multiplying a guess.
  it('renders nothing when the API states no billing basis', () => {
    const { container } = renderCalculator([
      product('Unknown CRM', 2000, undefined),
    ])
    expect(container).toBeEmptyDOMElement()
  })

  it('totals the page for five seats, cheapest first, and leaves flat prices alone', () => {
    renderCalculator([seatCRM, flatSuite, liteCRM])
    const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
    expect(rows.map((row) => row.textContent)).toEqual([
      expect.stringContaining('Flat Suite'),
      expect.stringContaining('Lite CRM'),
      expect.stringContaining('Seat CRM'),
    ])
    expect(rows[0]).toHaveTextContent('$9.00 flat')
    expect(rows[0]).toHaveTextContent('$9.00')
    expect(rows[1]).toHaveTextContent('$10.00 per seat')
    expect(rows[1]).toHaveTextContent('$50.00')
    expect(rows[2]).toHaveTextContent('$20.00 per seat')
    expect(rows[2]).toHaveTextContent('$100.00')
    expect(screen.getByRole('link', { name: 'Seat CRM' })).toHaveAttribute(
      'href',
      '/products/seat-crm',
    )
  })

  // The total column is headed "per month". A yearly-only vendor's per-seat
  // figure is a per-month equivalent on a different contract, and the row has
  // to say so or the column is comparing two promises as one.
  it('marks a per-seat price that is only sold on a yearly contract', () => {
    renderCalculator([
      seatCRM,
      product('Annual CRM', 1500, { ...perUser, period: 'annual' }),
    ])
    const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
    expect(rows[0]).toHaveTextContent('$15.00 per seat, billed yearly')
    expect(rows[0]).toHaveTextContent('$75.00')
    expect(rows[1]).toHaveTextContent('$20.00 per seat')
    expect(rows[1]).not.toHaveTextContent('billed yearly')
  })

  it('recomputes when the team grows or shrinks, within one to fifty seats', async () => {
    renderCalculator([seatCRM, flatSuite])
    const seats = screen.getByRole('spinbutton', { name: 'Team size' })
    await userEvent.click(screen.getByRole('button', { name: 'More seats' }))
    expect(seats).toHaveValue(6)
    expect(screen.getByRole('table')).toHaveTextContent('$120.00')
    // The flat price does not move with the team.
    expect(screen.getByRole('table')).toHaveTextContent('$9.00 flat')

    await userEvent.clear(seats)
    await userEvent.type(seats, '80')
    expect(seats).toHaveValue(50)
    expect(screen.getByRole('table')).toHaveTextContent('$1,000.00')
    expect(screen.getByRole('button', { name: 'More seats' })).toBeDisabled()
  })

  // The words the copy must never use: a trial is not a price, and a tier the
  // vendor did not publish is not a fact.
  it('states where the prices come from and invents nothing', () => {
    renderCalculator([seatCRM])
    expect(
      screen.getByText(/Prices read from vendor pages; see each product page/),
    ).toBeInTheDocument()
    expect(screen.queryByText(/free trial/i)).toBeNull()
    expect(screen.queryByText(/tier|plan/i)).toBeNull()
  })
})
