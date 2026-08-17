import { describe, expect, it } from 'vitest'

import { catalogRobots } from './catalogSeo'

describe('catalogRobots', () => {
  it('indexes only clean canonical catalog landing pages', () => {
    expect(catalogRobots(false, '')).toBe('index, follow')
    expect(catalogRobots(false, '?page=2')).toBe('noindex, follow')
    expect(catalogRobots(true, '')).toBe('noindex, follow')
  })
})
