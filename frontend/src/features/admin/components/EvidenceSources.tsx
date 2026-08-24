import { useState } from 'react'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Checkbox } from '../../../components/ui/Checkbox'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { Textarea } from '../../../components/ui/Textarea'
import { useToast } from '../../../components/ui/useToast'
import {
  useCreateEvidenceObservation,
  useCreateEvidenceSource,
  useEvidenceObservations,
  useEvidenceSources,
  useReviewEvidenceSource,
} from '../queries'
import type { EvidenceSource } from '../schemas'

// A product cannot be published without evidence, and evidence could only be
// written in SQL: the API had all eight write endpoints and the browser called
// none of them. These two panels are the first two steps of the chain — record
// where a fact came from, then record that the source was consulted about this
// product on a date.

const sourceTypes = [
  { value: 'manufacturer_documentation', label: 'Manufacturer documentation' },
  { value: 'verified_merchant_data', label: 'Verified merchant data' },
  { value: 'independent_testing', label: 'Independent testing' },
  { value: 'editorial_assessment', label: 'Editorial assessment' },
  { value: 'demo_fixture', label: 'Demo fixture (fictional)' },
] as const

const reviewTone = {
  verified: 'success',
  pending: 'warning',
  rejected: 'error',
} as const

export function EvidenceSourcesPanel() {
  const sources = useEvidenceSources()
  const [adding, setAdding] = useState(false)

  return (
    <section aria-labelledby="sources-heading" className="mt-12">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-editorial text-2xl" id="sources-heading">
          Sources
        </h2>
        <Button onClick={() => setAdding((open) => !open)} variant="secondary">
          {adding ? 'Cancel' : 'Add source'}
        </Button>
      </div>
      <p className="mt-2 max-w-2xl text-body-sm text-ink/70">
        Where a fact came from. A revision may only cite a source that has been
        reviewed and marked verified.
      </p>

      {adding && <SourceForm onDone={() => setAdding(false)} />}

      <div className="mt-5 border border-ink/10">
        {sources.isPending && (
          <p className="px-4 py-6 text-sm text-ink/70">Loading sources…</p>
        )}
        {sources.data?.length === 0 && (
          <p className="px-4 py-6 text-sm text-ink/70">
            No sources recorded yet. Add one before recording an observation.
          </p>
        )}
        <ul className="divide-y divide-ink/8">
          {sources.data?.map((source) => (
            <SourceRow key={source.id} source={source} />
          ))}
        </ul>
      </div>
    </section>
  )
}

function SourceRow({ source }: { source: EvidenceSource }) {
  const review = useReviewEvidenceSource()
  const { showToast } = useToast()

  const decide = (status: 'verified' | 'rejected') => {
    review.mutate(
      {
        id: source.id,
        status,
        note: `Marked ${status} from the evidence page.`,
      },
      {
        onSuccess: () =>
          showToast({ title: `Source ${status}`, variant: 'success' }),
        onError: () =>
          showToast({
            title: 'Review failed',
            description:
              'The same account may not be allowed to review a source it created.',
            variant: 'error',
          }),
      },
    )
  }

  return (
    <li className="flex flex-wrap items-start justify-between gap-4 px-4 py-4">
      <div className="min-w-0">
        <p className="font-semibold">{source.title}</p>
        <p className="mt-1 text-sm text-ink/70">
          {source.publisher} · {source.source_type.replaceAll('_', ' ')}
          {source.is_fictional && ' · fictional'}
        </p>
        {source.source_url && (
          <a
            className="mt-1 block break-all text-sm text-bronze-dark hover:text-ink"
            href={source.source_url}
            rel="noopener noreferrer"
            target="_blank"
          >
            {source.source_url}
          </a>
        )}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        <Badge variant={reviewTone[source.review_status]}>
          {source.review_status}
        </Badge>
        {source.review_status === 'pending' && (
          <>
            <Button
              disabled={review.isPending}
              onClick={() => decide('verified')}
              size="sm"
              variant="secondary"
            >
              Verify
            </Button>
            <Button
              disabled={review.isPending}
              onClick={() => decide('rejected')}
              size="sm"
              variant="danger"
            >
              Reject
            </Button>
          </>
        )}
      </div>
    </li>
  )
}

function SourceForm({ onDone }: { onDone: () => void }) {
  const create = useCreateEvidenceSource()
  const { showToast } = useToast()
  const [form, setForm] = useState({
    source_type: String(sourceTypes[0].value),
    title: '',
    publisher: '',
    source_url: '',
    is_fictional: false,
  })

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    create.mutate(
      { ...form, source_url: form.source_url.trim() || null },
      {
        onSuccess: () => {
          showToast({ title: 'Source created', variant: 'success' })
          onDone()
        },
        onError: () =>
          showToast({
            title: 'Could not create the source',
            description: 'Check the title, publisher and URL.',
            variant: 'error',
          }),
      },
    )
  }

  return (
    <form
      className="mt-5 grid gap-4 border border-ink/10 bg-paper p-5 sm:grid-cols-2"
      onSubmit={submit}
    >
      <Select
        label="Source type"
        name="source_type"
        onChange={(event) =>
          setForm({ ...form, source_type: event.target.value })
        }
        value={form.source_type}
      >
        {sourceTypes.map((type) => (
          <option key={type.value} value={type.value}>
            {type.label}
          </option>
        ))}
      </Select>
      <Input
        label="Publisher"
        name="publisher"
        onChange={(event) =>
          setForm({ ...form, publisher: event.target.value })
        }
        placeholder="Zoho Corporation"
        required
        value={form.publisher}
      />
      <Input
        containerClassName="sm:col-span-2"
        label="Title"
        name="title"
        onChange={(event) => setForm({ ...form, title: event.target.value })}
        placeholder="Zoho CRM pricing page"
        required
        value={form.title}
      />
      <Input
        containerClassName="sm:col-span-2"
        hint="The page the fact was read from. Leave empty only for an offline source."
        label="Source URL"
        name="source_url"
        onChange={(event) =>
          setForm({ ...form, source_url: event.target.value })
        }
        placeholder="https://www.zoho.com/crm/zohocrm-pricing.html"
        type="url"
        value={form.source_url}
      />
      <Checkbox
        checked={form.is_fictional}
        label="This source is fictional (development fixture)"
        name="is_fictional"
        onChange={(event) =>
          setForm({ ...form, is_fictional: event.target.checked })
        }
      />
      <div className="flex items-end justify-end gap-3">
        <Button onClick={onDone} type="button" variant="quiet">
          Cancel
        </Button>
        <Button disabled={create.isPending} type="submit">
          {create.isPending ? 'Saving…' : 'Create source'}
        </Button>
      </div>
    </form>
  )
}

export function EvidenceObservationsPanel({
  productID,
}: {
  productID: string
}) {
  const observations = useEvidenceObservations(productID)
  const sources = useEvidenceSources()
  const create = useCreateEvidenceObservation()
  const { showToast } = useToast()
  const verified = (sources.data ?? []).filter(
    (source) => source.review_status === 'verified',
  )
  const [form, setForm] = useState({
    source_id: '',
    observed_at: new Date().toISOString().slice(0, 10),
    confidence: 90,
    notes: '',
  })

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    create.mutate(
      {
        source_id: form.source_id,
        product_id: productID,
        // The API takes a timestamp; the field is a date, so it is anchored to
        // the start of that day in UTC rather than to the operator's clock.
        observed_at: new Date(`${form.observed_at}T00:00:00Z`).toISOString(),
        expires_at: null,
        confidence: form.confidence,
        notes: form.notes,
      },
      {
        onSuccess: () => {
          showToast({ title: 'Observation recorded', variant: 'success' })
          setForm({ ...form, notes: '' })
        },
        onError: () =>
          showToast({
            title: 'Could not record the observation',
            variant: 'error',
          }),
      },
    )
  }

  return (
    <section aria-labelledby="observations-heading" className="mt-12">
      <h2 className="font-editorial text-2xl" id="observations-heading">
        Observations
      </h2>
      <p className="mt-2 max-w-2xl text-body-sm text-ink/70">
        A verified source, consulted about this product, on a date. A revision
        cites one of these for every fact and every score.
      </p>

      {verified.length === 0 ? (
        <p className="mt-5 border border-ink/10 bg-paper px-4 py-6 text-sm text-ink/70">
          No verified source exists yet. Create a source above and mark it
          verified before recording an observation.
        </p>
      ) : (
        <form
          className="mt-5 grid gap-4 border border-ink/10 bg-paper p-5 sm:grid-cols-2"
          onSubmit={submit}
        >
          <Select
            label="Source"
            name="source_id"
            onChange={(event) =>
              setForm({ ...form, source_id: event.target.value })
            }
            required
            value={form.source_id}
          >
            <option value="">Select a verified source</option>
            {verified.map((source) => (
              <option key={source.id} value={source.id}>
                {source.title} — {source.publisher}
              </option>
            ))}
          </Select>
          <Input
            label="Observed on"
            name="observed_at"
            onChange={(event) =>
              setForm({ ...form, observed_at: event.target.value })
            }
            required
            type="date"
            value={form.observed_at}
          />
          <Input
            hint="How certain the reading is, 0 to 100."
            label="Confidence"
            max={100}
            min={0}
            name="confidence"
            onChange={(event) =>
              setForm({ ...form, confidence: Number(event.target.value) })
            }
            type="number"
            value={form.confidence}
          />
          <Textarea
            containerClassName="sm:col-span-2"
            label="Notes"
            name="notes"
            onChange={(event) =>
              setForm({ ...form, notes: event.target.value })
            }
            placeholder="Entry paid tier listed at $14 per user per month, billed annually."
            rows={3}
            value={form.notes}
          />
          <div className="flex items-end justify-end sm:col-span-2">
            <Button
              disabled={create.isPending || !form.source_id}
              type="submit"
            >
              {create.isPending ? 'Saving…' : 'Record observation'}
            </Button>
          </div>
        </form>
      )}

      <div className="mt-5 border border-ink/10">
        {observations.data?.length === 0 && (
          <p className="px-4 py-6 text-sm text-ink/70">
            Nothing observed about this product yet.
          </p>
        )}
        <ul className="divide-y divide-ink/8">
          {observations.data?.map((observation) => {
            const source = sources.data?.find(
              (candidate) => candidate.id === observation.source_id,
            )
            return (
              <li className="px-4 py-4" key={observation.id}>
                <p className="font-semibold">
                  {source?.title ?? 'Unknown source'}
                </p>
                <p className="mt-1 text-sm text-ink/70">
                  {new Date(observation.observed_at).toLocaleDateString()} ·
                  confidence {observation.confidence}
                </p>
                {observation.notes && (
                  <p className="mt-2 text-sm">{observation.notes}</p>
                )}
              </li>
            )
          })}
        </ul>
      </div>
    </section>
  )
}
