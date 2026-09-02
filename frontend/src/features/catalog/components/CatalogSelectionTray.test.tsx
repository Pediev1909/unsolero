import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { CatalogSelectionTray } from './CatalogSelectionTray'

function renderTray(props: { comparedCount: number; savedCount: number }) {
  const onOpenComparison = vi.fn()
  render(
    <MemoryRouter>
      <CatalogSelectionTray {...props} onOpenComparison={onOpenComparison} />
    </MemoryRouter>,
  )
  return { onOpenComparison }
}

describe('CatalogSelectionTray', () => {
  it('draws nothing while nothing is selected', () => {
    renderTray({ comparedCount: 0, savedCount: 0 })
    expect(screen.queryByRole('region')).toBeNull()
  })

  it('shows the comparison count and opens the comparison', async () => {
    const { onOpenComparison } = renderTray({
      comparedCount: 2,
      savedCount: 0,
    })
    expect(screen.getByText(/of 4 selected/)).toHaveTextContent(
      '2 of 4 selected',
    )
    await userEvent.click(screen.getByRole('button', { name: /Compare/ }))
    expect(onOpenComparison).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('link', { name: /Saved/ })).toBeNull()
  })

  it('shows the saved count with the way to the saved list', () => {
    renderTray({ comparedCount: 0, savedCount: 3 })
    const link = screen.getByRole('link', { name: /Saved 3/ })
    expect(link).toHaveAttribute('href', '/wishlist')
    expect(link).toHaveTextContent(/View/)
    expect(screen.queryByRole('button', { name: /Compare/ })).toBeNull()
  })

  it('carries both counts at once', () => {
    renderTray({ comparedCount: 1, savedCount: 5 })
    expect(screen.getByText(/of 4 selected/)).toHaveTextContent(
      '1 of 4 selected',
    )
    expect(screen.getByRole('link', { name: /Saved 5/ })).toBeInTheDocument()
  })
})
