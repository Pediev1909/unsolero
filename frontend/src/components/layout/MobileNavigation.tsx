import { ArrowUpRight } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'

import { ButtonLink } from '../ui/ButtonLink'
import { Drawer } from '../ui/Drawer'
import { isNavigationItemActive, type NavigationItem } from './navigationItems'

interface MobileNavigationProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  items: NavigationItem[]
}

export function MobileNavigation({
  open,
  onOpenChange,
  items,
}: MobileNavigationProps) {
  const location = useLocation()

  return (
    <Drawer
      description="Software decisions grounded in your goals, budget, and current stack."
      onOpenChange={onOpenChange}
      open={open}
      title="Navigate"
      footer={
        <ButtonLink fullWidth to="/login">
          Sign in
          <ArrowUpRight aria-hidden="true" size={16} />
        </ButtonLink>
      }
    >
      <nav aria-label="Mobile navigation">
        <ul className="divide-y divide-ink/10">
          {items.map((item) => (
            <li key={item.to}>
              <Link
                aria-current={
                  isNavigationItemActive(
                    location.pathname,
                    location.hash,
                    item.to,
                  )
                    ? 'page'
                    : undefined
                }
                className="flex items-center justify-between py-5 font-display text-2xl font-medium tracking-[-0.04em]"
                onClick={() => onOpenChange(false)}
                to={item.to}
              >
                {item.label}
                <ArrowUpRight
                  aria-hidden="true"
                  className="text-bronze"
                  size={18}
                />
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </Drawer>
  )
}
