import { Container } from '../../../components/ui/Container'
import type { ProductSection } from './productSections'

/**
 * A row of anchors to the page's sections.
 *
 * The product page is long — strip, profile, evidence, vendor, editorial,
 * alternatives — and a reader who wants the evidence should not have to
 * scroll past everything else to learn whether it is there. Every long-form
 * competitor page has this row; ours lists only the sections that exist,
 * because the page leaves out the ones that would be empty.
 *
 * Sticky from `lg` only. On a phone a second sticky bar under the header
 * takes a fifth of the screen, so there the row scrolls with the page, and
 * sideways within itself.
 */
export function ProductSectionNav({
  sections,
}: {
  sections: ProductSection[]
}) {
  if (sections.length === 0) return null

  return (
    <nav
      aria-label="Jump to"
      className="border-y border-ink/15 bg-canvas/95 backdrop-blur-sm lg:sticky lg:top-20 lg:z-30"
    >
      <Container className="flex items-center gap-4">
        <p
          aria-hidden="true"
          className="hidden shrink-0 text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65 sm:block"
        >
          Jump to
        </p>
        <ul className="flex min-w-0 flex-1 gap-1 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {sections.map((section) => (
            <li className="shrink-0" key={section.id}>
              <a
                className="inline-flex min-h-12 items-center px-3 text-sm font-semibold text-ink/70 hover:text-ink"
                href={`#${section.id}`}
              >
                {section.label}
              </a>
            </li>
          ))}
        </ul>
      </Container>
    </nav>
  )
}
