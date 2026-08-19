import {
  ArrowUpRight,
  BarChart3,
  CalendarClock,
  Contact,
  CreditCard,
  FileText,
  LifeBuoy,
  Mail,
  Workflow,
} from 'lucide-react'
import { Link } from 'react-router-dom'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'

const categories = [
  {
    name: 'CRM',
    slug: 'crm',
    note: 'One shared record of every client',
    icon: Contact,
  },
  {
    name: 'Project management',
    slug: 'project-management',
    note: 'Who is doing what, and what is late',
    icon: Workflow,
  },
  {
    name: 'Accounting and invoicing',
    slug: 'accounting-invoicing',
    note: 'Bill from what was agreed',
    icon: FileText,
  },
  {
    name: 'Email marketing',
    slug: 'email-marketing',
    note: 'Reach an audience you own',
    icon: Mail,
  },
  {
    name: 'Help desk',
    slug: 'help-desk',
    note: 'Shared inbox and ticketing',
    icon: LifeBuoy,
  },
  {
    name: 'Analytics',
    slug: 'analytics',
    note: 'Measure what actually happened',
    icon: BarChart3,
  },
  {
    name: 'Scheduling',
    slug: 'scheduling',
    note: 'Let clients book without email',
    icon: CalendarClock,
  },
  {
    name: 'Payments',
    slug: 'payments',
    note: 'Take money without friction',
    icon: CreditCard,
  },
] as const

export function CategoriesSection() {
  return (
    <Section className="scroll-mt-20" id="categories" space="lg">
      <Container>
        <div className="grid gap-7 lg:grid-cols-[0.85fr_1.15fr] lg:items-end lg:gap-24">
          <div>
            <p className="eyebrow">Explore categories</p>
            <Heading className="mt-5 max-w-xl" level={2} size="section">
              Start with the role, not the product.
            </Heading>
          </div>
          <p className="max-w-2xl text-lg leading-8 text-ink/65">
            Eight practical categories cover the catalog. A working stack
            usually needs only two or three of them.
          </p>
        </div>

        <ul className="mt-14 grid border-l border-t border-ink/15 sm:grid-cols-2 lg:grid-cols-4 lg:mt-20">
          {categories.map(({ name, slug, note, icon: Icon }) => (
            <li
              className="group min-h-52 border-b border-r border-ink/15"
              key={name}
            >
              <Link
                className="flex h-full min-h-52 flex-col p-6 transition-colors duration-180 hover:bg-paper/60 sm:p-7"
                to={`/categories/${slug}`}
              >
                <span className="flex items-start justify-between gap-4">
                  <Icon
                    aria-hidden="true"
                    className="text-bronze"
                    size={24}
                    strokeWidth={1.35}
                  />
                  <ArrowUpRight
                    aria-hidden="true"
                    className="text-ink/30 transition-colors group-hover:text-bronze-dark"
                    size={17}
                  />
                </span>
                <h3 className="mt-auto pt-12 font-display text-xl font-medium tracking-[-0.035em]">
                  {name}
                </h3>
                <p className="mt-2 text-sm leading-6 text-ink/55">{note}</p>
              </Link>
            </li>
          ))}
        </ul>
      </Container>
    </Section>
  )
}
