import { useState } from 'react'

import { Button } from '../../../components/ui/Button'
import { Select } from '../../../components/ui/Select'
import { Textarea } from '../../../components/ui/Textarea'
import { useToast } from '../../../components/ui/useToast'
import { ApiError } from '../../../lib/api/client'
import {
  useAdminProduct,
  useCreateEvidenceRevision,
  useEvidenceObservations,
  useEvidenceRevisionTransition,
} from '../queries'
import type { AdminProduct } from '../schemas'
import {
  factClassifications,
  requiredFactKeys,
  scoreRationaleKeys,
  type FactClassification,
} from '../schemas'

// A revision is the step that turns observations into a publishable product.
// The server requires provenance for every recommendation-critical fact and a
// written rationale for every score, and rejects the whole revision if one is
// missing — so the form mirrors that set rather than letting an operator
// discover it as a 422.

type RevisionAction = 'submit' | 'approve' | 'reject' | 'publish'

const actionLabels: Record<RevisionAction, string> = {
  submit: 'Submit for review',
  approve: 'Approve',
  reject: 'Reject',
  publish: 'Publish',
}

// Mirrors the server's transition graph so no button is offered for a move that
// will be refused.
const revisionActions: Record<string, RevisionAction[]> = {
  draft: ['submit'],
  in_review: ['approve', 'reject'],
  approved: ['publish'],
  rejected: ['submit'],
  published: [],
  superseded: [],
}

const scoreLabels: Record<(typeof scoreRationaleKeys)[number], string> = {
  quality: 'Quality',
  value: 'Value',
  durability: 'Durability',
  beginner: 'Beginner fit',
  advanced: 'Advanced fit',
  // Retained columns from the home-gym vertical. Their policy weights are zero
  // for software, but the evidence gate still demands a rationale, so the label
  // says what to write rather than leaving the operator guessing.
  apartment: 'Apartment fit (unused for software — state that)',
  noise: 'Noise (unused for software — state that)',
  portability: 'Portability',
}

export function RevisionActions({
  revisionID,
  status,
}: {
  revisionID: string
  status: string
}) {
  const transition = useEvidenceRevisionTransition()
  const { showToast } = useToast()
  const actions = revisionActions[status] ?? []

  if (actions.length === 0) {
    return <span className="text-xs text-ink/60">No transitions</span>
  }

  const run = (action: RevisionAction) => {
    // Reject is the only transition the server requires a note for.
    const note =
      action === 'reject'
        ? (window.prompt('Reason for rejection (required):') ?? '')
        : ''
    if (action === 'reject' && note.trim() === '') return

    transition.mutate(
      { revisionID, action, note },
      {
        onSuccess: () =>
          showToast({
            title: `${actionLabels[action]} succeeded`,
            variant: 'success',
          }),
        onError: (error) =>
          showToast({
            title: `${actionLabels[action]} failed`,
            description: transitionMessage(error),
            variant: 'error',
          }),
      },
    )
  }

  return (
    <div className="flex flex-wrap gap-2">
      {actions.map((action) => (
        <Button
          disabled={transition.isPending}
          key={action}
          onClick={() => run(action)}
          size="sm"
          variant={action === 'reject' ? 'danger' : 'secondary'}
        >
          {actionLabels[action]}
        </Button>
      ))}
    </div>
  )
}

function transitionMessage(error: unknown) {
  if (error instanceof ApiError && error.status === 409) {
    return 'Refused: the same account cannot review or publish a revision it authored. A second account with the reviewer role has to do this step.'
  }
  return 'The transition was refused from the current state.'
}

export function NewRevisionForm({ productID }: { productID: string }) {
  const product = useAdminProduct(productID)
  const observations = useEvidenceObservations(productID)
  const create = useCreateEvidenceRevision(productID)
  const { showToast } = useToast()
  const [open, setOpen] = useState(false)
  const [observationID, setObservationID] = useState('')
  const [classification, setClassification] =
    useState<FactClassification>('manufacturer')
  const [rationales, setRationales] = useState<Record<string, string>>({})

  const available = observations.data ?? []

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!product.data || !observationID) return

    create.mutate(
      {
        product: productInput(product.data),
        // One observation backs every required fact. Splitting provenance per
        // fact is possible in the API, and belongs in a later iteration; what
        // matters first is that a revision can be created at all.
        fact_links: requiredFactKeys.map((factKey) => ({
          fact_key: factKey,
          observation_id: observationID,
          classification,
        })),
        score_rationales: scoreRationaleKeys.map((scoreKey) => ({
          score_key: scoreKey,
          rationale: rationales[scoreKey]?.trim() ?? '',
          observation_id: observationID,
        })),
      },
      {
        onSuccess: () => {
          showToast({ title: 'Revision created as draft', variant: 'success' })
          setOpen(false)
          setRationales({})
        },
        onError: () =>
          showToast({
            title: 'The revision was rejected',
            description:
              'Every fact needs provenance and every score needs a rationale.',
            variant: 'error',
          }),
      },
    )
  }

  const missingRationale = scoreRationaleKeys.some(
    (key) => (rationales[key]?.trim() ?? '') === '',
  )

  return (
    <section aria-labelledby="revision-form-heading" className="mt-12">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-editorial text-2xl" id="revision-form-heading">
          New revision
        </h2>
        <Button onClick={() => setOpen((value) => !value)} variant="secondary">
          {open ? 'Cancel' : 'Start a revision'}
        </Button>
      </div>
      <p className="mt-2 max-w-2xl text-body-sm text-ink/70">
        Takes the product&rsquo;s current values and attaches evidence to them.
        The result is a draft; publishing it is a separate, reviewed step.
      </p>

      {open && available.length === 0 && (
        <p className="mt-5 border border-ink/10 bg-paper px-4 py-6 text-sm text-ink/70">
          Record an observation first. A revision must cite one.
        </p>
      )}

      {open && available.length > 0 && (
        <form
          className="mt-5 grid gap-5 border border-ink/10 bg-paper p-5"
          onSubmit={submit}
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <Select
              hint="Backs every fact in this revision."
              label="Observation"
              name="observation_id"
              onChange={(event) => setObservationID(event.target.value)}
              required
              value={observationID}
            >
              <option value="">Select an observation</option>
              {available.map((observation) => (
                <option key={observation.id} value={observation.id}>
                  {new Date(observation.observed_at).toLocaleDateString()} —{' '}
                  {observation.notes.slice(0, 60) || 'no notes'}
                </option>
              ))}
            </Select>
            <Select
              hint="How the fact was established."
              label="Classification"
              name="classification"
              onChange={(event) =>
                setClassification(event.target.value as FactClassification)
              }
              value={classification}
            >
              {factClassifications.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </Select>
          </div>

          <div>
            <h3 className="font-semibold">Score rationales</h3>
            <p className="mt-1 text-sm text-ink/70">
              All eight are required. One sentence each, saying what in the
              observation supports the score.
            </p>
            <div className="mt-4 grid gap-4 lg:grid-cols-2">
              {scoreRationaleKeys.map((key) => (
                <Textarea
                  key={key}
                  label={scoreLabels[key]}
                  name={`rationale_${key}`}
                  onChange={(event) =>
                    setRationales({ ...rationales, [key]: event.target.value })
                  }
                  rows={2}
                  value={rationales[key] ?? ''}
                />
              ))}
            </div>
          </div>

          <div className="flex flex-wrap items-center justify-end gap-3">
            {missingRationale && (
              <p className="text-sm text-ink/70">
                Every score needs a rationale before this can be saved.
              </p>
            )}
            <Button
              disabled={create.isPending || !observationID || missingRationale}
              type="submit"
            >
              {create.isPending ? 'Saving…' : 'Create draft revision'}
            </Button>
          </div>
        </form>
      )}
    </section>
  )
}

// The revision carries the product's own values. The admin product response is
// flat — category_id, price_minor — so this only fills in the physical fields a
// software product does not have.
function productInput(product: AdminProduct) {
  return {
    category_id: product.category_id,
    brand_id: product.brand_id,
    name: product.name,
    slug: product.slug,
    description: product.description,
    price_minor: product.price_minor,
    currency: product.currency,
    length_mm: product.length_mm,
    width_mm: product.width_mm,
    height_mm: product.height_mm,
    weight_grams: product.weight_grams,
    max_capacity_grams: product.max_capacity_grams,
    material: product.material,
    warranty_months: product.warranty_months,
    quality_score: product.scores.quality,
    value_score: product.scores.value,
    durability_score: product.scores.durability,
    beginner_score: product.scores.beginner,
    advanced_score: product.scores.advanced,
    apartment_score: product.scores.apartment,
    noise_score: product.scores.noise,
    portability_score: product.scores.portability,
  }
}
