import { BrandMark } from '../components/layout/BrandMark'

export function RouteLoadingState() {
  return (
    <main
      aria-label="Loading page"
      className="grid min-h-screen place-items-center bg-canvas px-6"
      role="status"
    >
      <div className="text-center">
        <BrandMark />
        <div className="mx-auto mt-6 h-px w-28 overflow-hidden bg-ink/15">
          <div className="h-full w-1/2 animate-pulse bg-bronze" />
        </div>
        <span className="sr-only">Loading page</span>
      </div>
    </main>
  )
}
