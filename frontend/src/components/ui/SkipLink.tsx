interface SkipLinkProps {
  href?: string
  label?: string
}

export function SkipLink({
  href = '#main-content',
  label = 'Skip to main content',
}: SkipLinkProps) {
  return (
    <a
      className="sr-only z-50 bg-ink px-4 py-3 text-sm font-semibold text-canvas focus:not-sr-only focus:fixed focus:left-3 focus:top-3"
      href={href}
    >
      {label}
    </a>
  )
}
