import { render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { ContentBody } from './ContentBody'
import type { Offer, ProductSummary } from '../../catalog/schemas'
import type { ContentBlock } from '../schemas'

// The offer block loads the product's live offer through the same query the
// product page uses. These tests are about what the block draws in each state
// of that query, not about how the offer is fetched, so the hook is replaced
// and each test sets the state it needs.
const offersState = vi.hoisted(() => ({
  current: { isPending: false, isError: false, data: [] as unknown[] },
}))

vi.mock('../../catalog/queries', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../catalog/queries')>()
  return { ...actual, useOffers: () => offersState.current }
})

const cta: ContentBlock = {
  type: 'cta',
  heading: 'If automation is why you are leaving',
  text: 'Their own comparison is the honest place to start.',
  label: 'See ActiveCampaign against Mailchimp',
  promotion: 'activecampaign-mailchimp-switch',
}

const zohoCampaigns: ProductSummary = {
  id: 'p1',
  name: 'Zoho Campaigns Standard',
  slug: 'zoho-campaigns-standard',
  brand: { name: 'Zoho', slug: 'zoho' },
  category: { name: 'Email marketing', slug: 'email-marketing' },
  price: { amount_minor: 525, currency: 'USD' },
  primary_image: null,
  key_specification: { label: 'Billing', value: 'Per month' },
  suitability: [],
  scores: {
    quality: 0,
    value: 0,
    durability: 0,
    beginner: 0,
    advanced: 0,
    apartment: 0,
    noise: 0,
    portability: 0,
  },
  is_demo: false,
}

const liveOffer: Offer = {
  id: 'o1',
  merchant: { name: 'Zoho', slug: 'zoho', country_code: 'IN', trust_score: 80 },
  price: { amount_minor: 525, currency: 'USD' },
  shipping_minor: 0,
  landed_price_minor: 525,
  availability: 'in_stock',
  condition: 'new',
  last_checked_at: '2026-08-26T10:48:13Z',
  observed_at: null,
  expires_at: null,
  freshness_status: 'fresh',
  purchase_path: '/api/affiliate/click/7e312958-2f21-42a8-ba82-fd240b858062',
  disclosure_label: 'Affiliate link',
}

const offer: ContentBlock = {
  type: 'offer',
  heading: 'Where to get it',
  text: 'Cheapest of the five, and already connected if you run Zoho.',
  product: 'zoho-campaigns-standard',
}

function renderWithRouter(blocks: ContentBlock[]) {
  return render(
    <MemoryRouter>
      <ContentBody blocks={blocks} products={[zohoCampaigns]} />
    </MemoryRouter>,
  )
}

describe('ContentBody CTA block', () => {
  it('routes through the tracked promotion path, never a raw vendor URL', () => {
    render(<ContentBody blocks={[cta]} />)
    const link = screen.getByRole('link', {
      name: /See ActiveCampaign against Mailchimp/,
    })
    // The block names a promotion slug and the href is built here. If a raw
    // destination could ever reach this attribute, an editor would be able to
    // publish an untracked or a stranger's affiliate link.
    expect(link.getAttribute('href')).toContain(
      '/api/affiliate/promotion/activecampaign-mailchimp-switch',
    )
    expect(link.getAttribute('href')).toContain('source=promotion')
    expect(link.getAttribute('href')).not.toContain('try.activecampaign.com')
  })

  it('marks the link sponsored and says so in words', () => {
    render(<ContentBody blocks={[cta]} />)
    const link = screen.getByRole('link', {
      name: /See ActiveCampaign against Mailchimp/,
    })
    expect(link).toHaveAttribute('rel', 'nofollow noopener sponsored')
    // A reader mid-article has not seen a disclosure anywhere else on screen,
    // so this sentence is the only one they get.
    expect(screen.getByText(/Affiliate link/i)).toBeInTheDocument()
    expect(screen.getByText(/pays us if you subscribe/i)).toBeInTheDocument()
  })

  it('draws nothing when the block is missing its promotion or label', () => {
    const { container: noPromotion } = render(
      <ContentBody blocks={[{ ...cta, promotion: undefined }]} />,
    )
    expect(noPromotion.querySelector('a')).toBeNull()

    const { container: noLabel } = render(
      <ContentBody blocks={[{ ...cta, label: undefined }]} />,
    )
    expect(noLabel.querySelector('a')).toBeNull()
  })

  it('leaves the other block types alone', () => {
    render(
      <ContentBody
        blocks={[
          { type: 'paragraph', text: 'Almost nobody leaves for the features.' },
          { type: 'callout', heading: 'The question', text: 'Do you pay?' },
        ]}
      />,
    )
    expect(
      screen.getByText('Almost nobody leaves for the features.'),
    ).toBeInTheDocument()
    expect(screen.getByText('The question')).toBeInTheDocument()
    expect(screen.queryByRole('link')).toBeNull()
  })
})

describe('ContentBody pros and cons block', () => {
  it('renders both columns, each item told apart for a screen reader', () => {
    render(
      <ContentBody
        blocks={[
          {
            type: 'pros_cons',
            heading: 'Brevo, weighed',
            pros: ['Priced on sends, not contacts', 'Free tier'],
            cons: ['Thinner automation than ActiveCampaign'],
          },
        ]}
      />,
    )
    expect(
      screen.getByRole('heading', { level: 3, name: 'Brevo, weighed' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Pros')).toBeInTheDocument()
    expect(screen.getByText('Cons')).toBeInTheDocument()

    const lists = screen.getAllByRole('list')
    expect(lists).toHaveLength(2)
    expect(
      within(lists[0] as HTMLElement).getAllByRole('listitem'),
    ).toHaveLength(2)
    expect(
      within(lists[1] as HTMLElement).getAllByRole('listitem'),
    ).toHaveLength(1)
    // The icons are decorative, so the column each item belongs to has to be
    // said in words somewhere a screen reader will read them.
    expect(screen.getAllByText('Pro:')).toHaveLength(2)
    expect(screen.getAllByText('Con:')).toHaveLength(1)
    expect(
      screen.getByText('Thinner automation than ActiveCampaign'),
    ).toBeInTheDocument()
  })
})

describe('ContentBody FAQ block', () => {
  const faq: ContentBlock = {
    type: 'faq',
    heading: 'Questions people ask',
    questions: [
      {
        question: 'Which Mailchimp alternative is the cheapest?',
        answer: 'Zoho Campaigns Standard, at 5.25 USD per month.',
      },
      {
        question: 'When is it not worth switching?',
        answer: 'If your list is under a few thousand and you send often.',
      },
    ],
  }

  it('renders every question as a disclosure with its answer in the document', () => {
    const { container } = render(<ContentBody blocks={[faq]} />)
    const details = container.querySelectorAll('details')
    expect(details).toHaveLength(2)
    expect(
      screen.getByText('Which Mailchimp alternative is the cheapest?'),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Zoho Campaigns Standard, at 5.25 USD per month.'),
    ).toBeInTheDocument()
    // The first is open so the section does not read as a list of things the
    // page refuses to say; the rest wait to be asked.
    expect(details[0]).toHaveAttribute('open')
    expect(details[1]).not.toHaveAttribute('open')
  })

  it('gives its heading the id the table of contents links to', () => {
    render(<ContentBody blocks={[faq]} />)
    expect(
      screen.getByRole('heading', { level: 2, name: 'Questions people ask' }),
    ).toHaveAttribute('id', 'questions-people-ask')
  })

  it('draws nothing for a block with no questions', () => {
    const { container } = render(
      <ContentBody blocks={[{ type: 'faq', heading: 'Empty' }]} />,
    )
    expect(container.querySelector('details')).toBeNull()
    expect(screen.queryByText('Empty')).toBeNull()
  })
})

describe('ContentBody offer block', () => {
  it('says so plainly when the product has no live offer, with no button and no price', () => {
    offersState.current = { isPending: false, isError: false, data: [] }
    renderWithRouter([offer])

    expect(screen.getByText('Where to get it')).toBeInTheDocument()
    expect(
      screen.getByText(/Cheapest of the five, and already connected/),
    ).toBeInTheDocument()
    const link = screen.getByRole('link', {
      name: 'See Zoho Campaigns Standard in the catalog',
    })
    expect(link).toHaveAttribute('href', '/products/zoho-campaigns-standard')
    expect(screen.queryByRole('link', { name: /View at/ })).toBeNull()
    expect(screen.queryByText(/\$/)).toBeNull()
    expect(screen.queryByText(/pays us/)).toBeNull()
  })

  it('treats a failed lookup the same as no offer', () => {
    offersState.current = { isPending: false, isError: true, data: [] }
    renderWithRouter([offer])
    expect(
      screen.getByRole('link', {
        name: 'See Zoho Campaigns Standard in the catalog',
      }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /View at/ })).toBeNull()
  })

  it('shows a compact loading state while the offer is fetched', () => {
    offersState.current = { isPending: true, isError: false, data: [] }
    renderWithRouter([offer])
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.queryByRole('link')).toBeNull()
  })

  it('renders the live offer: price, the date it was read, and a tracked vendor button', () => {
    offersState.current = {
      isPending: false,
      isError: false,
      data: [liveOffer],
    }
    renderWithRouter([offer])

    expect(
      screen.getByRole('link', { name: 'Zoho Campaigns Standard' }),
    ).toHaveAttribute('href', '/products/zoho-campaigns-standard')
    expect(screen.getByText('$5.25')).toBeInTheDocument()
    expect(screen.getByText(/Checked Aug 26, 2026/)).toBeInTheDocument()

    // The label defaults to the merchant, and the href is the offer's tracked
    // path with this surface as the source — never a vendor URL from the block.
    const button = screen.getByRole('link', { name: /View at Zoho/ })
    expect(button.getAttribute('href')).toContain(
      '/api/affiliate/click/7e312958-2f21-42a8-ba82-fd240b858062',
    )
    expect(button.getAttribute('href')).toContain('source=promotion')
    expect(button).toHaveAttribute('rel', 'nofollow noopener sponsored')
    expect(button).toHaveAttribute('target', '_blank')
  })

  it('uses the editor label when the block carries one', () => {
    offersState.current = {
      isPending: false,
      isError: false,
      data: [liveOffer],
    }
    renderWithRouter([{ ...offer, label: 'Start with Zoho Campaigns' }])
    expect(
      screen.getByRole('link', { name: /Start with Zoho Campaigns/ }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /View at Zoho/ })).toBeNull()
  })

  it('carries the same disclosure sentence as the CTA block', () => {
    offersState.current = {
      isPending: false,
      isError: false,
      data: [liveOffer],
    }
    renderWithRouter([cta, offer])
    // Two paid controls, one sentence, said twice. If the offer block ever
    // rewords it, the two would start telling readers different things about
    // the same arrangement.
    expect(
      screen.getAllByText(
        /Affiliate link\. It pays us if you subscribe, and it changed nothing about where this tool sits on this page\./,
      ),
    ).toHaveLength(2)
  })

  it('draws nothing for an offer block with no product', () => {
    offersState.current = {
      isPending: false,
      isError: false,
      data: [liveOffer],
    }
    const { container } = renderWithRouter([{ type: 'offer', text: 'Orphan' }])
    expect(container.querySelector('aside')).toBeNull()
  })
})
