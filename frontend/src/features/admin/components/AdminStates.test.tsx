import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { ApiError } from '../../../lib/api/client'
import { AdminQueryState } from './AdminStates'

function renderState(error: unknown) {
  return render(
    <AdminQueryState empty={false} error={error} onRetry={vi.fn()} pending={false}>
      <p>data</p>
    </AdminQueryState>,
  )
}

describe('AdminQueryState', () => {
  // Administration is gated on a verified address and a recent second factor.
  // Both refusals used to surface as "Something went wrong", which left an
  // administrator locked out of every page with nothing to act on.
  it('names the step that unlocks access when the email is unverified', () => {
    renderState(
      new ApiError(403, 'email_verification_required', 'A verified email address is required.'),
    )
    expect(screen.getByText(/Verify your email address first/i)).toBeInTheDocument()
    expect(screen.queryByText(/Something went wrong/i)).not.toBeInTheDocument()
  })

  it('names the step that unlocks access when a second factor is missing', () => {
    renderState(
      new ApiError(403, 'mfa_step_up_required', 'Recent multi-factor authentication is required.'),
    )
    expect(screen.getByText(/Two-factor authentication required/i)).toBeInTheDocument()
  })

  it('falls back to the general message for anything else', () => {
    renderState(new ApiError(500, 'internal_error', 'Boom'))
    expect(screen.getByText(/Something went wrong/i)).toBeInTheDocument()
  })

  it('renders the data when there is no error', () => {
    renderState(null)
    expect(screen.getByText('data')).toBeInTheDocument()
  })
})
