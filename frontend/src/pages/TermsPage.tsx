import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { Section } from '../components/ui/Section'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

// Affiliate programmes check for terms before they approve a site, and a
// visitor deciding whether to trust a recommendation is entitled to know what
// this site is claiming and what it is not. Every statement here describes how
// the application actually behaves; none of it is boilerplate carried over from
// another site.
export function TermsPage() {
  usePageMetadata({
    title: 'Terms of use | UNSOLERO',
    description:
      'What UNSOLERO is, what it does not promise, and the rules for using it.',
  })

  return (
    <>
      <SiteHeader />
      <main id="main-content">
        <Section space="lg">
          <Container>
            <div className="max-w-2xl">
              <p className="eyebrow">Terms</p>
              <Heading className="mt-5" level={1} size="section">
                What this site promises, and what it does not.
              </Heading>

              <div className="mt-10 space-y-8 text-base leading-7 text-ink/75">
                <p>
                  By using UNSOLERO you accept these terms. They are written to
                  be read, not to be survived.
                </p>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Who operates this site
                  </Heading>
                  <p>
                    UNSOLERO is built and run by Andon Pediev, an individual
                    based in Bulgaria, within the European Union. There is no
                    company behind it. Contact:{' '}
                    <a className="underline" href="mailto:hello@unsolero.com">
                      hello@unsolero.com
                    </a>
                    .
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    What the service is
                  </Heading>
                  <p>
                    UNSOLERO produces an explainable recommendation for a set of
                    business software, based on the goals, budget, team size and
                    existing tools you enter. It shows why each product was
                    chosen, what was rejected, and in what order to buy.
                  </p>
                  <p className="mt-3">
                    It is information to help you decide. It is not
                    professional, legal, financial, accounting or tax advice,
                    and it is not a substitute for evaluating a product
                    yourself. A free trial and a demo remain the only way to
                    know whether a tool fits the way your team actually works.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Accuracy of catalog data
                  </Heading>
                  <p>
                    Product facts are recorded against a source and a date, and
                    a product is only published once its recommendation-critical
                    facts have provenance. That process makes errors visible and
                    traceable. It does not make them impossible.
                  </p>
                  <p className="mt-3">
                    Vendors change prices, rename plans, move features between
                    tiers and discontinue products without notice.{' '}
                    <strong className="text-ink">
                      Always confirm the current price and terms on the
                      vendor&rsquo;s own site before you buy.
                    </strong>{' '}
                    Where what you see here differs from the vendor, the vendor
                    is right and we would like to be told:{' '}
                    <a className="underline" href="mailto:hello@unsolero.com">
                      hello@unsolero.com
                    </a>
                    .
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Commercial relationships
                  </Heading>
                  <p>
                    Some merchant links on this site earn a commission if you
                    buy through them. Those links are marked. Commission cannot
                    influence which products are recommended, their scores,
                    their order, or the reasons given — that separation is
                    enforced in the software itself, not by policy alone.
                  </p>
                  <p className="mt-3">
                    The full explanation is on the{' '}
                    <Link className="underline" to="/affiliate-disclosure">
                      affiliate disclosure
                    </Link>{' '}
                    page.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Your purchases are with the vendor
                  </Heading>
                  <p>
                    We sell nothing. When you follow a link and buy, the
                    contract is between you and that vendor or merchant, under
                    their terms. Billing, refunds, cancellation, support, data
                    processing and service availability are theirs to provide
                    and yours to agree. We are not a party to it and cannot
                    resolve a dispute about it.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Accounts
                  </Heading>
                  <p>
                    An account is optional; the recommendation works without
                    one. If you create one, keep your credentials to yourself
                    and tell us if you believe they have been used by someone
                    else. You may delete your account at any time from your
                    account settings.
                  </p>
                  <p className="mt-3">
                    We may suspend an account that is used to attack the
                    service, to scrape it in bulk, or to break the law.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Acceptable use
                  </Heading>
                  <p>
                    Read the site, use the recommendation, quote it with
                    attribution. Do not attempt to gain unauthorised access, do
                    not interfere with its availability, and do not reproduce
                    the catalog or the editorial content wholesale for a
                    competing service. Automated access at a rate that degrades
                    the service for other people may be blocked.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Intellectual property
                  </Heading>
                  <p>
                    The wording, structure, editorial content and software of
                    this site belong to its operator. Product names, logos and
                    trademarks belong to their respective owners and are used
                    here to identify those products, which neither implies nor
                    claims any endorsement by them.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Availability and liability
                  </Heading>
                  <p>
                    The service is provided as it is. It runs on modest
                    infrastructure and may be unavailable, interrupted, or
                    changed at any time. To the extent the law allows, we are
                    not liable for a purchasing decision you make, for a loss
                    arising from a product you bought through a link here, or
                    for indirect or consequential loss.
                  </p>
                  <p className="mt-3">
                    Nothing here limits a right you have as a consumer under
                    mandatory law, and nothing here excludes liability for fraud
                    or for anything that cannot lawfully be excluded.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Privacy
                  </Heading>
                  <p>
                    What is collected, why, for how long, and how to withdraw
                    consent is set out on the{' '}
                    <Link className="underline" to="/privacy">
                      privacy
                    </Link>{' '}
                    page.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Changes and governing law
                  </Heading>
                  <p>
                    These terms may change as the service does. Continuing to
                    use the site after a change means you accept it. Bulgarian
                    law governs these terms, and mandatory consumer protections
                    of your country of residence in the European Union continue
                    to apply.
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
