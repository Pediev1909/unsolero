import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { billingPeriods, billingUnits } from '../../catalog/schemas'
import {
  billingPeriodLabels,
  billingUnitLabels,
  unitNoteMaxLength,
  type BillingFieldName,
  type BillingFormValues,
} from '../billingForm'

interface BillingFieldsProps {
  value: BillingFormValues
  onChange: (next: BillingFormValues) => void
  errors?: Partial<Record<BillingFieldName, string | undefined>>
}

/**
 * The four billing controls, shared by the product editor and the evidence
 * revision form so the two cannot drift apart. Controlled rather than
 * registered: one form runs on react-hook-form and the other on local state,
 * and a value-and-onChange pair fits both. Renders a fragment so the caller's
 * grid lays the fields out with its own.
 */
export function BillingFields({
  value,
  onChange,
  errors = {},
}: BillingFieldsProps) {
  const monthly = value.billing_period === 'monthly'

  return (
    <>
      <Select
        error={errors.billing_period}
        hint="Which contract the price is quoted on."
        label="Billing period"
        name="billing_period"
        onChange={(event) => {
          const period = event.target
            .value as BillingFormValues['billing_period']
          // A yearly rate only exists beside a monthly one. Clearing it here
          // spares the operator a message about a field they can no longer
          // reach.
          onChange({
            ...value,
            billing_period: period,
            annual_price_minor:
              period === 'monthly' ? value.annual_price_minor : '',
          })
        }}
        value={value.billing_period}
      >
        {billingPeriods.map((period) => (
          <option key={period} value={period}>
            {billingPeriodLabels[period]}
          </option>
        ))}
      </Select>
      <Select
        error={errors.pricing_unit}
        hint="What one unit of the price buys."
        label="Pricing unit"
        name="pricing_unit"
        onChange={(event) =>
          onChange({
            ...value,
            pricing_unit: event.target
              .value as BillingFormValues['pricing_unit'],
          })
        }
        value={value.pricing_unit}
      >
        {billingUnits.map((unit) => (
          <option key={unit} value={unit}>
            {billingUnitLabels[unit]}
          </option>
        ))}
      </Select>
      <Input
        error={errors.unit_note}
        hint="The vendor's own words for the unit, such as “Up to 500 contacts”. Printed in place of the generic phrase for contact, transaction and usage units."
        label="Unit note"
        maxLength={unitNoteMaxLength}
        name="unit_note"
        onChange={(event) =>
          onChange({ ...value, unit_note: event.target.value })
        }
        value={value.unit_note}
      />
      <Input
        disabled={!monthly}
        error={errors.annual_price_minor}
        hint={
          monthly
            ? 'The per-month figure on the yearly contract, in minor units. Leave empty if the vendor sells no yearly plan.'
            : 'Only a product sold month to month has a separate yearly rate.'
        }
        inputMode="numeric"
        label="Annual price per month (minor units)"
        min={0}
        name="annual_price_minor"
        onChange={(event) =>
          onChange({ ...value, annual_price_minor: event.target.value })
        }
        step={1}
        type="number"
        value={value.annual_price_minor}
      />
    </>
  )
}
