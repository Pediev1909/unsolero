import { useCurrentUser } from '../auth/queries'
import { useDeleteSetup, useRenameSetup, useSetups } from './queries'
import { useLocalSetups } from './localSetups'

export function useSavedSetups() {
  const account = useCurrentUser()
  const authenticated = Boolean(account.data)
  const server = useSetups(authenticated)
  const local = useLocalSetups()
  const renameMutation = useRenameSetup()
  const deleteMutation = useDeleteSetup()
  const setups = !account.isSuccess
    ? []
    : authenticated
      ? (server.data?.setups ?? [])
      : local.setups.map((setup) => ({
          id: setup.id,
          name: setup.name,
          item_count: setup.result.recommended_products.length,
          total_cost: setup.result.total_cost,
          recommendation_score: setup.result.recommendation_score,
          created_at: setup.created_at,
          updated_at: setup.updated_at,
        }))
  return {
    authenticated,
    setups,
    local,
    isPending: account.isPending || (authenticated && server.isPending),
    isError: account.isError || (authenticated && server.isError),
    mutationPending: renameMutation.isPending || deleteMutation.isPending,
    refetch: server.refetch,
    async rename(id: string, name: string) {
      if (authenticated) await renameMutation.mutateAsync({ setupID: id, name })
      else local.rename(id, name)
    },
    async remove(id: string) {
      if (authenticated) await deleteMutation.mutateAsync(id)
      else local.remove(id)
    },
  }
}
