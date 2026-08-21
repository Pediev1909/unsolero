import { ChevronDown } from 'lucide-react'
import {
  useEffect,
  useId,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useLocation } from 'react-router-dom'

import { cn } from '../../lib/styles/cn'

interface NavigationMenuProps {
  label: string
  active?: boolean
  /** Rendered inside the panel. Receives a closer so links can dismiss it. */
  children: (close: () => void) => ReactNode
  /** A mega panel spans the viewport; a small one hangs off the trigger. */
  width?: 'wide' | 'auto'
}

/**
 * A disclosure menu for the header.
 *
 * It opens on click rather than on hover, deliberately. Hover menus open when
 * a cursor merely crosses them, they have no equivalent on a touchscreen, and
 * they need a delay tuned by guesswork to stop flickering. A click behaves the
 * same way for a mouse, a finger and a keyboard, which is the point when the
 * brief is that someone using a computer for the first time should manage it.
 *
 * Escape closes it and puts focus back on the trigger, because a keyboard user
 * who dismisses a menu and lands at the top of the document has been punished
 * for using the keyboard.
 */
export function NavigationMenu({
  label,
  active = false,
  children,
  width = 'auto',
}: NavigationMenuProps) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const panelId = useId()
  const location = useLocation()

  const close = () => setOpen(false)

  // Navigating away must close the menu, or a link inside the panel changes
  // the page underneath while the panel stays open over it. Links in the panel
  // close it themselves; this catches the rest, including the back button.
  //
  // Adjusted during render rather than in an effect. React documents this
  // pattern for state that has to follow a prop, and an effect here would
  // render the open menu once against the new page before closing it.
  const [renderedAt, setRenderedAt] = useState(
    () => location.pathname + location.search,
  )
  const currentLocation = location.pathname + location.search
  if (renderedAt !== currentLocation) {
    setRenderedAt(currentLocation)
    if (open) setOpen(false)
  }

  useEffect(() => {
    if (!open) return

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return
      setOpen(false)
      triggerRef.current?.focus()
    }
    const onPointerDown = (event: PointerEvent) => {
      if (!containerRef.current) return
      if (event.target instanceof Node && containerRef.current.contains(event.target)) {
        return
      }
      setOpen(false)
    }
    // Focus leaving the menu entirely closes it, so tabbing past the last link
    // does not leave a panel hanging open behind the rest of the page.
    const onFocusIn = (event: FocusEvent) => {
      if (!containerRef.current) return
      if (event.target instanceof Node && containerRef.current.contains(event.target)) {
        return
      }
      setOpen(false)
    }

    document.addEventListener('keydown', onKeyDown)
    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('focusin', onFocusIn)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('focusin', onFocusIn)
    }
  }, [open])

  return (
    <div className="relative" ref={containerRef}>
      <button
        aria-controls={panelId}
        // Browse has no page of its own, so colour alone would be the only
        // signal that you are inside it. aria-current says so out loud.
        aria-current={active ? 'true' : undefined}
        aria-expanded={open}
        className={cn(
          'nav-link inline-flex items-center gap-1',
          active && 'text-bronze',
        )}
        onClick={() => setOpen((value) => !value)}
        ref={triggerRef}
        type="button"
      >
        {label}
        <ChevronDown
          aria-hidden="true"
          className={cn('transition-transform', open && 'rotate-180')}
          size={15}
        />
      </button>

      {open && (
        <div
          className={cn(
            'absolute top-full z-50 mt-3 rounded-sm border border-ink/15 bg-surface p-6 shadow-overlay',
            width === 'wide'
              ? 'left-1/2 w-[min(58rem,calc(100vw-3rem))] -translate-x-1/2'
              : 'left-0 w-64',
          )}
          id={panelId}
        >
          {children(close)}
        </div>
      )}
    </div>
  )
}
