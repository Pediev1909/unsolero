import { StrictMode } from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'

import { VerifyEmailPage } from './VerifyEmailPage'

afterEach(() => {
  window.history.replaceState(null, '', '/')
  vi.unstubAllGlobals()
})

it('consumes a one-time verification token once under strict effects', async () => {
  window.history.replaceState(null, '', '/verify-email#one-time-token')
  const fetchMock = vi
    .fn()
    .mockResolvedValue(new Response(null, { status: 204 }))
  vi.stubGlobal('fetch', fetchMock)

  render(
    <StrictMode>
      <MemoryRouter>
        <VerifyEmailPage />
      </MemoryRouter>
    </StrictMode>,
  )

  expect(
    await screen.findByRole('heading', { name: 'Email verified' }),
  ).toBeInTheDocument()
  expect(fetchMock).toHaveBeenCalledTimes(1)
})
