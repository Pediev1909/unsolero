// Every build gives its chunks new hashed filenames and deletes the old ones.
// A visitor who already had the page open when a deploy landed is therefore
// holding a document that points at files the server no longer has, and the
// next lazy-loaded route fails on a 404 rather than anything being wrong with
// the application. Registering, which navigates to a lazy route on submit, hit
// this: the account was created and only the page after it failed to load.
//
// Reloading fetches the current document and its current chunk names, which
// resolves it. The guard is what makes that safe: if a reload does not fix the
// problem, the second attempt within the window falls through to the error
// page instead of reloading forever.

const reloadMarkerKey = 'unsolero:stale-build-reload'
const reloadWindowMs = 30_000

// Browsers word this differently and none of them give it a code, so matching
// the message is the only option available.
const staleChunkMessages = [
  'failed to fetch dynamically imported module',
  'error loading dynamically imported module',
  'importing a module script failed',
  'failed to load module script',
]

export function isStaleBuildError(error: unknown): boolean {
  const message =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
        ? error
        : ''
  if (!message) return false
  const normalized = message.toLowerCase()
  return staleChunkMessages.some((candidate) => normalized.includes(candidate))
}

export function recoverFromStaleBuild(): boolean {
  let previous: string | null
  try {
    previous = window.sessionStorage.getItem(reloadMarkerKey)
  } catch {
    // A blocked session store cannot record the attempt, so reloading could
    // loop. Showing the error page is the safe direction to fail in.
    return false
  }

  if (previous && Date.now() - Number(previous) < reloadWindowMs) {
    return false
  }

  try {
    window.sessionStorage.setItem(reloadMarkerKey, String(Date.now()))
  } catch {
    return false
  }

  window.location.reload()
  return true
}
