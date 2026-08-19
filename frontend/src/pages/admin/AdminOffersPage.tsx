import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Checkbox } from '../../components/ui/Checkbox'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'
import { PriceDisplay } from '../../components/ui/PriceDisplay'
import { Select } from '../../components/ui/Select'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import type { OfferInput } from '../../features/admin/api'
import {
  useAdminOffers,
  useAdminReferences,
  useOfferMutation,
} from '../../features/admin/queries'
import type { AdminOffer } from '../../features/admin/schemas'
import { ApiError } from '../../lib/api/client'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

const offerFormSchema = z.object({
  merchant_id: z.string().uuid(),
  product_id: z.string().uuid(),
  merchant_sku: z.string().trim().min(1).max(120),
  product_url: z.string().url().startsWith('https://'),
  price_minor: z.coerce.number().int().nonnegative(),
  shipping_minor: z.coerce.number().int().nonnegative(),
  currency: z.string().trim().length(3),
  availability: z.enum([
    'in_stock',
    'backorder',
    'out_of_stock',
    'discontinued',
  ]),
  condition: z.enum(['new', 'refurbished', 'used']),
  is_active: z.boolean(),
  affiliate_url: z.union([
    z.literal(''),
    z.string().url().startsWith('https://'),
  ]),
})
type OfferForm = z.input<typeof offerFormSchema>

const defaults: OfferForm = {
  merchant_id: '',
  product_id: '',
  merchant_sku: '',
  product_url: '',
  price_minor: 0,
  shipping_minor: 0,
  currency: 'USD',
  availability: 'in_stock',
  condition: 'new',
  is_active: true,
  affiliate_url: '',
}

export function AdminOffersPage() {
  usePageMetadata({
    title: 'Offers | UNSOLERO admin',
    description: 'Protected merchant offer administration.',
    robots: 'noindex, follow',
  })
  const [page, setPage] = useState(1)
  const query = useAdminOffers(page)
  const [editing, setEditing] = useState<AdminOffer | null | undefined>()
  return (
    <>
      <AdminPageHeader
        action={
          <Button onClick={() => setEditing(null)} size="sm">
            <Plus aria-hidden="true" size={15} /> Create offer
          </Button>
        }
        description="Maintain merchant price, availability, destination, and active state. Commission is never a recommendation input."
        eyebrow="Commerce"
        title="Offers"
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
                    'Price',
                    'Availability',
                    'Freshness',
                    'Links',
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
                {query.data.items.map((offer) => (
                  <tr key={offer.id}>
                    <td className={adminTableCell}>
                      <p className="font-medium">{offer.product_name}</p>
                      <p className="text-xs text-ink/65">
                        {offer.merchant_sku}
                      </p>
                    </td>
                    <td className={adminTableCell}>{offer.merchant_name}</td>
                    <td className={adminTableCell}>
                      <PriceDisplay
                        amountMinor={offer.price_minor + offer.shipping_minor}
                        currency={offer.currency}
                        size="sm"
                      />
                    </td>
                    <td className={adminTableCell}>
                      {offer.availability.replaceAll('_', ' ')}
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{offer.freshness_status.replace('_', ' ')}</Badge>
                      <p className="mt-1 text-xs text-ink/70">
                        {offer.expires_at
                          ? `Expires ${new Date(offer.expires_at).toLocaleString()}`
                          : `Checked ${new Date(offer.last_checked_at).toLocaleString()}`}
                      </p>
                    </td>
                    <td className={adminTableCell}>{offer.affiliate_links}</td>
                    <td className={adminTableCell}>
                      <Badge>{offer.is_active ? 'active' : 'inactive'}</Badge>
                    </td>
                    <td className={adminTableCell}>
                      <Button
                        onClick={() => setEditing(offer)}
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
      {editing !== undefined && (
        <OfferEditor offer={editing} onClose={() => setEditing(undefined)} />
      )}
    </>
  )
}

function OfferEditor({
  offer,
  onClose,
}: {
  offer: AdminOffer | null
  onClose: () => void
}) {
  const references = useAdminReferences()
  const mutation = useOfferMutation(offer?.id)
  const form = useForm<OfferForm>({ defaultValues: defaults })
  useEffect(() => {
    if (!offer) return
    form.reset({
      merchant_id: offer.merchant_id,
      product_id: offer.product_id,
      merchant_sku: offer.merchant_sku,
      product_url: offer.product_url,
      price_minor: offer.price_minor,
      shipping_minor: offer.shipping_minor,
      currency: offer.currency,
      availability: offer.availability as OfferForm['availability'],
      condition: offer.condition as OfferForm['condition'],
      is_active: offer.is_active,
      affiliate_url: '',
    })
  }, [form, offer])
  const submit = form.handleSubmit(async (raw) => {
    form.clearErrors()
    const parsed = offerFormSchema.safeParse(raw)
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        const key = issue.path[0]
        if (typeof key === 'string')
          form.setError(key as keyof OfferForm, { message: issue.message })
      }
      return
    }
    const { affiliate_url: affiliateURL, ...offerFields } = parsed.data
    const input: OfferInput = {
      ...offerFields,
      currency: offerFields.currency.toUpperCase(),
      affiliate: affiliateURL
        ? {
            provider: 'direct',
            destination_url: affiliateURL,
            external_reference: null,
            disclosure_label: 'Affiliate link',
            is_active: true,
            priority: 0,
            program_id: null,
            commission_type: 'unknown',
            commission_rate_bps: null,
            commission_amount_minor: null,
            commission_currency: null,
          }
        : null,
    }
    try {
      await mutation.mutateAsync(input)
      onClose()
    } catch (error) {
      form.setError('root.server', {
        message:
          error instanceof ApiError
            ? error.message
            : 'The offer could not be saved.',
      })
    }
  })
  return (
    <Modal
      description="Offer changes are recorded in the admin audit log."
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
      open
      title={offer ? 'Edit offer' : 'Create offer'}
    >
      <form
        className="grid gap-4 sm:grid-cols-2"
        onSubmit={(event) => void submit(event)}
      >
        <Select
          error={form.formState.errors.merchant_id?.message}
          label="Merchant"
          {...form.register('merchant_id')}
        >
          <option value="">Choose merchant</option>
          {references.data?.merchants.map((item) => (
            <option key={item.id} value={item.id}>
              {item.name}
            </option>
          ))}
        </Select>
        <Select
          error={form.formState.errors.product_id?.message}
          label="Product"
          {...form.register('product_id')}
        >
          <option value="">Choose product</option>
          {references.data?.products.map((item) => (
            <option key={item.id} value={item.id}>
              {item.name}
            </option>
          ))}
        </Select>
        <OfferInputField form={form} label="Merchant SKU" name="merchant_sku" />
        <OfferInputField
          form={form}
          label="Product destination"
          name="product_url"
        />
        <OfferInputField
          form={form}
          label="Price (minor units)"
          name="price_minor"
          number
        />
        <OfferInputField
          form={form}
          label="Shipping (minor units)"
          name="shipping_minor"
          number
        />
        <OfferInputField form={form} label="Currency" name="currency" />
        <Select label="Availability" {...form.register('availability')}>
          <option value="in_stock">In stock</option>
          <option value="backorder">Backorder</option>
          <option value="out_of_stock">Out of stock</option>
          <option value="discontinued">Discontinued</option>
        </Select>
        <Select label="Condition" {...form.register('condition')}>
          <option value="new">New</option>
          <option value="refurbished">Refurbished</option>
          <option value="used">Used</option>
        </Select>
        {!offer && (
          <OfferInputField
            form={form}
            label="Affiliate URL (optional)"
            name="affiliate_url"
          />
        )}
        <Checkbox label="Active offer" {...form.register('is_active')} />
        {form.formState.errors.root?.server?.message && (
          <p className="text-sm text-red-700 sm:col-span-2" role="alert">
            {form.formState.errors.root.server.message}
          </p>
        )}
        <div className="flex justify-end gap-3 border-t border-ink/10 pt-4 sm:col-span-2">
          <Button onClick={onClose} variant="quiet">
            Cancel
          </Button>
          <Button loading={mutation.isPending} type="submit">
            Save offer
          </Button>
        </div>
      </form>
    </Modal>
  )
}

function OfferInputField({
  form,
  name,
  label,
  number: numeric = false,
}: {
  form: ReturnType<typeof useForm<OfferForm>>
  name: keyof OfferForm
  label: string
  number?: boolean
}) {
  return (
    <Input
      error={form.formState.errors[name]?.message}
      label={label}
      type={numeric ? 'number' : 'text'}
      {...form.register(name, numeric ? { valueAsNumber: true } : undefined)}
    />
  )
}
