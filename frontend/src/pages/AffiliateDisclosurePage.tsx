import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { Section } from '../components/ui/Section'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function AffiliateDisclosurePage() {
  usePageMetadata({
    title: 'Affiliate disclosure | UNSOLERO',
    description:
      'How UNSOLERO earns money, and why commission cannot change what it recommends.',
  })

  return (
    <>
      <SiteHeader />
      <main id="main-content">
        <Section space="lg">
          <Container>
            <div className="max-w-2xl">
              <p className="eyebrow">Disclosure</p>
              <Heading className="mt-5" level={1} size="section">
                How we make money.
              </Heading>

              <div className="mt-10 space-y-8 text-base leading-7 text-ink/75">
                <p>
                  UNSOLERO may earn a commission when you subscribe to a product
                  after following a link from this site. That costs you nothing
                  extra — the price is the same as going direct.
                </p>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    What commission does not do
                  </Heading>
                  <p>
                    It does not affect which products are recommended, in what
                    order, or why. Recommendations are produced by a
                    deterministic engine that has no access to commercial data:
                    commission rates and merchant relationships live in a
                    separate part of the system and are never inputs to scoring.
                  </p>
                  <p className="mt-3">
                    This is enforced rather than promised. An automated test
                    fails the build if any commercial field is added to the data
                    the engine scores. A product that pays us nothing can and
                    does outrank one that pays well.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    What we do not do
                  </Heading>
                  <ul className="ml-5 list-disc space-y-2">
                    <li>We do not accept payment for placement or ranking.</li>
                    <li>
                      We do not hide products because they have no affiliate
                      programme.
                    </li>
                    <li>
                      We do not publish a product until its facts have a
                      recorded source, which is why the catalog grows slowly.
                    </li>
                  </ul>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    How to check
                  </Heading>
                  <p>
                    Every recommendation shows its reasons and the facts behind
                    them, and rejected products are shown with the reason they
                    were rejected rather than hidden. The method is described in
                    full in{' '}
                    <Link
                      className="underline underline-offset-4"
                      to="/articles/how-unsolero-ranks-software"
                    >
                      how we rank software
                    </Link>
                    .
                  </p>
                  <p className="mt-3">
                    If a recommendation does not make sense to you, it is either
                    a bug or a bad fact. Both are worth telling us about.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Prices
                  </Heading>
                  <p>
                    Prices are read from each vendor&apos;s own pricing page and
                    recorded with the date they were read. Software pricing
                    changes often; confirm the current price with the vendor
                    before subscribing.
                  </p>
                </div>
              </div>
            </div>
          </Container>
        </Section>
      </main>
      <SiteFooter />
    </>
  )
}
