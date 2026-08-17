import { Check, ListFilter, SlidersHorizontal } from 'lucide-react'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'
const steps = [
  {
    number: '01',
    icon: SlidersHorizontal,
    title: 'Define your constraints',
    copy: 'Share your goal, experience, room dimensions, budget, and equipment you already own.',
  },
  {
    number: '02',
    icon: ListFilter,
    title: 'Compare complete setups',
    copy: 'UNSOLERO weighs fit, quality, price, compatibility, and redundancy across the whole plan.',
  },
  {
    number: '03',
    icon: Check,
    title: 'Buy in the right order',
    copy: 'See what to buy now, what to skip, lower-cost options, and the upgrades worth considering later.',
  },
]

export function MethodSection() {
  return (
    <Section id="method" className="scroll-mt-20" space="lg" surface="paper">
      <Container>
        <div className="grid gap-7 lg:grid-cols-[0.85fr_1.15fr] lg:gap-24">
          <div>
            <p className="eyebrow">How it works</p>
            <Heading className="mt-5 max-w-xl" level={2} size="section">
              From an empty room to a clear plan.
            </Heading>
          </div>
          <p className="max-w-2xl text-lg leading-8 text-ink/70 lg:pt-10 lg:text-xl">
            This is structured decision support—not a chatbot guessing what
            might work. Every recommendation starts with your real constraints.
          </p>
        </div>

        <ol className="mt-14 grid border-l border-t border-ink/15 md:grid-cols-3 lg:mt-20">
          {steps.map(({ number, icon: Icon, title, copy }) => (
            <li
              className="border-b border-r border-ink/15 p-6 sm:p-8 lg:min-h-80 lg:p-10"
              key={number}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold tracking-[0.16em] text-bronze">
                  {number}
                </span>
                <Icon
                  aria-hidden="true"
                  className="text-ink/45"
                  size={22}
                  strokeWidth={1.4}
                />
              </div>
              <h3 className="mt-16 font-display text-2xl font-medium tracking-[-0.04em] lg:mt-24">
                {title}
              </h3>
              <p className="mt-4 text-sm leading-6 text-ink/65">{copy}</p>
            </li>
          ))}
        </ol>
      </Container>
    </Section>
  )
}
