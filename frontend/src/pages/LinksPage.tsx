import { Link, useSearchParams } from 'react-router-dom'

import { BrandMark } from '../components/layout/BrandMark'
import { Skeleton } from '../components/ui/Skeleton'
import { contentTypeLabel } from '../features/content/model'
import { useContent } from '../features/content/queries'
import { cn } from '../lib/styles/cn'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

// Mirrors the token rule the analytics attribution applies on arrival, so a
// value this page forwards is one the destination will accept and record.
const utmTokenPattern = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/

/**
 * Tags an internal path so a click from the bio page can be attributed.
 *
 * The platform is forwarded from this page's own URL: the TikTok bio points
 * at /links?utm_source=tiktok, the Instagram one at ?utm_source=instagram,
 * and each button on the page then carries that platform as its source with
 * `bio` as the medium. A visitor who reaches /links with no platform at all
 * is recorded as source `bio`, medium `links`, so the two cases stay
 * distinguishable in the attribution report. One helper, so the pattern lives
 * in one place.
 */
function bioLink(path: string, forwardedSource: string | null): string {
  const source =
    forwardedSource && utmTokenPattern.test(forwardedSource)
      ? forwardedSource.toLowerCase()
      : null
  const [pathname, search = ''] = path.split('?')
  const params = new URLSearchParams(search)
  params.set('utm_source', source ?? 'bio')
  params.set('utm_medium', source ? 'bio' : 'links')
  return `${pathname}?${params.toString()}`
}

interface BioLink {
  label: string
  to: string
  eyebrow?: string
  primary?: boolean
}

const leadingLinks: BioLink[] = [
  { label: 'Build my stack (free)', to: '/build', primary: true },
  {
    label: 'Mailchimp alternatives at 1,000 subscribers',
    to: '/guides/mailchimp-alternatives',
  },
  { label: 'Live vendor offers', to: '/offers' },
]

const trailingLinks: BioLink[] = [
  {
    label: 'How we rank software',
    to: '/articles/how-unsolero-ranks-software',
  },
  { label: 'Affiliate disclosure', to: '/affiliate-disclosure' },
]

const recentEntryCount = 4

/**
 * The bio landing page: the pages behind the current videos and posts, as a
 * column of large buttons a thumb can hit.
 *
 * It deliberately does not render the site header or footer. A bio page is
 * read on a phone in the seconds after a video, and its whole job is to fit
 * the links on one screen; the primary navigation would push half of them
 * below the fold. The logo at the top is the way into the rest of the site,
 * and the disclosure sits at the bottom where every other page keeps it.
 */
export function LinksPage() {
  const [searchParams] = useSearchParams()
  const forwardedSource = searchParams.get('utm_source')
  const recent = useContent({ limit: recentEntryCount })
  usePageMetadata({
    title: 'UNSOLERO — the pages behind our videos',
    description:
      'Comparisons, guides and the stack builder referenced in UNSOLERO videos and posts.',
    robots: 'noindex, follow',
  })

  const recentLinks: BioLink[] = (recent.data ?? []).map((entry) => ({
    label: entry.title,
    to: entry.path,
    eyebrow: contentTypeLabel(entry.type),
  }))

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-col bg-canvas px-5 py-8 sm:py-12">
      <main id="main-content">
        <header className="flex flex-col items-center text-center">
          <BrandMark />
          <p className="mt-4 text-sm leading-6 text-ink/70">
            Build the right software stack. Every price dated, every ranking
            commission-proof.
          </p>
        </header>

        <nav aria-label="Pages behind our videos" className="mt-7">
          <ul className="flex flex-col gap-2.5">
            {leadingLinks.map((link) => (
              <BioButton
                forwardedSource={forwardedSource}
                key={link.to}
                link={link}
              />
            ))}
            {recent.isPending &&
              Array.from({ length: recentEntryCount }, (_, index) => (
                <li aria-hidden="true" key={index}>
                  <Skeleton className="min-h-14 w-full" />
                </li>
              ))}
            {recent.isError && (
              <li className="px-1 py-2 text-center text-xs text-ink/60">
                The latest articles could not be loaded.
              </li>
            )}
            {recentLinks.map((link) => (
              <BioButton
                forwardedSource={forwardedSource}
                key={link.to}
                link={link}
              />
            ))}
            {trailingLinks.map((link) => (
              <BioButton
                forwardedSource={forwardedSource}
                key={link.to}
                link={link}
              />
            ))}
          </ul>
        </nav>
      </main>

      <footer className="mt-auto pt-8 text-center text-xs leading-5 text-ink/60">
        Some links are affiliate links. Commission never changes the ranking.{' '}
        <Link
          className="underline underline-offset-4 hover:text-bronze-dark"
          to={bioLink('/affiliate-disclosure', forwardedSource)}
        >
          How we earn
        </Link>
      </footer>
    </div>
  )
}

function BioButton({
  link,
  forwardedSource,
}: {
  link: BioLink
  forwardedSource: string | null
}) {
  return (
    <li>
      <Link
        className={cn(
          'flex min-h-14 w-full flex-col items-center justify-center px-4 py-3 text-center text-sm font-semibold tracking-[-0.01em] transition-colors duration-150',
          link.primary
            ? 'bg-ink text-canvas hover:bg-bronze-dark'
            : 'border border-ink/15 bg-surface text-ink hover:border-bronze hover:text-bronze-dark',
        )}
        to={bioLink(link.to, forwardedSource)}
      >
        {link.eyebrow && (
          <span className="text-[0.625rem] font-bold tracking-[0.13em] text-ink/55 uppercase">
            {link.eyebrow}
          </span>
        )}
        <span className="text-balance">{link.label}</span>
      </Link>
    </li>
  )
}
