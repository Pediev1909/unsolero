import { useState } from 'react'
import { Controller, type UseFormReturn } from 'react-hook-form'

import { Input } from '../../../components/ui/Input'
import { Slider } from '../../../components/ui/Slider'
import { formatMinorCurrency } from '../../../lib/money/format'
import type { BuilderValues } from '../schemas'
import {
  existingToolOptions,
  experienceOptions,
  goalOptions,
  preferenceOptions,
  priorityOptions,
} from '../options'
import { ChoiceCard } from './ChoiceCard'

interface RecommendationStepProps {
  step: number
  form: UseFormReturn<BuilderValues>
}

export function RecommendationStep({ step, form }: RecommendationStepProps) {
  const values = form.watch()

  if (step === 0) {
    return (
      <div className="grid gap-3 md:grid-cols-2">
        {goalOptions.map((option) => (
          <ChoiceCard
            checked={values.goal === option.value}
            description={option.description}
            key={option.value}
            label={option.label}
            name="goal"
            onChange={() =>
              form.setValue('goal', option.value, { shouldDirty: true })
            }
            value={option.value}
          />
        ))}
      </div>
    )
  }
  if (step === 1) {
    return (
      <div className="grid gap-3 md:grid-cols-3">
        {experienceOptions.map((option) => (
          <ChoiceCard
            checked={values.experience === option.value}
            description={option.description}
            key={option.value}
            label={option.label}
            name="experience"
            onChange={() =>
              form.setValue('experience', option.value, { shouldDirty: true })
            }
            value={option.value}
          />
        ))}
      </div>
    )
  }
  if (step === 2) {
    return (
      <div className="max-w-2xl space-y-8 border border-ink/15 bg-surface p-5 sm:p-8">
        <Controller
          control={form.control}
          name="budget_minor"
          render={({ field }) => (
            <Slider
              formatValue={(value) => formatMinorCurrency(value, 'USD')}
              hint="This is the total monthly spend for the whole stack, not a limit per tool."
              label="Complete setup budget"
              max={500_000}
              min={10_000}
              onChange={(event) => field.onChange(Number(event.target.value))}
              step={5_000}
              value={field.value}
            />
          )}
        />
        <Controller
          control={form.control}
          name="budget_minor"
          render={({ field }) => (
            <BudgetInput onChange={field.onChange} valueMinor={field.value} />
          )}
        />
      </div>
    )
  }
  if (step === 3) {
    return (
      <div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {existingToolOptions.map((option) => {
            const checked = values.existing_equipment.some(
              (item) => item.category_slug === option.category_slug,
            )
            return (
              <ChoiceCard
                checked={checked}
                key={option.category_slug}
                label={option.name}
                onChange={() =>
                  form.setValue(
                    'existing_equipment',
                    checked
                      ? values.existing_equipment.filter(
                          (item) => item.category_slug !== option.category_slug,
                        )
                      : [...values.existing_equipment, option],
                    { shouldDirty: true },
                  )
                }
                type="checkbox"
              />
            )
          })}
        </div>
        <p className="mt-5 text-sm text-ink/70">
          Nothing yet? Leave every option unselected and continue.
        </p>
      </div>
    )
  }
  if (step === 4) {
    return (
      <MultiChoice
        options={preferenceOptions}
        selected={values.training_preferences}
        onChange={(selected) =>
          form.setValue('training_preferences', selected, { shouldDirty: true })
        }
      />
    )
  }
  // Priorities is the last step. There was a seventh that collected free text,
  // and its own hint admitted the engine "does not interpret this note yet" --
  // which was true and had been true since the vertical changed. A question
  // that changes nothing is worse than no question, because the reader spends
  // real effort on it and reasonably assumes it counted.
  return (
    <MultiChoice
      options={priorityOptions}
      selected={values.priorities}
      onChange={(selected) =>
        form.setValue('priorities', selected, { shouldDirty: true })
      }
    />
  )
}

function MultiChoice<T extends string>({
  options,
  selected,
  onChange,
}: {
  options: readonly { value: T; label: string }[]
  selected: T[]
  onChange: (value: T[]) => void
}) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {options.map((option) => {
        const checked = selected.includes(option.value)
        return (
          <ChoiceCard
            checked={checked}
            key={option.value}
            label={option.label}
            onChange={() =>
              onChange(
                checked
                  ? selected.filter((value) => value !== option.value)
                  : [...selected, option.value],
              )
            }
            type="checkbox"
          />
        )
      })}
    </div>
  )
}

/**
 * The exact-budget field.
 *
 * It used to derive its text straight from the form value, which meant it
 * could never be empty: clear it and Number('') became 0, the field redrew as
 * "0", and typing 4000 after it produced "04000". You could not delete the
 * zero. On the question that decides everything else.
 *
 * The text is local state now, so it can be empty while you type, and only a
 * value that actually parses is pushed to the form. The clamp happens when you
 * leave the field rather than on every keystroke — clamping mid-typing is what
 * turns "1" into "100" before you have finished writing "1500".
 */
export function BudgetInput({
  valueMinor,
  onChange,
}: {
  valueMinor: number
  onChange: (value: number) => void
}) {
  const [text, setText] = useState(() => String(valueMinor / 100))
  const [editing, setEditing] = useState(false)

  // While the field is not being typed in, it follows the slider.
  if (!editing && text !== String(valueMinor / 100)) {
    setText(String(valueMinor / 100))
  }

  return (
    <Input
      hint="Between $100 and $5,000 a month."
      inputMode="numeric"
      label="Exact budget in dollars"
      max={5_000}
      min={100}
      onBlur={() => {
        setEditing(false)
        const parsed = Number(text)
        const clamped = Number.isFinite(parsed)
          ? Math.min(5_000, Math.max(100, Math.round(parsed)))
          : 100
        setText(String(clamped))
        onChange(clamped * 100)
      }}
      onChange={(event) => {
        const next = event.target.value
        setText(next)
        setEditing(true)
        const parsed = Number(next)
        if (next !== '' && Number.isFinite(parsed) && parsed >= 100) {
          onChange(Math.min(5_000, Math.round(parsed)) * 100)
        }
      }}
      onFocus={() => setEditing(true)}
      type="number"
      value={text}
    />
  )
}
