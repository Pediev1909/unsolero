import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { PriceRecord } from './PriceRecord'
import type { PriceRecordEntry } from '../schemas'

const audited: PriceRecordEntry = {
  observed_at: '2026-09-02T09:28:00Z',
  price_minor: 3900,
  currency: 'USD',
  billing: {
    period: 'monthly',
    unit: 'flat',
    unit_note: null,
    annual_price_minor: 2900,
  },
  is_current: true,
}

const earlier: PriceRecordEntry = {
  observed_at: '2026-08-21T11:05:00Z',
  price_minor: 2900,
  currency: 'USD',
  billing: null,
  note: 'Price read from the vendor pricing page on 2026-08-21.',
  is_current: false,
}

describe('PriceRecord', () => {
  it('lists both figures with the change between them and the date each was read', () => {
    render(<PriceRecord record={[audited, earlier]} />)

    expect(
      screen.getByRole('heading', { name: 'Price record' }),
    ).toBeInTheDocument()
    // A header row and one row per figure.
    const [, newest, oldest] = screen.getAllByRole('row')
    expect(screen.getAllByRole('row')).toHaveLength(3)
    if (!newest || !oldest) throw new Error('rows not rendered')

    const current = within(newest)
    expect(current.getByText('Sep 2, 2026')).toBeInTheDocument()
    expect(current.getByText('$39.00')).toBeInTheDocument()
    expect(current.getByText('Flat rate, monthly billing')).toBeInTheDocument()
    expect(current.getByText('Current')).toBeInTheDocument()
    expect(current.getByText('Up by')).toBeInTheDocument()
    expect(current.getByText('$10.00')).toBeInTheDocument()

    const previous = within(oldest)
    expect(previous.getByText('Aug 21, 2026')).toBeInTheDocument()
    expect(previous.getByText('$29.00')).toBeInTheDocument()
    expect(
      previous.getByText(
        'Price read from the vendor pricing page on 2026-08-21.',
      ),
    ).toBeInTheDocument()
    // Nothing older to compare the oldest figure against, and no basis was
    // recorded on that revision.
    expect(previous.getByText('No earlier price recorded')).toBeInTheDocument()

    expect(
      screen.getByText(
        "Each figure was read from the vendor's own page on the date shown.",
      ),
    ).toBeInTheDocument()
  })

  it('draws nothing for a price that has never moved, and nothing for no record', () => {
    const { container: single } = render(<PriceRecord record={[audited]} />)
    expect(single).toBeEmptyDOMElement()

    const { container: none } = render(<PriceRecord record={undefined} />)
    expect(none).toBeEmptyDOMElement()

    const { container: empty } = render(<PriceRecord record={[]} />)
    expect(empty).toBeEmptyDOMElement()
  })
})
