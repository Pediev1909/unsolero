import { Link, useLocation } from 'react-router-dom'

import { isNavigationItemActive, type NavigationItem } from './navigationItems'

interface NavigationProps {
  items: NavigationItem[]
  includeAccount?: boolean
}

export function Navigation({ items, includeAccount = true }: NavigationProps) {
  const location = useLocation()

  return (
    <nav
      aria-label="Primary navigation"
      className="flex items-center gap-4 lg:gap-6 xl:gap-8"
    >
      {items.map((item) => (
        <Link
          aria-current={
            isNavigationItemActive(location.pathname, location.hash, item.to)
              ? 'page'
              : undefined
          }
          className="nav-link"
          key={item.to}
          to={item.to}
        >
          {item.label}
        </Link>
      ))}
      {includeAccount && (
        <Link
          aria-current={
            location.pathname === '/login' || location.pathname === '/register'
              ? 'page'
              : undefined
          }
          className="nav-link"
          to="/login"
        >
          Sign in
        </Link>
      )}
    </nav>
  )
}
