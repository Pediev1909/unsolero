import { useMutation } from '@tanstack/react-query'
import { useId } from 'react'
import { useForm } from 'react-hook-form'
import { Link } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { ApiError } from '../../../lib/api/client'
import { cn } from '../../../lib/styles/cn'
import { subscribeToNewsletter } from '../api'
import {
  newsletterSubscriptionSchema,
  type NewsletterSubscription,
} from '../schemas'

interface NewsletterFormProps {
  /** Which surface asked: "footer", "article:<slug>". Stored with the row. */
  source: string
  /** Footer density: copy beside the field on wider screens instead of above. */
  compact?: boolean
}

// The consent sentence below is what the server records as consent text
// version 2026-09-02 (backend newsletter domain.ConsentTextVersion). Change the
// wording and bump that constant together.
export function NewsletterForm({
  source,
  compact = false,
}: NewsletterFormProps) {
  const headingId = useId()
  const form = useForm<NewsletterSubscription>({
    defaultValues: { email: '' },
  })
  const subscription = useMutation({
    mutationFn: (email: string) => subscribeToNewsletter(email, source),
  })
  const errors = form.formState.errors

  const submit = form.handleSubmit(async (values) => {
    form.clearErrors()
    const validation = newsletterSubscriptionSchema.safeParse(values)
    if (!validation.success) {
      form.setError('email', {
        message:
          validation.error.issues[0]?.message ?? 'Enter a valid email address.',
      })
      return
    }
    try {
      await subscription.mutateAsync(validation.data.email)
    } catch (error) {
      if (error instanceof ApiError && error.fields.email) {
        form.setError('email', { message: error.fields.email })
        return
      }
      form.setError('root.server', {
        message:
          error instanceof ApiError
            ? error.message
            : 'The subscription could not be recorded. Please try again.',
      })
    }
  })

  if (subscription.isSuccess) {
    return (
      <div role="status">
        <p className="text-lg font-medium leading-7">
          Check your inbox to confirm.
        </p>
        <p className="mt-2 text-sm leading-6 text-ink/70">
          We sent a one-time link to <strong>{subscription.variables}</strong>.
          It works for 48 hours, and nothing is sent until you use it.
        </p>
      </div>
    )
  }

  return (
    <form
      aria-labelledby={headingId}
      className={cn(
        compact &&
          'md:grid md:grid-cols-[minmax(0,20rem)_minmax(0,1fr)] md:items-start md:gap-10',
      )}
      noValidate
      onSubmit={(event) => void submit(event)}
    >
      <p
        className={cn(
          'font-medium leading-7',
          compact ? 'text-base' : 'text-lg',
        )}
        id={headingId}
      >
        One email when a price you care about changes. No list-building tricks,
        unsubscribe in one click.
      </p>
      <div className={compact ? 'mt-5 md:mt-0' : 'mt-6'}>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
          <Input
            {...form.register('email')}
            autoComplete="email"
            containerClassName="flex-1"
            error={errors.email?.message}
            inputMode="email"
            label="Email"
            required
            type="email"
          />
          {/* The label above the field is 18px of text plus its 8px margin;
              the button starts at that offset so it lines up with the field
              rather than with the label, and stays put when an error line
              appears under the field. */}
          <div className="sm:pt-[1.625rem]">
            <Button
              className="w-full sm:w-auto"
              loading={subscription.isPending}
              loadingLabel="Sending…"
              type="submit"
            >
              Send me the changes
            </Button>
          </div>
        </div>
        {errors.root?.server && (
          <p
            className="mt-3 border-l-2 border-bronze bg-paper px-4 py-3 text-sm"
            role="alert"
          >
            {errors.root.server.message}
          </p>
        )}
        <p className="mt-4 text-xs leading-5 text-ink/70">
          By subscribing you agree that UNSOLERO stores this address to send you
          price-change emails until you unsubscribe. Read the{' '}
          <Link
            className="underline decoration-ink/30 underline-offset-4 hover:text-ink"
            to="/privacy"
          >
            privacy notice
          </Link>
          .
        </p>
      </div>
    </form>
  )
}
