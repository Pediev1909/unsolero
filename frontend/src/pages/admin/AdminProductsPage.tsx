import { Plus } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { ButtonLink } from '../../components/ui/ButtonLink'
import { Input } from '../../components/ui/Input'
import { PriceDisplay } from '../../components/ui/PriceDisplay'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  useAdminProducts,
  useProductStatusMutation,
} from '../../features/admin/queries'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

export function AdminProductsPage() {
  usePageMetadata({
    title: 'Products | UNSOLERO admin',
    description: 'Protected product administration.',
    robots: 'noindex, follow',
  })
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const products = useAdminProducts(search, page)
  const status = useProductStatusMutation()

  return (
    <>
      <AdminPageHeader
        action={
          <ButtonLink size="sm" to="/admin/products/new">
            <Plus aria-hidden="true" size={16} /> Create product
          </ButtonLink>
        }
        description="Create structured product records, manage suitability facts, and control publication."
        eyebrow="Catalog"
        title="Products"
      />
      <Input
        className="max-w-md"
        containerClassName="mb-5 max-w-md"
        label="Search products"
        onChange={(event) => {
          setSearch(event.target.value)
          setPage(1)
        }}
        placeholder="Name, slug or brand"
        type="search"
        value={search}
      />
      <AdminQueryState
        empty={products.data?.items.length === 0}
        error={products.isError}
        onRetry={() => void products.refetch()}
        pending={products.isPending}
      >
        {products.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {[
                    'Product',
                    'Category',
                    'Price',
                    'Status',
                    'Updated',
                    '',
                  ].map((heading) => (
                    <th className={adminTableHead} key={heading} scope="col">
                      {heading}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {products.data.items.map((product) => (
                  <tr className="bg-surface hover:bg-paper" key={product.id}>
                    <td className={adminTableCell}>
                      <Link
                        className="font-semibold hover:underline"
                        to={`/admin/products/${product.id}`}
                      >
                        {product.name}
                      </Link>
                      <p className="mt-1 text-xs text-ink/65">
                        /{product.slug}
                      </p>
                    </td>
                    <td className={adminTableCell}>
                      <p>{product.brand_name}</p>
                      <p className="text-xs text-ink/65">
                        {product.category_name}
                      </p>
                    </td>
                    <td className={adminTableCell}>
                      <PriceDisplay
                        amountMinor={product.price_minor}
                        currency={product.currency}
                        size="sm"
                      />
                    </td>
                    <td className={adminTableCell}>
                      <Badge>{product.status}</Badge>
                    </td>
                    <td className={adminTableCell}>
                      {new Date(product.updated_at).toLocaleDateString()}
                    </td>
                    <td className={adminTableCell}>
                      <div className="flex min-w-32 flex-wrap justify-end gap-2">
                        <ButtonLink
                          size="sm"
                          to={`/admin/products/${product.id}`}
                          variant="secondary"
                        >
                          Edit
                        </ButtonLink>
                        {product.status === 'published' ? (
                          <Button
                            disabled={status.isPending}
                            onClick={() =>
                              void status.mutate({
                                id: product.id,
                                status: 'discontinued',
                              })
                            }
                            size="sm"
                            variant="quiet"
                          >
                            Archive
                          </Button>
                        ) : (
                          <ButtonLink
                            size="sm"
                            to={`/admin/evidence/${product.id}`}
                            variant="quiet"
                          >
                            Evidence
                          </ButtonLink>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </AdminTable>
            <AdminPagination
              onPageChange={setPage}
              page={products.data.page}
              total={products.data.total}
              totalPages={products.data.total_pages}
            />
          </>
        )}
      </AdminQueryState>
    </>
  )
}
