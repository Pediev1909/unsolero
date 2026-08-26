import type { ContentSummary } from '../schemas'
import { ContentCard } from './ContentCard'

export function ContentGrid({ entries }: { entries: ContentSummary[] }) {
  return (
    <div className="grid items-start gap-5 md:grid-cols-2 xl:grid-cols-3">
      {entries.map((entry) => (
        <ContentCard entry={entry} key={entry.id} />
      ))}
    </div>
  )
}
