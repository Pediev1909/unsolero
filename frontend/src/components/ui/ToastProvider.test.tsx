import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Button } from './Button'
import { ToastProvider } from './ToastProvider'
import { useToast } from './useToast'

function ToastHarness() {
  const { showToast } = useToast()
  return (
    <Button
      onClick={() =>
        showToast({
          title: 'Saved',
          description: 'Your preference was saved.',
          duration: 0,
        })
      }
    >
      Save preference
    </Button>
  )
}

describe('ToastProvider', () => {
  it('announces and dismisses notifications', async () => {
    const user = userEvent.setup()
    render(
      <ToastProvider>
        <ToastHarness />
      </ToastProvider>,
    )

    await user.click(screen.getByRole('button', { name: 'Save preference' }))
    expect(await screen.findByRole('status')).toHaveTextContent('Saved')

    await user.click(
      screen.getByRole('button', { name: 'Dismiss notification' }),
    )
    expect(
      screen.queryByText('Your preference was saved.'),
    ).not.toBeInTheDocument()
  })
})
