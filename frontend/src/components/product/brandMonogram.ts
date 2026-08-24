// The mark and the colour a vendor gets when no artwork exists for it. Kept
// apart from the component so both can be tested directly, and so the component
// file exports only a component.

// The variation lives in the mark, not the ground. Four tinted grounds at the
// same value were indistinguishable in a grid — the difference existed in the
// hex and nowhere in the eye. A small solid mark carries the same signal at a
// size where it reads, and keeps every tile on one ground.
//
// Semantic amber and ember are deliberately absent: a tile coloured like a
// warning reads as one.
export const monogramGround = '#eceef1'

export const monogramMarks = [
  '#14605a',
  '#2c6349',
  '#1c1f25',
  '#3f4a56',
] as const

// A stable, case-insensitive hash, so a category always lands on the same mark
// across sessions and across every surface that renders a card.
export function markColourFor(key: string) {
  let hash = 0
  for (const character of key.toLowerCase()) {
    hash = (hash * 31 + character.charCodeAt(0)) >>> 0
  }
  return monogramMarks[hash % monogramMarks.length] ?? monogramMarks[0]
}

// The vendor's own capitalisation is part of the name: monday.com and n8n are
// lowercase on purpose, and an uppercased monogram would be a small
// misspelling of a brand this site asks people to trust it about.
export function monogramFor(brand: string) {
  const words = brand.trim().split(/\s+/).filter(Boolean)
  const first = words[0]
  if (!first) return '—'

  // An initialism that is already the brand's short form: SE Ranking, IBM.
  if (first.length >= 2 && first.length <= 4 && first === first.toUpperCase()) {
    return first
  }

  const second = words[1]
  if (second) return first.slice(0, 1) + second.slice(0, 1)
  return first.slice(0, 1)
}
