import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { renderWithProviders } from '../../test/renderWithProviders'
import { NavigationMenu } from './NavigationMenu'

function Menu() {
  return (
    <>
      <NavigationMenu label="Browse">
        {(close) => (
          <button onClick={close} type="button">
            Inside
          </button>
        )}
      </NavigationMenu>
      <button type="button">Outside</button>
    </>
  )
}

describe('NavigationMenu', () => {
  it('starts closed and opens on click', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Menu />)

    const trigger = screen.getByRole('button', { name: /Browse/ })
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('button', { name: 'Inside' })).toBeNull()

    await user.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('button', { name: 'Inside' })).toBeInTheDocument()
  })

  // A keyboard user who dismisses a menu and lands at the top of the document
  // has been punished for using the keyboard.
  it('closes on Escape and puts focus back on the trigger', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Menu />)

    const trigger = screen.getByRole('button', { name: /Browse/ })
    await user.click(trigger)
    await user.keyboard('{Escape}')

    await waitFor(() =>
      expect(trigger).toHaveAttribute('aria-expanded', 'false'),
    )
    expect(trigger).toHaveFocus()
  })

  it('closes when a click lands outside it', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Menu />)

    await user.click(screen.getByRole('button', { name: /Browse/ }))
    await user.click(screen.getByRole('button', { name: 'Outside' }))

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Browse/ })).toHaveAttribute(
        'aria-expanded',
        'false',
      ),
    )
  })

  it('opens on the keyboard alone', async () => {
    const user = userEvent.setup()
    renderWithProviders(<Menu />)

    await user.tab()
    const trigger = screen.getByRole('button', { name: /Browse/ })
    expect(trigger).toHaveFocus()
    await user.keyboard('{Enter}')
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
  })

  it('marks itself as the current section when it is', () => {
    renderWithProviders(
      <NavigationMenu active label="Browse">
        {() => <span>Panel</span>}
      </NavigationMenu>,
    )
    expect(screen.getByRole('button', { name: /Browse/ })).toHaveAttribute(
      'aria-current',
      'true',
    )
  })
})
