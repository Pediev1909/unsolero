import { ArrowRight, Search } from 'lucide-react'

import { ButtonLink } from '../ui/ButtonLink'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'

export function FinalCtaSection() {
  return (
    <Section space="lg" surface="paper">
      <Container>
        <div className="border-y border-ink/15 py-14 text-center sm:py-20 lg:py-28">
          <p className="eyebrow">Start with what matters</p>
          <Heading className="mx-auto mt-5 max-w-5xl" level={2} size="display">
            Build a gym around your life—not the other way around.
          </Heading>
          <p className="mx-auto mt-7 max-w-xl text-base leading-7 text-ink/65 sm:text-lg">
            Start your brief now. Sign in when you want to save it across
            devices. No artificial countdowns, pressure, or commission-driven
            rankings.
          </p>
          <div className="mt-9 flex flex-col justify-center gap-3 sm:flex-row sm:flex-wrap">
            <ButtonLink to="/build">
              Build My Setup
              <ArrowRight aria-hidden="true" size={16} />
            </ButtonLink>
            <ButtonLink to="/#categories" variant="secondary">
              <Search aria-hidden="true" size={16} />
              Explore Equipment
            </ButtonLink>
          </div>
        </div>
      </Container>
    </Section>
  )
}
