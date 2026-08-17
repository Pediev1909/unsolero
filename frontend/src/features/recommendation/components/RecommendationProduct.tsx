import { Bookmark, Check, Scale } from 'lucide-react'
import { Link } from 'react-router-dom'

import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { ButtonLink } from '../../../components/ui/ButtonLink'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { MerchantAction } from '../../catalog/components/MerchantAction'
import type { ProductSummary } from '../../catalog/schemas'
import type { RecommendationResult } from '../schemas'

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
    <article className="grid overflow-hidden border border-ink/15 bg-surface lg:grid-cols-[minmax(16rem,0.72fr)_1.28fr]">
      <Link
        className="block min-h-64 bg-paper"
        to={`/products/${product.slug}`}
      >
        {product.primary_image ? (
          <img
            alt={product.primary_image.alt_text}
            className="size-full object-cover"
            height={product.primary_image.height_px}
            loading="lazy"
            src={product.primary_image.url}
            width={product.primary_image.width_px}
          />
        ) : (
          <span className="grid size-full place-items-center text-sm text-ink/45">
            Image unavailable
          </span>
        )}
      </Link>
      <div className="p-5 sm:p-7 lg:p-8">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="accent">Choice {item.rank}</Badge>
          <Badge variant="success">{item.score}/100 match</Badge>
          {product.is_demo && <Badge>Demo product</Badge>}
        </div>
        <p className="mt-5 text-xs font-bold uppercase tracking-[0.14em] text-ink/45">
          {product.brand.name}
        </p>
        <div className="mt-2 flex flex-wrap items-start justify-between gap-4">
          <h2 className="font-display text-3xl font-medium tracking-[-0.045em]">
            {product.name}
          </h2>
          <PriceDisplay
            amountMinor={product.price.amount_minor}
            currency={product.price.currency}
            size="lg"
          />
        </div>

        <ul className="mt-6 grid gap-2 text-sm leading-6 sm:grid-cols-2">
          {item.reasons.slice(0, 4).map((reason) => (
            <li className="border-l border-moss pl-3" key={reason.code}>
              {reason.message}
            </li>
          ))}
        </ul>

        <div className="mt-6 grid gap-4 border-y border-ink/10 py-5 text-sm sm:grid-cols-3">
          <Spec
            label={product.key_specification.label}
            value={product.key_specification.value}
          />
          <Spec label="Goal fit" value={`${item.breakdown.goal_match}/100`} />
          <Spec label="Space fit" value={`${item.breakdown.space_match}/100`} />
        </div>

        {alternative && (
          <div className="mt-5 bg-paper p-4 text-sm">
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

        <div className="mt-6 flex flex-wrap gap-2">
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
      <p className="text-[0.625rem] font-bold uppercase tracking-[0.12em] text-ink/40">
        {label}
      </p>
      <p className="mt-1 font-semibold">{value}</p>
    </div>
  )
}
