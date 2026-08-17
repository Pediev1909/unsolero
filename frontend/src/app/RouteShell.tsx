import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'

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
      <Outlet />
      <AnalyticsConsentBanner />
    </>
  )
}
