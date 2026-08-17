import { z } from 'zod'

const errorSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
    fields: z.record(z.string(), z.string()).optional(),
    request_id: z.string().optional(),
  }),
})

const apiBaseUrl = '/api'

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields: Record<string, string>

  constructor(
    status: number,
    code: string,
    message: string,
    fields: Record<string, string> = {},
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
  }
}

type RequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown
  timeoutMs?: number
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions,
  parse: (value: unknown) => T,
): Promise<T> {
  const { body, signal, timeoutMs = 15_000, ...requestOptions } = options
  const hasBody = body !== undefined
  const controller = new AbortController()
  let timedOut = false
  const timeout = window.setTimeout(() => {
    timedOut = true
    controller.abort()
  }, timeoutMs)
  const abortFromCaller = () => controller.abort(signal?.reason)
  if (signal?.aborted) abortFromCaller()
  else signal?.addEventListener('abort', abortFromCaller, { once: true })

  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}${path}`, {
      ...requestOptions,
      body: hasBody ? JSON.stringify(body) : undefined,
      credentials: 'include',
      headers: {
        ...(hasBody ? { 'Content-Type': 'application/json' } : {}),
        ...requestOptions.headers,
      },
      signal: controller.signal,
    })
  } catch (error) {
    if (timedOut) {
      throw new ApiError(
        0,
        'request_timeout',
        'The request took too long. Please try again.',
      )
    }
    if (error instanceof ApiError) throw error
    throw new ApiError(
      0,
      'network_unavailable',
      'We could not reach the service. Check your connection and try again.',
    )
  } finally {
    window.clearTimeout(timeout)
    signal?.removeEventListener('abort', abortFromCaller)
  }

  if (!response.ok) {
    throw await responseError(response)
  }
  if (response.status === 204) {
    return parse(undefined)
  }
  try {
    return parse(await response.json())
  } catch {
    throw new ApiError(
      response.status,
      'unexpected_response',
      'The service returned an unexpected response. Please try again.',
    )
  }
}

async function responseError(response: Response): Promise<ApiError> {
  try {
    const parsed = errorSchema.safeParse(await response.json())
    if (parsed.success) {
      return new ApiError(
        response.status,
        parsed.data.error.code,
        parsed.data.error.message,
        parsed.data.error.fields,
      )
    }
  } catch {
    // A safe generic error is returned when an upstream response is not JSON.
  }
  return new ApiError(
    response.status,
    'unexpected_response',
    'Something went wrong. Please try again.',
  )
}
