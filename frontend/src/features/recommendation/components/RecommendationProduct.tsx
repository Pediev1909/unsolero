import { Bookmark, Check, Scale } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { BrandMark } from '../../catalog/components/BrandMark'
import { MerchantAction } from '../../catalog/components/MerchantAction'
import type { ProductSummary } from '../../catalog/schemas'
import type { RecommendationResult } from '../schemas'
import { verticalHasSpatialConstraints } from '../options'

type RecommendedProduct = RecommendationResult['recommended_products'][number]

interface RecommendationProductProps {
  item: RecommendedProduct
  alternative?: RecommendationResult['alternatives'][number]
  compared: boolean
  saved: boolean
  savePending: boolean
  onCompare: (product: ProductSummary) => void
  onSave: (product: ProductSummary) => void
  merchantSource?: 'recommendation' | 'setup'
  recommendationID?: string | null
}

export function RecommendationProduct({
  item,
  alternative,
  compared,
  saved,
  savePending,
  onCompare,
  onSave,
  merchantSource = 'recommendation',
  recommendationID,
}: RecommendationProductProps) {
  const product = item.product
  return (
    <article className="flex h-full flex-col border border-ink/15 bg-surface">
      <div className="flex h-full flex-col p-5 sm:p-6">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="accent">Choice {item.rank}</Badge>
          <Badge variant="success">{item.score}/100 match</Badge>
          {product.is_demo && <Badge>Demo product</Badge>}
        </div>

        <div className="mt-4 flex items-start justify-between gap-4">
          <div className="flex min-w-0 items-center gap-3">
            <BrandMark
              brandName={product.brand.name}
              brandSlug={product.brand.slug}
              loading="eager"
              size="lg"
            />
            <div className="min-w-0">
              <Link
                className="text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65 hover:text-bronze-dark"
                to={`/brands/${product.brand.slug}`}
              >
                {product.brand.name}
              </Link>
              <h2 className="mt-1 font-display text-2xl font-medium leading-tight tracking-[-0.04em]">
                <Link
                  className="hover:text-bronze-dark"
                  to={`/products/${product.slug}`}
                >
                  {product.name}
                </Link>
              </h2>
            </div>
          </div>
          <PriceDisplay
            amountMinor={product.price.amount_minor}
            className="shrink-0"
            currency={product.price.currency}
            size="md"
          />
        </div>

        <ul className="mt-4 grid gap-2 text-sm leading-5 sm:grid-cols-2">
          {item.reasons.slice(0, 4).map((reason) => (
            <li className="border-l border-moss pl-3" key={reason.code}>
              {reason.message}
            </li>
          ))}
        </ul>

        <div className="mt-4 grid grid-cols-2 gap-4 border-y border-ink/10 py-3 text-sm">
          <Spec
            label={product.key_specification.label}
            value={product.key_specification.value}
          />
          <Spec label="Goal fit" value={`${item.breakdown.goal_match}/100`} />
          {verticalHasSpatialConstraints && (
            <Spec
              label="Space fit"
              value={`${item.breakdown.space_match}/100`}
            />
          )}
        </div>

        {alternative && (
          <div className="mt-4 bg-paper p-3 text-sm">
            <span className="font-semibold">
              {alternative.type === 'cheaper' ? 'Lower-cost' : 'Premium'}{' '}
              alternative:
            </span>{' '}
            <Link
              className="underline decoration-ink/25 underline-offset-4"
              to={`/products/${alternative.product.slug}`}
            >
              {alternative.product.name}
            </Link>
          </div>
        )}

        <div className="mt-auto flex flex-wrap gap-2 pt-5">
          <Button
            aria-pressed={compared}
            onClick={() => onCompare(product)}
            size="sm"
            variant={compared ? 'primary' : 'secondary'}
          >
            <Scale aria-hidden="true" size={15} />{' '}
            {compared ? 'Compared' : 'Compare'}
          </Button>
          <Button
            aria-pressed={saved}
            disabled={savePending}
            onClick={() => onSave(product)}
            size="sm"
            variant="secondary"
          >
            {saved ? (
              <Check aria-hidden="true" size={15} />
            ) : (
              <Bookmark aria-hidden="true" size={15} />
            )}
            {saved ? 'Saved' : 'Save'}
          </Button>
          <ButtonLink
            size="sm"
            to={`/products/${product.slug}`}
            variant="quiet"
          >
            View details
          </ButtonLink>
        </div>
        <MerchantAction
          className="mt-4"
          recommendationID={recommendationID}
          slug={product.slug}
          source={merchantSource}
        />
      </div>
    </article>
  )
}

function Spec({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-[0.625rem] font-bold uppercase tracking-[0.12em] text-ink/65">
        {label}
      </p>
      <p className="mt-1 font-semibold">{value}</p>
    </div>
  )
}
