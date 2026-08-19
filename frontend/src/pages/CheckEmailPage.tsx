import { MailCheck } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'

import { AuthLayout } from '../components/layout/AuthLayout'

export function CheckEmailPage() {
  const location = useLocation()
  const email = (location.state as { email?: string } | null)?.email
  return (
    <AuthLayout
      description="Registration responses are deliberately generic, so an address cannot be used to discover whether an account exists."
      eyebrow="Account security"
      title="Check your email flow."
    >
      <MailCheck aria-hidden="true" className="text-bronze" size={28} />
      <h2 className="mt-5 text-2xl font-medium tracking-[-0.035em]">
        Verification requested
      </h2>
      <p className="mt-4 text-sm leading-6 text-ink/70" role="status">
        If {email ? <strong>{email}</strong> : 'that address'} is eligible, a
        verification delivery intent has been recorded. In local development,
        inspect the documented development email sink.
      </p>
      <Link
        className="mt-7 inline-block font-semibold underline underline-offset-4"
        to="/login"
      >
        Return to sign in
      </Link>
    </AuthLayout>
  )
}
