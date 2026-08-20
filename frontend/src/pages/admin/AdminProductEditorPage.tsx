import { ArrowLeft } from 'lucide-react'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { z } from 'zod'

import { Button } from '../../components/ui/Button'
import { ButtonLink } from '../../components/ui/ButtonLink'
import { Input } from '../../components/ui/Input'
import { Select } from '../../components/ui/Select'
import { Textarea } from '../../components/ui/Textarea'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import { ProductAssetsManager } from '../../features/admin/components/ProductAssetsManager'
import { AdminQueryState } from '../../features/admin/components/AdminStates'
import {
  useAdminProduct,
  useAdminReferences,
  useProductMutation,
} from '../../features/admin/queries'
import { ApiError } from '../../lib/api/client'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

const number = z.coerce.number().nonnegative()
const positive = z.coerce.number().positive()
const score = z.coerce.number().int().min(0).max(100)
const formSchema = z.object({
  category_id: z.string().uuid('Choose a category.'),
  brand_id: z.string().uuid('Choose a brand.'),
  name: z.string().trim().min(1).max(180),
  slug: z
    .string()
    .trim()
    .regex(/^[a-z0-9]+(-[a-z0-9]+)*$/),
  description: z.string().trim().min(1),
  price_minor: number,
  currency: z.string().trim().length(3),
  // Kept in the payload, absent from the form. A software product has no
  // physical form, so these are always zero and asking an editor to type
  // "0" into seven boxes about length, weight and warranty before they can
  // save a subscription is asking them to fill in the previous catalog.
  // The backend still enforces real measurements on physical categories.
  length_mm: number,
  width_mm: number,
  height_mm: number,
  weight_grams: number,
  max_capacity_grams: z
    .union([positive, z.nan()])
    .transform((value) => (Number.isNaN(value) ? null : value)),
  material: z.string().trim().max(160),
  warranty_months: number,
  quality_score: score,
  value_score: score,
  durability_score: score,
  beginner_score: score,
  advanced_score: score,
  apartment_score: score,
  noise_score: score,
  portability_score: score,
})

type ProductForm = z.input<typeof formSchema>

const defaults: ProductForm = {
  category_id: '',
  brand_id: '',
  name: '',
  slug: '',
  description: '',
  price_minor: 0,
  currency: 'USD',
  length_mm: 0,
  width_mm: 0,
  height_mm: 0,
  weight_grams: 0,
  max_capacity_grams: Number.NaN,
  material: '',
  warranty_months: 0,
  quality_score: 50,
  value_score: 50,
  durability_score: 50,
  beginner_score: 50,
  advanced_score: 50,
  apartment_score: 0,
  noise_score: 0,
  portability_score: 50,
}

export function AdminProductEditorPage() {
  const { productID } = useParams()
  const editing = Boolean(productID)
  usePageMetadata({
    title: `${editing ? 'Edit' : 'Create'} product | UNSOLERO admin`,
    description: 'Protected structured product editor.',
    robots: 'noindex, follow',
  })
  const product = useAdminProduct(productID)
  const references = useAdminReferences()
  const mutation = useProductMutation(productID)
  const navigate = useNavigate()
  const form = useForm<ProductForm>({ defaultValues: defaults })

  useEffect(() => {
    if (!product.data) return
    const value = product.data
    form.reset({
      category_id: value.category_id,
      brand_id: value.brand_id,
      name: value.name,
      slug: value.slug,
      description: value.description,
      price_minor: value.price_minor,
      currency: value.currency,
      length_mm: value.length_mm,
      width_mm: value.width_mm,
      height_mm: value.height_mm,
      weight_grams: value.weight_grams,
      max_capacity_grams: value.max_capacity_grams ?? Number.NaN,
      material: value.material,
      warranty_months: value.warranty_months,
      quality_score: value.scores.quality,
      value_score: value.scores.value,
      durability_score: value.scores.durability,
      beginner_score: value.scores.beginner,
      advanced_score: value.scores.advanced,
      apartment_score: value.scores.apartment,
      noise_score: value.scores.noise,
      portability_score: value.scores.portability,
    })
  }, [form, product.data])

  const submit = form.handleSubmit(async (raw) => {
    form.clearErrors()
    const parsed = formSchema.safeParse(raw)
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        const field = issue.path[0]
        if (typeof field === 'string')
          form.setError(field as keyof ProductForm, { message: issue.message })
      }
      return
    }
    try {
      const saved = await mutation.mutateAsync(parsed.data)
      if (!editing)
        await navigate(`/admin/products/${saved.id}`, { replace: true })
    } catch (error) {
      form.setError('root.server', {
        message:
          error instanceof ApiError
            ? error.message
            : 'The product could not be saved.',
      })
    }
  })

  const pending = references.isPending || (editing && product.isPending)
  const failed = references.isError || product.isError
  const governed = product.data?.status === 'published'

  return (
    <>
      <Link
        className="mb-5 inline-flex items-center gap-2 text-sm text-ink/70 hover:text-ink"
        to="/admin/products"
      >
        <ArrowLeft aria-hidden="true" size={15} /> Products
      </Link>
      <AdminPageHeader
        description="Recommendation-critical specifications are structured and validated before persistence."
        eyebrow="Catalog"
        title={editing ? 'Edit product' : 'Create product'}
      />
      <AdminQueryState
        empty={false}
        error={failed}
        onRetry={() => {
          void references.refetch()
          if (editing) void product.refetch()
        }}
        pending={pending}
      >
        <form className="space-y-8" onSubmit={(event) => void submit(event)}>
          {governed && (
            <div className="border border-amber/30 bg-amber-soft p-4 text-sm leading-6 text-ink">
              This published projection is read-only. Create and publish a new
              evidence revision to change recommendation-critical facts.
            </div>
          )}
          <fieldset className="space-y-8" disabled={governed}>
            <EditorSection title="Identity">
              <InputField form={form} label="Name" name="name" />
              <InputField form={form} label="Slug" name="slug" />
              <Select
                error={form.formState.errors.category_id?.message}
                label="Category"
                {...form.register('category_id')}
              >
                <option value="">Choose category</option>
                {references.data?.categories.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </Select>
              <Select
                error={form.formState.errors.brand_id?.message}
                label="Brand"
                {...form.register('brand_id')}
              >
                <option value="">Choose brand</option>
                {references.data?.brands.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </Select>
              <Textarea
                containerClassName="sm:col-span-2"
                error={form.formState.errors.description?.message}
                label="Description"
                rows={5}
                {...form.register('description')}
              />
            </EditorSection>

            <EditorSection title="Price">
              <InputField
                form={form}
                label="Price (minor units)"
                name="price_minor"
                number
              />
              <InputField form={form} label="Currency" name="currency" />
            </EditorSection>

            {/* The score names are the ones the product page prints, so what is
                typed here is what a reader sees. Apartment suitability and
                quiet operation are not asked for: they describe a machine in a
                room and stay at zero, which is how the public pages know not
                to show them. */}
            <EditorSection title="Suitability scores">
              {(
                [
                  ['quality_score', 'Product quality'],
                  ['value_score', 'Value for money'],
                  ['durability_score', 'Vendor stability'],
                  ['beginner_score', 'Ease of adoption'],
                  ['advanced_score', 'Depth for power users'],
                  ['portability_score', 'Data portability'],
                ] as const
              ).map(([name, label]) => (
                <InputField
                  form={form}
                  key={name}
                  label={label}
                  name={name}
                  number
                />
              ))}
            </EditorSection>
          </fieldset>

          {form.formState.errors.root?.server?.message && (
            <p className="text-sm text-red-700" role="alert">
              {form.formState.errors.root.server.message}
            </p>
          )}
          <div className="flex justify-end gap-3 border-t border-ink/10 pt-6">
            <Button
              onClick={() => void navigate('/admin/products')}
              variant="quiet"
            >
              Cancel
            </Button>
            {governed && productID ? (
              <ButtonLink to={`/admin/evidence/${productID}`}>
                Inspect evidence
              </ButtonLink>
            ) : (
              <Button loading={mutation.isPending} type="submit">
                Save product
              </Button>
            )}
          </div>
        </form>

        {productID && product.data && (
          <ProductAssetsManager product={product.data} />
        )}
      </AdminQueryState>
    </>
  )
}

function EditorSection({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="border border-ink/10 bg-surface p-5 sm:p-7">
      <h2 className="mb-5 font-editorial text-xl">{title}</h2>
      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">{children}</div>
    </section>
  )
}

function InputField({
  form,
  name,
  label,
  number: numeric = false,
}: {
  form: ReturnType<typeof useForm<ProductForm>>
  name: keyof ProductForm
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
