import { useForm } from 'react-hook-form'
import { useLocation, useNavigate } from 'react-router-dom'

import { ApiError } from '../../lib/api/client'
import { synchronizeAnalyticsConsentAfterAuthentication } from '../analytics/consent'
import { credentialsSchema, type Credentials } from './schemas'
import { useLogin, useRegister } from './queries'

type AuthMode = 'login' | 'register'

interface LocationState {
  from?: string
}

export function useCredentialsForm(mode: AuthMode) {
  const form = useForm<Credentials>({
    defaultValues: { email: '', password: '' },
  })
  const login = useLogin()
  const register = useRegister()
  const mutation = mode === 'login' ? login : register
  const navigate = useNavigate()
  const location = useLocation()

  const submit = form.handleSubmit(async (values) => {
    form.clearErrors()
    const validation = credentialsSchema.safeParse(values)
    if (!validation.success) {
      for (const issue of validation.error.issues) {
        const field = issue.path[0]
        if (field === 'email' || field === 'password') {
          form.setError(field, { message: issue.message })
        }
      }
      return
    }

    try {
      const result = await mutation.mutateAsync(validation.data)
      if (mode === 'login' && !('mfa_required' in result)) {
        void synchronizeAnalyticsConsentAfterAuthentication().catch(() => {
          // Authentication succeeds independently; analytics remains fail-closed.
        })
      }
      if (mode === 'register') {
        await navigate('/check-email', {
          replace: true,
          state: { email: validation.data.email },
        })
        return
      }
      if ('mfa_required' in result) {
        await navigate('/login/mfa', { replace: true })
        return
      }
      const state = location.state as LocationState | null
      const destination = safeInternalPath(state?.from) ?? '/account'
      await navigate(destination, { replace: true })
    } catch (error) {
      if (error instanceof ApiError) {
        for (const [field, message] of Object.entries(error.fields)) {
          if (field === 'email' || field === 'password') {
            form.setError(field, { message })
          }
        }
        form.setError('root.server', { message: error.message })
        return
      }
      form.setError('root.server', {
        message: 'Authentication is temporarily unavailable.',
      })
    }
  })

  return {
    form,
    submit,
    isSubmitting: mutation.isPending,
  }
}

function safeInternalPath(path: string | undefined): string | null {
  if (!path || !path.startsWith('/') || path.startsWith('//')) {
    return null
  }
  return path
}
