import { Link } from 'react-router-dom'

import { Container } from '../../../components/ui/Container'
import { Heading } from '../../../components/ui/Heading'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { BrandMark } from '../../catalog/components/BrandMark'
import { MerchantAction } from '../../catalog/components/MerchantAction'
import type { ProductSummary } from '../../catalog/schemas'

/**
 * The products a piece covers, on the first screen.
 *
 * A reader arriving from a search for "X alternatives" wants the names, the
 * prices and a way out before they want the argument. This is that: one card
 * per product in the order the piece lists them, with the vendor control only
 * where a live offer exists. No card is promoted over another and the order
 * is the editor's, not the commission's.
 *
 * The compact MerchantAction drops its own disclosure, so the sentence under
 * the strip is the only one a reader gets here. If it goes, the disclosure
 * goes.
 */
export function AtAGlance({ products }: { products: ProductSummary[] }) {
  if (products.length === 0) return null

  return (
    <section aria-labelledby="at-a-glance" className="border-b border-ink/15">
      <Container className="py-7 sm:py-9">
        <Heading id="at-a-glance" level={2} size="subtitle">
          At a glance
        </Heading>
        {/* A scrolling row on a phone, a grid from the small breakpoint up.
            Cards keep their own hairline rather than sharing a background
            grid, because a shared grid leaves grey holes whenever the count
            does not fill the last row — five products on four columns. */}
        <ul className="mt-5 flex snap-x gap-3 overflow-x-auto pb-1 sm:grid sm:grid-cols-2 sm:overflow-visible sm:pb-0 lg:grid-cols-[repeat(auto-fit,minmax(14rem,1fr))]">
          {products.map((product) => (
            <li
              className="flex w-64 shrink-0 snap-start flex-col border border-ink/15 bg-surface p-4 sm:w-auto"
              key={product.id}
            >
              <div className="flex items-start gap-3">
                <BrandMark
                  brandName={product.brand.name}
                  brandSlug={product.brand.slug}
                  size="sm"
                />
                <div className="min-w-0">
                  <p className="text-[0.625rem] font-bold uppercase tracking-[0.13em] text-ink/65">
                    {product.brand.name}
                  </p>
                  <p className="mt-0.5 font-display text-lg font-medium leading-tight tracking-[-0.03em]">
                    <Link
                      className="hover:text-bronze-dark"
                      to={`/products/${product.slug}`}
                    >
                      {product.name}
                    </Link>
                  </p>
                </div>
              </div>
              <p className="mt-4 flex flex-wrap items-baseline gap-x-2 gap-y-1">
                <PriceDisplay
                  amountMinor={product.price.amount_minor}
                  currency={product.price.currency}
                  size="sm"
                />
                <span className="text-xs text-ink/65">
                  {product.key_specification.value}
                </span>
              </p>
              {/* Nothing rather than a disabled control when there is no live
                  offer: an empty slot reads as "no offer", a greyed button
                  reads as "we are hiding something". */}
              {product.purchase_path && (
                <MerchantAction
                  className="mt-4"
                  compact
                  slug={product.slug}
                  source="promotion"
                />
              )}
            </li>
          ))}
        </ul>
        <p className="mt-4 text-xs leading-5 text-ink/65">
          Prices read from vendor pages; the date is on each product page. The
          vendor buttons are affiliate links: they pay us if you subscribe, and
          they changed nothing about the order here.
        </p>
      </Container>
    </section>
  )
}
