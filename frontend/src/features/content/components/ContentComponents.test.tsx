import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import type { ContentSummary } from '../schemas'
import { ContentBody } from './ContentBody'
import { ContentCard } from './ContentCard'

const summary: ContentSummary = {
  id: '7da06367-69db-4fac-b1c4-e878b64a6a60',
  type: 'buying_guide',
  title: 'A structured buying guide for compact equipment',
  slug: 'structured-buying-guide',
  path: '/guides/structured-buying-guide',
  description: 'A concise description grounded in structured product facts.',
  hero_image: {
    url: '/images/saas-agency-stack.svg',
    alt_text: 'Diagram of a connected software stack',
    is_primary: false,
  },
  author_name: 'UNSOLERO Editorial',
  published_at: '2026-08-01T09:00:00Z',
  updated_at: '2026-08-01T09:00:00Z',
  covered: [],
}

describe('editorial content presentation', () => {
  it('links editorial cards to their canonical application route', () => {
    render(
      <MemoryRouter>
        <ContentCard entry={summary} />
      </MemoryRouter>,
    )

    expect(
      screen.getByRole('link', { name: `Read ${summary.title}` }),
    ).toHaveAttribute('href', summary.path)
    expect(screen.getByText('Buying guide')).toBeInTheDocument()
  })

  it('renders validated blocks as React content rather than arbitrary HTML', () => {
    const { container } = render(
      <ContentBody
        blocks={[
          { type: 'heading', heading: 'Measure the room' },
          { type: 'paragraph', text: '<script>unsafe()</script>' },
          { type: 'unordered_list', items: ['Width', 'Height'] },
        ]}
      />,
    )

    expect(
      screen.getByRole('heading', { name: 'Measure the room' }),
    ).toHaveAttribute('id', 'measure-the-room')
    expect(screen.getByText('<script>unsafe()</script>')).toBeInTheDocument()
    expect(container.querySelector('script')).not.toBeInTheDocument()
  })
})

describe('editorial card nameplate', () => {
  // Thirteen comparisons shared one illustration, so the card has to take its
  // identity from the products the piece covers rather than from a picture.
  it('names the card after the products the piece compares', () => {
    render(
      <MemoryRouter>
        <ContentCard
          entry={{
            ...summary,
            covered: [
              { name: 'Fathom', price_minor: 1500, currency: 'USD' },
              { name: 'Umami', price_minor: 2000, currency: 'USD' },
            ],
          }}
        />
      </MemoryRouter>,
    )
    expect(screen.getByText(/Fathom/)).toBeInTheDocument()
    expect(screen.getByText(/Umami/)).toBeInTheDocument()
    expect(screen.getByText('$15–$20')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('calls a free tool free rather than printing a zero', () => {
    render(
      <MemoryRouter>
        <ContentCard
          entry={{
            ...summary,
            covered: [
              { name: 'Wave', price_minor: 0, currency: 'USD' },
              { name: 'Zoho Books', price_minor: 2000, currency: 'USD' },
            ],
          }}
        />
      </MemoryRouter>,
    )
    expect(screen.getByText('Free–$20')).toBeInTheDocument()
  })

  // A piece with no product set keeps its illustration; the alternative is a
  // card with an empty band where a picture used to be.
  it('keeps the illustration when the piece covers nothing', () => {
    render(
      <MemoryRouter>
        <ContentCard entry={summary} />
      </MemoryRouter>,
    )
    expect(screen.getByRole('img')).toHaveAttribute(
      'src',
      '/images/saas-agency-stack.svg',
    )
  })
})
