import { ArrowUpRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { CardNameplate } from './CardNameplate'
import { contentTypeLabel, formatEditorialDate } from '../model'
import type { ContentSummary } from '../schemas'

export function ContentCard({ entry }: { entry: ContentSummary }) {
  return (
    <article className="group flex h-full flex-col border border-ink/15 bg-surface">
      <Link
        aria-label={`Read ${entry.title}`}
        className="flex h-full flex-col"
        to={entry.path}
      >
        <CardNameplate entry={entry} />
        <div className="flex flex-1 flex-col p-5 sm:p-6">
          <div className="flex items-center justify-between gap-4 text-[0.625rem] font-bold uppercase tracking-[0.14em] text-bronze-dark">
            <span>{contentTypeLabel(entry.type)}</span>
            <ArrowUpRight aria-hidden="true" size={16} />
          </div>
          <h2 className="mt-5 font-display text-2xl font-medium leading-tight tracking-[-0.04em]">
            {entry.title}
          </h2>
          <p className="mt-4 text-sm leading-6 text-ink/70">
            {entry.description}
          </p>
          <p className="mt-auto pt-7 text-xs text-ink/65">
            {entry.author_name} · {formatEditorialDate(entry.published_at)}
          </p>
        </div>
      </Link>
    </article>
  )
}
