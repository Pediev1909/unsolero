import { z } from 'zod'

import { billingPeriods, billingUnits, type Billing } from '../catalog/schemas'

/**
 * The billing basis as an operator types it, and the rules that turn it into
 * the object the API takes.
 *
 * Shared by the product editor and the evidence revision form. The second one
 * matters more than it looks: a published product's editor is read-only, so a
 * new revision is the only way to correct the twenty-five products whose
 * prices were recorded on a yearly basis and shown as monthly.
 */

export const billingPeriodLabels: Record<Billing['period'], string> = {
  monthly: 'Monthly — the vendor sells month to month',
  annual: 'Annual only — the price is the per-month equivalent',
  free: 'Free tier — the price is 0',
  usage: 'Usage-based — the price is the entry usage rate',
}

export const billingUnitLabels: Record<Billing['unit'], string> = {
  flat: 'Flat rate',
  per_user: 'Per user',
  per_contacts: 'Per contact tier',
  per_transaction: 'Per transaction',
  usage: 'Usage',
}

export const unitNoteMaxLength = 120

export const billingFieldNames = [
  'billing_period',
  'pricing_unit',
  'unit_note',
  'annual_price_minor',
] as const

export type BillingFieldName = (typeof billingFieldNames)[number]

/**
 * What the four controls hold, as typed. Strings throughout, because that is
 * what a select and a text box produce; the schema does the reading.
 */
export interface BillingFormValues {
  billing_period: Billing['period']
  pricing_unit: Billing['unit']
  unit_note: string
  annual_price_minor: string
}

/**
 * The per-field rules. Spread into a larger form schema, or used on their own
 * through `parseBillingForm`; the cross-field rules live in `billingIssues`
 * because they need the product's price, which is not one of these fields.
 */
export const billingFormFields = {
  billing_period: z.enum(billingPeriods, { error: 'Choose a billing period.' }),
  pricing_unit: z.enum(billingUnits, { error: 'Choose a pricing unit.' }),
  unit_note: z
    .string()
    .trim()
    .max(
      unitNoteMaxLength,
      `Keep the note to ${unitNoteMaxLength} characters.`,
    ),
  annual_price_minor: z
    .string()
    .trim()
    .transform((value, ctx) => {
      if (value === '') return null
      if (!/^\d+$/.test(value)) {
        ctx.addIssue({
          code: 'custom',
          message: 'Whole minor units — 1500 for $15.00.',
        })
        return z.NEVER
      }
      return Number(value)
    }),
}

const billingFieldsSchema = z.object(billingFormFields)

export type BillingFormParsed = z.output<typeof billingFieldsSchema>

export interface BillingIssue {
  path: BillingFieldName
  message: string
}

/**
 * The rules that span fields, mirrored from the server so an operator meets
 * them as a message beside the control rather than as a 422: a separate
 * yearly rate exists only beside a monthly one, and a free tier costs nothing.
 * `priceMinor` is null when the caller has no price to check against.
 */
export function billingIssues(
  values: BillingFormParsed,
  priceMinor: number | null,
): BillingIssue[] {
  const issues: BillingIssue[] = []
  if (
    values.annual_price_minor !== null &&
    values.billing_period !== 'monthly'
  ) {
    issues.push({
      path: 'annual_price_minor',
      message:
        'Only a product sold month to month has a separate yearly rate. Clear this, or set the period to monthly.',
    })
  }
  if (
    values.billing_period === 'free' &&
    priceMinor !== null &&
    priceMinor !== 0
  ) {
    issues.push({
      path: 'billing_period',
      message:
        'A free tier has a price of 0. Set the price to 0, or choose another period.',
    })
  }
  return issues
}

export function toBilling(values: BillingFormParsed): Billing {
  return {
    period: values.billing_period,
    unit: values.pricing_unit,
    unit_note: values.unit_note === '' ? null : values.unit_note,
    annual_price_minor: values.annual_price_minor,
  }
}

/** A product without the object starts as a monthly flat rate, which is what the catalog assumed for everything until now. */
export function fromBilling(billing: Billing | undefined): BillingFormValues {
  return {
    billing_period: billing?.period ?? 'monthly',
    pricing_unit: billing?.unit ?? 'flat',
    unit_note: billing?.unit_note ?? '',
    annual_price_minor:
      billing?.annual_price_minor == null
        ? ''
        : String(billing.annual_price_minor),
  }
}

export type BillingFormResult =
  | { ok: true; billing: Billing }
  | { ok: false; errors: Partial<Record<BillingFieldName, string>> }

/**
 * Reads the four controls against the product's price and returns either the
 * API object or one message per field — the first problem with each, which is
 * the one worth showing.
 */
export function parseBillingForm(
  values: BillingFormValues,
  priceMinor: number | null,
): BillingFormResult {
  const errors: Partial<Record<BillingFieldName, string>> = {}
  const parsed = billingFieldsSchema.safeParse(values)
  if (!parsed.success) {
    for (const issue of parsed.error.issues) {
      const field = issue.path[0]
      if (isBillingField(field)) errors[field] ??= issue.message
    }
    return { ok: false, errors }
  }
  for (const issue of billingIssues(parsed.data, priceMinor)) {
    errors[issue.path] ??= issue.message
  }
  if (Object.keys(errors).length > 0) return { ok: false, errors }
  return { ok: true, billing: toBilling(parsed.data) }
}

function isBillingField(value: unknown): value is BillingFieldName {
  return (
    typeof value === 'string' &&
    (billingFieldNames as readonly string[]).includes(value)
  )
}
