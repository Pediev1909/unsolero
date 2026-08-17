import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { authKeys } from '../auth/queries'
import { ProtectedAdminRoute } from './ProtectedAdminRoute'

function renderRoute(roles: string[]) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(authKeys.currentUser, {
    id: '97bfb760-6d09-4b96-8a39-d2bb16445537',
    email: 'operator@example.com',
    roles,
  })
  const router = createMemoryRouter(
    [
      {
        element: <ProtectedAdminRoute />,
        children: [{ path: '/admin', element: <p>Admin content</p> }],
      },
      { path: '/account', element: <p>Account content</p> },
      { path: '/login', element: <p>Login</p> },
    ],
    { initialEntries: ['/admin'] },
  )
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('ProtectedAdminRoute', () => {
  it('renders protected content for an administrator', async () => {
    vi.stubGlobal('fetch', vi.fn())
    renderRoute(['admin'])
    expect(await screen.findByText('Admin content')).toBeInTheDocument()
  })

  it('redirects an authenticated member without the role', async () => {
    vi.stubGlobal('fetch', vi.fn())
    renderRoute([])
    expect(await screen.findByText('Account content')).toBeInTheDocument()
  })
})
