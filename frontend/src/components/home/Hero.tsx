import { ArrowRight, Search } from 'lucide-react'

import { ButtonLink } from '../ui/ButtonLink'
import { HeroCatalogPanel } from './HeroCatalogPanel'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
export function Hero() {
  return (
    <section className="relative overflow-hidden bg-canvas pt-20 lg:min-h-[820px]">
      {/* The illustration used to be a static SVG wireframe: grey bars
          arranged to look like a dashboard. On a site whose whole argument is
          that its numbers are real and checked, drawing fake interface was the
          worst available choice — and its alt text described a small team
          reviewing tools, which is not what the picture showed either. It is
          the catalog now, live. */}
      <Container className="relative z-10 flex flex-col items-start gap-14 pt-14 pb-16 sm:pt-24 sm:pb-20 lg:min-h-[740px] lg:flex-row lg:items-center lg:justify-between lg:gap-16 lg:py-20">
        <div className="max-w-3xl lg:max-w-[620px]">
          <p className="eyebrow">Software stack intelligence</p>
          <Heading className="mt-5 max-w-2xl" level={1} size="hero">
            Build the right software stack.
          </Heading>
          <p className="mt-8 max-w-lg text-base leading-7 text-ink/70 sm:text-lg sm:leading-8">
            Tell us what your business does, what you already run, and what you
            can spend. We&apos;ll work out what you actually need.
          </p>
          <div className="mt-9 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <ButtonLink to="/build">
              Build My Setup
              <ArrowRight aria-hidden="true" size={16} />
            </ButtonLink>
            <ButtonLink to="/#categories" variant="secondary">
              <Search aria-hidden="true" size={16} />
              Explore Categories
            </ButtonLink>
          </div>
        </div>

        <HeroCatalogPanel />
      </Container>
    </section>
  )
}
