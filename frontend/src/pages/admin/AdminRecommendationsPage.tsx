import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { Badge } from '../../components/ui/Badge'
import { PriceDisplay } from '../../components/ui/PriceDisplay'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  useAdminRecommendation,
  useAdminRecommendations,
} from '../../features/admin/queries'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

export function AdminRecommendationsPage() {
  usePageMetadata({
    title: 'Recommendations | UNSOLERO admin',
    description: 'Protected recommendation inspection.',
    robots: 'noindex, follow',
  })
  const [page, setPage] = useState(1)
  const query = useAdminRecommendations(page)
  return (
    <>
      <AdminPageHeader
        description="Inspect deterministic sessions, objective scores, policy versions, and selected products."
        eyebrow="Decision engine"
        title="Recommendations"
      />
      <AdminQueryState
        empty={query.data?.items.length === 0}
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
                    'Created',
                    'Goal',
                    'Experience',
                    'Score',
                    'Total',
                    'User',
                    '',
                  ].map((header) => (
                    <th className={adminTableHead} key={header}>
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((item) => (
                  <tr key={item.id}>
                    <td className={adminTableCell}>
                      {new Date(item.created_at).toLocaleString()}
                    </td>
                    <td className={adminTableCell}>
                      {item.goal.replaceAll('_', ' ')}
                    </td>
                    <td className={adminTableCell}>{item.experience}</td>
                    <td className={adminTableCell}>
                      <Badge>{item.objective_score}/100</Badge>
                    </td>
                    <td className={adminTableCell}>
                      <PriceDisplay
                        amountMinor={item.total_price_minor}
                        currency={item.currency}
                        size="sm"
                      />
                    </td>
                    <td className={adminTableCell}>
                      {item.user_email ?? 'Anonymous'}
                    </td>
                    <td className={adminTableCell}>
                      <Link
                        className="font-semibold underline-offset-4 hover:underline"
                        to={`/admin/recommendations/${item.id}`}
                      >
                        Inspect
                      </Link>
                    </td>
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

export function AdminRecommendationDetailPage() {
  const { recommendationID } = useParams()
  const query = useAdminRecommendation(recommendationID)
  usePageMetadata({
    title: 'Recommendation detail | UNSOLERO admin',
    description: 'Protected deterministic recommendation diagnostics.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <Link
        className="mb-5 inline-block text-sm text-ink/68 hover:text-ink"
        to="/admin/recommendations"
      >
        ← Recommendations
      </Link>
      <AdminPageHeader
        description="Scoring dimensions and reasons shown exactly as persisted by the deterministic engine."
        eyebrow="Decision engine"
        title="Recommendation inspection"
      />
      <AdminQueryState
        empty={false}
        error={query.error}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && <RecommendationDetail data={query.data} />}
      </AdminQueryState>
    </>
  )
}

function RecommendationDetail({
  data,
}: {
  data: NonNullable<ReturnType<typeof useAdminRecommendation>['data']>
}) {
  return (
    <div className="space-y-8">
      <section className="grid gap-px border border-ink/10 bg-ink/10 sm:grid-cols-2 xl:grid-cols-4">
        <Fact
          label="Objective score"
          value={`${data.recommendation.objective_score}/100`}
        />
        <Fact label="Policy" value={data.recommendation.policy_version} />
        <Fact label="Engine" value={data.recommendation.engine_version} />
        <Fact
          label="User"
          value={data.recommendation.user_email ?? 'Anonymous'}
        />
      </section>
      <section>
        <h2 className="mb-4 font-editorial text-2xl">Score breakdown</h2>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Object.entries(data.scores).map(([name, score]) => (
            <div
              className="flex items-center justify-between border border-ink/10 bg-surface px-4 py-3"
              key={name}
            >
              <span className="capitalize">{name}</span>
              <strong>{score}/100</strong>
            </div>
          ))}
        </div>
      </section>
      <section>
        <h2 className="mb-4 font-editorial text-2xl">Products and reasons</h2>
        {data.items.length === 0 ? (
          <p className="border border-ink/10 bg-surface p-6 text-ink/68">
            No data
          </p>
        ) : (
          <div className="space-y-4">
            {data.items.map((item) => (
              <article
                className="border border-ink/10 bg-surface p-5"
                key={`${item.item_type}-${item.product_id}`}
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold">{item.product_name}</p>
                    <p className="text-xs text-ink/65">
                      {item.item_type.replaceAll('_', ' ')} · rank {item.rank}
                    </p>
                  </div>
                  <Badge>{item.objective_score}/100</Badge>
                </div>
                <p className="mt-4 text-sm">{item.reason_summary}</p>
                {item.reasons.length > 0 && (
                  <ul className="mt-3 space-y-2 text-sm text-ink/70">
                    {item.reasons.map((reason) => (
                      <li key={`${item.product_id}-${reason.code}`}>
                        {reason.message}{' '}
                        <span className="text-ink/35">
                          ({reason.dimension}: {reason.score})
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-surface p-5">
      <p className="text-xs text-ink/65">{label}</p>
      <p className="mt-2 font-semibold">{value}</p>
    </div>
  )
}
