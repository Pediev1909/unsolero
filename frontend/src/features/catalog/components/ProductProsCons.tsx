import { AlertTriangle, CheckCircle2, Target } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '../../../lib/styles/cn'
import type { ProductInsight } from '../schemas'

/**
 * Strengths against trade-offs, side by side.
 *
 * They were stacked — strengths, then trade-offs, then use cases, each a list
 * of "label 82/100" — which reads as one long list under three headings. Two
 * columns is how every reader already expects pros and cons to arrive, and it
 * puts a strength in the same eye-line as what it costs. The scores are
 * unchanged and still printed; the bar under each is the same number drawn, so
 * a column can be skimmed without reading every figure.
 */
export function ProductProsCons({
  strengths,
  weaknesses,
  useCases,
}: {
  strengths: ProductInsight[]
  weaknesses: ProductInsight[]
  useCases: ProductInsight[]
}) {
  return (
    <div>
      <div className="grid gap-10 md:grid-cols-2 md:gap-14">
        <InsightColumn
          empty="No standout strength crossed the current evidence threshold."
          icon={<CheckCircle2 aria-hidden="true" size={19} />}
          insights={strengths}
          title="Strengths"
          tone="strength"
        />
        <InsightColumn
          empty="No material weakness crossed the current evidence threshold."
          icon={<AlertTriangle aria-hidden="true" size={19} />}
          insights={weaknesses}
          title="Trade-offs"
          tone="trade-off"
        />
      </div>
      <div className="mt-10 md:mt-14">
        <InsightColumn
          empty="No use case crossed the current evidence threshold."
          icon={<Target aria-hidden="true" size={19} />}
          insights={useCases}
          title="Best use cases"
          tone="strength"
        />
      </div>
      <p className="mt-8 text-xs leading-5 text-ink/65">
        Scores are derived from structured catalog facts. They are not customer
        ratings or reviews.
      </p>
    </div>
  )
}

// A trade-off scored 80 is a strong weakness, so its bar is drawn in the
// warning colour rather than the accent that marks a strength of 80.
const meterTone = {
  strength: 'bg-bronze',
  'trade-off': 'bg-amber',
} as const

function InsightColumn({
  icon,
  insights,
  title,
  empty,
  tone,
}: {
  icon: ReactNode
  insights: ProductInsight[]
  title: string
  empty: string
  tone: keyof typeof meterTone
}) {
  return (
    <div>
      <h3 className="flex items-center gap-2 text-sm font-semibold">
        {icon}
        {title}
      </h3>
      {insights.length ? (
        <ul className="mt-4 space-y-4">
          {insights.map((insight) => (
            <li key={insight.key}>
              <div className="flex items-baseline justify-between gap-4 text-sm">
                <span className="text-ink/80">{insight.label}</span>
                <span className="font-semibold text-ink tabular-nums">
                  {insight.score}/100
                </span>
              </div>
              <span
                aria-hidden="true"
                className="mt-2 block h-1 w-full bg-ink/10"
              >
                <span
                  className={cn('block h-full', meterTone[tone])}
                  style={{ width: `${insight.score}%` }}
                />
              </span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-3 text-sm leading-6 text-ink/68">{empty}</p>
      )}
    </div>
  )
}
