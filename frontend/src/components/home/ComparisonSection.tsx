import { Check } from 'lucide-react'

import { Badge } from '../ui/Badge'
import { Container } from '../ui/Container'
import { Heading } from '../ui/Heading'
import { PriceDisplay } from '../ui/PriceDisplay'
import { Section } from '../ui/Section'
import { comparisonProducts } from './homeData'

export function ComparisonSection() {
  return (
    <Section
      className="scroll-mt-20"
      id="comparison"
      space="lg"
      surface="paper"
    >
      <Container>
        <div className="max-w-3xl">
          <p className="eyebrow">Product comparison</p>
          <Heading className="mt-5" level={2} size="section">
            Better depends on who is buying.
          </Heading>
          <p className="mt-7 max-w-2xl text-base leading-7 text-ink/65">
            For the example agency, deeper integrations are useful—but they do
            not automatically outweigh budget, ease of use, or the fact that
            nobody has time to configure it.
          </p>
        </div>

        <div
          aria-label="Mobile product comparison"
          className="mt-14 border-l border-t border-ink/15 bg-surface md:hidden"
        >
          {comparisonProducts.map((product) => (
            <article
              className={
                product.recommended
                  ? 'border-b border-r border-ink/15 bg-moss-soft/55 p-5'
                  : 'border-b border-r border-ink/15 p-5'
              }
              key={product.name}
            >
              <div className="flex items-start justify-between gap-4">
                <h3 className="font-display text-xl font-medium tracking-[-0.035em]">
                  {product.shortName}
                </h3>
                {product.recommended && (
                  <Badge className="shrink-0" variant="success">
                    Best fit
                  </Badge>
                )}
              </div>
              <PriceDisplay
                amountMinor={product.priceMinor}
                className="mt-4"
                currency="USD"
                size="md"
              />
              <dl className="mt-5 grid grid-cols-2 gap-3 border-t border-ink/10 pt-4">
                <div>
                  <dt className="text-[0.5625rem] font-bold uppercase tracking-[0.12em] text-ink/65">
                    Ease of use
                  </dt>
                  <dd className="mt-1 text-xs leading-5">
                    {product.easeScore}/100
                  </dd>
                </div>
                <div>
                  <dt className="text-[0.5625rem] font-bold uppercase tracking-[0.12em] text-ink/65">
                    Integrations
                  </dt>
                  <dd className="mt-1 text-xs leading-5">
                    {product.integrationScore}/100
                  </dd>
                </div>
              </dl>
              <p className="mt-4 text-xs leading-5 text-ink/70">
                {product.verdict}
              </p>
            </article>
          ))}
        </div>

        <div className="mt-14 hidden overflow-x-auto border border-ink/15 bg-surface md:block lg:mt-20">
          {/* Five columns rather than six: the width that used to overflow a
              768px viewport now fits inside one, and the scroller stays for
              anything narrower that still reaches this branch. */}
          <table className="w-full min-w-[640px] border-collapse text-left">
            <caption className="sr-only">
              Comparison of three business tools
            </caption>
            <thead>
              <tr className="border-b border-ink/15 text-[0.625rem] font-bold uppercase tracking-[0.14em] text-ink/65">
                <th className="p-5 font-bold sm:p-6" scope="col">
                  Product
                </th>
                <th className="p-5 font-bold sm:p-6" scope="col">
                  Reference price
                </th>
                <th className="p-5 font-bold sm:p-6" scope="col">
                  Ease of use
                </th>
                <th className="p-5 font-bold sm:p-6" scope="col">
                  Integrations
                </th>
                <th className="p-5 font-bold sm:p-6" scope="col">
                  Decision
                </th>
              </tr>
            </thead>
            <tbody>
              {comparisonProducts.map((product) => (
                <tr
                  className={
                    product.recommended
                      ? 'border-b border-ink/15 bg-moss-soft/55 last:border-b-0'
                      : 'border-b border-ink/15 last:border-b-0'
                  }
                  key={product.name}
                >
                  <th className="p-5 sm:p-6" scope="row">
                    <div className="flex items-center gap-3">
                      {product.recommended && (
                        <Check
                          aria-hidden="true"
                          className="shrink-0 text-moss"
                          size={17}
                        />
                      )}
                      <span className="font-display text-lg font-medium tracking-[-0.025em]">
                        {product.shortName}
                      </span>
                    </div>
                  </th>
                  <td className="p-5 sm:p-6">
                    <PriceDisplay
                      amountMinor={product.priceMinor}
                      currency="USD"
                      size="sm"
                    />
                  </td>
                  <td className="p-5 text-sm sm:p-6">
                    {product.easeScore}/100
                  </td>
                  <td className="p-5 text-sm sm:p-6">
                    {product.integrationScore}/100
                  </td>
                  <td className="p-5 sm:p-6">
                    {product.recommended ? (
                      <Badge variant="success">Best fit</Badge>
                    ) : (
                      <span className="text-sm text-ink/70">
                        {product.verdict}
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {/* The basis belongs here rather than in each row: it is the same for
            all three, and the catalog is what it has to agree with. Checked
            against the catalog on 2026-09-02 — every price below is per user,
            per month, on monthly billing. */}
        <p className="mt-4 text-xs leading-5 text-ink/68">
          Prices are the entry paid tier for one user at monthly billing, read
          from each vendor&apos;s own pricing page and recorded with the date.
          Suitability scores are our editorial assessment, not vendor claims.
        </p>
      </Container>
    </Section>
  )
}
