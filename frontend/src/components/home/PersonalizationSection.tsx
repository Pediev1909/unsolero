import { Boxes, House, WalletCards } from 'lucide-react'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'

const reasons = [
  {
    icon: House,
    title: 'Your space changes the answer',
    copy: 'Ceiling height, floor area, storage, neighbors, and shared rooms can rule out otherwise excellent equipment.',
  },
  {
    icon: WalletCards,
    title: 'Your budget needs a sequence',
    copy: 'A good plan protects room for the essentials and identifies the upgrades that can genuinely wait.',
  },
  {
    icon: Boxes,
    title: 'Your existing gear has value',
    copy: 'What you already own should reduce redundancy and change which movement patterns need equipment next.',
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
              There is no universal best home gym.
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
