import { Link } from 'react-router-dom'

import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import type { RecommendationResult } from '../schemas'

interface LivePreviewProps {
  result: RecommendationResult | undefined
  isPending: boolean
  isError: boolean
}

/**
 * The answer, while the questions are still being answered.
 *
 * Six forms and then a result is a bad trade to offer a stranger: they spend
 * the effort first and find out whether it was worth it last. Showing a real
 * set of tools after the second question inverts that. Every question after it
 * is then an improvement to something visible rather than another toll gate,
 * and people answer more of those, not fewer.
 */
export function LivePreview({ result, isPending, isError }: LivePreviewProps) {
  if (isError) return null

  const products = result?.recommended_products ?? []

  return (
    <section
      aria-label="Live suggestion"
      className="mt-8 border-t border-ink/15 pt-6"
    >
      <div className="flex items-baseline justify-between gap-3">
        <p className="eyebrow">So far we&rsquo;d suggest</p>
        {isPending && (
          <span className="text-xs text-ink/50" role="status">
            Updating…
          </span>
        )}
      </div>

      {/* Only when an answer came back and held nothing. Showing this while
          the first request is still in flight told people their budget was
          too small when in truth nobody had asked yet. */}
      {result && products.length === 0 && (
        <p className="mt-3 text-sm leading-6 text-ink/65">
          Nothing fits inside that budget yet. Raising it, or telling us what
          you already run, usually opens things up.
        </p>
      )}

      {!result && (
        <p className="mt-3 text-sm leading-6 text-ink/55">
          Working out a first suggestion&hellip;
        </p>
      )}

      {products.length > 0 && (
        <>
          <ul className="mt-4 flex flex-col gap-2.5">
            {products.map((item) => (
              <li
                className="flex items-baseline justify-between gap-3 text-sm"
                key={item.product.id}
              >
                <Link
                  className="font-medium underline-offset-4 hover:underline"
                  to={`/products/${item.product.slug}`}
                >
                  {item.product.name}
                </Link>
                <PriceDisplay
                  amountMinor={item.product.price.amount_minor}
                  className="shrink-0"
                  currency={item.product.price.currency}
                  size="sm"
                />
              </li>
            ))}
          </ul>

          {result && (
            <p className="mt-4 flex items-baseline justify-between gap-3 border-t border-ink/10 pt-3 text-sm font-semibold">
              <span>Per month</span>
              <PriceDisplay
                amountMinor={result.total_cost.amount_minor}
                className="shrink-0"
                currency={result.total_cost.currency}
                size="sm"
              />
            </p>
          )}

          <p className="mt-4 text-xs leading-5 text-ink/60">
            This updates with every answer. Keep going and it gets closer to
            your situation.
          </p>
        </>
      )}
    </section>
  )
}
