import {
  forwardRef,
  useEffect,
  useId,
  useState,
  type ChangeEvent,
  type InputHTMLAttributes,
  type ReactNode,
} from 'react'

import { cn } from '../../lib/styles/cn'

export interface SliderProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'type'
> {
  label: ReactNode
  hint?: ReactNode
  formatValue?: (value: number) => string
}

export const Slider = forwardRef<HTMLInputElement, SliderProps>(function Slider(
  {
    label,
    hint,
    formatValue = String,
    className,
    id: suppliedId,
    value,
    defaultValue,
    min = 0,
    max = 100,
    onChange,
    ...props
  },
  ref,
) {
  const generatedId = useId()
  const id = suppliedId ?? generatedId
  const [displayValue, setDisplayValue] = useState(() =>
    Number(value ?? defaultValue ?? min),
  )

  useEffect(() => {
    if (value !== undefined) {
      setDisplayValue(Number(value))
    }
  }, [value])

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    setDisplayValue(Number(event.currentTarget.value))
    onChange?.(event)
  }

  const lower = Number(min)
  const upper = Number(max)
  const span = upper - lower
  const filledPercent =
    span > 0
      ? Math.min(100, Math.max(0, ((displayValue - lower) / span) * 100))
      : 0

  return (
    <div>
      <div className="mb-3 flex items-baseline justify-between gap-4">
        <label
          className="text-label font-bold uppercase tracking-[0.12em]"
          htmlFor={id}
        >
          {label}
        </label>
        <output
          className="font-mono text-xs font-semibold text-bronze-dark"
          htmlFor={id}
        >
          {formatValue(displayValue)}
        </output>
      </div>
      {/* appearance-none strips the native track, and nothing was drawn to
          replace it — so the control rendered as a lone dot floating in white
          space with no indication it could be dragged. The rail is painted
          here as a gradient that fills to the current value, which also shows
          at a glance where in the range you are. */}
      <input
        {...props}
        aria-describedby={hint ? `${id}-hint` : undefined}
        className={cn('slider-input', className)}
        defaultValue={defaultValue}
        id={id}
        max={max}
        min={min}
        onChange={handleChange}
        ref={ref}
        style={{
          background:
            `linear-gradient(to right,` +
            ` var(--color-bronze) 0%, var(--color-bronze) ${filledPercent}%,` +
            ` var(--color-line) ${filledPercent}%, var(--color-line) 100%)`,
        }}
        type="range"
        value={value}
      />
      {hint && (
        <p className="mt-2 text-xs leading-5 text-ink/70" id={`${id}-hint`}>
          {hint}
        </p>
      )}
    </div>
  )
})
