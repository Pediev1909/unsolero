import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ProtectedRoute } from './ProtectedRoute'

function renderProtectedRoute(initialPath = '/account') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const router = createMemoryRouter(
    [
      {
        element: <ProtectedRoute />,
        children: [{ path: '/account', element: <div>Private account</div> }],
      },
      { path: '/login', element: <div>Login destination</div> },
    ],
    { initialEntries: [initialPath] },
  )
  return render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ProtectedRoute', () => {
  it('shows an account loading state while the session is checked', () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => new Promise(() => undefined)),
    )
    renderProtectedRoute()

    expect(screen.getByText(/checking your account/i)).toBeInTheDocument()
  })

  it('redirects an unauthenticated visitor to login', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: 'authentication_required',
              message: 'Sign in to continue.',
            },
          }),
          { status: 401, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    renderProtectedRoute()

    expect(await screen.findByText('Login destination')).toBeInTheDocument()
  })

  it('renders the protected child for an authenticated user', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            user: { id: 'user-1', email: 'person@example.com', roles: [] },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    renderProtectedRoute()

    expect(await screen.findByText('Private account')).toBeInTheDocument()
  })

  it('shows a recoverable error when authentication is unavailable', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: 'authentication_unavailable',
              message: 'Authentication is temporarily unavailable.',
            },
          }),
          { status: 503, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )
    renderProtectedRoute()

    expect(
      await screen.findByRole('heading', {
        name: /could not verify your session/i,
      }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /try again/i }),
    ).toBeInTheDocument()
  })
})
