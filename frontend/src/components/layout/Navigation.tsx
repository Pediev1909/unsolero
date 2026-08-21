import { Link, useLocation } from 'react-router-dom'

import { BrowseMenu } from './BrowseMenu'
import { NavigationMenu } from './NavigationMenu'
import {
  isBrowseActive,
  isLearnActive,
  isNavigationItemActive,
  learnNavigation,
  type NavigationItem,
} from './navigationItems'

interface NavigationProps {
  items: NavigationItem[]
  includeAccount?: boolean
}

export function Navigation({ items, includeAccount = true }: NavigationProps) {
  const location = useLocation()

  return (
    <nav
      aria-label="Primary navigation"
      className="flex items-center gap-4 lg:gap-6"
    >
      {/* Browse comes first because it is the catalog, and the catalog is what
          most people arrived for. The old bar led with the recommendation
          wizard and never mentioned the categories at all. */}
      <NavigationMenu
        active={isBrowseActive(location.pathname)}
        label="Browse"
        width="wide"
      >
        {(close) => <BrowseMenu onNavigate={close} />}
      </NavigationMenu>

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

      <NavigationMenu active={isLearnActive(location.pathname)} label="Learn">
        {(close) => (
          <ul className="flex flex-col">
            {learnNavigation.map((item) => (
              <li key={item.to}>
                <Link
                  className="block rounded-sm py-2 text-sm font-medium transition-colors hover:text-bronze focus-visible:text-bronze"
                  onClick={close}
                  to={item.to}
                >
                  {item.label}
                </Link>
              </li>
            ))}
          </ul>
        )}
      </NavigationMenu>

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
