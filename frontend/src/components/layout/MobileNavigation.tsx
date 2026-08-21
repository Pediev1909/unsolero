import { ArrowUpRight } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'

import { groupCategories } from '../../features/catalog/categoryGroups'
import { useCategories } from '../../features/catalog/queries'
import { ButtonLink } from '../ui/ButtonLink'
import { Drawer } from '../ui/Drawer'
import { HeaderSearch } from './HeaderSearch'
import {
  accountNavigation,
  browseShortcuts,
  isNavigationItemActive,
  learnNavigation,
  type NavigationItem,
} from './navigationItems'

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
  const categories = useCategories()
  const groups = groupCategories(categories.data ?? [])
  const close = () => onOpenChange(false)

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
        {/* Search first. On a phone the drawer is the whole navigation, and
            someone who already knows the name of the tool they want should
            not have to scroll past fifteen categories to type it. */}
        <div className="pb-5">
          <HeaderSearch onSubmitted={close} variant="block" />
        </div>

        <ul className="divide-y divide-ink/10 border-t border-ink/10">
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
                className="flex items-center justify-between py-4 font-display text-xl font-medium tracking-[-0.03em]"
                onClick={close}
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

        {/* The categories are laid out openly rather than behind a toggle.
            A collapsed section is one more thing to discover, and the whole
            point of this drawer is that nothing in the catalog is hidden. */}
        {groups.map((group) => (
          <section
            className="border-t border-ink/10 pt-5 pb-1"
            key={group.key}
          >
            <p className="eyebrow">{group.label}</p>
            <ul className="mt-2 flex flex-col">
              {group.categories.map((category) => (
                <li key={category.slug}>
                  <Link
                    className="flex items-baseline justify-between gap-3 py-2.5 text-base font-medium"
                    onClick={close}
                    to={`/categories/${category.slug}`}
                  >
                    <span>{category.name}</span>
                    {category.published_products !== undefined && (
                      <span className="shrink-0 text-sm text-ink/45 tabular-nums">
                        {category.published_products}
                      </span>
                    )}
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        ))}

        <section className="border-t border-ink/10 pt-5 pb-1">
          <p className="eyebrow">Everything</p>
          <ul className="mt-2 flex flex-col">
            {[...browseShortcuts, ...learnNavigation, ...accountNavigation].map(
              (item) => (
                <li key={item.to}>
                  <Link
                    className="block py-2.5 text-base font-medium"
                    onClick={close}
                    to={item.to}
                  >
                    {item.label}
                  </Link>
                </li>
              ),
            )}
          </ul>
        </section>
      </nav>
    </Drawer>
  )
}
