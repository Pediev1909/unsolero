import { screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { renderWithProviders } from '../test/renderWithProviders'
import { LinksPage } from './LinksPage'

const entries = [
  {
    id: '0d1c9a4e-1111-4a4a-8a8a-000000000001',
    type: 'comparison',
    title: 'ActiveCampaign vs Mailchimp',
    slug: 'activecampaign-vs-mailchimp',
    path: '/compare/activecampaign-vs-mailchimp',
    description: '',
    hero_image: { url: '', alt_text: '', is_primary: true },
    author_name: 'UNSOLERO Editorial',
    published_at: '2026-09-01T09:00:00Z',
    updated_at: '2026-09-01T09:00:00Z',
    covered: [],
  },
  {
    id: '0d1c9a4e-1111-4a4a-8a8a-000000000002',
    type: 'guide',
    title: 'Choosing a CRM for a five-person team',
    slug: 'choosing-a-crm',
    path: '/guides/choosing-a-crm',
    description: '',
    hero_image: { url: '', alt_text: '', is_primary: true },
    author_name: 'UNSOLERO Editorial',
    published_at: '2026-08-30T09:00:00Z',
    updated_at: '2026-08-30T09:00:00Z',
    covered: [],
  },
]

vi.mock('../features/content/queries', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../features/content/queries')>()
  return {
    ...actual,
    useContent: () => ({
      data: entries,
      isPending: false,
      isError: false,
      isSuccess: true,
    }),
  }
})

describe('LinksPage', () => {
  it('lists the pages behind the videos as tagged links, in order', () => {
    renderWithProviders(<LinksPage />, { route: '/links' })

    const nav = screen.getByRole('navigation', {
      name: 'Pages behind our videos',
    })
    const hrefs = Array.from(nav.querySelectorAll('a')).map((anchor) =>
      anchor.getAttribute('href'),
    )
    expect(hrefs).toEqual([
      '/build?utm_source=bio&utm_medium=links',
      '/guides/mailchimp-alternatives?utm_source=bio&utm_medium=links',
      '/offers?utm_source=bio&utm_medium=links',
      '/compare/activecampaign-vs-mailchimp?utm_source=bio&utm_medium=links',
      '/guides/choosing-a-crm?utm_source=bio&utm_medium=links',
      '/articles/how-unsolero-ranks-software?utm_source=bio&utm_medium=links',
      '/affiliate-disclosure?utm_source=bio&utm_medium=links',
    ])
    expect(
      screen.getByRole('link', { name: /build my stack \(free\)/i }),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/every price dated, every ranking commission-proof/i),
    ).toBeInTheDocument()
  })

  it('forwards the platform it was opened from into every link', () => {
    renderWithProviders(<LinksPage />, { route: '/links?utm_source=TikTok' })

    expect(
      screen.getByRole('link', { name: /build my stack \(free\)/i }),
    ).toHaveAttribute('href', '/build?utm_source=tiktok&utm_medium=bio')
    expect(
      screen.getByRole('link', { name: /activecampaign vs mailchimp/i }),
    ).toHaveAttribute(
      'href',
      '/compare/activecampaign-vs-mailchimp?utm_source=tiktok&utm_medium=bio',
    )
  })

  it('drops a forwarded source that is not a plain token', () => {
    renderWithProviders(<LinksPage />, {
      route: '/links?utm_source=%3Cscript%3Ealert(1)%3C%2Fscript%3E',
    })

    expect(
      screen.getByRole('link', { name: /build my stack \(free\)/i }),
    ).toHaveAttribute('href', '/build?utm_source=bio&utm_medium=links')
  })

  it('stays out of the index and keeps the disclosure in reach', () => {
    renderWithProviders(<LinksPage />, { route: '/links' })

    expect(document.title).toBe('UNSOLERO — the pages behind our videos')
    expect(document.querySelector('meta[name="robots"]')).toHaveAttribute(
      'content',
      'noindex, follow',
    )
    expect(screen.getByRole('contentinfo')).toHaveTextContent(
      /commission never changes the ranking/i,
    )
    expect(screen.getByRole('link', { name: /how we earn/i })).toHaveAttribute(
      'href',
      '/affiliate-disclosure?utm_source=bio&utm_medium=links',
    )
  })
})
