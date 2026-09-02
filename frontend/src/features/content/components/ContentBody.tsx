import {
  Check,
  ChevronDown,
  ExternalLink,
  X,
  type LucideIcon,
} from 'lucide-react'

import { affiliateClickPath } from '../../analytics/tracking'
import type { ProductSummary } from '../../catalog/schemas'
import { headingID } from '../model'
import type { ContentBlock } from '../schemas'
import { OfferBlock } from './OfferBlock'

export function ContentBody({
  blocks,
  // The piece's related products, so an `offer` block can print the product's
  // name without a request of its own. Optional: the body renders without it.
  products = [],
}: {
  blocks: ContentBlock[]
  products?: ProductSummary[]
}) {
  return (
    <div className="space-y-7 text-[1.0625rem] leading-8 text-ink/75 sm:text-lg sm:leading-9">
      {blocks.map((block, index) => {
        const key = `${block.type}-${index}`
        switch (block.type) {
          case 'heading':
            return (
              <h2
                className="scroll-mt-28 pt-7 font-display text-3xl font-medium leading-tight tracking-[-0.04em] text-ink sm:text-4xl"
                id={headingID(block.heading ?? `section-${index}`)}
                key={key}
              >
                {block.heading}
              </h2>
            )
          case 'unordered_list':
            return (
              <ul className="space-y-3 pl-6" key={key}>
                {block.items?.map((item) => (
                  <li className="list-disc pl-2 marker:text-bronze" key={item}>
                    {item}
                  </li>
                ))}
              </ul>
            )
          case 'ordered_list':
            return (
              <ol className="space-y-4 pl-7" key={key}>
                {block.items?.map((item) => (
                  <li
                    className="list-decimal pl-2 marker:font-semibold marker:text-bronze-dark"
                    key={item}
                  >
                    {item}
                  </li>
                ))}
              </ol>
            )
          case 'quote':
            return (
              <blockquote
                className="border-l-2 border-bronze py-1 pl-6 font-editorial text-2xl leading-9 text-ink"
                key={key}
              >
                <p>{block.text}</p>
                {block.attribution && (
                  <footer className="mt-3 font-sans text-xs uppercase tracking-[0.12em] text-ink/68">
                    {block.attribution}
                  </footer>
                )}
              </blockquote>
            )
          case 'callout':
            return (
              <aside
                className="border-y border-bronze/35 bg-paper px-5 py-6 sm:px-7"
                key={key}
              >
                <h3 className="font-semibold text-ink">{block.heading}</h3>
                <p className="mt-2 text-base leading-7">{block.text}</p>
              </aside>
            )
          // A vendor exit inside the article, at the point the argument has
          // been made rather than at the bottom where the reader has gone.
          //
          // It is deliberately not styled as a banner. This site's only asset
          // is that a reader believes the ranking is not for sale, and an ad
          // unit in the middle of a paragraph spends that belief for one
          // click. So it reads as part of the writing, says in its own text
          // that it pays us, and looks the same whether or not it does.
          //
          // `promotion` is a slug: the href is built here from an approved
          // promotion, never taken from the block. A block cannot name a
          // destination, only choose one.
          case 'cta': {
            if (!block.promotion || !block.label) return null
            return (
              <aside
                className="border border-ink/15 bg-paper px-5 py-6 sm:px-7"
                key={key}
              >
                {block.heading && (
                  <h3 className="font-semibold text-ink">{block.heading}</h3>
                )}
                <p className="mt-2 text-base leading-7">{block.text}</p>
                <a
                  className="mt-4 inline-flex min-h-11 items-center justify-center gap-2 bg-charcoal px-5 text-sm font-semibold text-canvas hover:bg-ink"
                  href={affiliateClickPath(
                    `/api/affiliate/promotion/${block.promotion}`,
                    'promotion',
                  )}
                  rel="nofollow noopener sponsored"
                  target="_blank"
                >
                  {block.label}
                  <ExternalLink aria-hidden="true" size={14} />
                </a>
                <p className="mt-3 text-xs text-ink/65">
                  Affiliate link. It pays us if you subscribe, and it changed
                  nothing about where this tool sits on this page.
                </p>
              </aside>
            )
          }
          // What a tool does well and what it costs you, side by side. Both
          // columns are required by the server, and the marks are ink like
          // everything else: green ticks and red crosses would praise and
          // condemn where the words already judge.
          case 'pros_cons':
            return (
              <section className="border-y border-ink/15 py-6" key={key}>
                {block.heading && (
                  <h3 className="font-semibold text-ink">{block.heading}</h3>
                )}
                {block.text && (
                  <p className="mt-2 text-base leading-7">{block.text}</p>
                )}
                <div className="mt-5 grid gap-6 sm:grid-cols-2 sm:gap-8">
                  <ProsConsList
                    icon={Check}
                    items={block.pros ?? []}
                    prefix="Pro"
                    title="Pros"
                  />
                  <ProsConsList
                    icon={X}
                    items={block.cons ?? []}
                    prefix="Con"
                    title="Cons"
                  />
                </div>
              </section>
            )
          // Question-and-answer pairs as native disclosure widgets, so they
          // work with no script and every answer is in the document for a
          // reader who searches the page. The first is open because a list
          // of closed questions is a list of things the page will not say.
          case 'faq': {
            if (!block.questions?.length) return null
            return (
              <section key={key}>
                {block.heading && (
                  // The same treatment as a heading block: this is a section
                  // the table of contents points at, under the same id.
                  <h2
                    className="scroll-mt-28 pt-7 font-display text-3xl font-medium leading-tight tracking-[-0.04em] text-ink sm:text-4xl"
                    id={headingID(block.heading)}
                  >
                    {block.heading}
                  </h2>
                )}
                {block.text && <p className="mt-7">{block.text}</p>}
                <div className="mt-7 divide-y divide-ink/15 border-y border-ink/15">
                  {block.questions.map((pair, position) => (
                    <details
                      className="group py-4"
                      key={pair.question}
                      open={position === 0}
                    >
                      <summary className="flex cursor-pointer list-none items-start justify-between gap-4 text-base font-semibold leading-7 text-ink [&::-webkit-details-marker]:hidden">
                        {pair.question}
                        <ChevronDown
                          aria-hidden="true"
                          className="mt-1.5 shrink-0 text-ink/60 transition-transform group-open:rotate-180 motion-reduce:transition-none"
                          size={18}
                        />
                      </summary>
                      <p className="mt-3 text-base leading-7">{pair.answer}</p>
                    </details>
                  ))}
                </div>
              </section>
            )
          }
          // A vendor exit for a catalog product. The block names a slug; the
          // price, the date and the destination come from the product's live
          // offer at render time. See OfferBlock for the honest empty state.
          case 'offer': {
            if (!block.product) return null
            return (
              <OfferBlock
                heading={block.heading}
                key={key}
                label={block.label}
                product={products.find(
                  (candidate) => candidate.slug === block.product,
                )}
                slug={block.product}
                text={block.text}
              />
            )
          }
          default:
            return <p key={key}>{block.text}</p>
        }
      })}
    </div>
  )
}

function ProsConsList({
  title,
  prefix,
  items,
  icon: Icon,
}: {
  title: string
  // Read out before each item, because an icon marked decorative tells a
  // screen reader nothing about which column it is hearing.
  prefix: string
  items: string[]
  icon: LucideIcon
}) {
  return (
    <div>
      <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
        {title}
      </p>
      <ul className="mt-3 space-y-3 text-base leading-7">
        {items.map((item) => (
          <li className="flex gap-3" key={item}>
            <Icon
              aria-hidden="true"
              className="mt-2 shrink-0 text-ink/70"
              size={16}
            />
            <span>
              <span className="sr-only">{prefix}: </span>
              {item}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
