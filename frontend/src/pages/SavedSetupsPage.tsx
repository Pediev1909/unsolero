import { ArrowRight, Dumbbell } from 'lucide-react'
import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { EmptyState } from '../components/ui/EmptyState'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { useSavedSetups } from '../features/recommendation/useSavedSetups'
import { formatMinorCurrency } from '../lib/money/format'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function SavedSetupsPage() {
  const saved = useSavedSetups()
  usePageMetadata({
    title: 'Saved Home Gym Setups | UNSOLERO',
    description: 'Reopen and manage personalized home gym plans.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <Container className="py-14 sm:py-20">
          <div className="flex flex-wrap items-end justify-between gap-5">
            <div>
              <p className="eyebrow">Decision history</p>
              <Heading className="mt-5" level={1} size="display">
                Saved setups.
              </Heading>
              <p className="mt-5 max-w-2xl text-ink/60">
                {saved.authenticated
                  ? 'These plans are securely saved to your account.'
                  : 'Guest plans stay only in this browser. Sign in before switching devices.'}
              </p>
            </div>
            <ButtonLink to="/build">Build a new setup</ButtonLink>
          </div>
          <div className="mt-10">
            {saved.isPending && <LoadingState title="Loading saved setups" />}
            {saved.isError && saved.setups.length === 0 && (
              <ErrorState
                description="Your saved setups could not be loaded."
                onRetry={() => void saved.refetch()}
                title="Setups unavailable"
              />
            )}
            {!saved.isPending && saved.setups.length === 0 && (
              <EmptyState
                action={
                  <ButtonLink to="/build">Build your first setup</ButtonLink>
                }
                description="Complete the recommendation brief, then save the result here."
                title="No saved setups yet"
              />
            )}
            {saved.setups.length > 0 && (
              <div className="divide-y divide-ink/10 border-y border-ink/10">
                {saved.setups.map((setup) => (
                  <Link
                    className="grid gap-4 py-6 hover:text-bronze-dark sm:grid-cols-[1fr_auto] sm:items-center"
                    key={setup.id}
                    to={`/setups/${setup.id}`}
                  >
                    <div>
                      <div className="flex items-center gap-3">
                        <Dumbbell
                          aria-hidden="true"
                          className="text-bronze"
                          size={18}
                        />
                        <h2 className="font-display text-2xl font-medium">
                          {setup.name}
                        </h2>
                      </div>
                      <p className="mt-2 text-xs text-ink/50">
                        {setup.item_count}{' '}
                        {setup.item_count === 1 ? 'product' : 'products'} ·{' '}
                        {setup.recommendation_score}/100 match · Updated{' '}
                        {new Intl.DateTimeFormat('en-US', {
                          dateStyle: 'medium',
                        }).format(new Date(setup.updated_at))}
                      </p>
                    </div>
                    <span className="flex items-center gap-3 font-semibold">
                      {formatMinorCurrency(
                        setup.total_cost.amount_minor,
                        setup.total_cost.currency,
                      )}
                      <ArrowRight aria-hidden="true" size={16} />
                    </span>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
