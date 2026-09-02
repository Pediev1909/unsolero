import { Minus, Plus } from 'lucide-react'
import { useId, useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '../../../components/ui/Button'
import { PriceDisplay } from '../../../components/ui/PriceDisplay'
import { formatMinorCurrency } from '../../../lib/money/format'
import type { ProductSummary } from '../schemas'
import {
  clampTeamSize,
  hasPerSeatPricing,
  seatCostLines,
  teamSizeRange,
} from '../seatCost'

interface SeatCostCalculatorProps {
  products: ProductSummary[]
}

const defaultTeamSize = 5

/**
 * What the listed products cost a team of a chosen size, per month, from the
 * reference prices on the page. It appears only where at least one product is
 * priced per seat; a category of flat-priced tools has nothing to multiply.
 */
export function SeatCostCalculator({ products }: SeatCostCalculatorProps) {
  const [teamSize, setTeamSize] = useState(defaultTeamSize)
  const id = useId()
  if (!hasPerSeatPricing(products)) return null
  const lines = seatCostLines(products, teamSize)

  return (
    <section
      aria-labelledby={`${id}-heading`}
      className="mt-10 border border-ink/15 bg-surface p-4 sm:p-6"
    >
      <div className="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.14em] text-bronze-dark">
            Cost for your team
          </p>
          <h2
            className="mt-2 font-display text-2xl leading-tight font-medium tracking-[-0.035em]"
            id={`${id}-heading`}
          >
            Monthly total at your team size
          </h2>
          <p className="mt-2 max-w-prose text-sm leading-6 text-ink/70">
            Per-seat prices are multiplied by the team size. Flat prices cover
            the whole team and are not.
          </p>
        </div>
        <div>
          <label
            className="text-label mb-2 block font-bold uppercase tracking-[0.12em]"
            htmlFor={`${id}-seats`}
          >
            Team size
          </label>
          <div className="flex items-center gap-2">
            <Button
              aria-label="Fewer seats"
              disabled={teamSize <= teamSizeRange.min}
              onClick={() => setTeamSize((size) => clampTeamSize(size - 1))}
              size="sm"
              variant="secondary"
            >
              <Minus aria-hidden="true" size={16} />
            </Button>
            <input
              className="control-base w-20 text-center tabular-nums"
              id={`${id}-seats`}
              inputMode="numeric"
              max={teamSizeRange.max}
              min={teamSizeRange.min}
              onChange={(event) => {
                const next = event.target.valueAsNumber
                if (Number.isFinite(next)) setTeamSize(clampTeamSize(next))
              }}
              type="number"
              value={teamSize}
            />
            <Button
              aria-label="More seats"
              disabled={teamSize >= teamSizeRange.max}
              onClick={() => setTeamSize((size) => clampTeamSize(size + 1))}
              size="sm"
              variant="secondary"
            >
              <Plus aria-hidden="true" size={16} />
            </Button>
          </div>
        </div>
      </div>

      <div className="mt-6 overflow-x-auto">
        <table className="w-full min-w-[28rem] border-t border-ink/15 text-sm">
          <thead>
            <tr className="text-left text-[0.625rem] font-bold uppercase tracking-[0.12em] text-ink/65">
              <th className="py-3 pr-4 font-bold" scope="col">
                Product
              </th>
              <th className="py-3 pr-4 font-bold" scope="col">
                Billing
              </th>
              <th className="py-3 text-right font-bold" scope="col">
                Per month, {teamSize} {teamSize === 1 ? 'seat' : 'seats'}
              </th>
            </tr>
          </thead>
          <tbody>
            {lines.map((line) => (
              <tr className="border-t border-ink/10" key={line.product.id}>
                <th className="py-3 pr-4 font-semibold" scope="row">
                  <Link
                    className="hover:text-bronze-dark"
                    to={`/products/${line.product.slug}`}
                  >
                    {line.product.name}
                  </Link>
                </th>
                <td className="py-3 pr-4 text-ink/70">
                  {/* Zero is a real price the PriceDisplay in the next column
                      already words carefully; "$0.00 flat" beside it would
                      undo that. */}
                  {line.unitMinor === 0
                    ? 'No monthly fee'
                    : `${formatMinorCurrency(line.unitMinor, line.currency)} ${line.basis === 'per_seat' ? 'per seat' : 'flat'}`}
                </td>
                <td className="py-3 text-right">
                  <PriceDisplay
                    amountMinor={line.totalMinor}
                    currency={line.currency}
                    size="sm"
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="mt-4 text-xs leading-5 text-ink/65">
        Prices read from vendor pages; see each product page for the date.
      </p>
    </section>
  )
}
