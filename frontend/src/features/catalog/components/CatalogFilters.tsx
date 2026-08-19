import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import type { Brand, Category } from '../schemas'
import type { CatalogFilterValues } from '../useCatalogUrlState'

const filterSchema = z
  .object({
    q: z.string().trim().max(100),
    category: z.string(),
    brand: z.string(),
    minPrice: z.string().regex(/^\d*(\.\d{0,2})?$/, 'Enter a valid price.'),
    maxPrice: z.string().regex(/^\d*(\.\d{0,2})?$/, 'Enter a valid price.'),
  })
  .refine(
    ({ minPrice, maxPrice }) =>
      !minPrice || !maxPrice || Number(minPrice) <= Number(maxPrice),
    {
      message: 'Maximum price must be at least the minimum.',
      path: ['maxPrice'],
    },
  )

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
  const { register, handleSubmit, reset, formState } =
    useForm<CatalogFilterValues>({
      defaultValues: values,
    })

  useEffect(() => reset(values), [reset, values])

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
        <legend className="text-sm font-semibold">Reference price, USD</legend>
        <div className="mt-3 grid grid-cols-2 gap-3">
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
      </fieldset>
      {(formError || formState.errors.root?.message) && (
        <p className="text-sm text-ember" role="alert">
          {formError ?? formState.errors.root?.message}
        </p>
      )}
      <div className="grid grid-cols-2 gap-3">
        <Button type="submit">Apply filters</Button>
        <Button
          onClick={() => {
            reset({
              q: '',
              category: values.category,
              brand: values.brand,
              minPrice: '',
              maxPrice: '',
            })
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
