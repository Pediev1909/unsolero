import { useEffect, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { z } from 'zod'

import { Button } from '../../../components/ui/Button'
import { Checkbox } from '../../../components/ui/Checkbox'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import type { Brand, Category } from '../schemas'
import {
  emptyCatalogFilters,
  matchPriceBucket,
  priceBuckets,
  type CatalogFilterValues,
  type PriceBucket,
} from '../useCatalogUrlState'

const filterSchema = z
  .object({
    q: z.string().trim().max(100),
    category: z.string(),
    brand: z.string(),
    minPrice: z.string().regex(/^\d*(\.\d{0,2})?$/, 'Enter a valid price.'),
    maxPrice: z.string().regex(/^\d*(\.\d{0,2})?$/, 'Enter a valid price.'),
    hasOffer: z.boolean(),
  })
  .refine(
    ({ minPrice, maxPrice }) =>
      !minPrice || !maxPrice || Number(minPrice) <= Number(maxPrice),
    {
      message: 'Maximum price must be at least the minimum.',
      path: ['maxPrice'],
    },
  )

const customBucketID = 'custom'

interface CatalogFiltersProps {
  values: CatalogFilterValues
  categories: Category[]
  brands: Brand[]
  fixedCategory?: boolean
  fixedBrand?: boolean
  onApply: (values: CatalogFilterValues) => void
  onClear: () => void
}

export function CatalogFilters({
  values,
  categories,
  brands,
  fixedCategory,
  fixedBrand,
  onApply,
  onClear,
}: CatalogFiltersProps) {
  const [formError, setFormError] = useState<string>()
  // The chips are presets for the two price fields, not a field of their own.
  // "Custom" is the one chip with no bounds to write, so the choice is kept
  // here — pinned to the URL values it was made against, so a new set of
  // values from the URL drops it without an effect having to.
  const [customPriceFor, setCustomPriceFor] =
    useState<CatalogFilterValues | null>(null)
  const customPrice = customPriceFor === values
  const { register, handleSubmit, reset, formState, control, setValue } =
    useForm<CatalogFilterValues>({
      defaultValues: values,
    })

  useEffect(() => reset(values), [reset, values])

  const [minPrice, maxPrice] = useWatch({
    control,
    name: ['minPrice', 'maxPrice'],
  })
  const selectedBucket = customPrice
    ? customBucketID
    : matchPriceBucket(minPrice, maxPrice)

  function chooseBucket(bucket: PriceBucket) {
    setCustomPriceFor(null)
    setValue('minPrice', bucket.minPrice, { shouldDirty: true })
    setValue('maxPrice', bucket.maxPrice, { shouldDirty: true })
  }

  return (
    <form
      className="space-y-6"
      onSubmit={(event) => {
        void handleSubmit((input) => {
          const parsed = filterSchema.safeParse(input)
          if (!parsed.success) {
            setFormError(
              parsed.error.issues[0]?.message ?? 'Check the filters.',
            )
            return
          }
          setFormError(undefined)
          onApply(parsed.data)
        })(event)
      }}
    >
      <Input
        label="Search software"
        placeholder="Product or brand"
        type="search"
        {...register('q')}
      />
      {!fixedCategory && (
        <Select label="Category" {...register('category')}>
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.id} value={category.slug}>
              {category.name}
            </option>
          ))}
        </Select>
      )}
      {!fixedBrand && (
        <Select label="Brand" {...register('brand')}>
          <option value="">All brands</option>
          {brands.map((brand) => (
            <option key={brand.id} value={brand.slug}>
              {brand.name}
            </option>
          ))}
        </Select>
      )}
      <fieldset>
        <legend className="text-sm font-semibold">
          Reference price, USD per month
        </legend>
        <div className="mt-3 flex flex-wrap gap-2">
          {priceBuckets.map((bucket) => (
            <PriceChip
              checked={selectedBucket === bucket.id}
              key={bucket.id}
              label={bucket.label}
              onChoose={() => chooseBucket(bucket)}
              value={bucket.id}
            />
          ))}
          <PriceChip
            checked={selectedBucket === customBucketID}
            label="Custom"
            onChoose={() => setCustomPriceFor(values)}
            value={customBucketID}
          />
        </div>
        {selectedBucket === customBucketID && (
          <div className="mt-4 grid grid-cols-2 gap-3">
            <Input
              inputMode="decimal"
              label="Minimum"
              placeholder="0"
              {...register('minPrice')}
            />
            <Input
              inputMode="decimal"
              label="Maximum"
              placeholder="1,500"
              {...register('maxPrice')}
            />
          </div>
        )}
      </fieldset>
      <Checkbox
        description="Shows only products whose card carries a vendor button."
        label="Only tools with a live vendor offer"
        {...register('hasOffer')}
      />
      {(formError || formState.errors.root?.message) && (
        <p className="text-sm text-ember" role="alert">
          {formError ?? formState.errors.root?.message}
        </p>
      )}
      <div className="grid grid-cols-2 gap-3">
        <Button type="submit">Apply filters</Button>
        <Button
          onClick={() => {
            reset(
              emptyCatalogFilters({
                category: values.category,
                brand: values.brand,
              }),
            )
            setCustomPriceFor(null)
            setFormError(undefined)
            onClear()
          }}
          type="button"
          variant="secondary"
        >
          Clear
        </Button>
      </div>
    </form>
  )
}

// A radio wearing a chip. The input stays in the document for keyboard and
// screen-reader users, so the group is arrowed through like any other set of
// radios, and the label draws the state from it.
function PriceChip({
  checked,
  label,
  value,
  onChoose,
}: {
  checked: boolean
  label: string
  value: string
  onChoose: () => void
}) {
  return (
    <label className="inline-flex min-h-9 cursor-pointer items-center border border-ink/30 px-3 text-xs font-semibold text-ink transition-colors hover:border-ink has-[:checked]:border-ink has-[:checked]:bg-ink has-[:checked]:text-canvas has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-bronze">
      <input
        checked={checked}
        className="sr-only"
        name="priceBucket"
        onChange={onChoose}
        type="radio"
        value={value}
      />
      {label}
    </label>
  )
}
