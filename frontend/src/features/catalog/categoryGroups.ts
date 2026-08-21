import type { Category } from './schemas'

/**
 * The catalog has fifteen categories, which is too many to list flat and too
 * few to hide behind a search box. Nielsen Norman's finding on mega menus is
 * that they work when the items are grouped into labelled columns rather than
 * poured into one — so these groups exist to make the list scannable.
 *
 * They are ordered by the question a visitor arrives with, not alphabetically.
 * Someone who needs a CRM is trying to find customers; someone who needs
 * Stripe is trying to get paid. Alphabetical order serves nobody who does not
 * already know the name of the thing they want.
 */
export interface CategoryGroup {
  key: string
  label: string
  /** Slugs in the order they should appear inside the group. */
  slugs: string[]
}

export const categoryGroups: CategoryGroup[] = [
  {
    key: 'find-customers',
    label: 'Find customers',
    slugs: ['crm', 'email-marketing', 'seo-tools', 'analytics'],
  },
  {
    key: 'sell-online',
    label: 'Sell online',
    slugs: ['ecommerce-platform', 'payments', 'course-platform'],
  },
  {
    key: 'build-and-design',
    label: 'Build and design',
    slugs: ['website-builder', 'design-tools', 'automation'],
  },
  {
    key: 'run-the-business',
    label: 'Run the business',
    slugs: [
      'project-management',
      'accounting-invoicing',
      'scheduling',
      'team-communication',
      'help-desk',
    ],
  },
]

export interface GroupedCategories {
  key: string
  label: string
  categories: Category[]
}

/**
 * Groups the live categories for display.
 *
 * A category the groups above have never heard of still appears, under "More",
 * rather than vanishing. Adding a category to the database and having it
 * silently disappear from the only page that lists them is exactly the kind of
 * failure nobody notices until a visitor cannot find something.
 *
 * A category with nothing published in it is left out entirely: its page is a
 * promise rather than a page, which is why the sitemap already excludes it.
 */
export function groupCategories(categories: Category[]): GroupedCategories[] {
  const stocked = categories.filter(
    (category) =>
      category.published_products === undefined ||
      category.published_products > 0,
  )
  const bySlug = new Map(stocked.map((category) => [category.slug, category]))
  const claimed = new Set<string>()

  const groups: GroupedCategories[] = []
  for (const group of categoryGroups) {
    const members: Category[] = []
    for (const slug of group.slugs) {
      const category = bySlug.get(slug)
      if (!category) continue
      members.push(category)
      claimed.add(slug)
    }
    if (members.length) {
      groups.push({ key: group.key, label: group.label, categories: members })
    }
  }

  const rest = stocked
    .filter((category) => !claimed.has(category.slug))
    .sort((left, right) => left.name.localeCompare(right.name))
  if (rest.length) {
    groups.push({ key: 'more', label: 'More', categories: rest })
  }
  return groups
}
