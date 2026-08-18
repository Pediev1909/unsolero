import { Button } from '../../../components/ui/Button'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Skeleton } from '../../../components/ui/Skeleton'
import {
  useRevokeOtherSessions,
  useRevokeSession,
  useSecuritySessions,
} from '../securityQueries'

const dateTime = new Intl.DateTimeFormat('en-US', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function SessionSecurity() {
  const sessions = useSecuritySessions()
  const revoke = useRevokeSession()
  const revokeOthers = useRevokeOtherSessions()
  return (
    <section aria-labelledby="sessions-heading">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h3 className="text-xl font-semibold" id="sessions-heading">
            Active sessions
          </h3>
          <p className="mt-2 text-sm text-ink/60">
            Only stable session IDs and safe timestamps are shown.
          </p>
        </div>
        <Button
          disabled={!sessions.data?.some((session) => !session.current)}
          loading={revokeOthers.isPending}
          onClick={() => revokeOthers.mutate()}
          size="sm"
          variant="secondary"
        >
          Revoke all others
        </Button>
      </div>
      {sessions.isPending && (
        <div className="mt-5 space-y-2">
          <Skeleton className="h-20" />
          <Skeleton className="h-20" />
        </div>
      )}
      {sessions.isError && (
        <div className="mt-5">
          <ErrorState
            compact
            onRetry={() => void sessions.refetch()}
            title="Sessions unavailable"
          />
        </div>
      )}
      {sessions.data && (
        <div className="mt-5 divide-y divide-ink/10 border-y border-ink/10">
          {sessions.data.map((session) => (
            <div
              className="flex flex-wrap items-center justify-between gap-4 py-4"
              key={session.id}
            >
              <div>
                <p className="font-semibold">
                  {session.current ? 'Current session' : 'Signed-in session'}
                </p>
                <p className="mt-1 text-xs text-ink/55">
                  Created {dateTime.format(new Date(session.created_at))} · Last
                  used {dateTime.format(new Date(session.last_seen_at))}
                </p>
                <p className="mt-1 font-mono text-[11px] text-ink/45">
                  {session.id}
                </p>
              </div>
              {!session.current && (
                <Button
                  disabled={revoke.isPending}
                  onClick={() => revoke.mutate(session.id)}
                  size="sm"
                  variant="quiet"
                >
                  Revoke
                </Button>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
