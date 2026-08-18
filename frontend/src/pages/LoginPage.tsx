import { AuthForm } from '../features/auth/AuthForm'
import { AuthLayout } from '../components/layout/AuthLayout'

export function LoginPage() {
  return (
    <AuthLayout
      description="Return to your saved decisions and build your stack with a clear record of every trade-off."
      eyebrow="Your account"
      title="Welcome back."
    >
      <h2 className="text-2xl font-medium tracking-[-0.035em]">Sign in</h2>
      <AuthForm mode="login" />
    </AuthLayout>
  )
}
