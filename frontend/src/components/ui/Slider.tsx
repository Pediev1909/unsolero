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
      <input
        {...props}
        aria-describedby={hint ? `${id}-hint` : undefined}
        className={cn(
          'h-5 w-full cursor-pointer appearance-none bg-transparent accent-bronze disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
        defaultValue={defaultValue}
        id={id}
        max={max}
        min={min}
        onChange={handleChange}
        ref={ref}
        type="range"
        value={value}
      />
      {hint && (
        <p className="mt-2 text-xs leading-5 text-ink/55" id={`${id}-hint`}>
          {hint}
        </p>
      )}
    </div>
  )
})
