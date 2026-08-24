import { useState } from 'react'

import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { Textarea } from '../../components/ui/Textarea'
import { useToast } from '../../components/ui/useToast'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  usePolicyTransition,
  useRecommendationPolicies,
} from '../../features/admin/queries'
import type {
  PolicyStatus,
  RecommendationPolicy,
} from '../../features/admin/schemas'
import { ApiError } from '../../lib/api/client'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

type PolicyAction = 'submit' | 'approve' | 'reject' | 'activate' | 'deactivate'

const actionLabels: Record<PolicyAction, string> = {
  submit: 'Submit for review',
  approve: 'Approve',
  reject: 'Reject',
  activate: 'Activate',
  deactivate: 'Retire',
}

// The transition graph enforced by the database, mirrored here so a button is
// never offered for a move the server will refuse.
const availableActions: Record<PolicyStatus, PolicyAction[]> = {
  draft: ['submit'],
  rejected: ['submit'],
  in_review: ['approve', 'reject'],
  approved: ['activate'],
  active: ['deactivate'],
  retired: [],
}

const statusTone: Record<
  PolicyStatus,
  'neutral' | 'accent' | 'success' | 'warning'
> = {
  draft: 'neutral',
  in_review: 'warning',
  approved: 'accent',
  active: 'success',
  retired: 'neutral',
  rejected: 'warning',
}

const statusLabels: Record<PolicyStatus, string> = {
  draft: 'Draft',
  in_review: 'In review',
  approved: 'Approved',
  active: 'Active',
  retired: 'Retired',
  rejected: 'Rejected',
}

// Activating a policy changes what every visitor is recommended, so the
// consequence is spelled out before the click rather than after it.
const actionConsequences: Record<PolicyAction, string> = {
  submit: 'The version is locked for review and can no longer be edited.',
  approve: 'The version becomes eligible for activation.',
  reject: 'The version returns to its author. A reason is required.',
  activate:
    'This immediately becomes the policy every visitor is recommended against, and the previous active version retires.',
  deactivate:
    'Recommendations stop using this version. Visitors get no results until another version is active.',
}

export function AdminPolicyPage() {
  usePageMetadata({
    title: 'Ranking policy | UNSOLERO admin',
    description: 'Protected recommendation policy lifecycle.',
    robots: 'noindex, follow',
  })
  const query = useRecommendationPolicies()
  const [pending, setPending] = useState<{
    policy: RecommendationPolicy
    action: PolicyAction
  }>()

  return (
    <>
      <AdminPageHeader
        description="Every recommendation is produced by exactly one active policy version per vertical. A version moves draft to review to approved to active, and an active version is immutable — corrections ship as a new version, never as an edit."
        eyebrow="Recommendation"
        title="Ranking policy"
      />
      <AdminQueryState
        empty={query.data?.length === 0}
        emptyDescription="No recommendation policy versions exist for this deployment."
        emptyTitle="No policy versions"
        error={query.error}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <AdminTable>
            <thead>
              <tr>
                {[
                  'Version',
                  'Vertical',
                  'Status',
                  'Categories',
                  'Products',
                  'Activated',
                  'Actions',
                ].map((heading) => (
                  <th className={adminTableHead} key={heading} scope="col">
                    {heading}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {query.data.map((policy) => (
                <tr className="bg-surface hover:bg-paper" key={policy.version}>
                  <td className={adminTableCell}>
                    <p className="font-semibold">{policy.version}</p>
                    {policy.review_note && (
                      <p className="mt-1 max-w-xs text-xs text-ink/65">
                        {policy.review_note}
                      </p>
                    )}
                  </td>
                  <td className={adminTableCell}>{policy.vertical_key}</td>
                  <td className={adminTableCell}>
                    <Badge variant={statusTone[policy.status]}>
                      {statusLabels[policy.status]}
                    </Badge>
                  </td>
                  <td className={adminTableCell}>{policy.category_count}</td>
                  <td className={adminTableCell}>{policy.product_count}</td>
                  <td className={adminTableCell}>
                    {policy.activated_at
                      ? new Date(policy.activated_at).toLocaleDateString()
                      : 'Never'}
                  </td>
                  <td className={adminTableCell}>
                    <div className="flex flex-wrap gap-2">
                      {availableActions[policy.status].map((action) => (
                        <Button
                          key={action}
                          onClick={() => setPending({ policy, action })}
                          size="sm"
                          variant={
                            action === 'reject' || action === 'deactivate'
                              ? 'danger'
                              : 'secondary'
                          }
                        >
                          {actionLabels[action]}
                        </Button>
                      ))}
                      {availableActions[policy.status].length === 0 && (
                        <span className="text-xs text-ink/60">
                          No transitions available
                        </span>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </AdminTable>
        )}
      </AdminQueryState>
      {pending && (
        <PolicyTransitionDialog
          action={pending.action}
          onClose={() => setPending(undefined)}
          policy={pending.policy}
        />
      )}
    </>
  )
}

function PolicyTransitionDialog({
  policy,
  action,
  onClose,
}: {
  policy: RecommendationPolicy
  action: PolicyAction
  onClose: () => void
}) {
  const [note, setNote] = useState('')
  const { showToast } = useToast()
  const transition = usePolicyTransition()
  const noteRequired = action === 'reject'
  const canSubmit = !noteRequired || note.trim().length > 0

  const run = () => {
    transition.mutate(
      { version: policy.version, action, note: note.trim() },
      {
        onSuccess: () => {
          showToast({
            title: `${actionLabels[action]} succeeded`,
            description: `${policy.version} is updated.`,
            variant: 'success',
          })
          onClose()
        },
        onError: (error) => {
          showToast({
            title: `${actionLabels[action]} failed`,
            description: transitionErrorMessage(error),
            variant: 'error',
          })
        },
      },
    )
  }

  return (
    <Modal
      description="Policy transitions are recorded in the admin audit log."
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
      open
      title={`${actionLabels[action]} \u2014 ${policy.version}`}
    >
      <div className="grid gap-5">
        <p className="text-sm leading-6 text-ink/75">
          {actionConsequences[action]}
        </p>
        <Textarea
          hint={
            noteRequired
              ? 'Required. The author sees this reason.'
              : 'Optional. Recorded with the transition.'
          }
          label="Review note"
          maxLength={2000}
          name="note"
          onChange={(event) => setNote(event.target.value)}
          rows={4}
          value={note}
        />
        <div className="flex flex-wrap justify-end gap-3">
          <Button onClick={onClose} type="button" variant="quiet">
            Cancel
          </Button>
          <Button
            disabled={!canSubmit || transition.isPending}
            onClick={run}
            type="button"
            variant={
              action === 'reject' || action === 'deactivate'
                ? 'danger'
                : 'primary'
            }
          >
            {transition.isPending ? 'Working…' : actionLabels[action]}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

// A single operator authoring, approving and activating the same version is
// refused by the database, and the generic conflict message does not say so.
function transitionErrorMessage(error: unknown) {
  if (!(error instanceof ApiError)) {
    return 'The transition could not be completed.'
  }
  if (error.status === 409) {
    return 'Refused: the same account cannot review or activate a version it authored or submitted. A second account with the reviewer role has to do this step.'
  }
  if (error.code === 'invalid_policy_transition') {
    return 'That transition is not allowed from the current state.'
  }
  return error.message
}
