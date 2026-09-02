import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { NewsletterForm } from './NewsletterForm'

function renderForm(source = 'footer') {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <NewsletterForm source={source} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NewsletterForm', () => {
  it('validates the address before making a request', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    renderForm()

    await user.type(
      screen.getByRole('textbox', { name: /email/i }),
      'not-an-address',
    )
    await user.click(
      screen.getByRole('button', { name: /send me the changes/i }),
    )

    expect(
      await screen.findByText(/enter a valid email address/i),
    ).toBeInTheDocument()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('posts the address with its source and asks the reader to check their inbox', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse(202, { recorded: true }))
    vi.stubGlobal('fetch', fetchMock)
    const user = userEvent.setup()
    renderForm('article:mailchimp-alternatives')

    await user.type(
      screen.getByRole('textbox', { name: /email/i }),
      'reader@example.com',
    )
    await user.click(
      screen.getByRole('button', { name: /send me the changes/i }),
    )

    const status = await screen.findByRole('status')
    expect(status).toHaveTextContent(/check your inbox to confirm/i)
    expect(status).toHaveTextContent('reader@example.com')
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toBe('/api/newsletter/subscriptions')
    expect(JSON.parse(init.body as string) as unknown).toEqual({
      email: 'reader@example.com',
      source: 'article:mailchimp-alternatives',
    })
  })

  it('shows the server message and keeps the form usable when recording fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(500, {
          error: {
            code: 'newsletter_unavailable',
            message:
              'The subscription could not be recorded. Please try again.',
          },
        }),
      ),
    )
    const user = userEvent.setup()
    renderForm()

    await user.type(
      screen.getByRole('textbox', { name: /email/i }),
      'reader@example.com',
    )
    await user.click(
      screen.getByRole('button', { name: /send me the changes/i }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /could not be recorded/i,
    )
    expect(
      screen.getByRole('button', { name: /send me the changes/i }),
    ).toBeEnabled()
    expect(screen.getByRole('textbox', { name: /email/i })).toHaveValue(
      'reader@example.com',
    )
  })

  it('puts a field error from the server on the field', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(400, {
          error: {
            code: 'validation_failed',
            message: 'Check the highlighted fields.',
            fields: { email: 'Enter a valid email address.' },
          },
        }),
      ),
    )
    const user = userEvent.setup()
    renderForm()

    await user.type(
      screen.getByRole('textbox', { name: /email/i }),
      'reader@example.com',
    )
    await user.click(
      screen.getByRole('button', { name: /send me the changes/i }),
    )

    expect(
      await screen.findByRole('textbox', { name: /email/i }),
    ).toHaveAccessibleDescription(/enter a valid email address/i)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('links the consent sentence to the privacy notice', () => {
    renderForm()
    expect(
      screen.getByRole('link', { name: /privacy notice/i }),
    ).toHaveAttribute('href', '/privacy')
  })
})
