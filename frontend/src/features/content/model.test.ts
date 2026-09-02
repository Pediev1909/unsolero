import { describe, expect, it } from 'vitest'

import { contentHub, contentTypeLabel } from './model'
import { contentTypeSchema } from './schemas'

// The breadcrumb on a piece and the BreadcrumbList the server emits for it
// (editorialHub in public_routes.go) must name the same hub. Every type gets
// one, so a new type cannot silently fall under Guides the way comparisons
// once did in the page's own ternary.
describe('contentHub', () => {
  it('sends every content type to its own index page', () => {
    expect(contentHub('article')).toEqual({
      label: 'Articles',
      to: '/articles',
    })
    expect(contentHub('guide')).toEqual({ label: 'Guides', to: '/guides' })
    expect(contentHub('buying_guide')).toEqual({
      label: 'Guides',
      to: '/guides',
    })
    expect(contentHub('comparison')).toEqual({
      label: 'Comparisons',
      to: '/comparisons',
    })
    expect(contentHub('stack')).toEqual({ label: 'Stacks', to: '/stacks' })
  })

  it('labels every type the schema accepts', () => {
    for (const type of contentTypeSchema.options) {
      expect(contentTypeLabel(type)).toBeTruthy()
      expect(contentHub(type).to.startsWith('/')).toBe(true)
    }
  })
})
