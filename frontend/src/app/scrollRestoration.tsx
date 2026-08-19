import { useEffect, useRef } from 'react'
import { NavigationType, useLocation, useNavigationType } from 'react-router-dom'

import { claimScrollControl } from './scrollControl'

// React Router's own ScrollRestoration restores a position the moment the
// route renders. This catalog fetches through React Query inside its
// components, so at that moment the listing is still a loading grid and the
// document is a fraction of its final height: asking for 2500px on a document
// 900px tall lands at 860 and stays there once the products arrive. Returning
// from a product to the listing dropped a reader most of the way back up.
//
// The browser's own guess is worse. With scrollRestoration left on 'auto', a
// full page load that follows a scrolled page opened the new page most of the
// way down it — a link from halfway down the catalog landed in the middle of
// the product page.
//
// So: the browser is told to stop guessing, a forward navigation always starts
// at the top, and a back or forward navigation keeps reaching for the stored
// position while the document is still growing.
const STORAGE_KEY = 'unsolero:scroll-positions'
// Long enough for a catalog page to fetch its products and lay them out on a
// slow connection. The retry stops early the moment it succeeds, and abandons
// immediately if the reader takes over.
const RETRY_WINDOW_MS = 3000


function readPositions(): Record<string, number> {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    return raw ? (JSON.parse(raw) as Record<string, number>) : {}
  } catch {
    // A private-mode session storage that throws is not a reason to break
    // navigation; the reader simply gets the top of the page.
    return {}
  }
}

function writePosition(key: string, offset: number) {
  try {
    const positions = readPositions()
    positions[key] = offset
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(positions))
  } catch {
    /* see readPositions */
  }
}

export function ScrollRestoration() {
  const location = useLocation()
  const navigationType = useNavigationType()
  const currentKey = useRef(location.key)
  // React Router names the first entry of every document "default", so the
  // key a fresh load asks about is the same one the previous document saved
  // under. Restoring on that first render made a followed link reopen at the
  // offset of the page it was followed from — the very thing this component
  // exists to prevent. A document that has just loaded has no position of its
  // own to return to.
  const hasNavigated = useRef(false)

  // Record where the reader was before the route changes, and again if they
  // close or hide the tab from a scrolled position.
  useEffect(() => {
    const key = location.key
    currentKey.current = key
    const save = () => writePosition(key, window.scrollY)
    window.addEventListener('pagehide', save)
    return () => {
      save()
      window.removeEventListener('pagehide', save)
    }
  }, [location.key])

  useEffect(() => {
    claimScrollControl()

    const isFirstRender = !hasNavigated.current
    hasNavigated.current = true

    // A link followed, or an address typed: this page has not been seen at
    // this position before, so it starts where a page starts.
    if (navigationType !== NavigationType.Pop || isFirstRender) {
      if (!location.hash) window.scrollTo(0, 0)
      return
    }

    const target = readPositions()[location.key]
    if (target === undefined) {
      window.scrollTo(0, 0)
      return
    }

    // The document grows as the route's queries resolve. Keep asking until it
    // is tall enough to honour the request, or until the window closes and the
    // reader is left wherever the content actually reaches.
    let frame = 0
    let abandoned = false
    const deadline = performance.now() + RETRY_WINDOW_MS
    // Someone who starts reading before the content settles has decided where
    // they want to be, and must not be dragged away from it.
    const abandon = () => {
      abandoned = true
    }
    const events = ['wheel', 'touchstart', 'keydown'] as const
    for (const event of events) {
      window.addEventListener(event, abandon, { passive: true })
    }

    const attempt = () => {
      if (abandoned) return
      window.scrollTo(0, target)
      const reached = Math.abs(window.scrollY - target) < 2
      if (!reached && performance.now() < deadline) {
        frame = window.requestAnimationFrame(attempt)
      }
    }
    attempt()

    return () => {
      window.cancelAnimationFrame(frame)
      for (const event of events) window.removeEventListener(event, abandon)
    }
  }, [location.key, location.hash, navigationType])

  return null
}
