/**
 * The sections of a product page a reader can jump to, in reading order.
 *
 * The page decides which of them exist — a product whose vendor has no
 * programme has no "Where to get it", one nobody has written about has no
 * "Compared in" — and the jump row follows that decision. An anchor to a
 * section that rendered nothing is a link to a blank.
 */
export interface ProductSection {
  id: string
  label: string
}

export const productSectionIDs = {
  glance: 'at-a-glance',
  priceRecord: 'price-record',
  profile: 'decision-profile',
  evidence: 'evidence',
  offers: 'where-to-get-it',
  editorial: 'compared-in',
  alternatives: 'alternatives',
} as const

/**
 * Room for what sits above a section when it scrolls into view: the sticky
 * header (4.5rem on phones, 5rem from `sm`), plus the jump row, which is only
 * sticky from `lg`. Without the offset the target heading lands under them.
 */
export const sectionAnchorClass = 'scroll-mt-28 lg:scroll-mt-36'

export interface PresentSections {
  evidence: boolean
  offers: boolean
  editorial: boolean
  /**
   * A price record exists only where a figure actually moved, which is seven
   * products today. Optional because most callers have no reason to know
   * about it, and an absent flag means the section is not there.
   */
  priceRecord?: boolean
}

export function productSections(present: PresentSections): ProductSection[] {
  const candidates: (ProductSection & { present: boolean })[] = [
    { id: productSectionIDs.glance, label: 'At a glance', present: true },
    {
      id: productSectionIDs.priceRecord,
      label: 'Price record',
      present: present.priceRecord === true,
    },
    { id: productSectionIDs.profile, label: 'Decision profile', present: true },
    {
      id: productSectionIDs.evidence,
      label: 'Evidence',
      present: present.evidence,
    },
    {
      id: productSectionIDs.offers,
      label: 'Where to get it',
      present: present.offers,
    },
    {
      id: productSectionIDs.editorial,
      label: 'Compared in',
      present: present.editorial,
    },
    {
      id: productSectionIDs.alternatives,
      label: 'Alternatives',
      present: true,
    },
  ]
  return candidates
    .filter((section) => section.present)
    .map(({ id, label }) => ({ id, label }))
}
