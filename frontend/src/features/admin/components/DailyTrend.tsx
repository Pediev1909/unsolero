import type { AnalyticsReportData } from '../schemas'

// Two charts rather than one with two lines. Views and clicks differ by an
// order of magnitude, and putting them on a shared axis would flatten clicks
// into the baseline; putting them on two axes would let the ratio between them
// be set by the scaling rather than the data. Each chart carries a single
// series, so its heading names it and no legend is needed.
//
// One hue for both, because identity comes from the heading. Two nearby greens
// were measured at ΔE 3.4 for normal vision — indistinguishable — so a second
// colour would have added confusion rather than information.

type Point = AnalyticsReportData['daily'][number]

export function DailyTrend({ daily }: { daily: Point[] }) {
  if (daily.length === 0) return null

  return (
    <section className="mt-10">
      <h2 className="font-editorial text-xl">Last {daily.length} days</h2>
      <p className="mt-2 max-w-2xl text-body-sm text-ink/70">
        A total says how many. Only the shape says whether it is growing.
      </p>
      <div className="mt-5 grid gap-5 lg:grid-cols-2">
        <TrendChart
          label="Product views"
          points={daily}
          value={(point) => point.product_views}
        />
        <TrendChart
          label="Affiliate clicks"
          points={daily}
          value={(point) => point.affiliate_clicks}
        />
      </div>
    </section>
  )
}

function TrendChart({
  label,
  points,
  value,
}: {
  label: string
  points: Point[]
  value: (point: Point) => number
}) {
  const total = points.reduce((sum, point) => sum + value(point), 0)
  // A single flat baseline is the honest picture of no activity. Scaling to a
  // maximum of zero would divide by zero; scaling to 1 would turn a single
  // event into a full-height bar on an otherwise empty chart.
  const peak = Math.max(1, ...points.map(value))
  const first = points[0]
  const last = points[points.length - 1]
  if (!first || !last) return null

  return (
    <figure className="border border-ink/15 bg-surface p-5">
      <figcaption className="flex items-baseline justify-between gap-4">
        <span className="text-label uppercase tracking-[0.12em] text-ink/70">
          {label}
        </span>
        <span className="font-numeric text-2xl tabular-nums">
          {total.toLocaleString()}
        </span>
      </figcaption>

      <div
        aria-label={`${label} per day. ${total.toLocaleString()} in total, peaking at ${peak.toLocaleString()}.`}
        className="mt-5 flex h-24 items-end gap-[2px]"
        role="img"
      >
        {points.map((point) => {
          const count = value(point)
          return (
            <div
              className="group relative flex h-full flex-1 items-end"
              key={point.day}
            >
              {/* A zero day keeps a hairline so the day is still visibly
                  present in the series rather than a gap in it. */}
              <div
                className="w-full rounded-t-[4px] bg-bronze transition-colors group-hover:bg-bronze-dark"
                style={{
                  height: count === 0 ? '2px' : `${(count / peak) * 100}%`,
                  opacity: count === 0 ? 0.25 : 1,
                }}
              />
              <div
                className="pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 hidden -translate-x-1/2 whitespace-nowrap border border-ink/15 bg-charcoal px-2 py-1 text-caption text-canvas group-hover:block"
                role="tooltip"
              >
                {formatDay(point.day)} · {count.toLocaleString()}
              </div>
            </div>
          )
        })}
      </div>

      <div className="mt-2 flex justify-between text-caption text-ink/65">
        <span>{formatDay(first.day)}</span>
        <span>{formatDay(last.day)}</span>
      </div>
    </figure>
  )
}

function formatDay(day: string) {
  const parsed = new Date(`${day}T00:00:00Z`)
  if (Number.isNaN(parsed.getTime())) return day
  return parsed.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    timeZone: 'UTC',
  })
}
