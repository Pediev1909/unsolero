import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ContentBody } from './ContentBody'
import type { ContentBlock } from '../schemas'

const cta: ContentBlock = {
  type: 'cta',
  heading: 'If automation is why you are leaving',
  text: 'Their own comparison is the honest place to start.',
  label: 'See ActiveCampaign against Mailchimp',
  promotion: 'activecampaign-mailchimp-switch',
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
