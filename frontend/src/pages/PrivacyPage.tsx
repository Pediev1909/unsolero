import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { Heading } from '../components/ui/Heading'
import { Section } from '../components/ui/Section'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

// The retention periods and cookie names below are the values the application
// actually runs with, not illustrative ones. If ANALYTICS_* or SESSION_* change
// in configuration, this page has to change with them or it becomes a false
// statement rather than a stale one.
export function PrivacyPage() {
  usePageMetadata({
    title: 'Privacy | UNSOLERO',
    description:
      'What UNSOLERO collects, why, how long it is kept, and how to withdraw consent.',
  })

  return (
    <>
      <SiteHeader />
      <main id="main-content">
        <Section space="lg">
          <Container>
            <div className="max-w-2xl">
              <p className="eyebrow">Privacy</p>
              <Heading className="mt-5" level={1} size="section">
                What we collect and why.
              </Heading>

              <div className="mt-10 space-y-8 text-base leading-7 text-ink/75">
                <p>
                  This page describes what the site actually does. Where a period or
                  a name is given, it is the value the running system uses.
                </p>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Who is responsible
                  </Heading>
                  <p className="rounded border border-ink/15 bg-surface p-4 text-sm">
                    <strong>To be completed before launch.</strong> Data protection
                    law requires the operator of this site to be identifiable by
                    name and contact address. Fill this in, or use a contact
                    address you are willing to publish.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    If you only browse
                  </Heading>
                  <p>
                    Pages can be read without an account. Analytics are collected
                    only after you consent, and you can withdraw that consent at any
                    time from the link in the footer. Declining does not restrict
                    anything on the site.
                  </p>
                  <p className="mt-3">
                    When you do consent, a cookie named{' '}
                    <code>unsolero_analytics_subject</code> distinguishes your
                    visits from other people&apos;s. Anonymous analytics events are
                    deleted after <strong>90 days</strong>.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    If you create an account
                  </Heading>
                  <p>
                    An account stores your email address, your saved setups,
                    wishlists and comparisons, and the recommendation briefs you
                    submit. A session cookie named <code>rigmark_session</code>{' '}
                    keeps you signed in for up to <strong>30 days</strong>, or{' '}
                    <strong>7 days</strong> of inactivity, whichever comes first.
                  </p>
                  <p className="mt-3">
                    Analytics events linked to an account are deleted after{' '}
                    <strong>397 days</strong>. Delivery receipts for account
                    security email are kept for <strong>30 days</strong>.
                  </p>
                  <p className="mt-3">
                    Account email is used for verification, password reset and
                    security notices only. It is not a marketing list and you will
                    not be added to one without asking.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    When you follow a link to a vendor
                  </Heading>
                  <p>
                    Following a link to a software vendor records which product,
                    which link and which page it came from, so we can tell what
                    people found useful. The vendor receives whatever their own site
                    collects; their privacy policy governs that, not this one.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Your rights
                  </Heading>
                  <p>
                    You can export your account data or delete your account from
                    account settings at any time. Deleting the account removes the
                    personal data attached to it.
                  </p>
                </div>

                <div>
                  <Heading className="mb-3" level={2} size="title">
                    Processors
                  </Heading>
                  <p>
                    The site runs on a server in Germany. Object storage is provided
                    by Cloudflare and transactional email by Resend. Each processes
                    data on our behalf in order to run the service.
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
