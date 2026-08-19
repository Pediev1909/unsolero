import { useEffect } from 'react'
import { Outlet, ScrollRestoration, useLocation } from 'react-router-dom'

import { trackEvent } from '../features/analytics/tracking'
import { AnalyticsConsentBanner } from '../features/analytics/AnalyticsConsentBanner'
import { analyticsConsentChangedEvent } from '../features/analytics/consent'

export function RouteShell() {
  const { hash, pathname } = useLocation()

  useEffect(() => {
    if (!pathname.startsWith('/admin')) trackEvent('page_view', 'route', {})
  }, [pathname])

  useEffect(() => {
    const trackCurrentPageAfterConsent = () => {
      if (!pathname.startsWith('/admin')) trackEvent('page_view', 'route', {})
    }
    window.addEventListener(
      analyticsConsentChangedEvent,
      trackCurrentPageAfterConsent,
    )
    return () =>
      window.removeEventListener(
        analyticsConsentChangedEvent,
        trackCurrentPageAfterConsent,
      )
  }, [pathname])

  useEffect(() => {
    if (!hash) return

    const animationFrame = window.requestAnimationFrame(() => {
      const target = document.getElementById(hash.slice(1))
      target?.scrollIntoView?.({ block: 'start' })
    })

    return () => window.cancelAnimationFrame(animationFrame)
  }, [hash])

  return (
    <>
      {/* Without this the browser keeps the scroll position across a route
          change, so following a link from halfway down a page lands halfway
          down the next one. ScrollRestoration is used rather than a plain
          scroll to the top because it also restores the previous position on
          back and forward, which a reader comparing products expects. */}
      <ScrollRestoration />
      <Outlet />
      <AnalyticsConsentBanner />
    </>
  )
}
