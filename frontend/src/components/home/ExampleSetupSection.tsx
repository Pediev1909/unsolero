import { ArrowRight, Check, Minus, Plus } from 'lucide-react'

import { Badge } from '../ui/Badge'
import { ButtonLink } from '../ui/ButtonLink'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { PriceDisplay } from '../ui/PriceDisplay'
import { Section } from '../ui/Section'
import { exampleSetup } from './homeData'

const profileItems = [
  ['Goal', exampleSetup.profile.goal],
  ['Team', exampleSetup.profile.teamSize],
  ['Experience', exampleSetup.profile.experience],
  ['Already running', exampleSetup.profile.owned],
] as const

export function ExampleSetupSection() {
  return (
    <Section
      className="scroll-mt-20"
      id="example-setup"
      space="lg"
      surface="ink"
    >
      <Container>
        <div className="grid gap-12 lg:grid-cols-[0.72fr_1.28fr] lg:gap-20">
          <div>
            <p className="eyebrow text-bronze-soft">Example plan</p>
            <Heading className="mt-5 max-w-lg" level={2} size="section">
              A working stack—not a pile of subscriptions.
            </Heading>
            <p className="mt-7 max-w-md text-base leading-7 text-canvas/65">
              One plan for the brief below, drawn from the live catalog. The
              engine produces it from your own answers, not from ours.
            </p>

            <dl className="mt-10 border-t border-canvas/15">
              {profileItems.map(([term, value]) => (
                <div
                  className="flex justify-between gap-6 border-b border-canvas/15 py-4 text-sm"
                  key={term}
                >
                  <dt className="text-canvas/45">{term}</dt>
                  <dd className="text-right font-medium">{value}</dd>
                </div>
              ))}
              <div className="flex justify-between gap-6 border-b border-canvas/15 py-4 text-sm">
                <dt className="text-canvas/45">Budget</dt>
                <dd className="font-medium">
                  <PriceDisplay
                    amountMinor={exampleSetup.profile.budgetMinor}
                    className="text-canvas"
                    currency="USD"
                    size="sm"
                  />
                </dd>
              </div>
            </dl>
          </div>

          <div className="bg-surface text-ink">
            <div className="flex flex-col gap-4 border-b border-ink/15 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-7">
              <div>
                <Badge variant="success">Within budget</Badge>
                <h3 className="mt-3 font-display text-2xl font-medium tracking-[-0.04em] sm:text-3xl">
                  Client services foundation
                </h3>
              </div>
              <div className="sm:text-right">
                <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/45">
                  Setup total
                </p>
                <PriceDisplay
                  amountMinor={exampleSetup.totalMinor}
                  className="mt-1"
                  currency="USD"
                  size="lg"
                />
              </div>
            </div>

            <ol>
              {exampleSetup.items.map((item, index) => (
                <li
                  className="grid grid-cols-[2rem_1fr] gap-3 border-b border-ink/15 p-5 sm:grid-cols-[2rem_1fr_auto] sm:items-center sm:gap-5 sm:p-7"
                  key={item.name}
                >
                  <span className="grid size-8 place-items-center rounded-full bg-moss-soft text-moss">
                    <Check aria-hidden="true" size={15} strokeWidth={2} />
                    <span className="sr-only">Included</span>
                  </span>
                  <div>
                    <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-bronze">
                      Buy {index + 1}
                    </p>
                    <h4 className="mt-1 font-medium">{item.name}</h4>
                    <p className="mt-2 max-w-xl text-sm leading-6 text-ink/60">
                      {item.reason}
                    </p>
                  </div>
                  <PriceDisplay
                    amountMinor={item.priceMinor}
                    className="col-start-2 sm:col-start-auto"
                    currency="USD"
                    size="sm"
                  />
                </li>
              ))}
            </ol>

            <div className="grid md:grid-cols-2">
              <div className="border-b border-ink/15 p-5 md:border-b-0 md:border-r sm:p-7">
                <div className="flex items-center gap-2 text-ember">
                  <Minus aria-hidden="true" size={16} />
                  <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em]">
                    Deliberately left out
                  </p>
                </div>
                <p className="mt-3 text-sm leading-6 text-ink/65">
                  {exampleSetup.rejected}
                </p>
              </div>
              <div className="p-5 sm:p-7">
                <div className="flex items-center gap-2 text-bronze-dark">
                  <Plus aria-hidden="true" size={16} />
                  <p className="text-[0.625rem] font-bold uppercase tracking-[0.14em]">
                    Upgrade later
                  </p>
                </div>
                <p className="mt-3 text-sm leading-6 text-ink/65">
                  {exampleSetup.upgrade}
                </p>
              </div>
            </div>

            <div className="flex flex-col gap-4 border-t border-ink/15 bg-paper p-5 sm:flex-row sm:items-center sm:justify-between sm:p-7">
              <p className="text-sm text-ink/65">
                <span className="font-semibold text-ink">
                  <PriceDisplay
                    amountMinor={exampleSetup.remainingMinor}
                    currency="USD"
                    size="sm"
                  />
                </span>{' '}
                stays unspent.
              </p>
              <ButtonLink size="sm" to="/build">
                Build My Setup
                <ArrowRight aria-hidden="true" size={15} />
              </ButtonLink>
            </div>
          </div>
        </div>
      </Container>
    </Section>
  )
}
