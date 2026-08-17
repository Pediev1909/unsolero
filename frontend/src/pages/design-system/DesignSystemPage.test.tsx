import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { ToastProvider } from '../../components/ui/ToastProvider'
import { DesignSystemPage } from './DesignSystemPage'

describe('DesignSystemPage', () => {
  it('renders the component inventory with honest demo labeling', () => {
    render(
      <MemoryRouter>
        <ToastProvider>
          <DesignSystemPage />
        </ToastProvider>
      </MemoryRouter>,
    )

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: /precision without noise/i,
      }),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/showcase records are fictional seed data/i),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/illustrative value only—not a product review/i),
    ).toBeInTheDocument()
  })
})
