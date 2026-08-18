import { CircleHelp } from 'lucide-react'
import { useState } from 'react'

import { Button } from '../../components/ui/Button'
import { Drawer } from '../../components/ui/Drawer'
import { EmptyState } from '../../components/ui/EmptyState'
import { ErrorState } from '../../components/ui/ErrorState'
import { LoadingState } from '../../components/ui/LoadingState'
import { Modal } from '../../components/ui/Modal'
import { Skeleton } from '../../components/ui/Skeleton'
import { Tabs } from '../../components/ui/Tabs'
import { Tooltip } from '../../components/ui/Tooltip'
import { useToast } from '../../components/ui/useToast'
import { ShowcaseBlock } from './ShowcaseBlock'

export function InteractionShowcase() {
  const [modalOpen, setModalOpen] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const { showToast } = useToast()

  return (
    <>
      <ShowcaseBlock
        description="Overlays use native dialog semantics, keyboard dismissal, focus containment, and controlled state."
        eyebrow="06 / Interaction"
        title="Overlays and disclosure"
      >
        <div className="flex flex-wrap items-center gap-3">
          <Button onClick={() => setModalOpen(true)}>Open modal</Button>
          <Button onClick={() => setDrawerOpen(true)} variant="secondary">
            Open drawer
          </Button>
          <Tooltip content="Recommendation scores never include affiliate commission.">
            <button
              aria-label="About objective scoring"
              className="grid size-11 place-items-center border border-ink/20 bg-surface"
              type="button"
            >
              <CircleHelp aria-hidden="true" size={18} />
            </button>
          </Tooltip>
          <Button
            onClick={() =>
              showToast({
                title: 'Preference saved',
                description: 'Your demo selection was updated locally.',
                variant: 'success',
              })
            }
            variant="quiet"
          >
            Show toast
          </Button>
        </div>

        <Tabs
          ariaLabel="Recommendation detail preview"
          className="mt-10"
          items={[
            {
              id: 'reason',
              label: 'Why it fits',
              content: (
                <p className="py-6 text-sm leading-6 text-ink/65">
                  Structured reasons explain goal, budget, space, and experience
                  fit independently.
                </p>
              ),
            },
            {
              id: 'tradeoffs',
              label: 'Trade-offs',
              content: (
                <p className="py-6 text-sm leading-6 text-ink/65">
                  Every decision can surface a limitation without hiding it
                  behind a single score.
                </p>
              ),
            },
            {
              id: 'future',
              label: 'Upgrade path',
              content: (
                <p className="py-6 text-sm leading-6 text-ink/65">
                  Future upgrades remain separate from what the user needs
                  today.
                </p>
              ),
            },
          ]}
        />

        <Modal
          description="This preview demonstrates hierarchy, focus behavior, and responsive actions."
          footer={
            <>
              <Button onClick={() => setModalOpen(false)} variant="quiet">
                Cancel
              </Button>
              <Button onClick={() => setModalOpen(false)}>
                Confirm choice
              </Button>
            </>
          }
          onOpenChange={setModalOpen}
          open={modalOpen}
          title="Confirm tool choice"
        >
          <p className="text-sm leading-6 text-ink/65">
            Product decisions should remain understandable before the user
            commits to an action.
          </p>
        </Modal>

        <Drawer
          footer={
            <Button fullWidth onClick={() => setDrawerOpen(false)}>
              Apply filters
            </Button>
          }
          onOpenChange={setDrawerOpen}
          open={drawerOpen}
          title="Refine selection"
        >
          <div className="space-y-6 text-sm leading-6 text-ink/65">
            <p>
              Drawer content remains usable at 320px without horizontal scroll.
            </p>
            <p>Escape closes the drawer and focus returns to its trigger.</p>
          </div>
        </Drawer>
      </ShowcaseBlock>

      <ShowcaseBlock
        description="Every asynchronous surface has a calm, explicit state rather than a blank panel or ambiguous spinner."
        eyebrow="07 / Feedback"
        title="Loading and outcomes"
      >
        <div className="grid gap-4 xl:grid-cols-3">
          <LoadingState compact title="Comparing tools" />
          <ErrorState
            compact
            onRetry={() =>
              showToast({
                title: 'Retry started',
                description: 'This is a showcase interaction only.',
              })
            }
            title="Results unavailable"
          />
          <EmptyState
            compact
            title="No saved setups"
            description="Saved stack plans will appear here."
          />
        </div>
        <div
          aria-label="Loading product card preview"
          className="mt-8 grid gap-5 border border-ink/15 bg-surface p-5 sm:grid-cols-[10rem_1fr]"
          role="status"
        >
          <Skeleton className="aspect-square min-h-0" />
          <div className="space-y-4 py-2">
            <Skeleton className="w-1/3" shape="line" />
            <Skeleton className="h-7 w-4/5" shape="line" />
            <Skeleton className="w-1/2" shape="line" />
          </div>
          <span className="sr-only">Loading product</span>
        </div>
      </ShowcaseBlock>
    </>
  )
}
