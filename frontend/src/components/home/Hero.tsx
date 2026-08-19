import { ArrowRight, Search } from 'lucide-react'

import { ButtonLink } from '../ui/ButtonLink'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
export function Hero() {
  return (
    <section className="relative overflow-hidden bg-canvas pt-20 lg:min-h-[820px]">
      {/* The illustration is a wireframe of a dashboard. Beside the copy on a
          wide screen it reads as what the site is about; behind it on a phone
          it reads as artefacts, and no amount of scrim fixes that — a fade
          heavy enough to protect the text leaves grey bars crossing the
          buttons, and a lighter one puts panels under the headline. A phone
          gets the type and the two things to do, which is the stronger first
          screen anyway. */}
      <div className="absolute inset-x-0 bottom-0 top-20 hidden lg:left-[44%] lg:block">
        <img
          className="size-full object-cover object-[66%_center]"
          src="/images/unsolero-saas-hero.svg"
          alt="A small team reviewing the tools their business runs on"
          fetchPriority="high"
        />
        <div className="absolute inset-0 bg-gradient-to-r from-canvas via-canvas/25 to-transparent" />
      </div>

      <Container className="relative z-10 flex items-start pb-16 pt-14 sm:pb-20 sm:pt-24 lg:min-h-[740px] lg:items-center lg:py-20">
        <div className="max-w-3xl lg:max-w-[720px]">
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
      </Container>

      <p className="absolute bottom-7 right-6 z-10 hidden max-w-[170px] text-xs leading-5 text-canvas/78 lg:block xl:right-10">
        Your budget and existing tools are constraints—not afterthoughts.
      </p>
    </section>
  )
}
