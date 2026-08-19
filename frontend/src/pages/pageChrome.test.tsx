import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { AboutPage } from './AboutPage'
import { AffiliateDisclosurePage } from './AffiliateDisclosurePage'
import { HomePage } from './HomePage'
import { NotFoundPage } from './NotFoundPage'
import { PrivacyPage } from './PrivacyPage'

// Each page renders its own header and footer, so a new one can ship without
// either and nothing complains. Three did: About, Privacy and the affiliate
// disclosure had no navigation at all, and they are exactly the pages a search
// result or an affiliate review drops a stranger onto. So did the 404, which is
// the page most in need of a way onward.
const standalonePages: [string, () => React.JSX.Element][] = [
  ['HomePage', HomePage],
  ['AboutPage', AboutPage],
  ['PrivacyPage', PrivacyPage],
  ['AffiliateDisclosurePage', AffiliateDisclosurePage],
  ['NotFoundPage', NotFoundPage],
]

describe.each(standalonePages)('%s', (_name, Page) => {
  it('offers a way to the rest of the site', () => {
    render(
      <MemoryRouter>
        <Page />
      </MemoryRouter>,
    )
    expect(screen.getByRole('banner')).toBeInTheDocument()
    expect(screen.getByRole('contentinfo')).toBeInTheDocument()
  })
})
