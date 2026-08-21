import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { renderWithProviders } from '../../test/renderWithProviders'
import { HeaderSearch } from './HeaderSearch'

describe('HeaderSearch', () => {
  it('labels its input', () => {
    renderWithProviders(<HeaderSearch />)
    expect(screen.getByLabelText('Search software')).toBeInTheDocument()
  })

  // The drawer keeps its own copy of this form mounted whether it is open or
  // not, so a fixed id put two elements with the same one on every page and
  // bound one of the labels to the wrong field.
  it('gives each instance its own input id', () => {
    renderWithProviders(
      <>
        <HeaderSearch />
        <HeaderSearch variant="block" />
      </>,
    )
    const inputs = screen.getAllByLabelText('Search software')
    expect(inputs).toHaveLength(2)
    expect(inputs[0]?.id).not.toBe(inputs[1]?.id)
    expect(inputs[0]?.id).toBeTruthy()
  })

  it('offers a submit control rather than relying on Enter alone', () => {
    renderWithProviders(<HeaderSearch />)
    expect(screen.getByRole('button', { name: 'Search' })).toHaveAttribute(
      'type',
      'submit',
    )
  })

  it('runs the callback when the search is submitted', async () => {
    const user = userEvent.setup()
    let submitted = 0
    renderWithProviders(<HeaderSearch onSubmitted={() => (submitted += 1)} />)

    await user.type(screen.getByLabelText('Search software'), 'stripe')
    await user.click(screen.getByRole('button', { name: 'Search' }))
    expect(submitted).toBe(1)
  })
})
