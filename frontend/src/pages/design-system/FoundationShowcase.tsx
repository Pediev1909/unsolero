import { Badge } from '../../components/ui/Badge'
import { Card } from '../../components/ui/Card'
import { Heading } from '../../components/ui/Heading'
import { PriceDisplay } from '../../components/ui/PriceDisplay'
import { Rating } from '../../components/ui/Rating'
import { ShowcaseBlock } from './ShowcaseBlock'

const swatches = [
  { name: 'Canvas', className: 'bg-canvas', value: '#F3F0E9' },
  { name: 'Surface', className: 'bg-surface', value: '#FCFAF5' },
  { name: 'Paper', className: 'bg-paper', value: '#E8E3D9' },
  { name: 'Ink', className: 'bg-ink text-canvas', value: '#171816' },
  { name: 'Bronze', className: 'bg-bronze text-surface', value: '#8C5D35' },
  { name: 'Moss', className: 'bg-moss text-surface', value: '#48604E' },
  { name: 'Ember', className: 'bg-ember text-surface', value: '#8A3F35' },
]

export function FoundationShowcase() {
  return (
    <>
      <ShowcaseBlock
        description="Warm neutrals create an editorial foundation. Bronze signals considered action; moss and ember are reserved for status."
        eyebrow="01 / Foundation"
        title="Color and surface"
      >
        <div className="grid grid-cols-2 gap-px overflow-hidden border border-ink/15 bg-ink/15 sm:grid-cols-3 xl:grid-cols-4">
          {swatches.map((swatch) => (
            <div className={swatch.className} key={swatch.name}>
              <div className="flex aspect-[4/3] flex-col justify-end p-4">
                <p className="text-xs font-bold uppercase tracking-[0.12em]">
                  {swatch.name}
                </p>
                <p className="mt-1 font-mono text-[0.625rem] opacity-60">
                  {swatch.value}
                </p>
              </div>
            </div>
          ))}
        </div>
      </ShowcaseBlock>

      <ShowcaseBlock
        description="Tight display tracking creates confidence; body copy remains calm and readable. Editorial serif is an accent, not the default."
        eyebrow="02 / Foundation"
        title="Typography"
      >
        <div className="space-y-10">
          <div>
            <p className="eyebrow">Display / 92% line height</p>
            <Heading className="mt-4" level={3} size="display">
              Strength, considered.
            </Heading>
          </div>
          <div className="grid gap-8 border-t border-ink/15 pt-8 md:grid-cols-2">
            <p className="text-body-lg text-ink/75">
              Equipment should earn its footprint through utility, quality, and
              long-term fit.
            </p>
            <blockquote className="font-editorial text-2xl leading-8 text-ink/75">
              “The quietest luxury is knowing exactly what belongs.”
            </blockquote>
          </div>
        </div>
      </ShowcaseBlock>

      <ShowcaseBlock
        description="Corners stay precise, shadows remain rare, and spacing carries more hierarchy than decoration."
        eyebrow="03 / Foundation"
        title="Shape and depth"
      >
        <div className="grid gap-4 sm:grid-cols-3">
          <Card variant="outlined">
            <p className="eyebrow">Outlined</p>
            <p className="mt-4 text-sm leading-6 text-ink/70">
              Default grouping with a quiet one-pixel boundary.
            </p>
          </Card>
          <Card variant="raised">
            <p className="eyebrow">Raised</p>
            <p className="mt-4 text-sm leading-6 text-ink/70">
              Reserved for objects that need temporary visual priority.
            </p>
          </Card>
          <Card variant="dark">
            <p className="text-caption font-bold uppercase tracking-[0.18em] text-canvas/75">
              Dark
            </p>
            <p className="mt-4 text-sm leading-6 text-canvas/78">
              Used for trust statements and decisive contrast.
            </p>
          </Card>
        </div>
        <div className="mt-8 flex flex-wrap items-center gap-3">
          <Badge>Neutral</Badge>
          <Badge variant="accent">Recommended</Badge>
          <Badge variant="success">Compatible</Badge>
          <Badge variant="warning">Trade-off</Badge>
          <Badge variant="error">Unavailable</Badge>
          <Badge variant="sponsored">Sponsored</Badge>
        </div>
        <div className="mt-8 flex flex-wrap items-center gap-8 border-t border-ink/15 pt-8">
          <PriceDisplay amountMinor={32900} currency="USD" size="lg" />
          <div>
            <Rating label="Illustrative component value" value={4.2} />
            <p className="mt-1 text-[0.625rem] text-ink/68">
              Illustrative value only—not a product review.
            </p>
          </div>
        </div>
      </ShowcaseBlock>
    </>
  )
}
