import { useEffect } from 'react'

interface PageMetadata {
  title: string
  description: string
  robots?: 'index, follow' | 'noindex, follow'
  type?: 'website' | 'product' | 'article'
  imagePath?: string
  canonicalURL?: string
  author?: string
  publishedAt?: string
  updatedAt?: string
}

export function usePageMetadata({
  title,
  description,
  robots = 'index, follow',
  type = 'website',
  imagePath = '/images/unsolero-saas-hero.webp',
  canonicalURL,
  author,
  publishedAt,
  updatedAt,
}: PageMetadata) {
  useEffect(() => {
    const canonical = new URL(
      canonicalURL ?? window.location.pathname,
      window.location.origin,
    )
    const imageURL = new URL(imagePath, window.location.origin)
    document.title = title
    setMeta('description', description)
    setMeta('robots', robots)
    setPropertyMeta('og:title', title)
    setPropertyMeta('og:description', description)
    setPropertyMeta('og:type', type)
    setPropertyMeta('og:url', canonical.href)
    setPropertyMeta('og:image', imageURL.href)
    setMeta('twitter:title', title)
    setMeta('twitter:description', description)
    setMeta('twitter:image', imageURL.href)
    setOptionalPropertyMeta('article:author', author)
    setOptionalPropertyMeta('article:published_time', publishedAt)
    setOptionalPropertyMeta('article:modified_time', updatedAt)
    setCanonical(canonical.href)
  }, [
    author,
    canonicalURL,
    description,
    imagePath,
    publishedAt,
    robots,
    title,
    type,
    updatedAt,
  ])
}

function setCanonical(href: string) {
  let element = document.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!element) {
    element = document.createElement('link')
    element.rel = 'canonical'
    document.head.append(element)
  }
  element.href = href
}

function setMeta(name: string, content: string) {
  let element = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!element) {
    element = document.createElement('meta')
    element.name = name
    document.head.append(element)
  }
  element.content = content
}

function setPropertyMeta(property: string, content: string) {
  let element = document.querySelector<HTMLMetaElement>(
    `meta[property="${property}"]`,
  )
  if (!element) {
    element = document.createElement('meta')
    element.setAttribute('property', property)
    document.head.append(element)
  }
  element.content = content
}

function setOptionalPropertyMeta(property: string, content?: string) {
  const element = document.querySelector<HTMLMetaElement>(
    `meta[property="${property}"]`,
  )
  if (!content) {
    element?.remove()
    return
  }
  setPropertyMeta(property, content)
}
