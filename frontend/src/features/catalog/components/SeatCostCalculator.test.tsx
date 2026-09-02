import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import type { ProductSummary } from '../schemas'
import { SeatCostCalculator } from './SeatCostCalculator'

function product(
  name: string,
  amountMinor: number,
  specification: string,
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
    key_specification: { label: 'Billing', value: specification },
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

const seatCRM = product('Seat CRM', 2000, 'Per user per month')
const liteCRM = product('Lite CRM', 1000, 'Per seat per month')
const flatSuite = product('Flat Suite', 900, 'Per month')

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
