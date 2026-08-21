import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ToastProvider } from '../../components/ui/ToastProvider'
import { renderWithProviders } from '../../test/renderWithProviders'
import { DesignSystemPage } from './DesignSystemPage'

describe('DesignSystemPage', () => {
  it('renders the component inventory with honest demo labeling', () => {
    renderWithProviders(
      <ToastProvider>
        <DesignSystemPage />
      </ToastProvider>,
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
