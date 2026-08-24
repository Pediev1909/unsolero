import { ProductCard, ProductGrid } from '../product'
import type { ProductCardData } from '../product/ProductCard'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { Section } from '../ui/Section'
import { Skeleton } from '../ui/Skeleton'
import { useProducts } from '../../features/catalog/queries'
import type { ProductSummary } from '../../features/catalog/schemas'

// These four cards used to be a hardcoded array carrying real vendor names and
// real prices. Nothing kept them in step with the catalog, so the first vendor
// price change would have made the home page contradict the product page it
// linked to — on the one site whose entire promise is that its facts are
// checked. They come from the catalog now.
const featuredCount = 4

export function FeaturedProductsSection() {
  const featured = useProducts({ sort: 'featured', pageSize: featuredCount })

  // Fictional fixture rows carry is_demo. They are development data and must
  // never reach the home page, whatever a database happens to hold.
  const products = (featured.data?.products ?? []).filter(
    (product) => !product.is_demo,
  )

  // A failed or empty catalog leaves the section out rather than showing an
  // apology. The home page still works; there is simply nothing to preview.
  if (featured.isError || (featured.isSuccess && products.length === 0)) {
    return null
  }

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
          <p className="max-w-md text-sm leading-6 text-ink/70">
            Entry paid tiers, at the price recorded in our catalog. Open any
            product to inspect its specifications, suitability, and currently
            available merchant offers.
          </p>
        </div>

        <ProductGrid className="mt-14 lg:mt-20" columns={4}>
          {featured.isPending
            ? Array.from({ length: featuredCount }, (_, index) => (
                <Skeleton
                  className="h-64 w-full"
                  key={`featured-placeholder-${index}`}
                />
              ))
            : products.map((product) => (
                <ProductCard key={product.id} product={toCardData(product)} />
              ))}
        </ProductGrid>
      </Container>
    </Section>
  )
}

function toCardData(product: ProductSummary): ProductCardData {
  return {
    id: product.id,
    href: `/products/${product.slug}`,
    name: product.name,
    brand: product.brand.name,
    category: product.category.name,
    priceMinor: product.price.amount_minor,
    currency: product.price.currency,
    image: product.primary_image
      ? {
          src: product.primary_image.url,
          alt: product.primary_image.alt_text,
        }
      : undefined,
  }
}
