import { describe, expect, it } from 'vitest'

import { isNavigationItemActive } from './navigationItems'

describe('isNavigationItemActive', () => {
  it('keeps catalog navigation active across product, category, and brand pages', () => {
    expect(
      isNavigationItemActive('/products/demo-product', '', '/products'),
    ).toBe(true)
    expect(isNavigationItemActive('/categories/benches', '', '/products')).toBe(
      true,
    )
    expect(isNavigationItemActive('/brands/demo-brand', '', '/products')).toBe(
      true,
    )
  })

  it('only marks an anchored homepage section when the hash matches', () => {
    expect(isNavigationItemActive('/', '#method', '/#method')).toBe(true)
    expect(isNavigationItemActive('/', '#categories', '/#method')).toBe(false)
  })
})
