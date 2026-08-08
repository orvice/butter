import { useMemo } from 'react'
import { useMigrationReadiness } from '@/api/agents'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { CheckCircle2, CircleDashed, Fingerprint, ListChecks, TriangleAlert } from 'lucide-react'
import type { AgentMigrationStatus, MigrationReadiness } from '@/types/api'

const READINESS_META: Record<
  MigrationReadiness,
  { label: string; className: string }
> = {
  MIGRATION_READINESS_READY: {
    label: 'Ready',
    className: 'border-transparent bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  },
  MIGRATION_READINESS_MISSING_ID: {
    label: 'Missing ID',
    className: 'border-transparent bg-amber-500/15 text-amber-700 dark:text-amber-400',
  },
  MIGRATION_READINESS_CONFLICT: {
    label: 'Conflict',
    className: 'border-transparent bg-destructive/15 text-destructive',
  },
  MIGRATION_READINESS_INCOMPLETE_DEPS: {
    label: 'Incomplete deps',
    className: 'border-transparent bg-orange-500/15 text-orange-700 dark:text-orange-400',
  },
  MIGRATION_READINESS_UNSPECIFIED: {
    label: 'Unknown',
    className: 'border-transparent bg-muted text-muted-foreground',
  },
}

function readinessOf(s: AgentMigrationStatus): MigrationReadiness {
  return s.readiness ?? 'MIGRATION_READINESS_UNSPECIFIED'
}

export function MigrationReadinessDialog({
  open,
  onOpenChange,
  onAssign,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onAssign: (agentName: string) => void
}) {
  const { data, isLoading } = useMigrationReadiness(open)
  const statuses = useMemo(() => data?.statuses ?? [], [data?.statuses])

  const readyCount = useMemo(
    () => statuses.filter((s) => readinessOf(s) === 'MIGRATION_READINESS_READY').length,
    [statuses],
  )
  const allReady = statuses.length > 0 && readyCount === statuses.length

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <ListChecks className='size-4' /> Agent ID Migration Readiness
          </DialogTitle>
          <DialogDescription>
            Every agent needs an immutable Agent ID before the identity migration. Assign IDs to
            referenced agents first — parents and workflows report incomplete dependencies until
            their targets have IDs.
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className='space-y-2'>
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className='h-12' />
            ))}
          </div>
        ) : statuses.length === 0 ? (
          <p className='py-8 text-center text-sm text-muted-foreground'>
            No agents in this workspace.
          </p>
        ) : (
          <>
            <div
              className={
                allReady
                  ? 'flex items-center gap-2 rounded-md border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-400'
                  : 'flex items-center gap-2 rounded-md border border-border bg-muted/40 p-3 text-sm text-muted-foreground'
              }
            >
              {allReady ? (
                <CheckCircle2 className='size-4 shrink-0' />
              ) : (
                <CircleDashed className='size-4 shrink-0' />
              )}
              <span>
                {readyCount} / {statuses.length} agents ready
                {allReady && ' — this workspace is ready for the identity migration.'}
              </span>
            </div>

            <div className='max-h-[50vh] divide-y divide-border overflow-y-auto rounded-md border border-border'>
              {statuses.map((s) => {
                const readiness = readinessOf(s)
                const meta = READINESS_META[readiness]
                return (
                  <div key={s.name} className='flex items-center gap-3 px-3 py-2.5'>
                    <div className='min-w-0 flex-1'>
                      <div className='flex items-center gap-2'>
                        <span className='truncate text-sm font-medium'>{s.name}</span>
                        {s.agent_id ? (
                          <span className='rounded border border-border bg-muted/50 px-1.5 py-0.5 font-mono text-[0.7rem] text-muted-foreground'>
                            {s.agent_id}
                          </span>
                        ) : null}
                      </div>
                      {s.detail && readiness !== 'MIGRATION_READINESS_READY' && (
                        <p className='mt-0.5 flex items-center gap-1 text-xs text-muted-foreground'>
                          <TriangleAlert className='size-3 shrink-0' />
                          <span className='truncate'>{s.detail}</span>
                        </p>
                      )}
                    </div>
                    <Badge variant='outline' className={meta.className}>
                      {meta.label}
                    </Badge>
                    {readiness === 'MIGRATION_READINESS_MISSING_ID' && (
                      <Button size='sm' variant='outline' onClick={() => onAssign(s.name)}>
                        <Fingerprint className='size-3.5' />
                        Assign ID
                      </Button>
                    )}
                  </div>
                )
              })}
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
