import { Bookmark, Scale } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { comparisonLimit } from '../productSelections'

interface CatalogSelectionTrayProps {
  /** How many products are in the comparison right now. */
  comparedCount: number
  /** How many products are on the saved list right now. */
  savedCount: number
  onOpenComparison: () => void
}

/**
 * The bar fixed to the bottom of a listing while the reader has something
 * selected: the comparison count with its Compare button, and the saved count
 * with the way to the saved list. It draws nothing when both are empty, so a
 * reader who has chosen nothing is not shown an empty tray.
 */
export function CatalogSelectionTray({
  comparedCount,
  savedCount,
  onOpenComparison,
}: CatalogSelectionTrayProps) {
  if (comparedCount === 0 && savedCount === 0) return null
  return (
    <div
      aria-label="Your selections"
      className="fixed inset-x-0 z-30 mx-auto flex w-[min(calc(100%-2rem),32rem)] flex-wrap items-center justify-between gap-x-4 gap-y-2 border border-ink/15 bg-ink px-4 py-2 text-canvas shadow-overlay"
      role="region"
      // Sits above the consent banner while that is showing, rather than
      // underneath it where its Compare button could not be clicked. The
      // safe-area term keeps it clear of a device's own bottom furniture;
      // the toolbar carries the same action in normal flow, so this bar
      // being obscured by anything is an inconvenience rather than a dead
      // end.
      style={{
        bottom:
          'calc(var(--bottom-bar-offset, 0px) + env(safe-area-inset-bottom, 0px) + 1rem)',
      }}
    >
      {comparedCount > 0 && (
        <div className="flex items-center gap-3">
          <p className="text-sm">
            <strong>{comparedCount}</strong> of {comparisonLimit} selected
          </p>
          <Button onClick={onOpenComparison} size="sm" variant="inverse">
            <Scale aria-hidden="true" size={16} /> Compare
          </Button>
        </div>
      )}
      {savedCount > 0 && (
        <Link
          className="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-canvas underline-offset-4 hover:underline"
          to="/wishlist"
        >
          <Bookmark aria-hidden="true" size={16} />
          Saved {savedCount}
          <span aria-hidden="true">·</span>
          View
        </Link>
      )}
    </div>
  )
}
