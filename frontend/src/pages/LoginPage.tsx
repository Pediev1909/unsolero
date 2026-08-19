import { AuthForm } from '../features/auth/AuthForm'
import { AuthLayout } from '../components/layout/AuthLayout'

export function LoginPage() {
  return (
    <AuthLayout
      description="Return to your saved decisions and build your stack with a clear record of every trade-off."
      documentDescription="Sign in to UNSOLERO to return to your saved software decisions and comparisons."
      documentTitle="Sign in | UNSOLERO"
      eyebrow="Your account"
      title="Welcome back."
    >
      <h2 className="text-2xl font-medium tracking-[-0.035em]">Sign in</h2>
      <AuthForm mode="login" />
    </AuthLayout>
  )
}
