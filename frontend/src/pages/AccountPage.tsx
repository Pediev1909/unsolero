import { ArrowRight, Layers3, LogOut } from 'lucide-react'
import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { ButtonLink } from '../components/ui/ButtonLink'
import { ErrorState } from '../components/ui/ErrorState'
import { Skeleton } from '../components/ui/Skeleton'
import { useCurrentUser, useLogout } from '../features/auth/queries'
import { SecuritySettings } from '../features/auth/components/SecuritySettings'
import { useSetups } from '../features/recommendation/queries'
import { formatMinorCurrency } from '../lib/money/format'

export function AccountPage() {
  const account = useCurrentUser()
  const logout = useLogout()
  const setups = useSetups()

  return (
    <>
      <SiteHeader
        position="static"
        showNavigation={false}
        actions={
          <Button
            loading={logout.isPending}
            loadingLabel="Signing out…"
            onClick={() => logout.mutate()}
            size="sm"
            variant="quiet"
          >
            <LogOut aria-hidden="true" size={16} />
            Sign out
          </Button>
        }
      />

      <main className="min-h-[calc(100vh-5rem)]" id="main-content">
        <Container className="py-16 sm:py-24">
          <p className="eyebrow">Account</p>
          <Heading className="mt-5" level={1} size="display">
            Your UNSOLERO.
          </Heading>
          <div className="mt-12 max-w-2xl border-t border-ink/15 pt-8">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-ink/45">
              Signed in as
            </p>
            <p className="mt-3 text-lg">{account.data?.email}</p>
            {account.data && account.data.roles.length > 0 && (
              <ButtonLink className="mt-5" size="sm" to="/admin">
                Open administration <ArrowRight aria-hidden="true" size={15} />
              </ButtonLink>
            )}
          </div>

          <section aria-labelledby="setups-heading" className="mt-16 max-w-4xl">
            <div className="flex flex-wrap items-end justify-between gap-4 border-b border-ink/15 pb-5">
              <div>
                <p className="eyebrow">Decision history</p>
                <Heading
                  className="mt-3"
                  id="setups-heading"
                  level={2}
                  size="title"
                >
                  Saved setups
                </Heading>
              </div>
              <ButtonLink size="sm" to="/build">
                Build a new setup <ArrowRight aria-hidden="true" size={15} />
              </ButtonLink>
            </div>

            {setups.isPending && (
              <div
                aria-label="Loading saved setups"
                className="mt-6 space-y-3"
                role="status"
              >
                <Skeleton className="h-28 w-full" />
                <Skeleton className="h-28 w-full" />
              </div>
            )}
            {setups.isError && (
              <div className="mt-6">
                <ErrorState
                  compact
                  description="Your saved setups could not be loaded."
                  onRetry={() => void setups.refetch()}
                  title="Setups unavailable"
                />
              </div>
            )}
            {setups.data?.setups.length === 0 && (
              <div className="mt-6 border border-ink/15 bg-surface p-6 sm:p-8">
                <Layers3
                  aria-hidden="true"
                  className="text-bronze"
                  size={24}
                />
                <h3 className="mt-4 text-lg font-semibold">
                  No saved setups yet
                </h3>
                <p className="mt-2 max-w-xl text-sm leading-6 text-ink/60">
                  Complete the decision brief to save a personalized product
                  plan and every trade-off behind it.
                </p>
              </div>
            )}
            {setups.data && setups.data.setups.length > 0 && (
              <div className="mt-6 divide-y divide-ink/10 border-y border-ink/10">
                {setups.data.setups.map((setup) => (
                  <Link
                    className="grid gap-4 py-5 hover:text-bronze-dark sm:grid-cols-[1fr_auto] sm:items-center"
                    key={setup.id}
                    to={`/setups/${setup.id}`}
                  >
                    <div>
                      <h3 className="font-display text-2xl font-medium tracking-[-0.035em]">
                        {setup.name}
                      </h3>
                      <p className="mt-2 text-xs text-ink/50">
                        {setup.item_count}{' '}
                        {setup.item_count === 1 ? 'product' : 'products'} ·{' '}
                        {setup.recommendation_score}/100 match · Saved{' '}
                        {new Intl.DateTimeFormat('en-US', {
                          dateStyle: 'medium',
                        }).format(new Date(setup.created_at))}
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
          </section>

          {account.data && <SecuritySettings user={account.data} />}

          {logout.isError && (
            <div
              className="mt-8 max-w-2xl border-l-2 border-bronze bg-paper px-4 py-3 text-sm"
              role="alert"
            >
              We could not sign you out. Please try again.
            </div>
          )}
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
