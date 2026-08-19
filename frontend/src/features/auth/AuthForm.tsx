import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { useCredentialsForm } from './useCredentialsForm'

interface AuthFormProps {
  mode: 'login' | 'register'
}

export function AuthForm({ mode }: AuthFormProps) {
  const { form, submit, isSubmitting } = useCredentialsForm(mode)
  const isLogin = mode === 'login'
  const errors = form.formState.errors

  return (
    <form
      className="mt-10 space-y-6"
      onSubmit={(event) => void submit(event)}
      noValidate
    >
      <Input
        {...form.register('email')}
        autoComplete="email"
        error={errors.email?.message}
        id="email"
        inputMode="email"
        label="Email"
        required
        type="email"
      />

      <Input
        {...form.register('password')}
        autoComplete={isLogin ? 'current-password' : 'new-password'}
        error={errors.password?.message}
        hint={!isLogin ? 'Use at least 12 characters.' : undefined}
        id="password"
        label="Password"
        required
        type="password"
      />

      {errors.root?.server && (
        <div
          className="border-l-2 border-bronze bg-paper px-4 py-3 text-sm"
          role="alert"
        >
          {errors.root.server.message}
        </div>
      )}

      <Button
        fullWidth
        loading={isSubmitting}
        loadingLabel={isLogin ? 'Signing in…' : 'Creating account…'}
        type="submit"
      >
        {isLogin ? 'Sign in' : 'Create account'}
        <ArrowRight aria-hidden="true" size={16} />
      </Button>

      <p className="text-center text-sm text-ink/70">
        {isLogin ? 'New to UNSOLERO?' : 'Already have an account?'}{' '}
        <Link
          className="py-1 font-semibold text-ink underline decoration-ink/25 underline-offset-4"
          to={isLogin ? '/register' : '/login'}
        >
          {isLogin ? 'Create an account' : 'Sign in'}
        </Link>
      </p>
      {isLogin && (
        <p className="text-center text-sm">
          <Link
            className="py-1 font-semibold text-ink underline decoration-ink/25 underline-offset-4"
            to="/forgot-password"
          >
            Forgot your password?
          </Link>
        </p>
      )}
    </form>
  )
}
