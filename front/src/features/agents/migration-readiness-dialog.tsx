import { useMemo, useState } from 'react'
import { toast } from 'sonner'
import {
  MigrateMode,
  useMigrateAgentsV2,
  useMigrationReadiness,
  type MigrateSummary,
} from '@/api/agents'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import {
  CheckCircle2,
  CircleDashed,
  Fingerprint,
  ListChecks,
  Play,
  Rocket,
  TriangleAlert,
} from 'lucide-react'
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

// Result actions returned by MigrateAgentsV2 (dry-run / apply / verify). Ok,
// expanded, and already_independent are healthy; the rest indicate an agent
// that was skipped or blocked.
const ACTION_META: Record<string, { label: string; className: string }> = {
  ok: {
    label: 'OK',
    className: 'border-transparent bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  },
  expanded: {
    label: 'Expanded',
    className: 'border-transparent bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  },
  already_independent: {
    label: 'Already independent',
    className: 'border-transparent bg-muted text-muted-foreground',
  },
  skipped: {
    label: 'Skipped',
    className: 'border-transparent bg-muted text-muted-foreground',
  },
  missing_id: {
    label: 'Missing ID',
    className: 'border-transparent bg-amber-500/15 text-amber-700 dark:text-amber-400',
  },
  migration_required: {
    label: 'Migration required',
    className: 'border-transparent bg-orange-500/15 text-orange-700 dark:text-orange-400',
  },
  error: {
    label: 'Error',
    className: 'border-transparent bg-destructive/15 text-destructive',
  },
}

function actionMeta(action: string) {
  return (
    ACTION_META[action] ?? {
      label: action || 'Unknown',
      className: 'border-transparent bg-muted text-muted-foreground',
    }
  )
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
  const migrate = useMigrateAgentsV2()
  const statuses = useMemo(() => data?.statuses ?? [], [data?.statuses])

  // Last run result (dry-run or apply), shown below the readiness list.
  const [result, setResult] = useState<MigrateSummary | null>(null)
  const [confirmApply, setConfirmApply] = useState(false)

  const readyCount = useMemo(
    () => statuses.filter((s) => readinessOf(s) === 'MIGRATION_READINESS_READY').length,
    [statuses],
  )
  const allReady = statuses.length > 0 && readyCount === statuses.length
  const hasReady = readyCount > 0

  function run(mode: MigrateMode) {
    migrate.mutate(mode, {
      onSuccess: (summary) => {
        setResult(summary)
        if (mode === MigrateMode.APPLY) {
          toast.success(
            summary.migrated > 0
              ? `Migrated ${summary.migrated} agent${summary.migrated === 1 ? '' : 's'}`
              : 'Migration ran — no agents changed',
          )
        }
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <ListChecks className='size-4' /> Agent ID Migration
          </DialogTitle>
          <DialogDescription>
            Every agent needs an immutable Agent ID before the identity migration. Assign IDs to
            referenced agents first — parents and workflows report incomplete dependencies until
            their targets have IDs. Then run the migration to expand legacy trees into independent
            records and clear the <span className='font-medium'>Legacy</span> badge.
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

            <div className='max-h-[40vh] divide-y divide-border overflow-y-auto rounded-md border border-border'>
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

            {result && (
              <div className='space-y-2'>
                <div className='flex items-center justify-between text-xs text-muted-foreground'>
                  <span className='font-medium text-foreground'>
                    {result.mode === MigrateMode.APPLY ? 'Migration result' : 'Dry run result'}
                  </span>
                  <span>
                    {result.migrated} migrated · {result.skipped} skipped · {result.errors} errors
                  </span>
                </div>
                <div className='max-h-[25vh] divide-y divide-border overflow-y-auto rounded-md border border-border'>
                  {result.results.map((r) => {
                    const meta = actionMeta(r.action)
                    return (
                      <div key={r.name} className='flex items-center gap-3 px-3 py-2'>
                        <div className='min-w-0 flex-1'>
                          <span className='truncate text-sm'>{r.name}</span>
                          {r.detail && (
                            <p className='mt-0.5 truncate text-xs text-muted-foreground'>{r.detail}</p>
                          )}
                        </div>
                        <Badge variant='outline' className={meta.className}>
                          {meta.label}
                        </Badge>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}
          </>
        )}

        {statuses.length > 0 && (
          <DialogFooter className='gap-2 sm:justify-between'>
            <Button
              variant='outline'
              onClick={() => run(MigrateMode.DRY_RUN)}
              disabled={migrate.isPending}
            >
              <Play className='size-4' />
              Dry run
            </Button>
            <Button
              onClick={() => setConfirmApply(true)}
              disabled={migrate.isPending || !hasReady}
              title={hasReady ? undefined : 'No agents are ready to migrate yet'}
            >
              <Rocket className='size-4' />
              {migrate.isPending && migrate.variables === MigrateMode.APPLY
                ? 'Migrating...'
                : 'Migrate now'}
            </Button>
          </DialogFooter>
        )}

        <AlertDialog open={confirmApply} onOpenChange={setConfirmApply}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Run the identity migration?</AlertDialogTitle>
              <AlertDialogDescription>
                This expands eligible legacy agents into independent records and promotes them to
                Active. Ready agents ({readyCount} / {statuses.length}) are migrated; agents missing
                an ID or with incomplete dependencies are skipped and remain unchanged. Legacy names
                are retained during the observation period.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={() => run(MigrateMode.APPLY)}>Migrate now</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </DialogContent>
    </Dialog>
  )
}
