import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from './app/App'
import { AppProviders } from './app/providers'
import { claimScrollControl } from './app/scrollControl'
import { captureLandingAttribution } from './features/analytics/tracking'
import './styles.css'

// Before React renders, and therefore before the browser has finished
// deciding where this document should sit. Left to itself it carried the
// previous page's offset into a fresh load.
claimScrollControl()

// Where the visit came from, read from the landing URL before React renders
// and before the first client-side navigation can discard it. See the
// function's comment for why this does not wait for analytics consent.
captureLandingAttribution()

const root = document.getElementById('root')

if (!root) {
  throw new Error('Root element not found')
}

createRoot(root).render(
  <StrictMode>
    <AppProviders>
      <App />
    </AppProviders>
  </StrictMode>,
)
