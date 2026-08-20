import { describe, expect, it } from 'vitest'

import { buttonStyles } from './buttonStyles'

// The compare bar reached production with its Compare button rendered as
// bg-ink on a bg-ink bar: 1.00:1 contrast. It was present, focusable and
// announced by a screen reader, and invisible to everyone else.
//
// The cause was cn() concatenating rather than merging. The component asked
// for variant="primary" and passed bg-canvas as a className to override it;
// both classes reached the element and the stylesheet's own order picked the
// winner. Overriding a variant's colour from outside cannot be relied on, so
// a surface that needs a different button gets a variant.
describe('buttonStyles', () => {
  it('does not paint the inverse variant in the ink it sits on', () => {
    const inverse = buttonStyles({ variant: 'inverse' })
    expect(inverse).toContain('bg-canvas')
    expect(inverse).not.toContain('bg-ink')
  })

  it('keeps every variant visually distinct from primary', () => {
    const primary = buttonStyles({ variant: 'primary' })
    for (const variant of ['secondary', 'quiet', 'danger', 'inverse'] as const) {
      expect(buttonStyles({ variant })).not.toEqual(primary)
    }
  })

  it('gives small buttons a 44px hit box', () => {
    expect(buttonStyles({ size: 'sm' })).toContain('min-h-11')
  })
})
