import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'

import { Button } from '../../../components/ui/Button'
import { Checkbox } from '../../../components/ui/Checkbox'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { adminApi } from '../api'
import { adminKeys } from '../queries'
import type { AdminProduct } from '../schemas'

export function ProductAssetsManager({ product }: { product: AdminProduct }) {
  return (
    <div className="mt-12 grid gap-8 xl:grid-cols-2">
      <ImageManager product={product} />
      <AttributeManager product={product} />
    </div>
  )
}

function ImageManager({ product }: { product: AdminProduct }) {
  const client = useQueryClient()
  const [url, setURL] = useState('')
  const [file, setFile] = useState<File>()
  const [alt, setAlt] = useState('')
  const [primary, setPrimary] = useState(product.images.length === 0)
  const mutation = useMutation({
    mutationFn: () =>
      file
        ? adminApi.uploadImage(product.id, {
            file,
            alt_text: alt,
            sort_order: product.images.length,
            is_primary: primary,
          })
        : adminApi.addImage(product.id, {
            url,
            alt_text: alt,
            sort_order: product.images.length,
            is_primary: primary,
          }),
    onSuccess: () => {
      setURL('')
      setFile(undefined)
      setAlt('')
      void client.invalidateQueries({ queryKey: adminKeys.product(product.id) })
    },
  })
  const remove = useMutation({
    mutationFn: (id: string) => adminApi.deleteImage(product.id, id),
    onSuccess: () =>
      void client.invalidateQueries({
        queryKey: adminKeys.product(product.id),
      }),
  })

  return (
    <section className="border border-ink/10 bg-surface p-5 sm:p-7">
      <h2 className="font-editorial text-xl">Images</h2>
      <p className="mt-1 text-xs leading-5 text-ink/50">
        Upload JPEG, PNG, or WebP files, or attach an existing HTTPS asset.
      </p>
      <div className="mt-5 space-y-3">
        {product.images.map((image) => (
          <div
            className="flex items-center gap-3 border-b border-ink/8 pb-3"
            key={image.id}
          >
            <img
              alt={image.alt_text}
              className="h-14 w-14 object-cover"
              src={image.url}
            />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{image.alt_text}</p>
              <p className="truncate text-xs text-ink/40">{image.url}</p>
            </div>
            <Button
              aria-label={`Delete ${image.alt_text}`}
              onClick={() => void remove.mutate(image.id)}
              size="sm"
              variant="quiet"
            >
              <Trash2 aria-hidden="true" size={15} />
            </Button>
          </div>
        ))}
      </div>
      <div className="mt-5 space-y-4">
        <Input
          accept="image/jpeg,image/png,image/webp"
          hint="Maximum file size: 5 MB."
          label="Upload image"
          onChange={(event) => setFile(event.target.files?.[0])}
          type="file"
        />
        <p className="text-center text-xs font-semibold tracking-[0.12em] text-ink/35 uppercase">
          or
        </p>
        <Input
          label="Image URL"
          onChange={(event) => setURL(event.target.value)}
          value={url}
        />
        <Input
          label="Alternative text"
          onChange={(event) => setAlt(event.target.value)}
          value={alt}
        />
        <Checkbox
          checked={primary}
          label="Primary image"
          onChange={(event) => setPrimary(event.target.checked)}
        />
        <Button
          disabled={(!url && !file) || !alt}
          loading={mutation.isPending}
          onClick={() => void mutation.mutate()}
          size="sm"
        >
          <Plus aria-hidden="true" size={15} /> Add image
        </Button>
        {mutation.isError ? (
          <p className="text-sm text-ember" role="alert">
            {mutation.error.message}
          </p>
        ) : null}
      </div>
    </section>
  )
}

function AttributeManager({ product }: { product: AdminProduct }) {
  const client = useQueryClient()
  const [key, setKey] = useState('')
  const [value, setValue] = useState('')
  const [type, setType] = useState<'text' | 'number' | 'boolean'>('text')
  const [unit, setUnit] = useState('')
  const [filterable, setFilterable] = useState(true)
  const mutation = useMutation({
    mutationFn: () => {
      const numericValue = type === 'number' ? Number(value) : null
      return adminApi.upsertAttribute(product.id, key, {
        type,
        numeric_value: numericValue,
        text_value: type === 'text' ? value : null,
        boolean_value: type === 'boolean' ? value === 'true' : null,
        unit: type === 'number' && unit.trim() ? unit.trim() : null,
        is_filterable: filterable,
      })
    },
    onSuccess: () => {
      setKey('')
      setValue('')
      setUnit('')
      void client.invalidateQueries({ queryKey: adminKeys.product(product.id) })
    },
  })
  const remove = useMutation({
    mutationFn: (attributeKey: string) =>
      adminApi.deleteAttribute(product.id, attributeKey),
    onSuccess: () =>
      void client.invalidateQueries({
        queryKey: adminKeys.product(product.id),
      }),
  })

  return (
    <section className="border border-ink/10 bg-surface p-5 sm:p-7">
      <h2 className="font-editorial text-xl">Structured attributes</h2>
      <p className="mt-1 text-xs text-ink/50">
        Typed product facts remain separate from recommendation-critical
        columns.
      </p>
      <div className="mt-5 space-y-3">
        {product.attributes.map((attribute) => (
          <div
            className="flex items-center gap-3 border-b border-ink/8 pb-3"
            key={attribute.key}
          >
            <div className="flex-1">
              <p className="text-sm font-medium">{attribute.key}</p>
              <p className="text-xs text-ink/45">
                {attribute.text_value ??
                  attribute.numeric_value?.toString() ??
                  String(attribute.boolean_value)}{' '}
                {attribute.unit ?? ''}
                <span className="ml-2 uppercase">{attribute.type}</span>
              </p>
            </div>
            <Button
              aria-label={`Delete ${attribute.key}`}
              onClick={() => void remove.mutate(attribute.key)}
              size="sm"
              variant="quiet"
            >
              <Trash2 aria-hidden="true" size={15} />
            </Button>
          </div>
        ))}
      </div>
      <div className="mt-5 grid gap-4 sm:grid-cols-2">
        <Input
          label="Attribute key"
          onChange={(event) => setKey(event.target.value)}
          placeholder="handle_finish"
          value={key}
        />
        <Select
          label="Value type"
          onChange={(event) => {
            setType(event.target.value as 'text' | 'number' | 'boolean')
            setValue(event.target.value === 'boolean' ? 'true' : '')
          }}
          value={type}
        >
          <option value="text">Text</option>
          <option value="number">Number</option>
          <option value="boolean">Yes / no</option>
        </Select>
        {type === 'boolean' ? (
          <Select
            label="Value"
            onChange={(event) => setValue(event.target.value)}
            value={value}
          >
            <option value="true">Yes</option>
            <option value="false">No</option>
          </Select>
        ) : (
          <Input
            label={type === 'number' ? 'Numeric value' : 'Text value'}
            onChange={(event) => setValue(event.target.value)}
            step={type === 'number' ? 'any' : undefined}
            type={type === 'number' ? 'number' : 'text'}
            value={value}
          />
        )}
        {type === 'number' ? (
          <Input
            label="Unit (optional)"
            onChange={(event) => setUnit(event.target.value)}
            placeholder="mm"
            value={unit}
          />
        ) : null}
        <Checkbox
          checked={filterable}
          label="Available as a catalog filter"
          onChange={(event) => setFilterable(event.target.checked)}
        />
      </div>
      <Button
        className="mt-4"
        disabled={
          !key.trim() ||
          value === '' ||
          (type === 'number' && !Number.isFinite(Number(value)))
        }
        loading={mutation.isPending}
        onClick={() => void mutation.mutate()}
        size="sm"
      >
        <Plus aria-hidden="true" size={15} /> Save attribute
      </Button>
      {mutation.isError ? (
        <p className="mt-3 text-sm text-ember" role="alert">
          {mutation.error.message}
        </p>
      ) : null}
    </section>
  )
}
