import { describe, expect, it } from 'vitest'

import {
  normalizeProductSelection,
  updateProductSelection,
} from './productSelections'

describe('product selection rules', () => {
  it('adds and removes products without changing the original array', () => {
    const original = ['one']
    expect(updateProductSelection(original, 'two', 4)).toEqual(['one', 'two'])
    expect(updateProductSelection(original, 'one', 4)).toEqual([])
    expect(original).toEqual(['one'])
  })

  it('enforces the comparison limit deterministically', () => {
    expect(
      updateProductSelection(['one', 'two', 'three', 'four'], 'five', 4),
    ).toEqual(['one', 'two', 'three', 'four'])
  })

  it('deduplicates persisted values while preserving order', () => {
    expect(
      normalizeProductSelection(['two', 'one', 'two', 'three'], 2),
    ).toEqual(['two', 'one'])
  })
})
