import { Search } from 'lucide-react'
import { useId, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'

import { cn } from '../../lib/styles/cn'

interface HeaderSearchProps {
  /** Full width inside the mobile drawer, fixed width in the header bar. */
  variant?: 'bar' | 'block'
  onSubmitted?: () => void
}

/**
 * Search in the header, going to the catalog's own filtered list.
 *
 * It is a real input rather than an icon that expands into one. An icon costs
 * a click and hides the single affordance that a visitor who does not
 * understand the navigation will reach for. The catalog already filters by
 * ?q=, so this adds an entry point rather than a second search system.
 */
export function HeaderSearch({
  variant = 'bar',
  onSubmitted,
}: HeaderSearchProps) {
  const [term, setTerm] = useState('')
  const navigate = useNavigate()
  // The drawer's copy of this form is in the DOM at all times, open or not,
  // so a fixed id put two inputs with the same one on every page and left a
  // label pointing at the wrong field. Generated per instance instead.
  const inputId = useId()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    const trimmed = term.trim()
    // An empty search means "show me everything", which is a legitimate thing
    // to ask for and lands on the unfiltered catalog.
    void navigate(
      trimmed ? `/products?q=${encodeURIComponent(trimmed)}` : '/products',
    )
    onSubmitted?.()
  }

  return (
    <form
      className={cn('relative', variant === 'bar' ? 'w-56 xl:w-64' : 'w-full')}
      onSubmit={submit}
      role="search"
    >
      <label className="sr-only" htmlFor={inputId}>
        Search software
      </label>
      <Search
        aria-hidden="true"
        className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink/45"
        size={16}
      />
      <input
        autoComplete="off"
        className={cn(
          'w-full rounded-sm border border-ink/20 bg-surface pr-3 pl-9 text-sm text-ink placeholder:text-ink/45 focus-visible:border-bronze',
          variant === 'bar' ? 'h-10' : 'h-12',
        )}
        id={inputId}
        name="q"
        onChange={(event) => setTerm(event.target.value)}
        placeholder="Search software"
        type="search"
        value={term}
      />
      {/* An explicit submit rather than relying on implicit submission. It
          costs nothing visually and means Enter, the on-screen keyboard's Go
          key and a screen reader's forms mode all do the same thing. */}
      <button className="sr-only" type="submit">
        Search
      </button>
    </form>
  )
}
