import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Checkbox } from '../../components/ui/Checkbox'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import type { AffiliateInput } from '../../features/admin/api'
import {
  useAdminAffiliateLinks,
  useAffiliateMutation,
} from '../../features/admin/queries'
import type { AdminAffiliateLink } from '../../features/admin/schemas'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

const schema = z.object({
  provider: z.string().regex(/^[a-z0-9][a-z0-9._-]{0,99}$/),
  destination_url: z.string().url().startsWith('https://'),
  external_reference: z.string(),
  disclosure_label: z.string().trim().min(1),
  is_active: z.boolean(),
  priority: z.coerce.number().int().min(-1000).max(1000),
  program_id: z.string(),
})
type FormData = z.input<typeof schema>

export function AdminAffiliateLinksPage() {
  usePageMetadata({
    title: 'Affiliate links | UNSOLERO admin',
    description: 'Protected affiliate destination administration.',
    robots: 'noindex, follow',
  })
  const [page, setPage] = useState(1)
  const query = useAdminAffiliateLinks(page)
  const [editing, setEditing] = useState<AdminAffiliateLink>()
  return (
    <>
      <AdminPageHeader
        description="Destinations and monetization metadata remain downstream of objective recommendation results."
        eyebrow="Commerce"
        title="Affiliate Links"
      />
      <AdminQueryState
        empty={query.data?.items.length === 0}
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Product',
                    'Merchant',
                    'Provider',
                    'Destination',
                    'Priority',
                    'Status',
                    '',
                  ].map((header) => (
                    <th className={adminTableHead} key={header}>
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {query.data.items.map((link) => (
                  <tr key={link.id}>
                    <td className={adminTableCell}>{link.product_name}</td>
                    <td className={adminTableCell}>{link.merchant_name}</td>
                    <td className={adminTableCell}>{link.provider}</td>
                    <td className={`${adminTableCell} max-w-xs truncate`}>
                      {link.destination_url}
                    </td>
                    <td className={adminTableCell}>{link.priority}</td>
                    <td className={adminTableCell}>
                      <Badge>{link.is_active ? 'active' : 'inactive'}</Badge>
                    </td>
                    <td className={adminTableCell}>
                      <Button
                        onClick={() => setEditing(link)}
                        size="sm"
                        variant="secondary"
                      >
                        Edit
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </AdminTable>
            <AdminPagination
              onPageChange={setPage}
              page={query.data.page}
              total={query.data.total}
              totalPages={query.data.total_pages}
            />
          </>
        )}
      </AdminQueryState>
      {editing && (
        <AffiliateEditor link={editing} onClose={() => setEditing(undefined)} />
      )}
    </>
  )
}

function AffiliateEditor({
  link,
  onClose,
}: {
  link: AdminAffiliateLink
  onClose: () => void
}) {
  const mutation = useAffiliateMutation(link.id)
  const form = useForm<FormData>({
    defaultValues: {
      provider: link.provider,
      destination_url: link.destination_url,
      external_reference: link.external_reference ?? '',
      disclosure_label: link.disclosure_label,
      is_active: link.is_active,
      priority: link.priority,
      program_id: link.program_id ?? '',
    },
  })
  const submit = form.handleSubmit(async (raw) => {
    const parsed = schema.safeParse(raw)
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        const key = issue.path[0]
        if (typeof key === 'string')
          form.setError(key as keyof FormData, { message: issue.message })
      }
      return
    }
    const value = parsed.data
    const input: AffiliateInput = {
      provider: value.provider,
      destination_url: value.destination_url,
      external_reference: value.external_reference || null,
      disclosure_label: value.disclosure_label,
      is_active: value.is_active,
      priority: value.priority,
      program_id: value.program_id || null,
      commission_type:
        link.commission_type as AffiliateInput['commission_type'],
      commission_rate_bps: link.commission_rate_bps,
      commission_amount_minor: link.commission_amount_minor,
      commission_currency: link.commission_currency,
    }
    await mutation.mutateAsync(input)
    onClose()
  })
  return (
    <Modal
      description="Raw destinations remain visible only to authorized administrators."
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
      open
      title="Edit affiliate link"
    >
      <form className="space-y-4" onSubmit={(event) => void submit(event)}>
        <Input
          error={form.formState.errors.provider?.message}
          label="Provider"
          {...form.register('provider')}
        />
        <Input
          error={form.formState.errors.destination_url?.message}
          label="Affiliate destination"
          {...form.register('destination_url')}
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <Input label="Program ID" {...form.register('program_id')} />
          <Input
            label="External reference"
            {...form.register('external_reference')}
          />
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <Input
            label="Disclosure label"
            {...form.register('disclosure_label')}
          />
          <Input
            label="Priority"
            type="number"
            {...form.register('priority', { valueAsNumber: true })}
          />
        </div>
        <Checkbox label="Active link" {...form.register('is_active')} />
        <p className="text-xs leading-5 text-ink/68">
          Commission: {link.commission_type}. It is retained for reporting and
          never used in recommendation scoring.
        </p>
        <div className="flex justify-end gap-3 border-t border-ink/10 pt-4">
          <Button onClick={onClose} variant="quiet">
            Cancel
          </Button>
          <Button loading={mutation.isPending} type="submit">
            Save link
          </Button>
        </div>
      </form>
    </Modal>
  )
}
