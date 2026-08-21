export interface NavigationItem {
  label: string
  to: string
}

/**
 * The primary navigation.
 *
 * Two items are gone from the old list. "How it works" pointed at /#method, a
 * hash anchor that scrolled the homepage instead of navigating — a nav item
 * that behaves differently from every other nav item is a small trap, and it
 * could not be linked to or found in search. It is now a page.
 *
 * "Wishlist" moved out of the top level because it is only meaningful once you
 * have saved something, and it was taking a slot from the catalog itself.
 */
export const primaryNavigation: NavigationItem[] = [
  { label: 'Build my setup', to: '/build' },
  { label: 'Compare', to: '/compare' },
  { label: 'How it works', to: '/how-it-works' },
]

/**
 * The Browse menu's own links, which sit under the category columns.
 * These are the ways into the catalog that are not a single category.
 */
export const browseShortcuts: NavigationItem[] = [
  { label: 'All software', to: '/products' },
  { label: 'All vendors', to: '/brands' },
  { label: 'Every category', to: '/categories' },
]

/**
 * Editorial. Small enough to be a plain list rather than a second mega menu.
 */
export const learnNavigation: NavigationItem[] = [
  { label: 'Guides', to: '/guides' },
  { label: 'Articles', to: '/articles' },
  { label: 'About UNSOLERO', to: '/about' },
]

/**
 * The account-shaped links that used to sit in the primary bar.
 */
export const accountNavigation: NavigationItem[] = [
  { label: 'Wishlist', to: '/wishlist' },
  { label: 'Saved setups', to: '/setups' },
]

export function isNavigationItemActive(
  pathname: string,
  hash: string,
  destination: string,
) {
  if (destination.startsWith('/#')) {
    return pathname === '/' && hash === destination.slice(1)
  }
  if (destination === '/products') {
    return (
      pathname === '/products' ||
      pathname.startsWith('/products/') ||
      pathname.startsWith('/categories/') ||
      pathname.startsWith('/brands/')
    )
  }
  return pathname === destination || pathname.startsWith(`${destination}/`)
}

/**
 * True when the Browse menu should read as the current section. Browse has no
 * page of its own, so it borrows the state of everything it contains.
 */
export function isBrowseActive(pathname: string) {
  return (
    pathname === '/products' ||
    pathname.startsWith('/products/') ||
    pathname === '/categories' ||
    pathname.startsWith('/categories/') ||
    pathname === '/brands' ||
    pathname.startsWith('/brands/')
  )
}

export function isLearnActive(pathname: string) {
  return learnNavigation.some((item) =>
    isNavigationItemActive(pathname, '', item.to),
  )
}
