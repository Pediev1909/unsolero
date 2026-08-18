import { Pencil, Scale, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { Button } from '../components/ui/Button'
import { ButtonLink } from '../components/ui/ButtonLink'
import { Container } from '../components/ui/Container'
import { ErrorState } from '../components/ui/ErrorState'
import { Input } from '../components/ui/Input'
import { LoadingState } from '../components/ui/LoadingState'
import { Modal } from '../components/ui/Modal'
import { useComparisonSelection } from '../features/catalog/useProductCollections'
import { RecommendationResults } from '../features/recommendation/components/RecommendationResults'
import { useSetup } from '../features/recommendation/queries'
import { useSavedSetups } from '../features/recommendation/useSavedSetups'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function SetupPage() {
  const { setupID = '' } = useParams()
  const navigate = useNavigate()
  const saved = useSavedSetups()
  const serverSetup = useSetup(setupID, saved.authenticated)
  const localSetup =
    saved.isPending || saved.isError || saved.authenticated
      ? undefined
      : saved.local.get(setupID)
  const result = saved.authenticated ? serverSetup.data : localSetup?.result
  const currentName = saved.authenticated
    ? result?.setup_name
    : localSetup?.name
  const comparison = useComparisonSelection()
  const [renameOpen, setRenameOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [name, setName] = useState('')
  const [actionError, setActionError] = useState<string | null>(null)
  usePageMetadata({
    title: `${currentName ?? 'Saved Software Stack'} | UNSOLERO`,
    description: 'Review and manage a saved personalized software stack.',
    robots: 'noindex, follow',
  })
  const pending =
    saved.isPending || (saved.authenticated && serverSetup.isPending)
  const failed =
    !pending &&
    (saved.isError || (saved.authenticated && serverSetup.isError) || !result)

  async function rename() {
    setActionError(null)
    try {
      await saved.rename(setupID, name)
      setRenameOpen(false)
    } catch (error) {
      setActionError(
        error instanceof Error
          ? error.message
          : 'The setup could not be renamed.',
      )
    }
  }

  async function remove() {
    setActionError(null)
    try {
      await saved.remove(setupID)
      void navigate('/setups', { replace: true })
    } catch {
      setActionError('The setup could not be deleted. Please try again.')
    }
  }

  function compare() {
    if (!result) return
    comparison.replace(
      result.recommended_products.slice(0, 4).map((item) => item.product.id),
    )
    void navigate('/compare')
  }

  return (
    <>
      <SiteHeader position="static" />
      <main id="main-content">
        <Container>
          {pending && (
            <div className="py-24">
              <LoadingState
                description="Loading the product decisions and trade-offs in your saved brief."
                title="Loading your setup"
              />
            </div>
          )}
          {failed && (
            <div className="py-24">
              <ErrorState
                description="It may have been removed or belong to another account or browser."
                onRetry={() => void serverSetup.refetch()}
                title="Setup unavailable"
              />
              <div className="mt-5 text-center">
                <ButtonLink to="/setups" variant="secondary">
                  Back to saved setups
                </ButtonLink>
              </div>
            </div>
          )}
          {result && (
            <>
              <div className="mt-8 flex flex-wrap items-center justify-between gap-4 border-y border-ink/15 py-4">
                <div>
                  <p className="text-xs uppercase tracking-[0.13em] text-ink/45">
                    Saved setup
                  </p>
                  <p className="mt-1 font-semibold">
                    {currentName ?? 'Personalized Stack'}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    onClick={() => {
                      setName(currentName ?? 'Personalized Stack')
                      setActionError(null)
                      setRenameOpen(true)
                    }}
                    size="sm"
                    variant="secondary"
                  >
                    <Pencil aria-hidden="true" size={14} /> Rename
                  </Button>
                  <Button
                    onClick={() =>
                      void navigate('/build', {
                        state: { editInput: result.input },
                      })
                    }
                    size="sm"
                    variant="secondary"
                  >
                    <Pencil aria-hidden="true" size={14} /> Edit
                  </Button>
                  <Button onClick={compare} size="sm" variant="secondary">
                    <Scale aria-hidden="true" size={14} /> Compare
                  </Button>
                  <Button
                    onClick={() => {
                      setActionError(null)
                      setDeleteOpen(true)
                    }}
                    size="sm"
                    variant="quiet"
                  >
                    <Trash2 aria-hidden="true" size={14} /> Delete
                  </Button>
                </div>
              </div>
              <RecommendationResults
                persistedLocally={!saved.authenticated}
                result={result}
              />
            </>
          )}
        </Container>
      </main>
      <SiteFooter />
      <Modal
        footer={
          <>
            <Button onClick={() => setRenameOpen(false)} variant="quiet">
              Cancel
            </Button>
            <Button
              loading={saved.mutationPending}
              onClick={() => void rename()}
            >
              Save name
            </Button>
          </>
        }
        onOpenChange={setRenameOpen}
        open={renameOpen}
        title="Rename setup"
      >
        <Input
          autoFocus
          error={actionError}
          label="Setup name"
          maxLength={120}
          onChange={(event) => setName(event.target.value)}
          value={name}
        />
      </Modal>
      <Modal
        footer={
          <>
            <Button onClick={() => setDeleteOpen(false)} variant="quiet">
              Cancel
            </Button>
            <Button
              loading={saved.mutationPending}
              onClick={() => void remove()}
            >
              Delete setup
            </Button>
          </>
        }
        onOpenChange={setDeleteOpen}
        open={deleteOpen}
        title="Delete this setup?"
      >
        <p className="text-sm leading-6 text-ink/65">
          This removes the saved plan. Your product wishlist is not affected.
        </p>
        {actionError && (
          <p className="mt-4 text-sm text-ember" role="alert">
            {actionError}
          </p>
        )}
      </Modal>
    </>
  )
}
