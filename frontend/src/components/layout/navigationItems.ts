export interface NavigationItem {
  label: string
  to: string
}

export const primaryNavigation: NavigationItem[] = [
  { label: 'Build my setup', to: '/build' },
  { label: 'How it works', to: '/#method' },
  { label: 'Equipment', to: '/products' },
  { label: 'Compare', to: '/compare' },
  { label: 'Wishlist', to: '/wishlist' },
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
