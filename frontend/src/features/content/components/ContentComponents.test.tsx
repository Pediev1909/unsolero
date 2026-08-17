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
    url: '/images/demo-adjustable-dumbbells.webp',
    alt_text: 'Fictional adjustable dumbbells',
    is_primary: false,
  },
  author_name: 'UNSOLERO Editorial',
  published_at: '2026-08-01T09:00:00Z',
  updated_at: '2026-08-01T09:00:00Z',
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
