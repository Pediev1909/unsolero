import {
  ArrowRight,
  CheckCircle2,
  ExternalLink,
  ShieldCheck,
} from 'lucide-react'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Badge } from '../components/ui/Badge'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { buttonStyles } from '../components/ui/buttonStyles'
import { affiliateClickPath } from '../features/analytics/tracking'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

const webinarPath = '/api/affiliate/promotion/funnel-hacking-secrets-webinar'
const orderPath = '/api/affiliate/promotion/funnel-hacking-secrets-order'

export function FunnelHackingSecretsPage() {
  usePageMetadata({
    title: 'Funnel Hacking Secrets free training | UNSOLERO',
    description:
      'What the free Funnel Hacking Secrets training covers, what happens after it, and the affiliate relationship behind the links.',
    type: 'article',
  })

  const webinarHref = affiliateClickPath(webinarPath, 'promotion')
  const orderHref = affiliateClickPath(orderPath, 'promotion')

  return (
    <>
      <SiteHeader position="sticky" />
      <main id="main-content">
        <header className="border-b border-ink/15 py-12 sm:py-16 lg:py-20">
          <Container>
            <div className="max-w-4xl">
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="neutral">Independent guide</Badge>
                <Badge variant="sponsored">Affiliate links</Badge>
              </div>
              <p className="eyebrow mt-7">Free funnel training</p>
              <Heading className="mt-5" level={1} size="display">
                Funnel Hacking Secrets, before you register.
              </Heading>
              <p className="mt-7 max-w-3xl text-lg leading-8 text-ink/68 sm:text-xl sm:leading-9">
                A free ClickFunnels® training about selecting, modelling and
                building sales funnels. This page explains the next step as
                clearly as the promise.
              </p>
              <div className="mt-9 flex flex-col gap-3 sm:flex-row sm:items-center">
                <a
                  className={buttonStyles({ size: 'lg', variant: 'primary' })}
                  href={webinarHref}
                  rel="nofollow noopener sponsored"
                  target="_blank"
                >
                  Register for the free training
                  <ExternalLink aria-hidden="true" size={17} />
                </a>
                <span className="text-xs leading-5 text-ink/65">
                  Opens the official Funnel Hacking Secrets registration page.
                </span>
              </div>
            </div>
          </Container>
        </header>

        <section className="py-14 sm:py-20">
          <Container>
            <div className="grid gap-12 lg:grid-cols-[minmax(0,42rem)_minmax(18rem,1fr)] lg:gap-20">
              <div>
                <p className="eyebrow">What it is</p>
                <Heading className="mt-4" level={2} size="title">
                  Training first. A paid offer may follow.
                </Heading>
                <div className="mt-7 space-y-5 text-base leading-8 text-ink/72 sm:text-lg">
                  <p>
                    ClickFunnels describes Funnel Hacking Secrets as a free
                    training course. After the training, it may present an
                    optional paid Funnel Hacking Secrets package.
                  </p>
                  <p>
                    Registration may ask for your email and an optional mobile
                    number. The confirmation message can contain an access code
                    or instructions for joining through the Webby app.
                  </p>
                  <p>
                    The training is educational. It does not guarantee sales,
                    revenue, profit, or that ClickFunnels® is the right tool for
                    your business.
                  </p>
                </div>
              </div>

              <aside className="border border-ink/15 bg-paper p-6 sm:p-7">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
                  What the class covers
                </p>
                <ul className="mt-6 space-y-5 text-sm leading-6 text-ink/75">
                  {[
                    'How to choose a funnel shape for an existing business.',
                    'How marketers study and model an existing funnel structure.',
                    'How funnel pages, an offer and traffic fit together.',
                  ].map((item) => (
                    <li className="flex gap-3" key={item}>
                      <CheckCircle2
                        aria-hidden="true"
                        className="mt-1 shrink-0 text-moss"
                        size={17}
                      />
                      <span>{item}</span>
                    </li>
                  ))}
                </ul>
              </aside>
            </div>
          </Container>
        </section>

        <section className="border-y border-ink/15 bg-paper py-14 sm:py-20">
          <Container>
            <div className="grid gap-8 lg:grid-cols-[1fr_auto] lg:items-center lg:gap-16">
              <div className="max-w-3xl">
                <div className="flex items-center gap-3 text-moss-dark">
                  <ShieldCheck aria-hidden="true" size={22} />
                  <p className="text-xs font-bold uppercase tracking-[0.14em]">
                    Clear commercial relationship
                  </p>
                </div>
                <Heading className="mt-4" level={2} size="title">
                  UNSOLERO may receive a referral payment.
                </Heading>
                <p className="mt-5 text-sm leading-7 text-ink/70 sm:text-base">
                  The buttons on this page contain affiliate attribution. If you
                  register and later make an eligible purchase, ClickFunnels may
                  pay UNSOLERO a commission at no additional cost to you. This
                  promotion is separate from software recommendation scoring.
                </p>
              </div>
              <a
                className={buttonStyles({ size: 'lg', variant: 'primary' })}
                href={webinarHref}
                rel="nofollow noopener sponsored"
                target="_blank"
              >
                Open the free class <ArrowRight aria-hidden="true" size={17} />
              </a>
            </div>
          </Container>
        </section>

        <section className="py-14 sm:py-20">
          <Container>
            <div className="max-w-3xl border-l-2 border-bronze pl-6 sm:pl-8">
              <p className="eyebrow">Already watched the training?</p>
              <Heading className="mt-4" level={2} size="title">
                Review the current package on the official order form.
              </Heading>
              <p className="mt-5 text-sm leading-7 text-ink/70 sm:text-base">
                This link skips the free class. UNSOLERO does not publish a
                package price here because contents and checkout terms can
                change. Read the current price, inclusions, renewal terms and
                refund conditions on the official form before paying.
              </p>
              <a
                className={buttonStyles({
                  className: 'mt-7',
                  size: 'md',
                  variant: 'secondary',
                })}
                href={orderHref}
                rel="nofollow noopener sponsored"
                target="_blank"
              >
                Review the current paid offer
                <ExternalLink aria-hidden="true" size={16} />
              </a>
            </div>
          </Container>
        </section>
      </main>
      <SiteFooter />
    </>
  )
}
