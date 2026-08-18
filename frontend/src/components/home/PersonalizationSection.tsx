import { Boxes, Plug, WalletCards } from 'lucide-react'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'

const reasons = [
  {
    icon: Plug,
    title: 'What you already run changes the answer',
    copy: 'A tool that does not connect to your stack costs somebody an afternoon a week, which is more than either subscription.',
  },
  {
    icon: WalletCards,
    title: 'Your budget needs a sequence',
    copy: 'A good plan covers what stops work or stops payment first, and names the tools that can genuinely wait.',
  },
  {
    icon: Boxes,
    title: 'Overlap is money you already spend',
    copy: 'Suites grow into each other. What you own should remove options, not sit unnoticed beside a second tool doing the same job.',
  },
]

export function PersonalizationSection() {
  return (
    <Section space="lg">
      <Container>
        <div className="grid gap-12 lg:grid-cols-[1fr_1.3fr] lg:gap-24">
          <div>
            <p className="eyebrow">Why personalized</p>
            <Heading className="mt-5 max-w-xl" level={2} size="section">
              There is no universal best stack.
            </Heading>
          </div>
          <div className="border-t border-ink/15">
            {reasons.map(({ icon: Icon, title, copy }) => (
              <article
                className="grid gap-5 border-b border-ink/15 py-8 sm:grid-cols-[3rem_1fr] sm:gap-6"
                key={title}
              >
                <Icon
                  aria-hidden="true"
                  className="text-bronze"
                  size={25}
                  strokeWidth={1.35}
                />
                <div>
                  <h3 className="font-display text-2xl font-medium tracking-[-0.035em]">
                    {title}
                  </h3>
                  <p className="mt-3 max-w-xl text-sm leading-6 text-ink/65">
                    {copy}
                  </p>
                </div>
              </article>
            ))}
          </div>
        </div>
      </Container>
    </Section>
  )
}
