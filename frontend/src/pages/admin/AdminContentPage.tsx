import { ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'

import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import { contentTypeLabel } from '../../features/content/model'
import { useContent } from '../../features/content/queries'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

export function AdminContentPage() {
  const content = useContent({ section: 'all', limit: 24 })
  usePageMetadata({
    title: 'Content | UNSOLERO admin',
    description: 'Protected editorial content inventory.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <AdminPageHeader
        description="Review the published editorial inventory and its public acquisition routes. Draft authoring remains repository-controlled in this foundation phase."
        eyebrow="Editorial"
        title="Content"
      />
      <AdminQueryState
        empty={content.data?.length === 0}
        emptyDescription="No reviewed editorial entries have been published."
        emptyTitle="No published content"
        error={content.error}
        onRetry={() => void content.refetch()}
        pending={content.isPending}
      >
        {content.data && (
          <AdminTable>
            <thead>
              <tr>
                {['Title', 'Type', 'Published', 'Updated', 'Public page'].map(
                  (heading) => (
                    <th className={adminTableHead} key={heading} scope="col">
                      {heading}
                    </th>
                  ),
                )}
              </tr>
            </thead>
            <tbody>
              {content.data.map((entry) => (
                <tr className="bg-surface hover:bg-paper" key={entry.id}>
                  <td className={adminTableCell}>
                    <p className="font-semibold">{entry.title}</p>
                    <p className="mt-1 text-xs text-ink/65">{entry.path}</p>
                  </td>
                  <td className={adminTableCell}>
                    {contentTypeLabel(entry.type)}
                  </td>
                  <td className={adminTableCell}>
                    {new Date(entry.published_at).toLocaleDateString()}
                  </td>
                  <td className={adminTableCell}>
                    {new Date(entry.updated_at).toLocaleDateString()}
                  </td>
                  <td className={adminTableCell}>
                    <Link
                      className="inline-flex items-center gap-2 font-semibold text-bronze-dark hover:text-ink"
                      rel="noopener noreferrer"
                      target="_blank"
                      to={entry.path}
                    >
                      View <ExternalLink aria-hidden="true" size={14} />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </AdminTable>
        )}
      </AdminQueryState>
    </>
  )
}
