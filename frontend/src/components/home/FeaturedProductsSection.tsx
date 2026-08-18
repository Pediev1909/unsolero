import { ProductCard, ProductGrid } from '../product'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'
import { featuredProducts } from './homeData'

export function FeaturedProductsSection() {
  return (
    <Section
      className="scroll-mt-20"
      id="featured"
      space="lg"
      surface="surface"
    >
      <Container>
        <div className="flex flex-col justify-between gap-7 lg:flex-row lg:items-end">
          <div>
            <p className="eyebrow">Featured tools</p>
            <Heading className="mt-5 max-w-3xl" level={2} size="section">
              A preview of the tools we evaluate.
            </Heading>
          </div>
          <p className="max-w-md text-sm leading-6 text-ink/60">
            Fictional demo products shown at their seeded reference prices. Open
            any product to inspect its specifications, suitability, and
            currently available merchant offers.
          </p>
        </div>

        <ProductGrid className="mt-14 lg:mt-20" columns={4}>
          {featuredProducts.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </ProductGrid>
      </Container>
    </Section>
  )
}
