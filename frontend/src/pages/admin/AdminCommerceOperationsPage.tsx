import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'
import { Select } from '../../components/ui/Select'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import { ConversionOperations } from '../../features/admin/components/ConversionOperations'
import {
  AdminPagination,
  AdminQueryState,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  useAdminMerchants,
  useCommerceImportFailures,
  useCommerceImports,
  useCommerceProviderLifecycle,
  useCommerceProviders,
  useCreateCommerceProvider,
  useRetryCommerceImport,
  useTriggerCommerceImport,
} from '../../features/admin/queries'
import type { CommerceImport } from '../../features/admin/schemas'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

const configurationSchema = z.object({
  merchant_id: z.string().uuid('Choose a merchant.'),
  provider_key: z.string().regex(/^[a-z0-9][a-z0-9._-]{0,99}$/),
  adapter_key: z.string().regex(/^[a-z0-9][a-z0-9._-]{0,99}$/),
  external_merchant_id: z.string().trim().min(1).max(200),
  credential_reference: z.string().trim().max(200),
  schedule_interval_minutes: z.number().int().min(5).max(10080),
  freshness_ttl_minutes: z.number().int().min(60).max(43200),
})

type ConfigurationForm = z.infer<typeof configurationSchema>

const defaults: ConfigurationForm = {
  merchant_id: '',
  provider_key: '',
  adapter_key: '',
  external_merchant_id: '',
  credential_reference: '',
  schedule_interval_minutes: 360,
  freshness_ttl_minutes: 4320,
}

export function AdminCommerceOperationsPage() {
  usePageMetadata({
    title: 'Commerce operations | UNSOLERO admin',
    description: 'Protected merchant provider and import operations.',
    robots: 'noindex, follow',
  })
  return (
    <>
      <AdminPageHeader
        description="Provider lifecycle, bounded imports, freshness, failures, and reconciliation. Commercial data remains downstream of recommendations."
        eyebrow="Operations"
        title="Commerce operations"
      />
      <ProviderConfiguration />
      <ImportHistory />
      <ConversionOperations />
    </>
  )
}

function ProviderConfiguration() {
  const providers = useCommerceProviders()
  const merchants = useAdminMerchants()
  const create = useCreateCommerceProvider()
  const lifecycle = useCommerceProviderLifecycle()
  const trigger = useTriggerCommerceImport()
  const form = useForm<ConfigurationForm>({ defaultValues: defaults })
  const [formError, setFormError] = useState<string>()

  const submit = form.handleSubmit(async (input) => {
    const parsed = configurationSchema.safeParse(input)
    if (!parsed.success) {
      setFormError('Review the provider configuration fields.')
      return
    }
    setFormError(undefined)
    try {
      await create.mutateAsync({
        ...parsed.data,
        credential_reference: parsed.data.credential_reference || null,
      })
      form.reset(defaults)
    } catch (error) {
      setFormError(
        error instanceof Error
          ? error.message
          : 'The provider could not be created.',
      )
    }
  })

  return (
    <section aria-labelledby="provider-configurations" className="mb-12">
      <div className="mb-5">
        <h2 className="font-editorial text-2xl" id="provider-configurations">
          Provider configurations
        </h2>
        <p className="mt-1 max-w-3xl text-sm leading-6 text-ink/60">
          Credential references name entries in an external secret manager;
          secrets are never stored here. A provider cannot become active until
          its runtime adapter verifies the referenced configuration.
        </p>
      </div>

      <form
        className="mb-7 border border-ink/10 bg-surface p-5"
        onSubmit={(event) => void submit(event)}
      >
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Select label="Merchant" {...form.register('merchant_id')} required>
            <option value="">Choose merchant</option>
            {merchants.data?.map((merchant) => (
              <option key={merchant.id} value={merchant.id}>
                {merchant.name}
              </option>
            ))}
          </Select>
          <Input
            label="Provider key"
            placeholder="awin"
            {...form.register('provider_key')}
            required
          />
          <Input
            label="Adapter key"
            placeholder="awin"
            {...form.register('adapter_key')}
            required
          />
          <Input
            label="External merchant ID"
            {...form.register('external_merchant_id')}
            required
          />
          <Input
            hint="Secret-manager reference only"
            label="Credential reference"
            placeholder="secret/commerce/awin"
            {...form.register('credential_reference')}
          />
          <Input
            label="Schedule interval (minutes)"
            min={5}
            type="number"
            {...form.register('schedule_interval_minutes', {
              valueAsNumber: true,
            })}
          />
          <Input
            label="Freshness TTL (minutes)"
            min={60}
            type="number"
            {...form.register('freshness_ttl_minutes', { valueAsNumber: true })}
          />
        </div>
        {formError && (
          <p className="mt-4 text-sm text-danger" role="alert">
            {formError}
          </p>
        )}
        <Button className="mt-5" loading={create.isPending} type="submit">
          Add disabled provider
        </Button>
      </form>

      <AdminQueryState
        empty={providers.data?.items.length === 0}
        emptyDescription="No merchant provider is configured. No automated offer imports can run."
        error={providers.isError}
        onRetry={() => void providers.refetch()}
        pending={providers.isPending}
      >
        {providers.data && (
          <AdminTable>
            <thead>
              <tr>
                {[
                  'Provider',
                  'Merchant',
                  'Lifecycle',
                  'Freshness',
                  'Last success',
                  'Failures',
                  'Actions',
                ].map((label) => (
                  <th className={adminTableHead} key={label} scope="col">
                    {label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {providers.data.items.map((provider) => (
                <tr key={provider.id}>
                  <td className={adminTableCell}>
                    <p className="font-medium">{provider.provider_key}</p>
                    <p className="text-xs text-ink/55">
                      Adapter: {provider.adapter_key}
                    </p>
                  </td>
                  <td className={adminTableCell}>{provider.merchant_name}</td>
                  <td className={adminTableCell}>
                    <Badge>{provider.lifecycle_status}</Badge>
                  </td>
                  <td className={adminTableCell}>
                    {formatDuration(provider.freshness_ttl_minutes)}
                  </td>
                  <td className={adminTableCell}>
                    {formatDate(provider.last_import_succeeded_at)}
                  </td>
                  <td className={adminTableCell}>
                    {provider.consecutive_failures || 'None'}
                    {provider.last_error_code && (
                      <p className="text-xs text-danger">
                        {provider.last_error_code}
                      </p>
                    )}
                  </td>
                  <td className={adminTableCell}>
                    <div className="flex min-w-44 flex-wrap gap-2">
                      {provider.lifecycle_status === 'active' ||
                      provider.lifecycle_status === 'degraded' ? (
                        <>
                          <Button
                            loading={trigger.isPending}
                            onClick={() =>
                              void trigger.mutateAsync(provider.id)
                            }
                            size="sm"
                          >
                            Import now
                          </Button>
                          <Button
                            onClick={() =>
                              lifecycle.mutate({
                                id: provider.id,
                                status: 'disabled',
                              })
                            }
                            size="sm"
                            variant="secondary"
                          >
                            Disable
                          </Button>
                        </>
                      ) : (
                        <Button
                          onClick={() =>
                            lifecycle.mutate({
                              id: provider.id,
                              status: 'active',
                            })
                          }
                          size="sm"
                          variant="secondary"
                        >
                          Verify &amp; activate
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </AdminTable>
        )}
      </AdminQueryState>
      {(lifecycle.error || trigger.error) && (
        <p className="mt-4 text-sm text-danger" role="alert">
          {(lifecycle.error ?? trigger.error)?.message}
        </p>
      )}
    </section>
  )
}

function ImportHistory() {
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<CommerceImport>()
  const imports = useCommerceImports(page)
  const retry = useRetryCommerceImport()
  return (
    <section aria-labelledby="import-history">
      <div className="mb-5">
        <h2 className="font-editorial text-2xl" id="import-history">
          Import history
        </h2>
        <p className="mt-1 text-sm text-ink/60">
          Partial and failed imports remain visible and never reconcile missing
          offers.
        </p>
      </div>
      <AdminQueryState
        empty={imports.data?.items.length === 0}
        emptyDescription="No provider import has run. This is expected while every provider is disabled."
        error={imports.isError}
        onRetry={() => void imports.refetch()}
        pending={imports.isPending}
      >
        {imports.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Provider',
                    'Status',
                    'Trigger',
                    'Records',
                    'Attempts',
                    'Started',
                    'Actions',
                  ].map((label) => (
                    <th className={adminTableHead} key={label} scope="col">
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {imports.data.items.map((item) => (
                  <tr key={item.id}>
                    <td className={adminTableCell}>
                      {item.provider_configuration.provider_key}
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{item.status}</Badge>
                    </td>
                    <td className={adminTableCell}>{item.trigger}</td>
                    <td className={adminTableCell}>
                      {item.records_applied}/{item.records_received} applied
                      {item.records_rejected > 0 && (
                        <p className="text-xs text-danger">
                          {item.records_rejected} rejected
                        </p>
                      )}
                    </td>
                    <td className={adminTableCell}>
                      {item.attempt_count}/{item.max_attempts}
                    </td>
                    <td className={adminTableCell}>
                      {formatDate(item.started_at)}
                    </td>
                    <td className={adminTableCell}>
                      <div className="flex gap-2">
                        {item.records_rejected > 0 && (
                          <Button
                            onClick={() => setSelected(item)}
                            size="sm"
                            variant="secondary"
                          >
                            Failures
                          </Button>
                        )}
                        {['failed', 'partial', 'cancelled'].includes(
                          item.status,
                        ) && (
                          <Button
                            loading={retry.isPending}
                            onClick={() => void retry.mutateAsync(item.id)}
                            size="sm"
                          >
                            Retry
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </AdminTable>
            <AdminPagination
              onPageChange={setPage}
              page={imports.data.page}
              total={imports.data.total}
              totalPages={imports.data.total_pages}
            />
          </>
        )}
      </AdminQueryState>
      {retry.error && (
        <p className="mt-4 text-sm text-danger" role="alert">
          {retry.error.message}
        </p>
      )}
      {selected && (
        <ImportFailures
          importRun={selected}
          onClose={() => setSelected(undefined)}
        />
      )}
    </section>
  )
}

function ImportFailures({
  importRun,
  onClose,
}: {
  importRun: CommerceImport
  onClose: () => void
}) {
  const query = useCommerceImportFailures(importRun.id)
  return (
    <Modal
      description={`${importRun.provider_configuration.provider_key} · ${importRun.id}`}
      onOpenChange={(open) => !open && onClose()}
      open
      size="lg"
      title="Import failures"
    >
      <AdminQueryState
        empty={query.data?.items.length === 0}
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <ul className="mt-6 divide-y divide-ink/10 border border-ink/10">
            {query.data.items.map((failure) => (
              <li className="p-4" key={failure.id}>
                <p className="font-medium">{failure.error_code}</p>
                <p className="mt-1 text-sm text-ink/65">
                  {failure.error_message}
                </p>
                <p className="mt-2 text-xs text-ink/55">
                  External record:{' '}
                  {failure.external_record_id ?? 'Not supplied'}
                </p>
              </li>
            ))}
          </ul>
        )}
      </AdminQueryState>
    </Modal>
  )
}

function formatDate(value: string | null) {
  return value ? new Date(value).toLocaleString() : 'No data'
}

function formatDuration(minutes: number) {
  if (minutes % 1440 === 0) return `${minutes / 1440}d`
  if (minutes % 60 === 0) return `${minutes / 60}h`
  return `${minutes}m`
}
