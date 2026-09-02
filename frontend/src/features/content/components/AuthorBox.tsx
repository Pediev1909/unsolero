import { Link } from 'react-router-dom'

import { authorInitial, authorRole, formatEditorialDate } from '../model'
import type { ContentDetail } from '../schemas'

interface AuthorBoxProps {
  author: ContentDetail['author']
  publishedAt: string
  updatedAt: string
}

/**
 * Who wrote this and how to check on them.
 *
 * A name and two dates used to be the whole byline. The two links are the
 * point: the ranking method and the affiliate terms are one click from every
 * piece, not buried in a footer, because a reader deciding whether to trust a
 * recommendation should not have to go looking for how it was made or who
 * pays for it.
 */
export function AuthorBox({ author, publishedAt, updatedAt }: AuthorBoxProps) {
  const updated = updatedAt.slice(0, 10) !== publishedAt.slice(0, 10)

  return (
    <div className="text-xs leading-6 text-ink/70">
      <div className="flex items-center gap-3">
        {author.avatar_url ? (
          <img
            alt=""
            className="size-11 shrink-0 object-cover"
            height={44}
            src={author.avatar_url}
            width={44}
          />
        ) : (
          <span
            aria-hidden="true"
            className="flex size-11 shrink-0 items-center justify-center bg-paper font-display text-lg font-semibold text-ink/80"
          >
            {authorInitial(author.name)}
          </span>
        )}
        <div className="min-w-0">
          <p className="font-semibold text-ink">
            <Link
              className="underline decoration-ink/25 underline-offset-4 hover:decoration-ink"
              to={`/author/${author.slug}`}
            >
              {author.name}
            </Link>
          </p>
          <p>{authorRole}</p>
        </div>
      </div>
      <p className="mt-3">Published {formatEditorialDate(publishedAt)}</p>
      {updated && <p>Updated {formatEditorialDate(updatedAt)}</p>}
      <ul className="mt-3 flex flex-wrap gap-x-4 gap-y-1">
        <li>
          <Link
            className="hover:text-bronze-dark"
            to="/articles/how-unsolero-ranks-software"
          >
            How we rank software
          </Link>
        </li>
        <li>
          <Link className="hover:text-bronze-dark" to="/affiliate-disclosure">
            Affiliate disclosure
          </Link>
        </li>
      </ul>
    </div>
  )
}
