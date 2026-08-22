import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { renderWithProviders } from '../../../test/renderWithProviders'
import { BudgetInput } from './RecommendationStep'

function Harness({ start = 12_000 }: { start?: number }) {
  const [minor, setMinor] = useState(start)
  return (
    <>
      <BudgetInput onChange={setMinor} valueMinor={minor} />
      <output data-testid="minor">{minor}</output>
    </>
  )
}

describe('BudgetInput', () => {
  // It used to derive its text straight from the form value, so clearing it
  // made Number('') into 0, the field redrew as "0", and typing 4000 produced
  // "04000". You could not delete the zero, on the question that decides
  // everything else on the page.
  it('can be emptied and retyped', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Harness />)
    const field = screen.getByLabelText(/Exact budget/i)

    await user.clear(field)
    expect(field).toHaveValue(null)

    await user.type(field, '4000')
    expect(field).toHaveValue(4000)
    expect(screen.getByTestId('minor')).toHaveTextContent('400000')
  })

  // Clamping while someone types turns "1" into "100" before they finish
  // writing "1500", so it waits for them to leave the field.
  it('clamps on blur, not on every keystroke', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Harness />)
    const field = screen.getByLabelText(/Exact budget/i)

    await user.clear(field)
    await user.type(field, '1')
    expect(field).toHaveValue(1)

    await user.tab()
    expect(field).toHaveValue(100)
  })

  it('follows the value when it changes elsewhere', () => {
    renderWithProviders(<Harness start={25_000} />)
    expect(screen.getByLabelText(/Exact budget/i)).toHaveValue(250)
  })
})
