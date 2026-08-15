import { useMemo, useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import type {
  Agent,
  AgentLifecycleStatus,
  AgentRuntimeStatus,
} from '@/types/api'
import {
  Bot,
  ArchiveRestore,
  CircleAlert,
  CircleCheck,
  CircleDashed,
  Fingerprint,
  ListTodo,
  LoaderCircle,
  MessageSquarePlus,
  MoreVertical,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Search,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  useAgents,
  useDeleteAgent,
  useRestoreAgent,
  useReloadAgents,
  useInvokeAgent,
  useAgentRuntimeStatuses,
} from '@/api/agents'
import { AGENT_TYPE_LABELS } from '@/lib/constants'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import {
  AgentAvatar,
  StatusBadge,
  type RunStatus,
} from '@/components/butter/primitives'
import { DeleteDialog } from '@/components/delete-dialog'
import { agentIconUrl } from './icon-utils'

type RuntimeState = 'running' | 'idle' | 'failed'
type RuntimeFilter = 'all' | RuntimeState

function runtimeStatusOf(rt?: AgentRuntimeStatus): RuntimeState {
  switch (rt?.state) {
    case 'AGENT_RUNTIME_STATE_RUNNING':
      return 'running'
    case 'AGENT_RUNTIME_STATE_FAILED':
      return 'failed'
    default:
      return 'idle'
  }
}

const RUNTIME_TO_BADGE: Record<
  RuntimeState,
  { status: RunStatus; label: string }
> = {
  running: { status: 'running', label: 'Running' },
  idle: { status: 'success', label: 'Available' },
  failed: { status: 'failed', label: 'Failed' },
}

const ACTIVE_LIFECYCLES = new Set<AgentLifecycleStatus | undefined>([
  undefined,
  'AGENT_LIFECYCLE_STATUS_UNSPECIFIED',
  'AGENT_LIFECYCLE_STATUS_ACTIVE',
])

const LIFECYCLE_BADGES: Record<
  AgentLifecycleStatus,
  { label: string; icon: typeof CircleCheck; className: string }
> = {
  AGENT_LIFECYCLE_STATUS_UNSPECIFIED: {
    label: 'Legacy',
    icon: CircleDashed,
    className: 'border-border bg-muted text-muted-foreground',
  },
  AGENT_LIFECYCLE_STATUS_ACTIVE: {
    label: 'Active',
    icon: CircleCheck,
    className:
      'border-emerald-500/35 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  },
  AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED: {
    label: 'Migration required',
    icon: CircleAlert,
    className:
      'border-amber-500/35 bg-amber-500/10 text-amber-700 dark:text-amber-300',
  },
  AGENT_LIFECYCLE_STATUS_PROVISIONING: {
    label: 'Provisioning',
    icon: LoaderCircle,
    className: 'border-sky-500/35 bg-sky-500/10 text-sky-700 dark:text-sky-300',
  },
  AGENT_LIFECYCLE_STATUS_ERROR: {
    label: 'Needs attention',
    icon: CircleAlert,
    className: 'border-destructive/35 bg-destructive/10 text-destructive',
  },
  AGENT_LIFECYCLE_STATUS_DELETING: {
    label: 'Deleting',
    icon: LoaderCircle,
    className: 'border-border bg-muted text-muted-foreground',
  },
  AGENT_LIFECYCLE_STATUS_DELETED: {
    label: 'Deleted',
    icon: CircleDashed,
    className: 'border-border bg-muted text-muted-foreground',
  },
}

function isRunnable(agent: Agent): boolean {
  return ACTIVE_LIFECYCLES.has(agent.lifecycle_status)
}

function LifecycleBadge({ agent }: { agent: Agent }) {
  const lifecycle =
    agent.lifecycle_status ?? 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED'
  const config = LIFECYCLE_BADGES[lifecycle]
  const Icon = config.icon
  const isTransitioning =
    lifecycle === 'AGENT_LIFECYCLE_STATUS_PROVISIONING' ||
    lifecycle === 'AGENT_LIFECYCLE_STATUS_DELETING'
  return (
    <Badge variant='outline' className={config.className}>
      <Icon
        className={cn('size-3', isTransitioning && 'motion-safe:animate-spin')}
      />
      {config.label}
    </Badge>
  )
}

function timeAgo(ts?: string): string | null {
  if (!ts) return null
  const d = Date.now() - new Date(ts).getTime()
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`
  return `${Math.floor(d / 86_400_000)}d ago`
}

function AgentCard({
  agent,
  runtime,
  onDelete,
  onRun,
  onRestore,
  restoring,
}: {
  agent: Agent
  runtime?: AgentRuntimeStatus
  onDelete: () => void
  onRun: () => void
  onRestore: () => void
  restoring: boolean
}) {
  const navigate = useNavigate()
  const status = runtimeStatusOf(runtime)
  const badge = RUNTIME_TO_BADGE[status]
  const lastRun = timeAgo(runtime?.last_run_at)
  const runnable = isRunnable(agent)
  const deleted = agent.lifecycle_status === 'AGENT_LIFECYCLE_STATUS_DELETED'
  const canDelete =
    runnable || agent.lifecycle_status === 'AGENT_LIFECYCLE_STATUS_ERROR'

  return (
    <div className='group relative flex flex-col rounded-lg border border-transparent bg-card p-4 shadow-card transition-[background-color,border-color,box-shadow] duration-150 ease-out hover:border-ring/40 hover:shadow-card-hover'>
      <div className='flex items-start gap-3'>
        <AgentAvatar
          name={agent.name}
          iconUrl={agentIconUrl(agent)}
          size='lg'
        />
        <div className='min-w-0 flex-1'>
          <h3 className='truncate text-sm font-semibold'>{agent.name}</h3>
          <p className='mt-1 line-clamp-2 text-xs leading-relaxed text-muted-foreground'>
            {agent.description || 'No description.'}
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger
            className='rounded-md p-1.5 text-muted-foreground hover:bg-muted'
            aria-label={`Open actions for ${agent.name}`}
          >
            <MoreVertical className='size-4' />
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' sideOffset={6}>
            {runnable && (
              <>
                <DropdownMenuItem
                  onClick={() =>
                    navigate({
                      to: '/agents/$name/edit',
                      params: { name: agent.agent_id ?? '' },
                    })
                  }
                >
                  <Pencil /> Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={onRun}>
                  <Play /> Run once
                </DropdownMenuItem>
              </>
            )}
            {deleted && (
              <DropdownMenuItem disabled={restoring} onClick={onRestore}>
                <ArchiveRestore /> Restore agent
              </DropdownMenuItem>
            )}
            {!runnable && (
              <DropdownMenuItem asChild>
                <Link to='/operations'>
                  <ListTodo /> View operations
                </Link>
              </DropdownMenuItem>
            )}
            {canDelete && <DropdownMenuSeparator />}
            {canDelete && (
              <DropdownMenuItem variant='destructive' onClick={onDelete}>
                <Trash2 /> Delete
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className='mt-3 flex items-center gap-2 text-xs'>
        <LifecycleBadge agent={agent} />
        {runnable && <StatusBadge status={badge.status} label={badge.label} />}
        {runnable && (runtime?.in_flight ?? 0) > 0 && (
          <span className='text-muted-foreground'>
            ×{runtime!.in_flight} in flight
          </span>
        )}
      </div>

      <div className='mt-3 flex flex-wrap gap-1'>
        <span className='rounded border border-border bg-muted/50 px-1.5 py-0.5 font-mono text-[0.7rem] text-muted-foreground'>
          {AGENT_TYPE_LABELS[agent.type ?? 'AGENT_TYPE_UNSPECIFIED']}
        </span>
        <span className='inline-flex items-center gap-1 rounded border border-border bg-muted/50 px-1.5 py-0.5 font-mono text-[0.7rem] text-muted-foreground'>
          <Fingerprint className='size-3' />
          {agent.agent_id}
        </span>
        {agent.enable_a2a && (
          <span className='rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[0.7rem] text-muted-foreground'>
            A2A
          </span>
        )}
        {agent.enable_openai_api && (
          <span className='rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[0.7rem] text-muted-foreground'>
            OpenAI API
          </span>
        )}
        {agent.enable_agui && (
          <span className='rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[0.7rem] text-muted-foreground'>
            AG-UI
          </span>
        )}
      </div>

      <div className='mt-3 flex items-center justify-between border-t border-border pt-3'>
        <span className='text-xs text-muted-foreground'>
          {runnable
            ? lastRun
              ? `Last run ${lastRun}`
              : 'Not run yet'
            : deleted
              ? 'Content and Agent ID retained'
              : 'Open operations for details'}
        </span>
        {runnable ? (
          <Button size='sm' asChild>
            <Link
              to='/chat'
              search={{ agent: agent.agent_id }}
            >
              <MessageSquarePlus />
              Start chat
            </Link>
          </Button>
        ) : deleted ? (
          <Button size='sm' disabled={restoring} onClick={onRestore}>
            <ArchiveRestore
              className={cn('size-4', restoring && 'motion-safe:animate-spin')}
            />
            {restoring ? 'Restoring...' : 'Restore agent'}
          </Button>
        ) : (
          <Button size='sm' variant='outline' asChild>
            <Link to='/operations'>
              <ListTodo /> View operations
            </Link>
          </Button>
        )}
      </div>
    </div>
  )
}

export function AgentList() {
  const { data, isLoading } = useAgents()
  const agents = useMemo(() => data?.agents ?? [], [data?.agents])
  // Query runtime status by immutable agent_id.
  const statusParams = useMemo(
    () => ({
      agent_ids: agents.flatMap((a) => (a.agent_id ? [a.agent_id] : [])),
    }),
    [agents]
  )
  const { data: runtimeData } = useAgentRuntimeStatuses(statusParams)

  const runtimeMap = useMemo(() => {
    const m = new Map<string, AgentRuntimeStatus>()
    for (const s of runtimeData?.statuses ?? []) {
      if (s.agent_id) m.set(s.agent_id, s)
    }
    return m
  }, [runtimeData])

  const deleteMutation = useDeleteAgent()
  const restoreMutation = useRestoreAgent()
  const reloadMutation = useReloadAgents()
  const invokeMutation = useInvokeAgent()

  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState<RuntimeFilter>('all')
  const [deleteTarget, setDeleteTarget] = useState<Agent | null>(null)
  const [invokeTarget, setInvokeTarget] = useState<Agent | null>(null)
  const [invokeInput, setInvokeInput] = useState('')
  const [invokeResult, setInvokeResult] = useState<{
    session_id: string
    response: string
  } | null>(null)

  const filtered = useMemo(
    () =>
      agents.filter(
        (a) =>
          (filter === 'all' ||
            runtimeStatusOf(runtimeMap.get(a.agent_id ?? '')) === filter) &&
          (a.name.toLowerCase().includes(query.toLowerCase()) ||
            (a.description ?? '').toLowerCase().includes(query.toLowerCase()))
      ),
    [agents, query, filter, runtimeMap]
  )

  const filters: { key: RuntimeFilter; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'running', label: 'Running' },
    { key: 'idle', label: 'Idle' },
    { key: 'failed', label: 'Failed' },
  ]

  return (
    <Page>
      <PageHeader
        title='Agents'
        subtitle='Browse agents and start a conversation, or configure how they work.'
        actions={
          <>
            <Button
              variant='outline'
              size='sm'
              onClick={() =>
                reloadMutation.mutate(undefined, {
                  onSuccess: () => toast.success('Agents reloaded'),
                  onError: (err) => toast.error(err.message),
                })
              }
              disabled={reloadMutation.isPending}
            >
              <RefreshCw
                className={cn(
                  'size-4',
                  reloadMutation.isPending && 'motion-safe:animate-spin'
                )}
              />
              Hot-reload
            </Button>
            <Button size='sm' asChild>
              <Link to='/agents/create'>
                <Plus className='size-4' />
                Create Agent
              </Link>
            </Button>
          </>
        }
      />
      <PageScroll>
        {isLoading ? (
          <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className='h-48' />
            ))}
          </div>
        ) : agents.length === 0 ? (
          <div className='mx-auto max-w-md rounded-lg border border-dashed border-border bg-card/40 px-6 py-14 text-center'>
            <div className='mx-auto flex size-11 items-center justify-center rounded-lg bg-muted text-muted-foreground'>
              <Bot className='size-5' />
            </div>
            <h2 className='mt-4 text-base font-semibold'>No agents yet</h2>
            <p className='mx-auto mt-1 max-w-xs text-sm text-pretty text-muted-foreground'>
              Agents are configurable assistants with their own model, tools,
              and instructions. Create one to start chatting and automating.
            </p>
            <Button className='mt-5' asChild>
              <Link to='/agents/create'>
                <Plus />
                Create your first Agent
              </Link>
            </Button>
          </div>
        ) : (
          <>
            <div className='mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
              <div className='relative w-full sm:max-w-xs'>
                <Search className='pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground' />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder='Search agents'
                  className='w-full rounded-md border border-border bg-card py-2 pr-3 pl-9 text-sm outline-none focus:border-ring'
                />
              </div>
              <div className='flex flex-wrap items-center gap-1 rounded-md border border-border bg-card p-0.5'>
                {filters.map((f) => (
                  <button
                    key={f.key}
                    type='button'
                    onClick={() => setFilter(f.key)}
                    className={cn(
                      'rounded px-2.5 py-1 text-xs font-medium transition-colors',
                      filter === f.key
                        ? 'bg-secondary text-secondary-foreground'
                        : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    {f.label}
                  </button>
                ))}
              </div>
            </div>

            {filtered.length === 0 ? (
              <p className='py-16 text-center text-sm text-muted-foreground'>
                No agents match your filters.
              </p>
            ) : (
              <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
                {filtered.map((a) => (
                  <AgentCard
                    key={a.agent_id}
                    agent={a}
                    runtime={runtimeMap.get(a.agent_id ?? '')}
                    onDelete={() => setDeleteTarget(a)}
                    onRun={() => {
                      setInvokeTarget(a)
                      setInvokeResult(null)
                      setInvokeInput('')
                    }}
                    onRestore={() => {
                      if (!a.agent_id) return
                      restoreMutation.mutate(a.agent_id, {
                        onSuccess: () => toast.success('Agent restored'),
                        onError: (err) => toast.error(err.message),
                      })
                    }}
                    restoring={
                      restoreMutation.isPending &&
                      restoreMutation.variables === a.agent_id
                    }
                  />
                ))}
              </div>
            )}
          </>
        )}
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete agent'
        description={`Delete "${deleteTarget?.name}"? The agent will stop running, but its Agent ID and repository content will be retained for restoration.`}
        loading={deleteMutation.isPending}
        confirmLabel='Delete agent'
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(
              {
                agent_id: deleteTarget.agent_id,
                operation_id: crypto.randomUUID(),
              },
              {
                onSuccess: () => {
                  toast.success('Agent deleted')
                  setDeleteTarget(null)
                },
                onError: (err) => toast.error(err.message),
              }
            )
          }
        }}
      />

      {/* Invoke dialog */}
      <Dialog
        open={!!invokeTarget}
        onOpenChange={(o) => {
          if (!o) {
            setInvokeTarget(null)
            setInvokeResult(null)
          }
        }}
      >
        <DialogContent className='max-w-2xl'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <Play className='size-4' /> Run {invokeTarget?.name}
            </DialogTitle>
            <DialogDescription>
              Sends a one-off invocation via the API. Creates an ephemeral
              session.
            </DialogDescription>
          </DialogHeader>

          {!invokeResult ? (
            <div className='space-y-2'>
              <Label htmlFor='invoke-input'>Input</Label>
              <Textarea
                id='invoke-input'
                rows={5}
                placeholder='What should the agent do?'
                value={invokeInput}
                onChange={(e) => setInvokeInput(e.target.value)}
              />
            </div>
          ) : (
            <div className='space-y-2'>
              <div className='text-xs text-muted-foreground'>
                Session:{' '}
                <span className='font-mono'>{invokeResult.session_id}</span>
              </div>
              <div className='rounded-md border bg-muted p-3 text-sm whitespace-pre-wrap'>
                {invokeResult.response || (
                  <span className='text-muted-foreground italic'>
                    (empty response)
                  </span>
                )}
              </div>
            </div>
          )}

          <DialogFooter>
            {!invokeResult ? (
              <>
                <Button variant='outline' onClick={() => setInvokeTarget(null)}>
                  Cancel
                </Button>
                <Button
                  disabled={!invokeInput.trim() || invokeMutation.isPending}
                  onClick={() =>
                    invokeTarget &&
                    invokeMutation.mutate(
                      {
                        agent_id: invokeTarget.agent_id ?? '',
                        input: invokeInput.trim(),
                      },
                      {
                        onSuccess: (res) => setInvokeResult(res),
                        onError: (err) => toast.error(err.message),
                      }
                    )
                  }
                >
                  {invokeMutation.isPending ? 'Running…' : 'Run'}
                </Button>
              </>
            ) : (
              <Button onClick={() => setInvokeTarget(null)}>Done</Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Page>
  )
}
