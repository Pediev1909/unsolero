import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { groupCategories } from '../../features/catalog/categoryGroups'
import { useCategories } from '../../features/catalog/queries'
import { browseShortcuts } from './navigationItems'

interface BrowseMenuProps {
  onNavigate: () => void
}

/**
 * The catalog laid out in labelled columns.
 *
 * Fifteen categories is the awkward middle: too many to list flat in a bar,
 * few enough that a search box alone would hide them. Grouping them by the
 * job the visitor is trying to do lets the whole catalog be taken in at a
 * glance, which is what a mega menu is actually for.
 */
export function BrowseMenu({ onNavigate }: BrowseMenuProps) {
  const categories = useCategories()
  const groups = groupCategories(categories.data ?? [])

  return (
    <div>
      {categories.isPending && (
        <p className="py-6 text-sm text-ink/60">Loading categories…</p>
      )}

      {categories.isError && (
        <p className="py-6 text-sm text-ink/70">
          The category list could not be loaded.{' '}
          <Link
            className="underline underline-offset-4"
            onClick={onNavigate}
            to="/products"
          >
            Browse all software instead
          </Link>
          .
        </p>
      )}

      {groups.length > 0 && (
        <div className="grid gap-x-8 gap-y-7 sm:grid-cols-2 lg:grid-cols-4">
          {groups.map((group) => (
            <div key={group.key}>
              <p className="eyebrow">{group.label}</p>
              <ul className="mt-3 flex flex-col">
                {group.categories.map((category) => (
                  <li key={category.slug}>
                    <Link
                      className="flex items-baseline justify-between gap-3 rounded-sm py-2 text-sm font-medium transition-colors hover:text-bronze focus-visible:text-bronze"
                      onClick={onNavigate}
                      to={`/categories/${category.slug}`}
                    >
                      <span>{category.name}</span>
                      {category.published_products !== undefined && (
                        <span className="shrink-0 text-xs text-ink/45 tabular-nums">
                          {category.published_products}
                        </span>
                      )}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      )}

      <div className="mt-6 flex flex-wrap gap-x-6 gap-y-2 border-t border-ink/15 pt-5">
        {browseShortcuts.map((shortcut) => (
          <Link
            className="group inline-flex items-center gap-1.5 text-sm font-semibold text-bronze"
            key={shortcut.to}
            onClick={onNavigate}
            to={shortcut.to}
          >
            {shortcut.label}
            <ArrowRight
              aria-hidden="true"
              className="transition-transform group-hover:translate-x-0.5"
              size={14}
            />
          </Link>
        ))}
      </div>
    </div>
  )
}
