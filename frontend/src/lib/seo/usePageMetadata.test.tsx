import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { usePageMetadata } from './usePageMetadata'

function MetadataHarness() {
  usePageMetadata({
    title: 'Editorial title | UNSOLERO',
    description: 'A useful editorial description.',
    type: 'article',
    canonicalURL: 'https://rigmark.example/guides/example',
    author: 'UNSOLERO Editorial',
    publishedAt: '2026-08-01T09:00:00Z',
    updatedAt: '2026-08-02T09:00:00Z',
  })
  return null
}

describe('usePageMetadata', () => {
  it('sets canonical, Open Graph, Twitter, and article metadata', () => {
    render(<MetadataHarness />)

    expect(document.querySelector('link[rel="canonical"]')).toHaveAttribute(
      'href',
      'https://rigmark.example/guides/example',
    )
    expect(document.querySelector('meta[property="og:type"]')).toHaveAttribute(
      'content',
      'article',
    )
    expect(
      document.querySelector('meta[name="twitter:title"]'),
    ).toHaveAttribute('content', 'Editorial title | UNSOLERO')
    expect(
      document.querySelector('meta[property="article:author"]'),
    ).toHaveAttribute('content', 'UNSOLERO Editorial')
    expect(document.querySelector('meta[property="og:image"]')).toHaveAttribute(
      'content',
      'http://localhost:3000/images/rigmark-home-gym-hero.webp',
    )
  })
})
