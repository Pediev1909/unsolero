import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ProductSectionNav } from './ProductSectionNav'
import { productSections } from './productSections'

describe('productSections', () => {
  it('lists every section when all of them render', () => {
    expect(
      productSections({ evidence: true, offers: true, editorial: true }).map(
        (section) => section.id,
      ),
    ).toEqual([
      'at-a-glance',
      'decision-profile',
      'evidence',
      'where-to-get-it',
      'compared-in',
      'alternatives',
    ])
  })

  // An anchor to a section the page did not render is a link to nothing.
  it('drops the anchors for sections that render nothing', () => {
    expect(
      productSections({ evidence: false, offers: false, editorial: false }).map(
        (section) => section.label,
      ),
    ).toEqual(['At a glance', 'Decision profile', 'Alternatives'])
  })
})

describe('ProductSectionNav', () => {
  it('renders one in-page link per present section', () => {
    render(
      <ProductSectionNav
        sections={productSections({
          evidence: true,
          offers: false,
          editorial: true,
        })}
      />,
    )

    const nav = screen.getByRole('navigation', { name: 'Jump to' })
    expect(nav).toBeInTheDocument()
    expect(screen.getAllByRole('link')).toHaveLength(5)
    expect(screen.getByRole('link', { name: 'Evidence' })).toHaveAttribute(
      'href',
      '#evidence',
    )
    expect(screen.getByRole('link', { name: 'Compared in' })).toHaveAttribute(
      'href',
      '#compared-in',
    )
    expect(screen.queryByRole('link', { name: 'Where to get it' })).toBeNull()
  })

  it('renders nothing with no sections', () => {
    const { container } = render(<ProductSectionNav sections={[]} />)
    expect(container).toBeEmptyDOMElement()
  })
})
