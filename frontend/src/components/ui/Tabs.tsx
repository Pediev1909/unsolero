import { useId, useState, type KeyboardEvent, type ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'

export interface TabItem {
  id: string
  label: string
  content: ReactNode
  disabled?: boolean
}

interface TabsProps {
  items: TabItem[]
  value?: string
  defaultValue?: string
  onValueChange?: (value: string) => void
  ariaLabel: string
  className?: string
}

export function Tabs({
  items,
  value,
  defaultValue,
  onValueChange,
  ariaLabel,
  className,
}: TabsProps) {
  const firstEnabled = items.find((item) => !item.disabled)?.id ?? ''
  const [internalValue, setInternalValue] = useState(
    defaultValue ?? firstEnabled,
  )
  const activeValue = value ?? internalValue
  const baseId = useId()

  function select(nextValue: string) {
    if (value === undefined) setInternalValue(nextValue)
    onValueChange?.(nextValue)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) {
      return
    }
    event.preventDefault()
    const tabs = Array.from(
      event.currentTarget.parentElement?.querySelectorAll<HTMLButtonElement>(
        '[role="tab"]:not(:disabled)',
      ) ?? [],
    )
    const currentIndex = tabs.indexOf(event.currentTarget)
    let nextIndex = currentIndex
    if (event.key === 'Home') nextIndex = 0
    if (event.key === 'End') nextIndex = tabs.length - 1
    if (event.key === 'ArrowRight') {
      nextIndex = (currentIndex + 1) % tabs.length
    }
    if (event.key === 'ArrowLeft') {
      nextIndex = (currentIndex - 1 + tabs.length) % tabs.length
    }
    const nextTab = tabs[nextIndex]
    nextTab?.focus()
    if (nextTab?.dataset.value) select(nextTab.dataset.value)
  }

  return (
    <div className={className}>
      <div
        aria-label={ariaLabel}
        className="flex gap-6 overflow-x-auto border-b border-ink/15"
        role="tablist"
      >
        {items.map((item) => {
          const selected = item.id === activeValue
          return (
            <button
              aria-controls={`${baseId}-panel-${item.id}`}
              aria-selected={selected}
              className={cn(
                'relative shrink-0 border-0 bg-transparent px-0 pb-3 pt-1 text-sm font-semibold text-ink/70 transition-colors duration-150 hover:text-ink disabled:cursor-not-allowed disabled:opacity-35',
                selected &&
                  'text-ink after:absolute after:inset-x-0 after:bottom-0 after:h-0.5 after:bg-bronze',
              )}
              data-value={item.id}
              disabled={item.disabled}
              id={`${baseId}-tab-${item.id}`}
              key={item.id}
              onClick={() => select(item.id)}
              onKeyDown={handleKeyDown}
              role="tab"
              tabIndex={selected ? 0 : -1}
              type="button"
            >
              {item.label}
            </button>
          )
        })}
      </div>
      {items.map((item) => (
        <div
          aria-labelledby={`${baseId}-tab-${item.id}`}
          hidden={item.id !== activeValue}
          id={`${baseId}-panel-${item.id}`}
          key={item.id}
          role="tabpanel"
          tabIndex={0}
        >
          {item.content}
        </div>
      ))}
    </div>
  )
}
