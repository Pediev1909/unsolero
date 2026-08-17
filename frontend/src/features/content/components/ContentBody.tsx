import { headingID } from '../model'
import type { ContentBlock } from '../schemas'

export function ContentBody({ blocks }: { blocks: ContentBlock[] }) {
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
                  <footer className="mt-3 font-sans text-xs uppercase tracking-[0.12em] text-ink/50">
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
          default:
            return <p key={key}>{block.text}</p>
        }
      })}
    </div>
  )
}
