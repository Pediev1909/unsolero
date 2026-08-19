import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from './app/App'
import { AppProviders } from './app/providers'
import { claimScrollControl } from './app/scrollControl'
import './styles.css'

// Before React renders, and therefore before the browser has finished
// deciding where this document should sit. Left to itself it carried the
// previous page's offset into a fresh load.
claimScrollControl()

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
