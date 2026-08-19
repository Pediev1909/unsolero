import { useState } from 'react'

import { Select } from '../../components/ui/Select'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import { useAdminEvents } from '../../features/admin/queries'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

const eventNames = [
  '',
  'page_view',
  'onboarding_started',
  'onboarding_completed',
  'recommendation_generated',
  'product_viewed',
  'affiliate_clicked',
  'product_saved',
  'comparison_created',
  'setup_saved',
]

export function AdminEventsPage() {
  usePageMetadata({
    title: 'Events | UNSOLERO admin',
    description: 'Protected first-party event inspection.',
    robots: 'noindex, follow',
  })
  const [name, setName] = useState('')
  const [page, setPage] = useState(1)
  const query = useAdminEvents(name, page)
  return (
    <>
      <AdminPageHeader
        description="Validated first-party events only. Conversion, revenue, and customer claims are never inferred."
        eyebrow="Analytics"
        title="Events"
      />
      <Select
        containerClassName="mb-5 max-w-sm"
        label="Event name"
        onChange={(event) => {
          setName(event.target.value)
          setPage(1)
        }}
        value={name}
      >
        {eventNames.map((value) => (
          <option key={value || 'all'} value={value}>
            {value ? value.replaceAll('_', ' ') : 'All events'}
          </option>
        ))}
      </Select>
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="No real events match this filter."
        error={query.error}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Time',
                    'Event',
                    'Surface',
                    'Traffic source',
                    'Subject',
                    'Properties',
                    'Consent',
                  ].map((header) => (
                    <th className={adminTableHead} key={header}>
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((event) => (
                  <tr key={event.id}>
                    <td className={adminTableCell}>
                      {new Date(event.occurred_at).toLocaleString()}
                    </td>
                    <td className={adminTableCell}>{event.name}</td>
                    <td className={adminTableCell}>{event.surface}</td>
                    <td className={adminTableCell}>
                      {event.traffic_source ?? event.referrer_host ?? 'No data'}
                    </td>
                    <td className={adminTableCell}>
                      {event.user_id
                        ? 'Authenticated'
                        : event.anonymous_id
                          ? 'Anonymous'
                          : 'Unattributed'}
                    </td>
                    <td className={`${adminTableCell} max-w-sm`}>
                      <code className="break-all text-xs text-ink/70">
                        {JSON.stringify(event.properties)}
                      </code>
                    </td>
                    <td className={adminTableCell}>{event.consent_state}</td>
                  </tr>
                ))}
              </tbody>
            </AdminTable>
            <AdminPagination
              onPageChange={setPage}
              page={query.data.page}
              total={query.data.total}
              totalPages={query.data.total_pages}
            />
          </>
        )}
      </AdminQueryState>
    </>
  )
}
