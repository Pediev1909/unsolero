// Left to itself the browser restores a scroll offset on a fresh document, and
// on this site that meant carrying the previous page's offset into the next
// one: a link followed from halfway down the catalog opened the product page
// most of the way down it. Claiming control has to happen before React renders,
// so it lives here rather than in the component that uses it.
export function claimScrollControl() {
  if ('scrollRestoration' in history) history.scrollRestoration = 'manual'
}
