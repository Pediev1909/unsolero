export type ClassValue = string | false | null | undefined

// Concatenation, not merging. Two conflicting Tailwind classes both survive,
// and which one wins is decided by their order in the generated stylesheet
// rather than by the order they were written here.
//
// So a className passed into a component CANNOT be relied on to override that
// component's own colours — add a variant instead. Learned from a Compare
// button that reached production as bg-ink on bg-ink: present, focusable,
// announced by a screen reader, and invisible to everyone else.
export function cn(...values: ClassValue[]): string {
  return values.filter(Boolean).join(' ')
}
