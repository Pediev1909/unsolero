import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  DataValue,
} from '../../features/admin/components/AdminStates'
import {
  useAdminAnalytics,
  useAdminDashboard,
} from '../../features/admin/queries'
import type { AnalyticsReportData } from '../../features/admin/schemas'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'
import { CampaignAttribution } from '../../features/admin/components/CampaignAttribution'
import { DailyTrend } from '../../features/admin/components/DailyTrend'
import { MonetizationReadiness } from '../../features/admin/components/MonetizationReadiness'

export function AdminDashboardPage() {
  usePageMetadata({
    title: 'Admin dashboard | UNSOLERO',
    description: 'Protected UNSOLERO administration and first-party analytics.',
    robots: 'noindex, follow',
  })
  const dashboard = useAdminDashboard()
  const analytics = useAdminAnalytics()

  return (
    <>
      <AdminPageHeader
        description="Did anyone come, where from, did they click through, and what is stopping the catalog from earning. Observed first-party behaviour only — no modeled conversions, commission, or revenue."
        eyebrow="Overview"
        title="Dashboard"
      />
      <AdminQueryState
        empty={false}
        error={dashboard.error ?? analytics.error}
        onRetry={() => {
          void dashboard.refetch()
          void analytics.refetch()
        }}
        pending={dashboard.isPending || analytics.isPending}
      >
        {dashboard.data && analytics.data ? (
          <div className="space-y-12">
            {/* Ordered by the question each answers: can it earn, did anyone
                come, where from, and only then how much of everything exists. */}
            <MonetizationReadiness readiness={dashboard.data.readiness} />
            <AnalyticsSummary data={analytics.data} />
            <DailyTrend daily={analytics.data.daily} />
            <RankingGrid data={analytics.data} />
            <CampaignAttribution data={analytics.data} />
            <MetricSection
              items={[
                ['Products', dashboard.data.counts.products],
                ['Published', dashboard.data.counts.published],
                ['Offers', dashboard.data.counts.offers],
                ['Active offers', dashboard.data.counts.active_offers],
              ]}
              title="Inventory"
            />
            <ReportingStatus data={analytics.data} />
            <IngestionSummary data={analytics.data} />
            <UnavailableMetrics />
          </div>
        ) : null}
      </AdminQueryState>
    </>
  )
}

function ReportingStatus({ data }: { data: AnalyticsReportData }) {
  const state =
    data.window.data_state === 'no_data'
      ? 'No data'
      : data.window.data_state === 'insufficient_data'
        ? 'Insufficient data'
        : 'Data available'
  return (
    <section className="border border-ink/10 bg-paper p-5 text-sm leading-6 sm:p-6">
      <h2 className="font-semibold">Validated analytics window</h2>
      <p className="mt-1 text-ink/70">
        {state}. Coverage is {data.window.coverage}; reporting uses the{' '}
        validated and filtered layer from{' '}
        {new Date(data.window.from).toLocaleDateString()} to{' '}
        {new Date(data.window.to).toLocaleDateString()}. Rates remain hidden
        below {data.window.minimum_sample_size} eligible observations.
      </p>
    </section>
  )
}

function AnalyticsSummary({ data }: { data: AnalyticsReportData }) {
  return (
    <section>
      <div className="mb-4 max-w-3xl">
        <h2 className="text-sm font-semibold tracking-[0.15em] uppercase">
          Decision journey
        </h2>
        <p className="mt-2 text-xs leading-5 text-ink/68">
          Completion pairs unique onboarding attempts. Affiliate CTR is the
          share of observed product-detail sessions that also produced a
          countable merchant click for that product. Raw automated and unknown
          requests remain separately auditable.
        </p>
      </div>
      <div className="grid grid-cols-1 gap-px overflow-hidden border border-ink/10 bg-ink/10 sm:grid-cols-2 xl:grid-cols-4">
        <Metric
          dataState={data.window.data_state}
          label="Users"
          value={data.summary.users}
        />
        <Metric
          label="Recommendation sessions"
          value={data.summary.recommendation_sessions}
          dataState={data.window.data_state}
        />
        <Metric
          label="Onboarding starts"
          value={data.summary.onboarding_started}
          dataState={data.window.data_state}
        />
        <Metric
          label="Completed onboarding"
          value={data.summary.onboarding_completed}
          dataState={data.window.data_state}
        />
        <RateMetric
          label="Recommendation completion rate"
          value={data.summary.recommendation_completion_rate}
          dataState={data.window.data_state}
        />
        <Metric
          dataState={data.window.data_state}
          label="Product views"
          value={data.summary.product_views}
        />
        <Metric
          label="Countable affiliate clicks"
          value={data.summary.affiliate_clicks}
          dataState={data.window.data_state}
        />
        <Metric
          label="Raw affiliate requests"
          value={data.summary.affiliate_clicks_raw}
          dataState={data.window.data_state}
        />
        <RateMetric
          dataState={data.window.data_state}
          label="Affiliate CTR"
          value={data.summary.affiliate_ctr}
        />
      </div>
    </section>
  )
}

function IngestionSummary({ data }: { data: AnalyticsReportData }) {
  return (
    <MetricSection
      items={[
        ['Received', data.ingestion.received],
        ['Accepted', data.ingestion.accepted],
        ['Rejected', data.ingestion.rejected],
        ['Privacy filtered', data.ingestion.privacy_filtered],
        ['Bot or prefetch filtered', data.ingestion.bot_filtered],
        ['Deduplicated', data.ingestion.deduplicated],
      ]}
      title="Payload-free ingestion outcomes"
    />
  )
}

function RankingGrid({ data }: { data: AnalyticsReportData }) {
  const rankings: Array<
    [
      string,
      Array<{ id?: string; name?: string; source?: string; count: number }>,
    ]
  > = [
    ['Most recommended products', data.most_recommended_products],
    ['Most viewed products', data.most_viewed_products],
    ['Most clicked products', data.most_clicked_products],
    ['Top merchants by clicks', data.top_merchants],
    ['Top categories by recommendations', data.top_categories],
    ['Traffic sources by session', data.traffic_sources],
  ]
  return (
    <section>
      <h2 className="mb-4 text-sm font-semibold tracking-[0.15em] uppercase">
        Rankings
      </h2>
      <div className="grid gap-4 xl:grid-cols-2">
        {rankings.map(([title, items]) => (
          <article className="border border-ink/10 bg-surface p-5" key={title}>
            <h3 className="font-editorial text-xl">{title}</h3>
            {items.length ? (
              <ol className="mt-4 divide-y divide-ink/8">
                {items.map((item, index) => (
                  <li
                    className="flex items-center gap-3 py-3 text-sm"
                    key={item.id ?? item.source}
                  >
                    <span className="w-5 text-xs text-ink/35">{index + 1}</span>
                    <span className="min-w-0 flex-1 truncate">
                      {item.name ?? item.source}
                    </span>
                    <span className="font-semibold">
                      {item.count.toLocaleString()}
                    </span>
                  </li>
                ))}
              </ol>
            ) : (
              <p className="mt-5 text-sm font-medium text-ink/35">No data</p>
            )}
          </article>
        ))}
      </div>
    </section>
  )
}

function UnavailableMetrics() {
  return (
    <section className="border border-ink/10 bg-paper p-5 sm:p-7">
      <h2 className="font-editorial text-xl">Metrics awaiting verified data</h2>
      <p className="mt-2 max-w-3xl text-sm leading-6 text-ink/70">
        Affiliate conversion rate, commission, EPC, revenue per visitor, revenue
        per recommendation, revenue per user, customer acquisition cost,
        lifetime value, and repeat user rate remain unavailable. They will only
        appear after their provider or lifecycle data is collected and verified.
      </p>
    </section>
  )
}

function MetricSection({
  title,
  items,
}: {
  title: string
  items: Array<[string, number]>
}) {
  return (
    <section>
      <h2 className="mb-4 text-sm font-semibold tracking-[0.15em] uppercase">
        {title}
      </h2>
      <div className="grid grid-cols-1 gap-px overflow-hidden border border-ink/10 bg-ink/10 sm:grid-cols-2 xl:grid-cols-4">
        {items.map(([label, value]) => (
          <Metric key={label} label={label} value={value} />
        ))}
      </div>
    </section>
  )
}

function Metric({
  label,
  value,
  dataState,
}: {
  label: string
  value: number
  dataState?: AnalyticsReportData['window']['data_state']
}) {
  return (
    <div className="bg-surface p-5 sm:p-6">
      <p className="text-xs font-medium text-ink/68">{label}</p>
      <p className="mt-3 font-editorial text-3xl">
        {dataState === 'no_data' ? (
          <span className="text-sm font-medium text-ink/35">No data</span>
        ) : (
          <DataValue value={value} />
        )}
      </p>
    </div>
  )
}

function RateMetric({
  label,
  value,
  dataState,
}: {
  label: string
  value: number | null
  dataState: AnalyticsReportData['window']['data_state']
}) {
  return (
    <div className="bg-surface p-5 sm:p-6">
      <p className="text-xs font-medium text-ink/68">{label}</p>
      <p className="mt-3 font-editorial text-3xl">
        {value === null ? (
          <span className="text-sm font-medium text-ink/35">
            {dataState === 'insufficient_data'
              ? 'Insufficient data'
              : 'No data'}
          </span>
        ) : (
          `${value.toLocaleString(undefined, { maximumFractionDigits: 2 })}%`
        )}
      </p>
    </div>
  )
}
