import { cloneElement, useId, type ReactElement, type ReactNode } from 'react'

interface TooltipChildProps {
  'aria-describedby'?: string
}

interface TooltipProps {
  content: ReactNode
  children: ReactElement<TooltipChildProps>
  placement?: 'top' | 'bottom'
}

export function Tooltip({
  content,
  children,
  placement = 'top',
}: TooltipProps) {
  const id = useId()
  const describedBy = [children.props['aria-describedby'], id]
    .filter(Boolean)
    .join(' ')

  return (
    <span className="group/tooltip relative inline-flex">
      {cloneElement(children, { 'aria-describedby': describedBy })}
      <span className={cnTooltip(placement)} id={id} role="tooltip">
        {content}
      </span>
    </span>
  )
}

function cnTooltip(placement: 'top' | 'bottom'): string {
  const position =
    placement === 'top'
      ? 'bottom-[calc(100%+0.5rem)]'
      : 'top-[calc(100%+0.5rem)]'
  return `${position} pointer-events-none absolute left-1/2 z-50 w-max max-w-56 -translate-x-1/2 rounded-xs bg-ink px-2.5 py-1.5 text-center text-xs leading-5 text-canvas opacity-0 shadow-raised transition-opacity duration-150 group-hover/tooltip:opacity-100 group-focus-within/tooltip:opacity-100`
}
