import {
  ArrowUpRight,
  Bike,
  Blocks,
  CircleDot,
  Dumbbell,
  Gauge,
  Layers3,
  StretchHorizontal,
  Weight,
} from 'lucide-react'
import { Link } from 'react-router-dom'

import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'

const categories = [
  {
    name: 'Adjustable dumbbells',
    slug: 'adjustable-dumbbells',
    note: 'Strength with a small footprint',
    icon: Dumbbell,
  },
  {
    name: 'Benches',
    slug: 'benches',
    note: 'Flat, incline, and folding',
    icon: Layers3,
  },
  {
    name: 'Power racks',
    slug: 'power-racks',
    note: 'Barbell training foundations',
    icon: Blocks,
  },
  {
    name: 'Barbells',
    slug: 'barbells',
    note: 'Technique and general-purpose bars',
    icon: StretchHorizontal,
  },
  {
    name: 'Weight plates',
    slug: 'weight-plates',
    note: 'Iron, bumper, and coated sets',
    icon: CircleDot,
  },
  {
    name: 'Kettlebells',
    slug: 'kettlebells',
    note: 'Fixed and adjustable loads',
    icon: Weight,
  },
  {
    name: 'Resistance bands',
    slug: 'resistance-bands',
    note: 'Portable, versatile resistance',
    icon: Gauge,
  },
  {
    name: 'Cardio machines',
    slug: 'cardio-machines',
    note: 'Home-oriented conditioning',
    icon: Bike,
  },
]

export function CategoriesSection() {
  return (
    <Section className="scroll-mt-20" id="categories" space="lg">
      <Container>
        <div className="grid gap-7 lg:grid-cols-[0.85fr_1.15fr] lg:items-end lg:gap-24">
          <div>
            <p className="eyebrow">Explore equipment</p>
            <Heading className="mt-5 max-w-xl" level={2} size="section">
              Start with the role, not the product.
            </Heading>
          </div>
          <p className="max-w-2xl text-lg leading-8 text-ink/65">
            Eight practical categories cover the demo catalog. The right setup
            may use only two or three of them.
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
