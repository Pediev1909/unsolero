import type { ContentBlock, ContentType } from './schemas'

export function contentTypeLabel(type: ContentType) {
  switch (type) {
    case 'article':
      return 'Article'
    case 'guide':
      return 'Guide'
    case 'buying_guide':
      return 'Buying guide'
    case 'comparison':
      return 'Product comparison'
    case 'stack':
      return 'Stack'
  }
}

export interface ContentHubLink {
  label: string
  to: string
}

// The index page a piece belongs under, for its breadcrumb. Comparisons live
// at /compare/{slug} but their index is /comparisons: the former is the tool a
// visitor drives, the latter is the writing. The server emits the same trail
// as BreadcrumbList (editorialHub in public_routes.go), so the two must agree.
export function contentHub(type: ContentType): ContentHubLink {
  switch (type) {
    case 'article':
      return { label: 'Articles', to: '/articles' }
    case 'comparison':
      return { label: 'Comparisons', to: '/comparisons' }
    case 'stack':
      return { label: 'Stacks', to: '/stacks' }
    case 'guide':
    case 'buying_guide':
      return { label: 'Guides', to: '/guides' }
  }
}

export function formatEditorialDate(value: string) {
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'long' }).format(
    new Date(value),
  )
}

export function headingID(value: string) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}

// The site has one author and the API does not carry a role, so the byline's
// second line is a constant rather than a column. When a second writer
// arrives this moves to the author record and the constant goes.
export const authorRole = 'Founder, UNSOLERO'

export function authorInitial(name: string) {
  return name.trim().charAt(0).toUpperCase()
}

// The one sentence every paid control inside an article carries. A reader
// mid-article has not seen a disclosure anywhere else on screen, so this is
// the only one they get; the `cta` block in ContentBody says the same words.
export const affiliateDisclosure =
  'Affiliate link. It pays us if you subscribe, and it changed nothing about where this tool sits on this page.'

export interface TableOfContentsEntry {
  id: string
  label: string
}

// The "In this piece" list. Section headings, plus a FAQ block's heading when
// it carries one — the questions are a section a reader jumps to as often as
// any other, and the block renders its heading with the same id.
export function tableOfContents(
  blocks: ContentBlock[],
): TableOfContentsEntry[] {
  return blocks.flatMap((block) => {
    if (!block.heading) return []
    if (block.type !== 'heading' && block.type !== 'faq') return []
    return [{ id: headingID(block.heading), label: block.heading }]
  })
}
