import { Skeleton } from '../../../components/ui/Skeleton'

export function CatalogLoadingGrid() {
  return (
    <div
      aria-label="Loading software"
      className="grid grid-cols-1 border-l border-t border-ink/15 xs:grid-cols-2 xl:grid-cols-3"
      role="status"
    >
      {Array.from({ length: 6 }, (_, index) => (
        <div className="border-b border-r border-ink/15 p-4 sm:p-5" key={index}>
          <Skeleton className="aspect-[4/3] w-full" />
          <Skeleton className="mt-5 h-3 w-20" />
          <Skeleton className="mt-3 h-7 w-4/5" />
          <Skeleton className="mt-5 h-16 w-full" />
          <Skeleton className="mt-6 h-10 w-full" />
        </div>
      ))}
      <span className="sr-only">Loading software catalog.</span>
    </div>
  )
}
