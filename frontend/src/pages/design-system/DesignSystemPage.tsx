import { ArrowLeft } from 'lucide-react'

import { SiteFooter } from '../../components/layout/SiteFooter'
import { SiteHeader } from '../../components/layout/SiteHeader'
import { Badge } from '../../components/ui/Badge'
import { ButtonLink } from '../../components/ui/ButtonLink'
import { Container } from '../../components/ui/Container'
import { Heading } from '../../components/ui/Heading'
import { CommerceShowcase } from './CommerceShowcase'
import { ControlsShowcase } from './ControlsShowcase'
import { FoundationShowcase } from './FoundationShowcase'
import { InteractionShowcase } from './InteractionShowcase'

export function DesignSystemPage() {
  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="pb-16 pt-14 sm:pb-20 sm:pt-18 lg:pb-28 lg:pt-24">
          <Container>
            <Badge variant="warning">Temporary showcase</Badge>
            <p className="eyebrow mt-8">UNSOLERO / Design system 01</p>
            <Heading className="mt-5 max-w-5xl" level={1} size="display">
              Precision without noise.
            </Heading>
            <div className="mt-8 grid gap-8 border-t border-ink/15 pt-8 md:grid-cols-2 md:gap-16">
              <p className="max-w-xl text-body-lg text-ink/70">
                A premium, editorial system for trustworthy fitness decisions
                across every screen size.
              </p>
              <p className="max-w-xl text-sm leading-6 text-ink/70">
                This route exists to verify component behavior and responsive
                composition. It uses no live product claims or analytics.
              </p>
            </div>
          </Container>
        </section>

        <Container>
          <FoundationShowcase />
          <ControlsShowcase />
          <InteractionShowcase />
          <div id="products">
            <CommerceShowcase />
          </div>
          <div className="border-t border-ink/15 py-14">
            <ButtonLink to="/" variant="secondary">
              <ArrowLeft aria-hidden="true" size={16} />
              Back to home
            </ButtonLink>
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  )
}
