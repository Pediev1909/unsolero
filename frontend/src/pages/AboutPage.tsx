import { Link } from 'react-router-dom'

import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { Section } from '../components/ui/Section'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function AboutPage() {
  usePageMetadata({
    title: 'About UNSOLERO | Who runs this site',
    description:
      'UNSOLERO is built and run by Andon Pediev. Who writes it, where the facts come from, and what it does not yet claim to know.',
  })

  return (
    <Section space="lg">
      <Container>
        <div className="max-w-2xl">
          <p className="eyebrow">About</p>
          <Heading className="mt-5" level={1} size="section">
            Who is behind this site.
          </Heading>

          <div className="mt-10 space-y-8 text-base leading-7 text-ink/75">
            <p>
              UNSOLERO is built and run by <strong>Andon Pediev</strong>. Not a
              team, not an agency — one person who writes the software, chooses
              which products are listed, and records where every fact came from.
            </p>

            <div>
              <Heading className="mb-3" level={2} size="title">
                Why it exists
              </Heading>
              <p>
                Most software comparison sites rank by whoever pays the most and
                do not tell you. You read a list, you cannot see how it was
                ordered, and the order is for sale.
              </p>
              <p className="mt-3">
                This site was built the other way around. The ranking is
                produced by a deterministic engine, the same inputs always give
                the same output, and commercial data is not one of the inputs.
                You can read{' '}
                <Link className="underline" to="/articles/how-unsolero-ranks-software">
                  the full method
                </Link>{' '}
                and check it against what the site actually shows you.
              </p>
            </div>

            <div>
              <Heading className="mb-3" level={2} size="title">
                Where the facts come from
              </Heading>
              <p>
                Prices and plan limits are read from each vendor&rsquo;s own
                pricing and documentation pages. Every one of them records which
                page it came from and the date it was read, and that record is
                published next to the fact rather than kept internally.
              </p>
              <p className="mt-3">
                Suitability scores are a separate thing: they are editorial
                judgements, they are labelled as such, and they are attributed
                apart from the vendor&rsquo;s own claims. A number a vendor
                supplied and a number this site formed are never presented as
                the same kind of thing.
              </p>
            </div>

            <div>
              <Heading className="mb-3" level={2} size="title">
                What this site does not claim
              </Heading>
              {/* Stating the limit is the point. A comparison site that implies
                  hands-on testing it has not done is the exact pattern search
                  engines have been demoting, and a reader who later discovers
                  the gap stops trusting everything else on the page. */}
              <p>
                Current scoring is built from documented specifications and
                pricing, not from months of daily use of every product. Where
                that changes, the page will say so: hands-on notes carry the
                date they were written and what was actually done, so you can
                tell a tested opinion from a documented fact.
              </p>
              <p className="mt-3">
                Software pricing changes often. Each price here records when it
                was read — confirm the current figure with the vendor before
                committing to it.
              </p>
            </div>

            <div>
              <Heading className="mb-3" level={2} size="title">
                How it pays for itself
              </Heading>
              <p>
                Through affiliate commission, disclosed in full on the{' '}
                <Link className="underline" to="/affiliate-disclosure">
                  affiliate disclosure
                </Link>{' '}
                page. Commission cannot move a product up the list, and that is
                enforced by an automated test rather than promised in a
                sentence.
              </p>
            </div>

            <div>
              <Heading className="mb-3" level={2} size="title">
                Contact
              </Heading>
              <p>
                Corrections are welcome, particularly on prices and plan limits.
                If something here is wrong, write to{' '}
                <a
                  className="underline"
                  href="mailto:pedievandon1909@gmail.com"
                >
                  pedievandon1909@gmail.com
                </a>{' '}
                and say which page and which figure.
              </p>
            </div>
          </div>
        </div>
      </Container>
    </Section>
  )
}
