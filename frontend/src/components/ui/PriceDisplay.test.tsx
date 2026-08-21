import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { PriceDisplay } from './PriceDisplay'

describe('PriceDisplay', () => {
  it('formats a normal price as currency', () => {
    render(<PriceDisplay amountMinor={2900} currency="USD" locale="en-US" />)
    expect(screen.getByText('$29.00')).toBeInTheDocument()
  })

  // Six products in the catalog charge nothing per month and take a
  // percentage of each sale instead. Rendering that as "$0.00" made a real
  // fact look like missing data in a comparison column.
  it('names a zero price instead of printing $0.00', () => {
    render(<PriceDisplay amountMinor={0} currency="USD" locale="en-US" />)
    expect(screen.getByText('No monthly fee')).toBeInTheDocument()
    expect(screen.queryByText('$0.00')).not.toBeInTheDocument()
  })

  // The claim has to stay narrow: Stripe has no monthly fee but is not free,
  // so the label must never say "Free".
  it('does not claim a zero-price product is free', () => {
    const { container } = render(
      <PriceDisplay amountMinor={0} currency="USD" locale="en-US" />,
    )
    expect(container.textContent).not.toMatch(/\bFree\b/)
  })
})
