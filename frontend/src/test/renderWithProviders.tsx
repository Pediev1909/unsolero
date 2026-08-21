import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'

/**
 * Renders with the providers the real application supplies.
 *
 * The header now reads the category list so the Browse menu can show it, which
 * means every page that renders a header needs a QueryClient. Tests that
 * supplied only a router started failing all at once — a good signal that each
 * test was rebuilding the app shell by hand and would drift from it again.
 */
export function renderWithProviders(
  ui: ReactElement,
  options: RenderOptions & { route?: string } = {},
) {
  const { route = '/', ...renderOptions } = options
  const queryClient = new QueryClient({
    defaultOptions: {
      // A test asserting an error state should not wait through three retries.
      queries: { retry: false, gcTime: 0 },
    },
  })

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions })
}
