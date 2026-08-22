import { BrandMark } from '../../features/catalog/components/BrandMark'
import { useProducts } from '../../features/catalog/queries'

/**
 * The hero illustration, drawn from the catalog rather than invented.
 *
 * It used to be a static SVG wireframe: grey bars arranged to look like a
 * dashboard. On a site whose entire argument is that its numbers are real and
 * checked, drawing fake interface is the worst available choice — and its own
 * alt text described "a small team reviewing tools", which is not what the
 * picture showed either.
 *
 * This shows four actual products at their actual prices. If the catalog
 * changes, so does the picture, because it is the catalog.
 *
 * The height is fixed so the hero does not jump when the data lands, and the
 * loading state keeps the same shape rather than collapsing.
 */
export function HeroCatalogPanel() {
  const products = useProducts({ pageSize: 4, sort: 'featured' })
  const rows: PanelRow[] = (products.data?.products ?? [])
    .slice(0, 4)
    .map((product) => ({
      id: product.id,
      brandName: product.brand.name,
      brandSlug: product.brand.slug,
      name: product.name,
      category: product.category.name,
      price:
        product.price.amount_minor === 0
          ? 'Free'
          : new Intl.NumberFormat('en-US', {
              style: 'currency',
              currency: product.price.currency,
              maximumFractionDigits: 0,
            }).format(product.price.amount_minor / 100),
      billing: product.key_specification.value,
    }))

  return (
    <div className="w-full max-w-sm">
      <p className="mb-3 text-[0.625rem] font-bold tracking-[0.16em] text-ink/60 uppercase">
        Live from the catalog
      </p>

      <div className="divide-y divide-ink/10 border border-ink/15 bg-surface shadow-raised">
        {(rows.length > 0 ? rows : placeholderRows).map((row, index) => (
          <div
            className="flex items-center gap-3 px-4 py-3.5"
            key={row.id ?? index}
          >
            {row.brandSlug ? (
              <BrandMark
                brandName={row.brandName}
                brandSlug={row.brandSlug}
                size="sm"
              />
            ) : (
              <span
                aria-hidden="true"
                className="h-7 w-7 shrink-0 rounded-sm bg-paper"
              />
            )}
            <span className="min-w-0 flex-1">
              <span className="block truncate text-sm font-semibold">
                {row.name || ' '}
              </span>
              <span className="block truncate text-xs text-ink/65">
                {row.category || ' '}
              </span>
            </span>
            <span className="shrink-0 text-right">
              <span className="block font-display text-sm font-semibold tabular-nums">
                {row.price || ' '}
              </span>
              <span className="block text-[0.625rem] text-ink/60 uppercase">
                {row.billing || ' '}
              </span>
            </span>
          </div>
        ))}
      </div>

      <p className="mt-3 text-xs leading-5 text-ink/65">
        Real products, real prices, each read from the vendor&rsquo;s own page.
      </p>
    </div>
  )
}

interface PanelRow {
  id?: string
  brandName: string
  brandSlug: string | null
  name: string
  category: string
  price: string
  billing: string
}

// Same shape, no content: keeps the hero from resizing while the catalog
// loads, without inventing a product that does not exist.
const placeholderRows: PanelRow[] = Array.from({ length: 4 }, () => ({
  brandName: '',
  brandSlug: null,
  name: '',
  category: '',
  price: '',
  billing: '',
}))
