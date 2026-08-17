import { Search } from 'lucide-react'
import { useState } from 'react'

import { Button } from '../../components/ui/Button'
import { Checkbox } from '../../components/ui/Checkbox'
import { Input } from '../../components/ui/Input'
import { Radio } from '../../components/ui/Radio'
import { Select } from '../../components/ui/Select'
import { Slider } from '../../components/ui/Slider'
import { Textarea } from '../../components/ui/Textarea'
import { ShowcaseBlock } from './ShowcaseBlock'

export function ControlsShowcase() {
  const [budget, setBudget] = useState(700)

  return (
    <>
      <ShowcaseBlock
        description="Native controls retain keyboard, pointer, and assistive-technology behavior while sharing one precise visual language."
        eyebrow="04 / Controls"
        title="Fields and choices"
      >
        <div className="grid gap-6 md:grid-cols-2">
          <Input
            leadingIcon={<Search size={17} />}
            label="Search equipment"
            placeholder="Adjustable dumbbells"
            required
          />
          <Select
            hint="Used to prioritize the right capabilities."
            label="Primary goal"
            required
            defaultValue=""
          >
            <option disabled value="">
              Select a goal
            </option>
            <option value="muscle">Build muscle</option>
            <option value="strength">Build strength</option>
            <option value="fitness">General fitness</option>
          </Select>
          <Input
            error="Enter a budget greater than zero."
            label="Budget"
            required
            type="number"
            value="0"
            readOnly
          />
          <Textarea
            hint="Do not include sensitive personal information."
            label="Space notes"
            placeholder="Shared living room; equipment must store against a wall."
          />
        </div>
        <div className="mt-8 grid gap-8 border-t border-ink/15 pt-8 md:grid-cols-2">
          <div className="space-y-4">
            <Checkbox
              defaultChecked
              description="Exclude equipment that duplicates this capability."
              label="I already own a pull-up bar"
            />
            <Checkbox
              description="Prioritize quiet operation and floor protection."
              label="Apartment living"
            />
          </div>
          <fieldset className="space-y-4">
            <legend className="mb-4 text-label font-bold uppercase tracking-[0.12em]">
              Experience
            </legend>
            <Radio defaultChecked label="Beginner" name="experience-preview" />
            <Radio label="Intermediate" name="experience-preview" />
            <Radio label="Advanced" name="experience-preview" />
          </fieldset>
        </div>
        <div className="mt-8 border-t border-ink/15 pt-8">
          <Slider
            formatValue={(value) => `$${value}`}
            hint="Keyboard arrow keys adjust in $50 increments."
            label="Setup budget"
            max={2000}
            min={200}
            onChange={(event) => setBudget(Number(event.currentTarget.value))}
            step={50}
            value={budget}
          />
        </div>
      </ShowcaseBlock>

      <ShowcaseBlock
        description="Variants communicate hierarchy without relying on oversized shadows, gradients, or novelty motion."
        eyebrow="05 / Controls"
        title="Button hierarchy"
      >
        <div className="flex flex-wrap items-center gap-3">
          <Button>Build my setup</Button>
          <Button variant="secondary">Compare options</Button>
          <Button variant="quiet">Learn more</Button>
          <Button variant="danger">Remove item</Button>
          <Button loading loadingLabel="Calculating…">
            Calculate
          </Button>
          <Button disabled>Unavailable</Button>
        </div>
        <div className="mt-6 flex flex-wrap items-center gap-3">
          <Button size="sm">Small</Button>
          <Button size="md">Medium</Button>
          <Button size="lg">Large</Button>
        </div>
      </ShowcaseBlock>
    </>
  )
}
