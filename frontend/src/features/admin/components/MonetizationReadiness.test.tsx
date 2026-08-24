import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { MonetizationReadiness } from './MonetizationReadiness'
import type { DashboardData } from '../schemas'

type Readiness = DashboardData['readiness']

function readiness(overrides: Partial<Readiness> = {}): Readiness {
  return {
    published_products: 0,
    without_active_offer: 0,
    without_affiliate_link: 0,
    earning_ready: 0,
    commerce_providers: 0,
    published_content: 0,
    blocked: [],
    ...overrides,
  }
}

function renderPanel(value: Readiness) {
  render(
    <MemoryRouter>
      <MonetizationReadiness readiness={value} />
    </MemoryRouter>,
  )
}

describe('MonetizationReadiness', () => {
  it('reports the share of the catalog that can take money', () => {
    renderPanel(
      readiness({
        published_products: 53,
        earning_ready: 4,
        without_active_offer: 49,
      }),
    )
    expect(screen.getByText(/8% of the published catalog/)).toBeInTheDocument()
    expect(screen.getByText('/ 53')).toBeInTheDocument()
  })

  // An empty catalog must not divide by zero and must not claim 0% as if the
  // catalog were failing; it has simply not started.
  it('states that nothing is published rather than reporting zero percent', () => {
    renderPanel(readiness())
    expect(screen.getByText('Nothing is published yet.')).toBeInTheDocument()
  })

  it('links each blocked product to its editor and names the reason', () => {
    renderPanel(
      readiness({
        published_products: 2,
        earning_ready: 1,
        without_affiliate_link: 1,
        blocked: [
          {
            id: '2f1a4c9e-6b3d-4f2a-9c8e-1d5b7a3f6c20',
            name: 'Pipedrive Lite',
            slug: 'pipedrive-lite',
            reason: 'no_affiliate_link',
          },
        ],
      }),
    )
    const link = screen.getByRole('link', { name: 'Pipedrive Lite' })
    expect(link).toHaveAttribute(
      'href',
      '/admin/products/2f1a4c9e-6b3d-4f2a-9c8e-1d5b7a3f6c20',
    )
    expect(screen.getByText('No affiliate link')).toBeInTheDocument()
  })

  // The call to action is the point of the tile; a zero gap has nothing to fix.
  it('offers a fix only where there is a gap', () => {
    renderPanel(
      readiness({
        published_products: 10,
        earning_ready: 10,
      }),
    )
    expect(screen.queryByRole('link', { name: 'Add an offer' })).toBeNull()
    expect(screen.queryByRole('link', { name: 'Add a link' })).toBeNull()
  })
})
