import { Bookmark, CheckCircle2, Edit3, LockKeyhole } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { ProductComparison } from '../../catalog/components/ProductComparison'
import { useCatalogActions } from '../../catalog/useCatalogActions'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { EmptyState } from '../../../components/ui/EmptyState'
import { Heading } from '../../../components/ui/Heading'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import type { RecommendationResult } from '../schemas'
import { RecommendationProduct } from './RecommendationProduct'
import { useLocalSetups } from '../localSetups'
import { useToast } from '../../../components/ui/useToast'

interface RecommendationResultsProps {
  result: RecommendationResult
  onEdit?: (step: number) => void
  persistedLocally?: boolean
}

export function RecommendationResults({
  result,
  onEdit,
  persistedLocally = false,
}: RecommendationResultsProps) {
  const actions = useCatalogActions()
  const localSetups = useLocalSetups()
  const { showToast } = useToast()
  const [locallySaved, setLocallySaved] = useState(false)
  function saveLocally() {
    try {
      localSetups.save(result)
      setLocallySaved(true)
      showToast({
        title: 'Setup saved on this device',
        description: 'You can reopen it from Saved setups.',
        variant: 'success',
      })
    } catch {
      showToast({
        title: 'Setup could not be saved',
        description: 'Browser storage is unavailable or full.',
        variant: 'error',
      })
    }
  }
  if (
    result.status === 'no_suitable_products' ||
    result.recommended_products.length === 0
  ) {
    return (
      <div className="py-16 sm:py-24">
        <EmptyState
          action={
            onEdit ? (
              <Button onClick={() => onEdit(3)}>Edit budget or space</Button>
            ) : (
              <ButtonLink to="/build">Start a new brief</ButtonLink>
            )
          }
          description="The deterministic engine found no complete, compatible setup inside every hard constraint. We will not recommend a poor fit just to fill the page."
          title="No responsible setup found"
        />
      </div>
    )
  }
  return (
    <div className="pb-20 pt-10 sm:pb-28 sm:pt-16">
      <div className="max-w-4xl">
        <p className="eyebrow">Your decision brief</p>
        <Heading className="mt-5" level={1} size="display">
          Your Personalized Home Gym
        </Heading>
        <p className="mt-6 max-w-2xl text-base leading-7 text-ink/65 sm:text-lg">
          A complete setup selected from structured product facts, your
          constraints, and a deterministic scoring policy.
        </p>
      </div>

      <div className="mt-10 grid border border-ink/15 bg-surface sm:grid-cols-2 lg:grid-cols-5">
        <Metric label="Total cost">
          <PriceDisplay
            amountMinor={result.total_cost.amount_minor}
            currency={result.total_cost.currency}
            size="lg"
          />
        </Metric>
        <Metric
          label="Recommendation"
          value={`${result.recommendation_score}/100`}
        />
        <Metric label="Goal fit" value={`${result.fit.goal_match}/100`} />
        <Metric label="Space fit" value={`${result.fit.space_match}/100`} />
        <Metric label="Budget fit" value={`${result.fit.budget_match}/100`} />
      </div>

      <div className="mt-6 flex flex-wrap items-center justify-between gap-4 border-y border-ink/10 py-4 text-sm">
        <p className="flex items-center gap-2 text-ink/65">
          {result.saved || persistedLocally ? (
            <CheckCircle2 aria-hidden="true" className="text-moss" size={18} />
          ) : (
            <LockKeyhole aria-hidden="true" size={18} />
          )}
          {result.saved
            ? 'Saved to your account and available to revisit.'
            : persistedLocally
              ? 'Saved in this browser and available to revisit on this device.'
              : 'This guest result is not stored on the server.'}
        </p>
        <div className="flex gap-2">
          {!result.saved && !persistedLocally && !locallySaved && (
            <Button onClick={saveLocally} size="sm" variant="secondary">
              <Bookmark aria-hidden="true" size={15} /> Save setup
            </Button>
          )}
          {!result.saved && (
            <ButtonLink
              state={{ from: '/build' }}
              to="/login"
              size="sm"
              variant="quiet"
            >
              Sign in to sync
            </ButtonLink>
          )}
          {locallySaved && (
            <ButtonLink size="sm" to="/setups" variant="secondary">
              View saved setups
            </ButtonLink>
          )}
          {onEdit && (
            <Button onClick={() => onEdit(0)} size="sm" variant="quiet">
              <Edit3 aria-hidden="true" size={15} /> Edit answers
            </Button>
          )}
        </div>
      </div>

      <div className="mt-12 space-y-6">
        {result.recommended_products.map((item) => (
          <RecommendationProduct
            alternative={result.alternatives.find(
              (alternative) => alternative.for_product_id === item.product.id,
            )}
            compared={actions.comparedIDs.has(item.product.id)}
            item={item}
            key={item.product.id}
            recommendationID={result.recommendation_id}
            merchantSource={
              result.saved || persistedLocally ? 'setup' : 'recommendation'
            }
            onCompare={actions.compare}
            onSave={actions.save}
            savePending={actions.savePending}
            saved={actions.savedIDs.has(item.product.id)}
          />
        ))}
      </div>

      {result.alternatives.length > 0 && (
        <section className="mt-20" aria-labelledby="alternatives-heading">
          <p className="eyebrow">Trade-offs</p>
          <Heading className="mt-4" id="alternatives-heading" size="title">
            Considered alternatives
          </Heading>
          <div className="mt-7 grid gap-px bg-ink/15 md:grid-cols-2">
            {result.alternatives.map((alternative) => (
              <article
                className="bg-surface p-5 sm:p-6"
                key={`${alternative.for_product_id}-${alternative.product.id}`}
              >
                <Badge
                  variant={
                    alternative.type === 'cheaper' ? 'success' : 'accent'
                  }
                >
                  {alternative.type}
                </Badge>
                <h3 className="mt-4 font-display text-2xl font-medium tracking-[-0.04em]">
                  <Link to={`/products/${alternative.product.slug}`}>
                    {alternative.product.name}
                  </Link>
                </h3>
                <p className="mt-3 text-sm leading-6 text-ink/60">
                  {alternative.reasons[0]?.message ??
                    'A viable trade-off for the same role.'}
                </p>
                <PriceDisplay
                  className="mt-5"
                  amountMinor={alternative.product.price.amount_minor}
                  currency={alternative.product.price.currency}
                />
              </article>
            ))}
          </div>
        </section>
      )}

      <section
        className="mt-20 border-t border-ink/15 pt-10"
        aria-labelledby="rejected-heading"
      >
        <Heading id="rejected-heading" size="title">
          Products we deliberately rejected
        </Heading>
        <p className="mt-3 max-w-2xl text-sm leading-6 text-ink/60">
          Transparency includes what did not make the setup. These are
          deterministic constraint and redundancy decisions, not negative
          reviews.
        </p>
        {result.rejected_alternatives.length ? (
          <div className="mt-6 divide-y divide-ink/10 border-y border-ink/10">
            {result.rejected_alternatives.slice(0, 8).map((item) => (
              <div
                className="grid gap-2 py-4 sm:grid-cols-[minmax(0,0.8fr)_1.2fr]"
                key={item.product.id}
              >
                <Link
                  className="font-semibold"
                  to={`/products/${item.product.slug}`}
                >
                  {item.product.name}
                </Link>
                <p className="text-sm text-ink/60">{item.reason}</p>
              </div>
            ))}
          </div>
        ) : (
          <p className="mt-6 text-sm text-ink/55">
            No additional catalog products needed an explicit rejection.
          </p>
        )}
      </section>

      <ProductComparison
        error={actions.comparisonError}
        loading={actions.comparisonPending}
        onOpenChange={actions.setComparisonOpen}
        onRemove={actions.compare}
        open={actions.comparisonOpen}
        products={actions.compared}
      />
      {actions.comparedIDs.size > 0 && !actions.comparisonOpen && (
        <Button
          className="fixed bottom-5 right-5 z-30 shadow-overlay"
          onClick={() => actions.setComparisonOpen(true)}
        >
          Compare {actions.comparedIDs.size} products
        </Button>
      )}
    </div>
  )
}

function Metric({
  label,
  value,
  children,
}: {
  label: string
  value?: string
  children?: React.ReactNode
}) {
  return (
    <div className="border-b border-ink/10 p-5 last:border-b-0 sm:p-6 lg:border-b-0 lg:border-r lg:last:border-r-0">
      <p className="text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/45">
        {label}
      </p>
      <div className="mt-2 font-display text-3xl font-medium tracking-[-0.04em]">
        {children ?? value}
      </div>
    </div>
  )
}
