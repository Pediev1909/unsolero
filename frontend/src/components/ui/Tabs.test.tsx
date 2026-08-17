import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Tabs } from './Tabs'

describe('Tabs', () => {
  it('moves selection with arrow keys', async () => {
    const user = userEvent.setup()
    render(
      <Tabs
        ariaLabel="Product details"
        items={[
          { id: 'fit', label: 'Fit', content: <p>Fit details</p> },
          {
            id: 'tradeoffs',
            label: 'Trade-offs',
            content: <p>Trade-off details</p>,
          },
        ]}
      />,
    )

    const fit = screen.getByRole('tab', { name: 'Fit' })
    fit.focus()
    await user.keyboard('{ArrowRight}')

    expect(screen.getByRole('tab', { name: 'Trade-offs' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
    expect(screen.getByText('Trade-off details')).toBeVisible()
  })
})
