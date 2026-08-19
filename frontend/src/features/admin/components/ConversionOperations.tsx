import { useState } from 'react'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import {
  AdminPagination,
  AdminQueryState,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from './AdminStates'
import {
  useCommerceProviders,
  useConversionImports,
  useConversionProviderState,
  useConversionReconciliations,
  useMonetizationMetrics,
  useReconcileConversionImport,
  useRetryConversionImport,
  useTriggerConversionImport,
  useVerifiedConversions,
} from '../queries'
import type { MonetizationReport } from '../schemas'

export function ConversionOperations() {
  return (
    <div className="mt-14 border-t border-ink/10 pt-12">
      <header className="mb-8 max-w-3xl">
        <p className="text-xs font-semibold uppercase tracking-[0.16em] text-ink/70">
          Verified commercial facts
        </p>
        <h2 className="mt-2 font-editorial text-3xl">
          Conversions &amp; monetization
        </h2>
        <p className="mt-2 text-sm leading-6 text-ink/70">
          Provider-verified events only. Clicks are not conversions, pending
          commission is not earned revenue, and currencies are never combined
          without an external FX source.
        </p>
      </header>
      <MonetizationMetrics />
      <ConversionProviders />
      <VerifiedConversions />
      <ConversionImports />
      <ReconciliationHistory />
    </div>
  )
}

function MonetizationMetrics() {
  const query = useMonetizationMetrics()
  return (
    <section aria-labelledby="monetization-metrics" className="mb-12">
      <h3 className="font-editorial text-2xl" id="monetization-metrics">
        Monetization metrics
      </h3>
      <AdminQueryState
        empty={false}
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && <MetricGrid report={query.data} />}
      </AdminQueryState>
    </section>
  )
}

function MetricGrid({ report }: { report: MonetizationReport }) {
  const cards = [
    {
      label: 'Affiliate conversion rate',
      value: formatRatio(report.affiliate_conversion_rate),
      detail: report.affiliate_conversion_rate.definition,
    },
    {
      label: 'Earnings per click',
      value: formatCurrencyMetric(report.earnings_per_click),
      detail: report.earnings_per_click.definition,
    },
    {
      label: 'Revenue per visitor',
      value: formatCurrencyMetric(report.revenue_per_visitor),
      detail: report.revenue_per_visitor.definition,
    },
    {
      label: 'Revenue per recommendation',
      value: formatCurrencyMetric(report.revenue_per_recommendation),
      detail: report.revenue_per_recommendation.definition,
    },
    {
      label: 'Earned commission',
      value: formatCurrencyMetric(report.commission, true),
      detail: report.commission.definition,
    },
    {
      label: 'Reversal rate',
      value: formatRatio(report.reversal_rate),
      detail: report.reversal_rate.definition,
    },
    {
      label: 'Repeat user rate',
      value: formatRatio(report.repeat_user_rate),
      detail: report.repeat_user_rate.definition,
    },
  ]
  return (
    <>
      <div className="mt-5 grid gap-px border border-ink/10 bg-ink/10 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map((card) => (
          <article className="min-h-36 bg-surface p-5" key={card.label}>
            <p className="text-xs font-semibold uppercase tracking-[0.12em] text-ink/68">
              {card.label}
            </p>
            <p className="mt-3 text-xl font-semibold text-ink">{card.value}</p>
            <p className="mt-3 text-xs leading-5 text-ink/70">{card.detail}</p>
          </article>
        ))}
      </div>
      <p className="mt-3 text-xs text-ink/70">
        Fresh through: {formatDate(report.fresh_through)}.{' '}
        {report.currency_policy}
      </p>
    </>
  )
}

function ConversionProviders() {
  const query = useCommerceProviders()
  const state = useConversionProviderState()
  const trigger = useTriggerConversionImport()
  return (
    <section aria-labelledby="conversion-providers" className="mb-12">
      <h3 className="font-editorial text-2xl" id="conversion-providers">
        Provider status
      </h3>
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="No provider is configured. Conversion ingestion is disabled."
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <AdminTable>
            <thead>
              <tr>
                {[
                  'Provider',
                  'Ingestion',
                  'Last success',
                  'Health',
                  'Actions',
                ].map((label) => (
                  <th className={adminTableHead} key={label} scope="col">
                    {label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {query.data.items.map((provider) => (
                <tr key={provider.id}>
                  <td className={adminTableCell}>
                    <p className="font-medium">{provider.provider_key}</p>
                    <p className="text-xs text-ink/70">
                      {provider.merchant_name}
                    </p>
                  </td>
                  <td className={adminTableCell}>
                    <Badge>
                      {provider.conversion_ingestion_enabled
                        ? 'verified & enabled'
                        : 'not configured'}
                    </Badge>
                  </td>
                  <td className={adminTableCell}>
                    {formatDate(provider.last_conversion_import_succeeded_at)}
                  </td>
                  <td className={adminTableCell}>
                    {provider.conversion_consecutive_failures || 'No failures'}
                    {provider.last_conversion_error_code && (
                      <p className="text-xs text-danger">
                        {provider.last_conversion_error_code}
                      </p>
                    )}
                  </td>
                  <td className={adminTableCell}>
                    <div className="flex min-w-44 flex-wrap gap-2">
                      {provider.conversion_ingestion_enabled && (
                        <Button
                          loading={trigger.isPending}
                          onClick={() => void trigger.mutateAsync(provider.id)}
                          size="sm"
                        >
                          Import conversions
                        </Button>
                      )}
                      <Button
                        loading={state.isPending}
                        onClick={() =>
                          void state.mutateAsync({
                            id: provider.id,
                            enabled: !provider.conversion_ingestion_enabled,
                          })
                        }
                        size="sm"
                        variant="secondary"
                      >
                        {provider.conversion_ingestion_enabled
                          ? 'Disable'
                          : 'Verify & enable'}
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </AdminTable>
        )}
      </AdminQueryState>
      {(state.error || trigger.error) && (
        <p className="mt-3 text-sm text-danger" role="alert">
          {(state.error ?? trigger.error)?.message}
        </p>
      )}
    </section>
  )
}

function VerifiedConversions() {
  const [page, setPage] = useState(1)
  const query = useVerifiedConversions(page)
  return (
    <section aria-labelledby="verified-conversions" className="mb-12">
      <h3 className="font-editorial text-2xl" id="verified-conversions">
        Verified conversions
      </h3>
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="No data. No provider-verified conversion has been ingested."
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Provider',
                    'Event',
                    'Order',
                    'Commission',
                    'Attribution',
                    'Reconciliation',
                  ].map((label) => (
                    <th className={adminTableHead} key={label} scope="col">
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((item) => (
                  <tr key={item.id}>
                    <td className={adminTableCell}>{item.provider}</td>
                    <td className={adminTableCell}>
                      {formatDate(item.event_timestamp)}
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{item.order_status}</Badge>
                      <p className="mt-1 text-xs text-ink/70">
                        {formatMoney(
                          item.order_value_minor,
                          item.order_currency,
                        )}
                      </p>
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{item.commission_status ?? 'unreported'}</Badge>
                      <p className="mt-1 text-xs text-ink/70">
                        {formatMoney(
                          item.commission_amount_minor,
                          item.commission_currency,
                        )}
                      </p>
                    </td>
                    <td className={adminTableCell}>
                      {item.attribution_status}
                    </td>
                    <td className={adminTableCell}>
                      {item.reconciliation_status ?? 'unresolved'}
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
    </section>
  )
}

function ConversionImports() {
  const [page, setPage] = useState(1)
  const query = useConversionImports(page)
  const retry = useRetryConversionImport()
  const reconcile = useReconcileConversionImport()
  return (
    <section aria-labelledby="conversion-imports" className="mb-12">
      <h3 className="font-editorial text-2xl" id="conversion-imports">
        Conversion import health
      </h3>
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="No data. Conversion imports have not run."
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Provider',
                    'Status',
                    'Records',
                    'Coverage',
                    'Error',
                    'Actions',
                  ].map((label) => (
                    <th className={adminTableHead} key={label} scope="col">
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((run) => (
                  <tr key={run.id}>
                    <td className={adminTableCell}>
                      {run.provider_configuration.provider_key}
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{run.status}</Badge>
                    </td>
                    <td className={adminTableCell}>
                      {run.records_applied}/{run.records_received} applied
                    </td>
                    <td className={adminTableCell}>
                      {run.coverage_start
                        ? `${formatDate(run.coverage_start)} – ${formatDate(run.coverage_end)}`
                        : 'No data'}
                    </td>
                    <td className={adminTableCell}>
                      {run.error_code ?? 'None'}
                    </td>
                    <td className={adminTableCell}>
                      <div className="flex gap-2">
                        {['failed', 'partial', 'cancelled'].includes(
                          run.status,
                        ) && (
                          <Button
                            loading={retry.isPending}
                            onClick={() => void retry.mutateAsync(run.id)}
                            size="sm"
                          >
                            Retry
                          </Button>
                        )}
                        {run.status === 'succeeded' && run.coverage_start && (
                          <Button
                            loading={reconcile.isPending}
                            onClick={() => void reconcile.mutateAsync(run.id)}
                            size="sm"
                            variant="secondary"
                          >
                            Reconcile
                          </Button>
                        )}
                      </div>
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
    </section>
  )
}

function ReconciliationHistory() {
  const [page, setPage] = useState(1)
  const query = useConversionReconciliations(page)
  return (
    <section aria-labelledby="reconciliation-history">
      <h3 className="font-editorial text-2xl" id="reconciliation-history">
        Reconciliation history
      </h3>
      <AdminQueryState
        empty={query.data?.items.length === 0}
        emptyDescription="No data. No complete provider snapshot has been reconciled."
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Provider',
                    'Status',
                    'Matched',
                    'Missing',
                    'Conflicting',
                    'Stale',
                    'Coverage',
                  ].map((label) => (
                    <th className={adminTableHead} key={label} scope="col">
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((run) => (
                  <tr key={run.id}>
                    <td className={adminTableCell}>
                      {run.provider_configuration.provider_key}
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{run.status}</Badge>
                    </td>
                    <td className={adminTableCell}>{run.matched}</td>
                    <td className={adminTableCell}>{run.missing}</td>
                    <td className={adminTableCell}>{run.conflicting}</td>
                    <td className={adminTableCell}>{run.stale}</td>
                    <td className={adminTableCell}>
                      {formatDate(run.coverage_end)}
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
    </section>
  )
}

function formatRatio(metric: MonetizationReport['affiliate_conversion_rate']) {
  if (metric.status === 'no_data') return 'No data'
  if (metric.status === 'insufficient_data' || metric.value === null)
    return 'Insufficient data'
  return `${(metric.value * 100).toFixed(2)}%`
}

function formatCurrencyMetric(
  metric: MonetizationReport['commission'],
  total = false,
) {
  if (metric.status === 'no_data') return 'No data'
  if (metric.status === 'insufficient_data') return 'Insufficient data'
  if (metric.values.length === 0) return '0 verified'
  return metric.values
    .map((value) =>
      formatMoney(
        total ? value.amount_minor : value.value_minor,
        value.currency,
      ),
    )
    .join(' · ')
}

function formatMoney(value: number | null, currency: string | null) {
  if (value === null || !currency) return 'Not reported'
  return `${currency} ${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 4,
  }).format(value)} minor units`
}

function formatDate(value: string | null) {
  return value ? new Date(value).toLocaleString() : 'No data'
}
