import { GitCompare, LayoutGrid, Wand2 } from 'lucide-react'
import { Link } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { Section } from '../components/ui/Section'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

const description =
  'Three ways to use UNSOLERO, where every price comes from, how the scores are decided, and why commission cannot move a ranking.'

/**
 * Three doors, in the order of how sure the visitor is about what they want.
 * Someone who knows they need a CRM should not be made to answer a
 * questionnaire first, and someone who has no idea should not be dropped into
 * a list of fifty-three products.
 */
const routes = [
  {
    icon: LayoutGrid,
    title: 'I know what kind of tool I need',
    copy: 'Pick the category — CRM, invoicing, help desk — and read the tools in it side by side. Every one shows its price, what that price includes, and where the figure came from.',
    action: 'Browse the categories',
    to: '/categories',
  },
  {
    icon: Wand2,
    title: 'I have no idea where to start',
    copy: 'Answer a few plain questions about what your business does, what you already pay for, and what you can spend. We work out a whole set of tools that fit together, and tell you what to skip.',
    action: 'Build my setup',
    to: '/build',
  },
  {
    icon: GitCompare,
    title: 'I am stuck between two or three',
    copy: 'Put them next to each other. Price, what each tier actually gives you, and where they differ on the things that decide it.',
    action: 'Open the comparison',
    to: '/compare',
  },
]

const method = [
  {
    heading: 'Every price was read from the vendor, on a date we tell you',
    copy: 'Nobody here retypes a number from another comparison site. Each price is read from the vendor’s own pricing page, and the site records which page, what day it was read, and how confident that reading is. If a price cannot be verified, the product does not go in the catalog at all — which is why some well-known tools are missing.',
  },
  {
    heading: 'The billing basis is always stated, because it is where the trick lives',
    copy: 'One vendor quotes per month, the next quotes the same plan billed annually and shows a smaller number. They are not comparable, so we say which is which on every product. Where a vendor was running a promotion, we publish the standing rate, not the discount, so the comparison does not quietly go stale when the offer ends.',
  },
  {
    heading: 'The scores are opinions, and they are labelled as opinions',
    copy: 'Prices are facts. Whether a tool suits a beginner is a judgement. Those are kept apart: every score carries a written reason you can read, and none of them is presented as something the vendor said.',
  },
  {
    heading: 'We earn a commission, and it cannot move a ranking',
    copy: 'Some links here earn money if you buy through them, and every one of those is labelled where it appears. The ranking is produced by an engine that is never given the commission figure, so it has nothing to weigh even if it wanted to. A tool that pays us nothing can and does beat one that pays us well.',
  },
]

export function HowItWorksPage() {
  usePageMetadata({ title: 'How UNSOLERO works | UNSOLERO', description })

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <section className="border-b border-ink/15 py-16 sm:py-24 lg:py-28">
          <Container>
            <p className="eyebrow">How it works</p>
            <Heading className="mt-5 max-w-4xl" level={1} size="display">
              Start wherever you actually are.
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/65">
              You do not need to know what you are looking for to use this
              site. Pick whichever of these three sounds like you.
            </p>
          </Container>
        </section>

        <Container className="py-12 sm:py-18">
          <ul className="grid gap-5 lg:grid-cols-3">
            {routes.map(({ icon: Icon, title, copy, action, to }) => (
              <li className="flex" key={to}>
                <Link
                  className="group flex flex-col rounded-sm border border-ink/15 bg-surface p-6 transition-colors hover:border-bronze sm:p-8"
                  to={to}
                >
                  <Icon
                    aria-hidden="true"
                    className="text-bronze"
                    size={26}
                    strokeWidth={1.4}
                  />
                  <h2 className="mt-8 font-display text-xl font-medium tracking-[-0.03em] group-hover:text-bronze sm:text-2xl">
                    {title}
                  </h2>
                  <p className="mt-4 grow text-sm leading-6 text-ink/65">
                    {copy}
                  </p>
                  <span className="mt-7 text-sm font-semibold text-bronze underline underline-offset-4">
                    {action}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </Container>

        <Section space="lg" surface="paper">
          <Container>
            <p className="eyebrow">Where the numbers come from</p>
            <Heading className="mt-5 max-w-3xl" level={2} size="section">
              Why you should believe any of this.
            </Heading>
            <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/70">
              Every comparison site says it is independent. Here is what that
              claim is worth here, in detail, so you can check it.
            </p>

            <dl className="mt-14 grid border-l border-t border-ink/15 md:grid-cols-2">
              {method.map(({ heading, copy }) => (
                <div
                  className="border-b border-r border-ink/15 p-6 sm:p-8 lg:p-10"
                  key={heading}
                >
                  <dt className="font-display text-xl font-medium tracking-[-0.03em] sm:text-2xl">
                    {heading}
                  </dt>
                  <dd className="mt-4 text-sm leading-6 text-ink/70">{copy}</dd>
                </div>
              ))}
            </dl>

            <p className="mt-10 text-sm text-ink/65">
              The long version is in{' '}
              <Link
                className="underline underline-offset-4"
                to="/articles/how-unsolero-ranks-software"
              >
                how UNSOLERO ranks software
              </Link>
              , and the money side is set out in the{' '}
              <Link
                className="underline underline-offset-4"
                to="/affiliate-disclosure"
              >
                affiliate disclosure
              </Link>
              .
            </p>
          </Container>
        </Section>
      </main>
      <SiteFooter />
    </>
  )
}
