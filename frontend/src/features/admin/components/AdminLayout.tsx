import {
  BarChart3,
  Boxes,
  Building2,
  ClipboardList,
  FolderTree,
  Gauge,
  FileCheck2,
  Link2,
  PackageSearch,
  Settings,
  ShoppingBag,
  Tags,
  Users,
} from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'

import { BrandMark } from '../../../components/layout/BrandMark'
import { SkipLink } from '../../../components/ui/SkipLink'
import { cn } from '../../../lib/styles/cn'

const items = [
  { to: '/admin', label: 'Dashboard', icon: Gauge, end: true },
  { to: '/admin/products', label: 'Products', icon: Boxes, end: false },
  { to: '/admin/evidence', label: 'Evidence', icon: FileCheck2, end: false },
  {
    to: '/admin/categories',
    label: 'Categories',
    icon: FolderTree,
    end: false,
  },
  { to: '/admin/brands', label: 'Brands', icon: Tags, end: false },
  { to: '/admin/merchants', label: 'Merchants', icon: Building2, end: false },
  { to: '/admin/offers', label: 'Offers', icon: ShoppingBag, end: false },
  {
    to: '/admin/affiliate-links',
    label: 'Affiliate Links',
    icon: Link2,
    end: false,
  },
  {
    to: '/admin/recommendations',
    label: 'Recommendations',
    icon: PackageSearch,
    end: false,
  },
  { to: '/admin/users', label: 'Users', icon: Users, end: false },
  { to: '/admin/events', label: 'Events', icon: BarChart3, end: false },
  { to: '/admin/content', label: 'Content', icon: ClipboardList, end: false },
  { to: '/admin/settings', label: 'Settings', icon: Settings, end: false },
] as const

export function AdminLayout() {
  return (
    <div className="min-h-screen bg-canvas text-ink">
      <SkipLink />
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-64 flex-col bg-charcoal px-5 py-7 text-canvas lg:flex">
        <BrandMark className="text-canvas" />
        <p className="mt-2 text-[0.65rem] font-semibold tracking-[0.24em] text-canvas/45 uppercase">
          Administration
        </p>
        <nav aria-label="Admin" className="mt-9 flex flex-1 flex-col gap-1">
          {items.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              className={({ isActive }) =>
                cn(
                  'flex min-h-11 items-center gap-3 border-l-2 px-3 text-sm transition-colors',
                  isActive
                    ? 'border-bronze-soft bg-surface/8 text-canvas'
                    : 'border-transparent text-canvas/60 hover:bg-surface/5 hover:text-canvas',
                )
              }
              end={end}
              key={to}
              to={to}
            >
              <Icon aria-hidden="true" size={17} />
              {label}
            </NavLink>
          ))}
        </nav>
        <NavLink
          className="text-xs text-canvas/50 hover:text-canvas"
          to="/account"
        >
          Return to account
        </NavLink>
      </aside>

      <div className="lg:pl-64">
        <header className="sticky top-0 z-20 border-b border-ink/10 bg-canvas/95 px-4 py-3 backdrop-blur lg:hidden">
          <div className="flex items-center justify-between">
            <BrandMark />
            <span className="text-xs font-semibold tracking-[0.18em] uppercase">
              Admin
            </span>
          </div>
          <nav
            aria-label="Admin"
            className="-mx-4 mt-3 flex gap-1 overflow-x-auto px-4 pb-1"
          >
            {items.map(({ to, label, end }) => (
              <NavLink
                className={({ isActive }) =>
                  cn(
                    'shrink-0 border-b-2 px-3 py-2 text-xs font-medium',
                    isActive
                      ? 'border-charcoal text-charcoal'
                      : 'border-transparent text-ink/50',
                  )
                }
                end={end}
                key={to}
                to={to}
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </header>
        <main
          className="mx-auto min-h-screen max-w-[100rem] px-4 py-8 sm:px-6 lg:px-10 lg:py-10"
          id="main-content"
        >
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export function AdminPageHeader({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow: string
  title: string
  description: string
  action?: React.ReactNode
}) {
  return (
    <div className="mb-8 flex flex-col justify-between gap-5 border-b border-ink/10 pb-7 sm:flex-row sm:items-end">
      <div>
        <p className="text-xs font-semibold tracking-[0.2em] text-ink/45 uppercase">
          {eyebrow}
        </p>
        <h1 className="mt-2 font-editorial text-3xl tracking-tight sm:text-4xl">
          {title}
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-ink/60">
          {description}
        </p>
      </div>
      {action}
    </div>
  )
}
