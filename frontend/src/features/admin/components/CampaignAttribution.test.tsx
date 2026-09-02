import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { CampaignAttribution } from './CampaignAttribution'

const campaign = '2026-09-crm-shootout'

const report = {
  campaigns: [
    {
      campaign,
      traffic_source: 'youtube',
      traffic_medium: 'shorts',
      sessions: 1234,
      page_views: 3100,
      affiliate_clicks: 37,
    },
    // A visitor who declined analytics still produces a countable merchant
    // click, so a row with clicks and no sessions is a real outcome.
    {
      campaign,
      traffic_source: 'tiktok',
      traffic_medium: null,
      sessions: 0,
      page_views: 0,
      affiliate_clicks: 2,
    },
  ],
  landing_pages: [{ campaign, page_path: '/tools/crm', sessions: 900 }],
  sources_by_medium: [],
}

describe('CampaignAttribution', () => {
  it('renders each campaign with its source, medium, and counts', () => {
    render(<CampaignAttribution data={report} />)
    const table = within(screen.getByRole('article', { name: 'Campaigns' }))
    const rows = table.getAllByRole('row')
    expect(rows).toHaveLength(3)
    expect(
      within(rows[1]!)
        .getAllByRole('cell')
        .map((c) => c.textContent),
    ).toEqual([
      campaign,
      'youtube',
      'shorts',
      (1234).toLocaleString(),
      (3100).toLocaleString(),
      '37',
    ])
  })

  it('shows a dash for a missing medium instead of an empty cell', () => {
    render(<CampaignAttribution data={report} />)
    const table = within(screen.getByRole('article', { name: 'Campaigns' }))
    const tiktok = table.getByText('tiktok').closest('tr')
    expect(tiktok).not.toBeNull()
    expect(within(tiktok!).getByText('—')).toBeInTheDocument()
    expect(within(tiktok!).getByText('2')).toBeInTheDocument()
  })

  it('lists landing pages per campaign', () => {
    render(<CampaignAttribution data={report} />)
    const table = within(
      screen.getByRole('article', { name: 'Landing pages by campaign' }),
    )
    expect(table.getByText('/tools/crm')).toBeInTheDocument()
    expect(table.getByText((900).toLocaleString())).toBeInTheDocument()
  })

  it('reports no data for an empty section rather than an empty table', () => {
    render(<CampaignAttribution data={report} />)
    const section = within(
      screen.getByRole('article', { name: 'Sources by medium' }),
    )
    expect(section.getByText('No data')).toBeInTheDocument()
    expect(section.queryByRole('table')).toBeNull()
  })

  it('keeps wide tables scrollable inside their own container', () => {
    render(<CampaignAttribution data={report} />)
    for (const table of screen.getAllByRole('table')) {
      expect(table.parentElement?.className).toContain('overflow-x-auto')
    }
  })
})
