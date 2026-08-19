import { Link, useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Heading } from '../components/ui/Heading'
import { LoadingState } from '../components/ui/LoadingState'
import { Section } from '../components/ui/Section'
import { useContentAuthor } from '../features/content/queries'
import { formatEditorialDate } from '../features/content/model'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

// The page behind a byline. Attribution to a named person is the signal a
// reader weighing a ranking and a search engine weighing a page both look for,
// and it is worth nothing if the name leads nowhere.
export function AuthorPage() {
  const { slug = '' } = useParams()
  const author = useContentAuthor(slug)

  usePageMetadata({
    title: author.data ? `${author.data.author.name} | UNSOLERO` : 'Author | UNSOLERO',
    description: author.data?.author.bio ?? 'Who writes UNSOLERO.',
  })

  return (
    <>
      <SiteHeader />
      <main id="main-content">
        <Section space="lg">
          <Container>
            {author.isPending && (
              <LoadingState description="Loading author." title="Loading" />
            )}
            {author.isError && (
              <ErrorState
                description="This author page could not be loaded."
                onRetry={() => void author.refetch()}
                title="Author unavailable"
              />
            )}
            {author.data && (
              <div className="max-w-3xl">
                <p className="eyebrow">Author</p>
                <Heading className="mt-5" level={1} size="section">
                  {author.data.author.name}
                </Heading>
                <p className="mt-7 max-w-2xl text-lg leading-8 text-ink/70">
                  {author.data.author.bio}
                </p>
                <p className="mt-6 text-base text-ink/70">
                  More about how this site is run, funded and sourced is on the{' '}
                  <Link className="underline" to="/about">
                    about page
                  </Link>
                  .
                </p>

                <Heading className="mt-14" level={2} size="title">
                  Published work
                </Heading>
                {author.data.entries.length === 0 ? (
                  <p className="mt-5 text-base text-ink/70">
                    Nothing published yet.
                  </p>
                ) : (
                  <ul className="mt-6 divide-y divide-ink/10 border-y border-ink/10">
                    {author.data.entries.map((entry) => (
                      <li key={entry.id}>
                        <Link
                          className="group block py-5 transition-colors hover:bg-paper/60"
                          to={entry.path}
                        >
                          <p className="text-lg font-semibold tracking-[-0.01em] group-hover:text-bronze-dark">
                            {entry.title}
                          </p>
                          <p className="mt-1 text-base leading-7 text-ink/70">
                            {entry.description}
                          </p>
                          <p className="mt-2 text-label uppercase tracking-[0.12em] text-ink/65">
                            {formatEditorialDate(entry.published_at)}
                          </p>
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </Container>
        </Section>
      </main>
      <SiteFooter />
    </>
  )
}
