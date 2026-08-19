import { ArrowLeft, ArrowRight, Check, CircleUserRound } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { RecommendationResults } from '../features/recommendation/components/RecommendationResults'
import { RecommendationStep } from '../features/recommendation/components/RecommendationStep'
import { startOnboarding } from '../features/analytics/tracking'
import { steps } from '../features/recommendation/options'
import { useRecommendationBuilder } from '../features/recommendation/useRecommendationBuilder'
import { recommendationInputSchema } from '../features/recommendation/schemas'
import { usePageMetadata } from '../lib/seo/usePageMetadata'
import { cn } from '../lib/styles/cn'

export function RecommendationBuilderPage() {
  const location = useLocation()
  const state = location.state as { editInput?: unknown } | null
  const parsedInput = recommendationInputSchema.safeParse(state?.editInput)
  const builder = useRecommendationBuilder(
    parsedInput.success ? parsedInput.data : undefined,
  )
  const hasPresentedStep = useRef(false)
  useEffect(() => {
    startOnboarding()
  }, [])
  useEffect(() => {
    if (!hasPresentedStep.current) {
      hasPresentedStep.current = true
      return
    }
    document.getElementById('recommendation-step-title')?.focus()
  }, [builder.currentStep])
  usePageMetadata({
    title: 'Build Your Software Stack | UNSOLERO',
    description:
      'Build a software stack around what your business does, the tools you already run, and your monthly budget.',
    robots: 'noindex, follow',
  })

  if (builder.result) {
    return (
      <>
        <SiteHeader position="static" />
        <main id="main-content">
          <Container>
            <RecommendationResults
              onEdit={builder.edit}
              result={builder.result}
            />
          </Container>
        </main>
        <SiteFooter />
      </>
    )
  }

  const step = steps[builder.currentStep] ?? steps[0]
  return (
    <>
      <SiteHeader
        actions={
          <ButtonLink
            size="sm"
            to={builder.account.data ? '/account' : '/login'}
            variant="quiet"
          >
            <CircleUserRound aria-hidden="true" size={16} />
            {builder.account.data ? 'Account' : 'Sign in'}
          </ButtonLink>
        }
        position="static"
        showNavigation={false}
      />
      <main className="min-h-[calc(100vh-5rem)]" id="main-content">
        <Container className="py-8 sm:py-12 lg:py-16">
          <div className="grid gap-10 lg:grid-cols-[15rem_minmax(0,1fr)] xl:gap-16">
            <aside
              aria-label="Recommendation progress"
              className="lg:sticky lg:top-8 lg:self-start"
            >
              <div className="flex items-center justify-between">
                <p className="eyebrow">Your brief</p>
                <span className="text-xs font-semibold text-ink/50">
                  {builder.currentStep + 1} / {steps.length}
                </span>
              </div>
              <div
                aria-label="Recommendation progress"
                aria-valuemax={steps.length}
                aria-valuemin={1}
                aria-valuenow={builder.currentStep + 1}
                aria-valuetext={`Question ${builder.currentStep + 1} of ${steps.length}`}
                className="mt-4 h-1 bg-ink/10"
                role="progressbar"
              >
                <div
                  className="h-full bg-bronze transition-[width]"
                  style={{
                    width: `${((builder.currentStep + 1) / steps.length) * 100}%`,
                  }}
                />
              </div>
              <ol className="mt-6 hidden space-y-1 lg:block">
                {steps.map((item, index) => (
                  <li key={item.label}>
                    <button
                      aria-current={
                        index === builder.currentStep ? 'step' : undefined
                      }
                      className={cn(
                        'flex w-full items-center gap-3 py-2 text-left text-sm text-ink/40',
                        index === builder.currentStep &&
                          'font-semibold text-ink',
                        index < builder.currentStep &&
                          'text-ink/65 hover:text-ink',
                      )}
                      disabled={index > builder.currentStep}
                      onClick={() => builder.edit(index)}
                      type="button"
                    >
                      <span
                        className={cn(
                          'grid size-6 place-items-center rounded-full border border-ink/20 text-[0.625rem]',
                          index <= builder.currentStep && 'border-ink',
                        )}
                      >
                        {index < builder.currentStep ? (
                          <Check aria-hidden="true" size={12} />
                        ) : (
                          index + 1
                        )}
                      </span>
                      {item.label}
                    </button>
                  </li>
                ))}
              </ol>
              <p className="mt-6 text-xs leading-5 text-ink/50" role="status">
                {builder.account.isPending
                  ? 'Checking for an account and saved progress…'
                  : builder.account.isError
                    ? 'Your account could not be checked. Answers remain in this browser.'
                    : builder.account.data
                      ? builder.draft.isPending
                        ? 'Loading your saved progress…'
                        : builder.draft.isError
                          ? 'Saved progress is unavailable. Answers remain in this browser.'
                          : builder.saveDraft.isError
                            ? 'Your latest change could not be saved. Your answers remain in this browser.'
                            : builder.saveDraft.isPending
                              ? 'Saving progress…'
                              : 'Progress saves securely to your account.'
                      : 'Guest progress stays in this browser session. Sign in to save it to your account.'}
              </p>
            </aside>

            <form
              className="min-w-0"
              noValidate
              onSubmit={(event) => {
                event.preventDefault()
                void builder.next()
              }}
            >
              <p className="eyebrow">Question {builder.currentStep + 1}</p>
              <Heading
                className="mt-4 max-w-4xl focus:outline-none"
                id="recommendation-step-title"
                level={1}
                size="section"
                tabIndex={-1}
              >
                {step.title}
              </Heading>
              <p className="mt-5 max-w-2xl text-sm leading-6 text-ink/60 sm:text-base sm:leading-7">
                {step.supporting}
              </p>

              <fieldset className="mt-9 border-0 p-0 sm:mt-12">
                <legend className="sr-only">{step.title}</legend>
                <RecommendationStep
                  form={builder.form}
                  step={builder.currentStep}
                />
              </fieldset>

              {builder.stepError && (
                <p
                  className="mt-6 border-l-2 border-ember bg-ember-soft px-4 py-3 text-sm text-ember"
                  role="alert"
                >
                  {builder.stepError}
                </p>
              )}
              {builder.generate.isError && (
                <div className="mt-6">
                  <ErrorState
                    compact
                    description="Your answers are intact. Please try again."
                    title="We could not build your setup"
                  />
                </div>
              )}

              <div className="mt-10 flex flex-col-reverse gap-3 border-t border-ink/15 pt-6 sm:flex-row sm:items-center sm:justify-between">
                <Button
                  disabled={
                    builder.currentStep === 0 || builder.generate.isPending
                  }
                  onClick={builder.back}
                  variant="quiet"
                >
                  <ArrowLeft aria-hidden="true" size={16} /> Back
                </Button>
                <Button
                  loading={builder.generate.isPending}
                  loadingLabel="Building your setup…"
                  type="submit"
                >
                  {builder.currentStep === steps.length - 1
                    ? 'Build my setup'
                    : 'Next'}
                  <ArrowRight aria-hidden="true" size={16} />
                </Button>
              </div>
            </form>
          </div>
        </Container>
      </main>
    </>
  )
}
