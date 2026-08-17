import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { AuthForm } from './AuthForm'

function renderForm(mode: 'login' | 'register') {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <AuthForm mode={mode} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('AuthForm', () => {
  it('validates credentials before making a request', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    renderForm('register')

    await user.click(screen.getByRole('button', { name: /create account/i }))

    expect(
      await screen.findByText(/enter a valid email address/i),
    ).toBeInTheDocument()
    expect(screen.getByText(/use at least 12 characters/i)).toBeInTheDocument()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('shows safe server authentication errors', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: 'invalid_credentials',
              message: 'The email or password is incorrect.',
            },
          }),
          { status: 401, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    const user = userEvent.setup()
    renderForm('login')

    await user.type(screen.getByLabelText(/email/i), 'person@example.com')
    await user.type(
      screen.getByLabelText(/password/i),
      'a long secure password',
    )
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /email or password is incorrect/i,
    )
  })
})
