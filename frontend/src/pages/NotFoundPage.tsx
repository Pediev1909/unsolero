import { ArrowLeft } from 'lucide-react'

import { ButtonLink } from '../components/ui/ButtonLink'
import { Heading } from '../components/ui/Heading'
export function NotFoundPage() {
  return (
    <main className="grid min-h-screen place-items-center bg-canvas px-6">
      <div className="max-w-xl text-center">
        <p className="eyebrow">404 / Page not found</p>
        <Heading className="mt-5" level={1} size="display">
          This path needs a rethink.
        </Heading>
        <p className="mx-auto mt-6 max-w-md leading-7 text-ink/65">
          The page you requested does not exist or has moved.
        </p>
        <ButtonLink className="mt-8" to="/">
          <ArrowLeft aria-hidden="true" size={16} />
          Back to UNSOLERO
        </ButtonLink>
      </div>
    </main>
  )
}
