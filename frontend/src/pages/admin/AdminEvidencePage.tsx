import { ArrowLeft, ExternalLink } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'

import { Badge } from '../../components/ui/Badge'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  useGovernedProducts,
  useProductGovernance,
} from '../../features/admin/queries'

export function AdminEvidencePage() {
  const [search, setSearch] = useSearchParams()
  const page = Math.max(1, Number(search.get('page') ?? 1))
  const query = useGovernedProducts(page)
  return (
    <>
      <AdminPageHeader
        description="Inspect publication state and immutable fact and score revisions. Commercial data is deliberately absent."
        eyebrow="Recommendation governance"
        title="Product evidence"
      />
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="Create evidence before publishing catalog facts."
        emptyTitle="No governed products"
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <div className="overflow-x-auto border border-ink/10">
            <table className="w-full min-w-[48rem] text-left text-sm">
              <thead className={adminTableHead}>
                <tr>
                  <th className="px-4 py-3">Product</th>
                  <th className="px-4 py-3">State</th>
                  <th className="px-4 py-3">Fact revision</th>
                  <th className="px-4 py-3">Score revision</th>
                  <th className="px-4 py-3">Inspection</th>
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((item) => (
                  <tr key={item.product_id}>
                    <td className={adminTableCell}>{item.product_name}</td>
                    <td className={adminTableCell}>
                      <Badge
                        variant={
                          item.status === 'published' ? 'success' : 'neutral'
                        }
                      >
                        {item.status}
                      </Badge>
                    </td>
                    <td className={adminTableCell}>
                      {shortID(item.published_fact_revision_id)}
                    </td>
                    <td className={adminTableCell}>
                      {shortID(item.published_score_revision_id)}
                    </td>
                    <td className={adminTableCell}>
                      <Link
                        className="font-semibold text-bronze-dark hover:text-ink"
                        to={`/admin/evidence/${item.product_id}`}
                      >
                        Inspect provenance
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {query.data.total_pages > 1 && (
              <div className="flex justify-end gap-3 p-4 text-sm">
                <button
                  className="min-h-11 px-3 disabled:opacity-40"
                  disabled={page === 1}
                  onClick={() => setSearch({ page: String(page - 1) })}
                  type="button"
                >
                  Previous
                </button>
                <button
                  className="min-h-11 px-3 disabled:opacity-40"
                  disabled={page >= query.data.total_pages}
                  onClick={() => setSearch({ page: String(page + 1) })}
                  type="button"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        )}
      </AdminQueryState>
    </>
  )
}

export function AdminEvidenceDetailPage() {
  const { productID } = useParams()
  const query = useProductGovernance(productID)
  return (
    <>
      <Link
        className="mb-6 inline-flex min-h-11 items-center gap-2 text-sm font-semibold"
        to="/admin/evidence"
      >
        <ArrowLeft aria-hidden="true" size={16} /> Back to evidence
      </Link>
      <AdminQueryState
        empty={false}
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminPageHeader
              description="Source observations, classifications, revision workflow, and append-only audit history."
              eyebrow="Evidence inspection"
              title={query.data.product_name}
            />
            <section aria-labelledby="revision-heading">
              <h2 className="font-editorial text-2xl" id="revision-heading">
                Revisions
              </h2>
              <div className="mt-5 overflow-x-auto border border-ink/10">
                <table className="w-full min-w-[44rem] text-left text-sm">
                  <thead className={adminTableHead}>
                    <tr>
                      <th className="px-4 py-3">Version</th>
                      <th className="px-4 py-3">Status</th>
                      <th className="px-4 py-3">Created</th>
                      <th className="px-4 py-3">Valid until</th>
                    </tr>
                  </thead>
                  <tbody>
                    {query.data.revisions.map((revision) => (
                      <tr key={revision.fact_revision_id}>
                        <td className={adminTableCell}>
                          Facts v{revision.fact_version} · Scores v
                          {revision.score_version}
                        </td>
                        <td className={adminTableCell}>
                          <Badge
                            variant={
                              revision.status === 'published'
                                ? 'success'
                                : 'neutral'
                            }
                          >
                            {revision.status.replaceAll('_', ' ')}
                          </Badge>
                        </td>
                        <td className={adminTableCell}>
                          {formatDate(revision.created_at)}
                        </td>
                        <td className={adminTableCell}>
                          {formatDate(revision.valid_until)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
            <section aria-labelledby="provenance-heading" className="mt-12">
              <h2 className="font-editorial text-2xl" id="provenance-heading">
                Provenance
              </h2>
              <div className="mt-5 grid gap-4 lg:grid-cols-2">
                {query.data.provenance.map((item, index) => (
                  <article
                    className="border border-ink/10 bg-paper p-5"
                    key={`${item.observation.id}-${item.fact_key}-${item.score_key}-${index}`}
                  >
                    <div className="flex flex-wrap gap-2">
                      <Badge>{item.fact_key || item.score_key}</Badge>
                      <Badge
                        variant={
                          item.source.is_fictional ? 'warning' : 'success'
                        }
                      >
                        {item.source.is_fictional
                          ? 'Fictional demo'
                          : item.classification || 'Score rationale'}
                      </Badge>
                    </div>
                    <h3 className="mt-4 font-semibold">{item.source.title}</h3>
                    <p className="mt-1 text-sm text-ink/55">
                      {item.source.publisher}
                    </p>
                    {item.rationale && (
                      <p className="mt-4 text-sm leading-6">{item.rationale}</p>
                    )}
                    <p className="mt-4 text-xs text-ink/45">
                      Observed {formatDate(item.observation.observed_at)} ·
                      Confidence {item.observation.confidence}/100
                    </p>
                    {item.source.source_url && (
                      <a
                        className="mt-3 inline-flex min-h-11 items-center gap-2 text-xs font-semibold text-bronze-dark"
                        href={item.source.source_url}
                        rel="noreferrer"
                        target="_blank"
                      >
                        Open source{' '}
                        <ExternalLink aria-hidden="true" size={14} />
                      </a>
                    )}
                  </article>
                ))}
              </div>
            </section>
            <section aria-labelledby="audit-heading" className="mt-12">
              <h2 className="font-editorial text-2xl" id="audit-heading">
                Audit history
              </h2>
              {query.data.audit.length === 0 ? (
                <p className="mt-4 text-sm text-ink/55">
                  No audit events recorded.
                </p>
              ) : (
                <ol className="mt-5 divide-y divide-ink/10 border-y border-ink/10">
                  {query.data.audit.map((event, index) => (
                    <li
                      className="py-4 text-sm"
                      key={`${event.occurred_at}-${event.action}-${index}`}
                    >
                      <span className="font-semibold">{event.action}</span>
                      <span className="text-ink/50">
                        {' '}
                        · {event.actor_email ?? 'System migration'} ·{' '}
                        {formatDate(event.occurred_at)}
                      </span>
                    </li>
                  ))}
                </ol>
              )}
            </section>
          </>
        )}
      </AdminQueryState>
    </>
  )
}

function shortID(value: string | null) {
  return value ? value.slice(0, 8) : 'Not published'
}
function formatDate(value: string | null) {
  return value ? new Date(value).toLocaleString() : 'No expiry'
}
