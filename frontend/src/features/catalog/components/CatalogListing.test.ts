import { describe, expect, it } from 'vitest'

import { catalogRobots } from './catalogSeo'

describe('catalogRobots', () => {
  it('indexes only clean canonical catalog landing pages', () => {
    expect(catalogRobots(false, '')).toBe('index, follow')
    expect(catalogRobots(false, '?page=2')).toBe('noindex, follow')
    // The live-offer and price-band filters are query strings like any other.
    expect(catalogRobots(false, '?has_offer=true')).toBe('noindex, follow')
    expect(catalogRobots(false, '?minPrice=10&maxPrice=20')).toBe(
      'noindex, follow',
    )
    expect(catalogRobots(true, '')).toBe('noindex, follow')
  })
})
