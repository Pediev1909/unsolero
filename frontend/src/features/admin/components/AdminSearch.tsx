import { Search } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { cn } from '../../../lib/styles/cn'
import { useAdminReferences } from '../queries'

// Every section searched only itself, so finding one product meant knowing
// which section it lived in first. The references endpoint already returns
// products, categories, brands and merchants in a single call, so the whole
// catalog is searchable without adding anything to the API.
//
// Section names are searchable too: an administrator looking for "offers" wants
// the page, not a record, and typing the word is faster than reading a sidebar.

interface Destination {
  group: string
  label: string
  detail?: string
  to: string
}

export function AdminSearch({
  sections,
}: {
  sections: { label: string; to: string }[]
}) {
  const references = useAdminReferences()
  const navigate = useNavigate()
  const [term, setTerm] = useState('')
  const [highlight, setHighlight] = useState(0)
  const input = useRef<HTMLInputElement>(null)
  const container = useRef<HTMLDivElement>(null)

  // The shortcut every search field in a tool of this shape answers to.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        input.current?.focus()
        input.current?.select()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  // A click anywhere else dismisses the results, which is what a reader who has
  // finished looking expects, and stops the panel covering the page underneath.
  useEffect(() => {
    const onPointerDown = (event: MouseEvent) => {
      if (!container.current?.contains(event.target as Node)) setTerm('')
    }
    window.addEventListener('mousedown', onPointerDown)
    return () => window.removeEventListener('mousedown', onPointerDown)
  }, [])

  const results = useMemo(() => {
    const needle = term.trim().toLowerCase()
    if (needle.length < 2) return []
    const data = references.data
    const matches: Destination[] = []
    const take = (list: Destination[]) => {
      for (const item of list) {
        if (matches.length >= 12) return
        if (item.label.toLowerCase().includes(needle)) matches.push(item)
      }
    }

    take(sections.map((s) => ({ group: 'Section', label: s.label, to: s.to })))
    if (data) {
      take(
        data.products.map((p) => ({
          group: 'Product',
          label: p.name,
          detail: p.slug,
          to: `/admin/products/${p.id}`,
        })),
      )
      take(
        data.categories.map((c) => ({
          group: 'Category',
          label: c.name,
          detail: c.slug,
          to: '/admin/categories',
        })),
      )
      take(
        data.brands.map((b) => ({
          group: 'Brand',
          label: b.name,
          detail: b.slug,
          to: '/admin/brands',
        })),
      )
      take(
        data.merchants.map((m) => ({
          group: 'Merchant',
          label: m.name,
          to: '/admin/merchants',
        })),
      )
    }
    return matches
  }, [term, references.data, sections])

  // Clamped at the point of use rather than reset from an effect: the result
  // list also shrinks when the references finish loading, and a stored index
  // would then point past the end.
  const selected = Math.min(highlight, Math.max(0, results.length - 1))

  const go = (destination: Destination | undefined) => {
    if (!destination) return
    setTerm('')
    void navigate(destination.to)
  }

  return (
    <div className="relative" ref={container}>
      <label className="sr-only" htmlFor="admin-search">
        Search products, categories, brands, merchants and sections
      </label>
      <div className="flex items-center gap-2 border border-ink/15 bg-surface px-3">
        <Search aria-hidden="true" className="text-ink/50" size={16} />
        <input
          autoComplete="off"
          className="min-h-10 w-full bg-transparent text-sm outline-none placeholder:text-ink/50"
          id="admin-search"
          onChange={(event) => {
            setTerm(event.target.value)
            setHighlight(0)
          }}
          onKeyDown={(event) => {
            if (event.key === 'ArrowDown') {
              event.preventDefault()
              setHighlight(Math.min(selected + 1, results.length - 1))
            } else if (event.key === 'ArrowUp') {
              event.preventDefault()
              setHighlight(Math.max(selected - 1, 0))
            } else if (event.key === 'Enter') {
              event.preventDefault()
              go(results[selected])
            } else if (event.key === 'Escape') {
              setTerm('')
            }
          }}
          placeholder="Search everything"
          ref={input}
          type="search"
          value={term}
        />
        <kbd className="hidden shrink-0 text-caption text-ink/50 sm:block">
          Ctrl K
        </kbd>
      </div>

      {term.trim().length >= 2 && (
        <div
          className="absolute inset-x-0 top-full z-30 mt-1 max-h-96 overflow-y-auto border border-ink/15 bg-surface shadow-overlay"
          role="listbox"
        >
          {results.length === 0 ? (
            <p className="px-3 py-3 text-sm text-ink/70">
              {references.isPending ? 'Loading records…' : 'Nothing matches.'}
            </p>
          ) : (
            results.map((result, index) => (
              <button
                aria-selected={index === selected}
                className={cn(
                  'flex w-full items-center justify-between gap-3 px-3 py-2.5 text-left text-sm',
                  index === selected ? 'bg-paper' : 'hover:bg-paper/60',
                )}
                key={`${result.group}-${result.to}-${result.label}`}
                onClick={() => go(result)}
                onMouseEnter={() => setHighlight(index)}
                role="option"
                type="button"
              >
                <span className="truncate">
                  {result.label}
                  {result.detail && (
                    <span className="ml-2 text-ink/60">{result.detail}</span>
                  )}
                </span>
                <span className="shrink-0 text-caption uppercase tracking-[0.12em] text-ink/60">
                  {result.group}
                </span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}
