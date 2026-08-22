import { describe, expect, it } from 'vitest'

import { groupCategories } from './categoryGroups'
import type { Category } from './schemas'

function category(slug: string, published = 3): Category {
  return {
    id: slug,
    name: slug,
    slug,
    description: '',
    published_products: published,
  }
}

describe('groupCategories', () => {
  it('orders groups by the question a visitor arrives with', () => {
    const groups = groupCategories([
      category('help-desk'),
      category('crm'),
      category('payments'),
    ])
    expect(groups.map((group) => group.label)).toEqual([
      'Find customers',
      'Sell online',
      'Run the business',
    ])
  })

  // A category added to the database that nobody remembered to place must
  // still be reachable. Silently dropping it from the only page that lists
  // categories is a failure nobody notices until a visitor cannot find it.
  it('keeps an unplaced category under More instead of dropping it', () => {
    const groups = groupCategories([category('crm'), category('warehouse')])
    const more = groups.find((group) => group.label === 'More')
    expect(more?.categories.map((c) => c.slug)).toEqual(['warehouse'])
  })

  it('leaves out a category with nothing published in it', () => {
    const groups = groupCategories([category('crm'), category('payments', 0)])
    const slugs = groups.flatMap((group) => group.categories.map((c) => c.slug))
    expect(slugs).toEqual(['crm'])
  })

  // The count is optional so that a cached response without it still renders.
  it('keeps a category whose count is unknown', () => {
    const groups = groupCategories([
      { id: 'crm', name: 'CRM', slug: 'crm', description: '' },
    ])
    expect(groups[0]?.categories).toHaveLength(1)
  })

  it('omits an empty group entirely', () => {
    const groups = groupCategories([category('crm')])
    expect(groups.map((group) => group.label)).toEqual(['Find customers'])
  })
})
