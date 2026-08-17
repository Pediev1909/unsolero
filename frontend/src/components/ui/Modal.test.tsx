import { useState } from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Button } from './Button'
import { Modal } from './Modal'

function ModalHarness() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open confirmation</Button>
      <Modal onOpenChange={setOpen} open={open} title="Confirm selection">
        <p>Selection details</p>
      </Modal>
    </>
  )
}

describe('Modal', () => {
  it('opens with an accessible name and closes from its control', async () => {
    const user = userEvent.setup()
    render(<ModalHarness />)

    await user.click(screen.getByRole('button', { name: 'Open confirmation' }))
    const dialog = screen.getByRole('dialog', { name: 'Confirm selection' })
    expect(dialog).toHaveAttribute('open')

    await user.click(screen.getByRole('button', { name: 'Close dialog' }))
    expect(dialog).not.toHaveAttribute('open')
  })
})
