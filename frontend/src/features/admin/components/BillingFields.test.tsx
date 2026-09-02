import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { fromBilling, type BillingFormValues } from '../billingForm'
import { BillingFields } from './BillingFields'

// The control is controlled; this is the smallest owner it can have.
function Harness({ initial }: { initial: BillingFormValues }) {
  const [value, setValue] = useState(initial)
  return (
    <>
      <BillingFields onChange={setValue} value={value} />
      <output data-testid="value">{JSON.stringify(value)}</output>
    </>
  )
}

function current(): BillingFormValues {
  return JSON.parse(
    screen.getByTestId('value').textContent ?? '{}',
  ) as BillingFormValues
}

describe('BillingFields', () => {
  it('starts a product without a billing object as a monthly flat rate', () => {
    render(<Harness initial={fromBilling(undefined)} />)
    expect(screen.getByLabelText('Billing period')).toHaveValue('monthly')
    expect(screen.getByLabelText('Pricing unit')).toHaveValue('flat')
    expect(
      screen.getByLabelText('Annual price per month (minor units)'),
    ).toBeEnabled()
  })

  // The yearly rate is a second price beside a monthly one; on any other
  // period the box is closed, and whatever was in it goes with it.
  it('disables and clears the yearly rate when the period is not monthly', async () => {
    render(
      <Harness
        initial={{ ...fromBilling(undefined), annual_price_minor: '1500' }}
      />,
    )
    const annual = screen.getByLabelText('Annual price per month (minor units)')
    expect(annual).toHaveValue(1500)

    await userEvent.selectOptions(
      screen.getByLabelText('Billing period'),
      'annual',
    )
    expect(annual).toBeDisabled()
    expect(current().annual_price_minor).toBe('')
    expect(current().billing_period).toBe('annual')

    await userEvent.selectOptions(
      screen.getByLabelText('Billing period'),
      'monthly',
    )
    expect(annual).toBeEnabled()
  })

  it('reports each typed value to its owner', async () => {
    render(<Harness initial={fromBilling(undefined)} />)
    await userEvent.selectOptions(
      screen.getByLabelText('Pricing unit'),
      'per_contacts',
    )
    await userEvent.type(screen.getByLabelText('Unit note'), 'Up to 500')
    await userEvent.type(
      screen.getByLabelText('Annual price per month (minor units)'),
      '1300',
    )
    expect(current()).toEqual({
      billing_period: 'monthly',
      pricing_unit: 'per_contacts',
      unit_note: 'Up to 500',
      annual_price_minor: '1300',
    })
  })

  it('shows a message beside the field it belongs to', () => {
    render(
      <BillingFields
        errors={{ annual_price_minor: 'Whole minor units — 1500 for $15.00.' }}
        onChange={() => undefined}
        value={fromBilling(undefined)}
      />,
    )
    const annual = screen.getByLabelText('Annual price per month (minor units)')
    expect(annual).toHaveAttribute('aria-invalid', 'true')
    expect(annual).toHaveAccessibleDescription(/Whole minor units/)
  })
})
