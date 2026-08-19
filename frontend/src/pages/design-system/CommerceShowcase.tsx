import {
  ProductCard,
  type ProductCardData,
} from '../../components/product/ProductCard'
import { ProductGrid } from '../../components/product/ProductGrid'
import { ShowcaseBlock } from './ShowcaseBlock'

const demoProducts: ProductCardData[] = [
  {
    id: 'demo-northline-nest-24-pair',
    name: 'Demo Northwind CRM',
    brand: 'Demo Northline',
    category: 'CRM',
    priceMinor: 32900,
    currency: 'USD',
    href: '/design-system#products',
    badge: { label: 'Fictional demo', variant: 'accent' },
  },
  {
    id: 'demo-quietform-dial-20-pair',
    name: 'Demo Pocket CRM',
    brand: 'Demo QuietForm',
    category: 'CRM',
    priceMinor: 27900,
    currency: 'USD',
    href: '/design-system#products',
    badge: { label: 'Fictional demo', variant: 'neutral' },
  },
  {
    id: 'demo-oak-iron-foldaway-flat-bench',
    name: 'Demo Oak & Iron Foldaway Flat Bench',
    brand: 'Demo Oak & Iron',
    category: 'Benches',
    priceMinor: 14900,
    currency: 'USD',
    href: '/design-system#products',
    badge: { label: 'Fictional demo', variant: 'neutral' },
  },
]

export function CommerceShowcase() {
  return (
    <ShowcaseBlock
      description="Product surfaces emphasize identity, price, and clearly sourced labels. Recommendation reasons and affiliate details remain separate concerns."
      eyebrow="08 / Commerce"
      title="Product composition"
    >
      <p className="mb-5 text-xs leading-5 text-ink/55">
        Showcase records are fictional seed data. No reviews, availability, or
        merchant claims are implied.
      </p>
      <ProductGrid columns={3}>
        {demoProducts.map((product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </ProductGrid>
    </ShowcaseBlock>
  )
}
