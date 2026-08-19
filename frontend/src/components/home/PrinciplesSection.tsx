import {
  ArrowUpRight,
  CircleDollarSign,
  ListChecks,
  ShieldCheck,
} from 'lucide-react'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'
const principles = [
  {
    number: '01',
    icon: ListChecks,
    title: 'Fit before features',
    copy: 'A great product is still the wrong product when it misses your goals, dimensions, or experience level.',
  },
  {
    number: '02',
    icon: ShieldCheck,
    title: 'Rejections included',
    copy: 'The products we leave out—and why—matter as much as the products that make the final plan.',
  },
  {
    number: '03',
    icon: CircleDollarSign,
    title: 'Commerce stays separate',
    copy: 'Commission never changes objective ranking. Sponsored placement will always be clearly labeled.',
  },
]

export function PrinciplesSection() {
  return (
    <Section id="trust" className="scroll-mt-20" space="lg" surface="ink">
      <Container>
        <div className="flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
          <div>
            <p className="eyebrow text-bronze-soft">Trust, by design</p>
            <Heading className="mt-5 max-w-2xl" level={2} size="section">
              The recommendation earns your confidence.
            </Heading>
          </div>
          <ArrowUpRight
            aria-hidden="true"
            className="hidden text-bronze-soft sm:block"
            size={42}
            strokeWidth={1}
          />
        </div>

        <div className="mt-16 border-t border-canvas/15">
          {principles.map(({ number, icon: Icon, title, copy }) => (
            <article
              className="grid gap-5 border-b border-canvas/15 py-8 sm:grid-cols-[80px_1fr_1.4fr] sm:items-start sm:gap-8 sm:py-10"
              key={number}
            >
              <div className="flex items-center justify-between sm:block">
                <span className="text-xs font-semibold tracking-[0.14em] text-bronze-soft">
                  {number}
                </span>
                <Icon
                  aria-hidden="true"
                  className="text-canvas/45 sm:mt-8"
                  size={22}
                  strokeWidth={1.4}
                />
              </div>
              <h3 className="text-2xl font-medium tracking-[-0.035em]">
                {title}
              </h3>
              <p className="max-w-xl text-base leading-7 text-canvas/75">
                {copy}
              </p>
            </article>
          ))}
        </div>
      </Container>
    </Section>
  )
}
