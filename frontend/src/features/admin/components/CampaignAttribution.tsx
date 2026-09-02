import { useId } from 'react'

import type { AnalyticsReportData } from '../schemas'
import { AdminTable, adminTableCell, adminTableHead } from './AdminStates'

type Attribution = Pick<
  AnalyticsReportData,
  'campaigns' | 'landing_pages' | 'sources_by_medium'
>

type Column<Row> = {
  header: string
  value: (row: Row) => string
  numeric?: boolean
}

// A missing UTM parameter is a dash rather than an empty cell, so a row with a
// campaign but no medium reads as a complete visit, not a broken row.
const missing = '—'

const campaignColumns: Column<Attribution['campaigns'][number]>[] = [
  { header: 'Campaign', value: (row) => row.campaign },
  { header: 'Source', value: (row) => row.traffic_source ?? missing },
  { header: 'Medium', value: (row) => row.traffic_medium ?? missing },
  {
    header: 'Sessions',
    value: (row) => row.sessions.toLocaleString(),
    numeric: true,
  },
  {
    header: 'Page views',
    value: (row) => row.page_views.toLocaleString(),
    numeric: true,
  },
  {
    header: 'Affiliate clicks',
    value: (row) => row.affiliate_clicks.toLocaleString(),
    numeric: true,
  },
]

const landingPageColumns: Column<Attribution['landing_pages'][number]>[] = [
  { header: 'Campaign', value: (row) => row.campaign },
  { header: 'Landing page', value: (row) => row.page_path },
  {
    header: 'Sessions',
    value: (row) => row.sessions.toLocaleString(),
    numeric: true,
  },
]

const sourceMediumColumns: Column<Attribution['sources_by_medium'][number]>[] =
  [
    { header: 'Source', value: (row) => row.traffic_source },
    { header: 'Medium', value: (row) => row.traffic_medium ?? missing },
    {
      header: 'Sessions',
      value: (row) => row.sessions.toLocaleString(),
      numeric: true,
    },
  ]

// Three tables rather than one, because they answer three different
// questions: which post brought people (Campaigns), where its link dropped
// them (Landing pages), and which format on a platform works (Sources by
// medium). One joined grid would repeat the campaign on every landing-page row
// and bury the per-medium split under it.
export function CampaignAttribution({ data }: { data: Attribution }) {
  return (
    <section>
      <div className="mb-4 max-w-3xl">
        <h2 className="text-sm font-semibold tracking-[0.15em] uppercase">
          Campaign attribution
        </h2>
        <p className="mt-2 text-xs leading-5 text-ink/68">
          Attribution is the utm_source, utm_medium and utm_campaign a visit
          arrived with. Sessions and page views count visitors who accepted
          analytics; affiliate clicks are countable merchant redirects and need
          no consent, so a campaign can show clicks without sessions.
        </p>
      </div>
      <div className="space-y-6">
        <AttributionTable
          columns={campaignColumns}
          rowKey={(row) =>
            `${row.campaign}|${row.traffic_source ?? ''}|${row.traffic_medium ?? ''}`
          }
          rows={data.campaigns}
          title="Campaigns"
        />
        <div className="grid gap-6 xl:grid-cols-2">
          <AttributionTable
            columns={landingPageColumns}
            rowKey={(row) => `${row.campaign}|${row.page_path}`}
            rows={data.landing_pages}
            title="Landing pages by campaign"
          />
          <AttributionTable
            columns={sourceMediumColumns}
            rowKey={(row) =>
              `${row.traffic_source}|${row.traffic_medium ?? ''}`
            }
            rows={data.sources_by_medium}
            title="Sources by medium"
          />
        </div>
      </div>
    </section>
  )
}

function AttributionTable<Row>({
  title,
  columns,
  rows,
  rowKey,
}: {
  title: string
  columns: readonly Column<Row>[]
  rows: readonly Row[]
  rowKey: (row: Row) => string
}) {
  const headingID = useId()
  return (
    <article aria-labelledby={headingID}>
      <h3 className="mb-3 font-editorial text-xl" id={headingID}>
        {title}
      </h3>
      {rows.length ? (
        <AdminTable>
          <thead>
            <tr>
              {columns.map((column) => (
                <th
                  className={`${adminTableHead} ${column.numeric ? 'text-right' : ''}`}
                  key={column.header}
                  scope="col"
                >
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={rowKey(row)}>
                {columns.map((column) => (
                  <td
                    className={`${adminTableCell} ${column.numeric ? 'font-numeric text-right tabular-nums' : ''}`}
                    key={column.header}
                  >
                    {column.value(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </AdminTable>
      ) : (
        <p className="border border-ink/10 bg-surface p-5 text-sm font-medium text-ink/35">
          No data
        </p>
      )}
    </article>
  )
}
