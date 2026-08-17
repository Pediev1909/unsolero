import { useState, type ReactNode } from 'react'

import { Badge } from '../../components/ui/Badge'
import { AdminPageHeader } from '../../features/admin/components/AdminLayout'
import {
  AdminQueryState,
  AdminPagination,
  AdminTable,
  adminTableCell,
  adminTableHead,
} from '../../features/admin/components/AdminStates'
import {
  useAdminBrands,
  useAdminCategories,
  useAdminMerchants,
  useAdminUsers,
} from '../../features/admin/queries'
import { usePageMetadata } from '../../lib/seo/usePageMetadata'

export function AdminCategoriesPage() {
  const query = useAdminCategories()
  useAdminMeta('Categories')
  return (
    <ReferencePage
      description="Category activation and product coverage from the live catalog."
      headers={['Category', 'Slug', 'Products', 'Status']}
      query={query}
      render={(item) => (
        <tr key={item.id}>
          <td className={adminTableCell}>{item.name}</td>
          <td className={adminTableCell}>/{item.slug}</td>
          <td className={adminTableCell}>{item.products}</td>
          <td className={adminTableCell}>
            <Badge>{item.is_active ? 'active' : 'inactive'}</Badge>
          </td>
        </tr>
      )}
      title="Categories"
    />
  )
}

export function AdminBrandsPage() {
  const query = useAdminBrands()
  useAdminMeta('Brands')
  return (
    <ReferencePage
      description="Brand records and current product coverage."
      headers={['Brand', 'Slug', 'Products', 'Status']}
      query={query}
      render={(item) => (
        <tr key={item.id}>
          <td className={adminTableCell}>{item.name}</td>
          <td className={adminTableCell}>/{item.slug}</td>
          <td className={adminTableCell}>{item.products}</td>
          <td className={adminTableCell}>
            <Badge>{item.is_active ? 'active' : 'inactive'}</Badge>
          </td>
        </tr>
      )}
      title="Brands"
    />
  )
}

export function AdminMerchantsPage() {
  const query = useAdminMerchants()
  useAdminMeta('Merchants')
  return (
    <ReferencePage
      description="Merchant eligibility, trust policy, and offer coverage."
      headers={['Merchant', 'Country', 'Trust', 'Offers', 'Status']}
      query={query}
      render={(item) => (
        <tr key={item.id}>
          <td className={adminTableCell}>
            <p className="font-medium">{item.name}</p>
            <a
              className="text-xs text-ink/45 hover:underline"
              href={item.website_url}
              rel="noopener noreferrer"
              target="_blank"
            >
              {item.website_url}
            </a>
          </td>
          <td className={adminTableCell}>{item.country_code}</td>
          <td className={adminTableCell}>{item.trust_score}/100</td>
          <td className={adminTableCell}>{item.offers}</td>
          <td className={adminTableCell}>
            <Badge>{item.status}</Badge>
          </td>
        </tr>
      )}
      title="Merchants"
    />
  )
}

export function AdminUsersPage() {
  const [page, setPage] = useState(1)
  const query = useAdminUsers(page)
  useAdminMeta('Users')
  return (
    <ReferencePage
      description="Account status and assigned roles. Password credentials and session tokens are never exposed."
      headers={['Account', 'Status', 'Roles', 'Last login', 'Created']}
      pagination={
        query.data ? (
          <AdminPagination
            onPageChange={setPage}
            page={query.data.page}
            total={query.data.total}
            totalPages={query.data.total_pages}
          />
        ) : undefined
      }
      query={{ ...query, data: query.data?.items }}
      render={(item) => (
        <tr key={item.id}>
          <td className={adminTableCell}>{item.email}</td>
          <td className={adminTableCell}>
            <Badge>{item.status}</Badge>
          </td>
          <td className={adminTableCell}>
            {item.roles.length ? item.roles.join(', ') : 'Member'}
          </td>
          <td className={adminTableCell}>
            {item.last_login_at
              ? new Date(item.last_login_at).toLocaleString()
              : 'Never'}
          </td>
          <td className={adminTableCell}>
            {new Date(item.created_at).toLocaleDateString()}
          </td>
        </tr>
      )}
      title="Users"
    />
  )
}

function ReferencePage<T>({
  title,
  description,
  headers,
  query,
  render,
  pagination,
}: {
  title: string
  description: string
  headers: string[]
  query: {
    data: T[] | undefined
    isPending: boolean
    isError: boolean
    refetch: () => Promise<unknown>
  }
  render: (item: T) => ReactNode
  pagination?: ReactNode
}) {
  return (
    <>
      <AdminPageHeader
        description={description}
        eyebrow="Operations"
        title={title}
      />
      <AdminQueryState
        empty={query.data?.length === 0}
        error={query.isError}
        onRetry={() => void query.refetch()}
        pending={query.isPending}
      >
        {query.data && (
          <>
            <AdminTable>
              <thead>
                <tr>
                  {headers.map((header) => (
                    <th className={adminTableHead} key={header} scope="col">
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>{query.data.map(render)}</tbody>
            </AdminTable>
            {pagination}
          </>
        )}
      </AdminQueryState>
    </>
  )
}

function useAdminMeta(title: string) {
  usePageMetadata({
    title: `${title} | UNSOLERO admin`,
    description: `Protected ${title.toLowerCase()} administration.`,
    robots: 'noindex, follow',
  })
}
