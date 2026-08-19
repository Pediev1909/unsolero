import { AuthLayout } from '../components/layout/AuthLayout'
import { AuthForm } from '../features/auth/AuthForm'

export function RegisterPage() {
  return (
    <AuthLayout
      description="Create a private account for future saved setups and recommendations. Your software decisions stay grounded in your goals—not commission."
      documentDescription="Create a free UNSOLERO account to save software setups, comparisons and recommendations."
      documentTitle="Create an account | UNSOLERO"
      eyebrow="Start with trust"
      title="Build with confidence."
    >
      <h2 className="text-2xl font-medium tracking-[-0.035em]">
        Create your account
      </h2>
      <AuthForm mode="register" />
    </AuthLayout>
  )
}
