import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { AboutPage } from './AboutPage'
import { AffiliateDisclosurePage } from './AffiliateDisclosurePage'
import { HomePage } from './HomePage'
import { LoginPage } from './LoginPage'
import { NotFoundPage } from './NotFoundPage'
import { PrivacyPage } from './PrivacyPage'
import { RegisterPage } from './RegisterPage'

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

// React Router swaps the view without touching the document, so a page that
// never sets a title inherits whichever one the previous page left behind. In
// production that meant the home page introduced itself as "About UNSOLERO"
// to anyone who reached it from the footer, and the 404 claimed to be the home
// page. Rendering these in sequence is what catches it: each assertion only
// means something because the page before it set a different title.
const titledPages: [string, () => React.JSX.Element, RegExp][] = [
  ['AboutPage', AboutPage, /^About UNSOLERO/],
  ['HomePage', HomePage, /^UNSOLERO —/],
  ['LoginPage', LoginPage, /^Sign in \| UNSOLERO$/],
  ['RegisterPage', RegisterPage, /^Create an account \| UNSOLERO$/],
  ['NotFoundPage', NotFoundPage, /^Page not found \| UNSOLERO$/],
  ['PrivacyPage', PrivacyPage, /^Privacy/],
]

describe('document title', () => {
  it.each(titledPages)('%s names itself', (_name, Page, expected) => {
    // The sign-in and registration forms reach for the query client, so these
    // pages cannot be rendered on the router alone.
    render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <Page />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(document.title).toMatch(expected)
  })
})
