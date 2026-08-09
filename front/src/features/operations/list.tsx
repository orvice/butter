import { Fragment, useMemo, useState } from 'react'
import {
  AgentOperationStatus,
  AgentOperationStepKind,
  AgentOperationStepStatus,
  AgentOperationType,
  type AgentOperation,
} from '@/gen/agents/v1/agent_operation_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import {
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleDot,
  Clock3,
  LoaderCircle,
  RefreshCw,
  RotateCcw,
  XCircle,
} from 'lucide-react'
import { toast } from 'sonner'
import { useAgentOperations, useRetryAgentOperation } from '@/api/agents'
import { cn } from '@/lib/utils'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'

type StatusFilter = 'all' | 'pending' | 'running' | 'succeeded' | 'failed'

const STATUS_FILTERS: Array<{
  key: StatusFilter
  label: string
  value: AgentOperationStatus
}> = [
  { key: 'all', label: 'All', value: AgentOperationStatus.UNSPECIFIED },
  { key: 'pending', label: 'Pending', value: AgentOperationStatus.PENDING },
  { key: 'running', label: 'Running', value: AgentOperationStatus.RUNNING },
  {
    key: 'succeeded',
    label: 'Succeeded',
    value: AgentOperationStatus.SUCCEEDED,
  },
  { key: 'failed', label: 'Failed', value: AgentOperationStatus.FAILED },
]

const OPERATION_LABELS: Record<AgentOperationType, string> = {
  [AgentOperationType.UNSPECIFIED]: 'Unknown',
  [AgentOperationType.CREATE]: 'Create',
  [AgentOperationType.UPDATE_CONFIGURATION]: 'Update configuration',
  [AgentOperationType.DELETE]: 'Delete',
  [AgentOperationType.RESTORE]: 'Restore',
}

const STEP_LABELS: Record<AgentOperationStepKind, string> = {
  [AgentOperationStepKind.UNSPECIFIED]: 'Unknown step',
  [AgentOperationStepKind.DB_PROVISION]: 'Provision database record',
  [AgentOperationStepKind.CONTENT_COMMIT]: 'Commit agent content',
  [AgentOperationStepKind.SYNC_PUBLISH]: 'Sync and publish',
  [AgentOperationStepKind.ACTIVATE]: 'Activate agent',
  [AgentOperationStepKind.TOMBSTONE]: 'Tombstone agent',
  [AgentOperationStepKind.RESTORE_DB]: 'Restore database record',
  [AgentOperationStepKind.DB_PATCH]: 'Apply configuration patch',
}

function formatTimestamp(value: AgentOperation['updatedAt']): string {
  if (!value) return 'Not recorded'
  return timestampDate(value).toLocaleString()
}

function shortID(value: string): string {
  if (value.length <= 18) return value
  return `${value.slice(0, 8)}…${value.slice(-6)}`
}

function OperationStatusBadge({ status }: { status: AgentOperationStatus }) {
  const config = {
    [AgentOperationStatus.UNSPECIFIED]: {
      label: 'Unknown',
      icon: CircleDot,
      className: 'border-border bg-muted text-muted-foreground',
    },
    [AgentOperationStatus.PENDING]: {
      label: 'Pending',
      icon: Clock3,
      className:
        'border-amber-500/35 bg-amber-500/10 text-amber-700 dark:text-amber-300',
    },
    [AgentOperationStatus.RUNNING]: {
      label: 'Running',
      icon: LoaderCircle,
      className:
        'border-sky-500/35 bg-sky-500/10 text-sky-700 dark:text-sky-300',
    },
    [AgentOperationStatus.SUCCEEDED]: {
      label: 'Succeeded',
      icon: CheckCircle2,
      className:
        'border-emerald-500/35 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
    },
    [AgentOperationStatus.FAILED]: {
      label: 'Failed',
      icon: XCircle,
      className: 'border-destructive/35 bg-destructive/10 text-destructive',
    },
  }[status]
  const Icon = config.icon
  return (
    <Badge variant='outline' className={config.className}>
      <Icon
        className={cn(
          status === AgentOperationStatus.RUNNING && 'motion-safe:animate-spin'
        )}
      />
      {config.label}
    </Badge>
  )
}

function StepStatusIcon({ status }: { status: AgentOperationStepStatus }) {
  if (status === AgentOperationStepStatus.SUCCEEDED) {
    return (
      <CheckCircle2
        className='size-4 text-emerald-600 dark:text-emerald-400'
        aria-label='Succeeded'
      />
    )
  }
  if (status === AgentOperationStepStatus.FAILED) {
    return <XCircle className='size-4 text-destructive' aria-label='Failed' />
  }
  if (status === AgentOperationStepStatus.PENDING) {
    return (
      <Clock3
        className='size-4 text-amber-600 dark:text-amber-400'
        aria-label='Pending'
      />
    )
  }
  return (
    <CircleDot
      className='size-4 text-muted-foreground'
      aria-label='Not started'
    />
  )
}

function StepDetails({ operation }: { operation: AgentOperation }) {
  if (operation.steps.length === 0) {
    return (
      <p className='text-sm text-muted-foreground'>
        No step records are available.
      </p>
    )
  }
  return (
    <ol
      className='grid gap-2'
      aria-label={`Steps for operation ${operation.id}`}
    >
      {operation.steps.map((step, index) => (
        <li
          key={`${step.kind}-${index}`}
          className='grid grid-cols-[1rem_minmax(0,1fr)] gap-3 rounded-md border border-border/70 bg-background px-3 py-2.5'
        >
          <div className='pt-0.5'>
            <StepStatusIcon status={step.status} />
          </div>
          <div className='min-w-0'>
            <div className='flex flex-wrap items-center justify-between gap-x-4 gap-y-1'>
              <span className='text-sm font-medium'>
                {STEP_LABELS[step.kind]}
              </span>
              <span className='text-xs text-muted-foreground'>
                Attempt {step.attemptCount}
              </span>
            </div>
            <p className='mt-0.5 text-xs text-muted-foreground'>
              {step.finishedAt
                ? `Finished ${formatTimestamp(step.finishedAt)}`
                : step.startedAt
                  ? `Started ${formatTimestamp(step.startedAt)}`
                  : 'Not started'}
            </p>
            {step.error && (
              <p className='mt-2 text-xs break-words text-destructive'>
                {step.error}
              </p>
            )}
          </div>
        </li>
      ))}
    </ol>
  )
}

function RetryButton({ operation }: { operation: AgentOperation }) {
  const retry = useRetryAgentOperation()
  if (operation.status !== AgentOperationStatus.FAILED) return null
  return (
    <Button
      size='sm'
      variant='outline'
      disabled={retry.isPending}
      onClick={() =>
        retry.mutate(operation.id, {
          onSuccess: () => toast.success('Operation retry completed'),
          onError: (error) => toast.error(error.message),
        })
      }
    >
      <RotateCcw
        className={cn('size-4', retry.isPending && 'motion-safe:animate-spin')}
      />
      Retry
    </Button>
  )
}

function MobileOperation({
  operation,
  expanded,
  onToggle,
}: {
  operation: AgentOperation
  expanded: boolean
  onToggle: () => void
}) {
  const detailsID = `operation-mobile-${operation.id}`
  return (
    <article className='rounded-lg border bg-card p-4 shadow-xs'>
      <div className='flex items-start justify-between gap-3'>
        <div className='min-w-0'>
          <p className='truncate text-sm font-semibold'>
            {operation.agentName || operation.agentId}
          </p>
          <p
            className='mt-0.5 font-mono text-xs text-muted-foreground'
            title={operation.id}
          >
            {shortID(operation.id)}
          </p>
        </div>
        <OperationStatusBadge status={operation.status} />
      </div>
      <dl className='mt-4 grid grid-cols-2 gap-x-4 gap-y-3 text-xs'>
        <div>
          <dt className='text-muted-foreground'>Operation</dt>
          <dd className='mt-0.5 font-medium'>
            {OPERATION_LABELS[operation.type]}
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground'>Attempts</dt>
          <dd className='mt-0.5 font-medium'>{operation.attemptCount}</dd>
        </div>
        <div className='col-span-2'>
          <dt className='text-muted-foreground'>Updated</dt>
          <dd className='mt-0.5 font-medium'>
            {formatTimestamp(operation.updatedAt)}
          </dd>
        </div>
      </dl>
      {operation.error && (
        <p className='mt-3 text-xs break-words text-destructive'>
          {operation.error}
        </p>
      )}
      <div className='mt-4 flex items-center justify-between gap-2 border-t pt-3'>
        <Button
          type='button'
          size='sm'
          variant='ghost'
          aria-expanded={expanded}
          aria-controls={detailsID}
          onClick={onToggle}
        >
          <ChevronDown
            className={cn(
              'size-4 transition-transform',
              expanded && 'rotate-180'
            )}
          />
          {expanded ? 'Hide steps' : 'Show steps'}
        </Button>
        <RetryButton operation={operation} />
      </div>
      {expanded && (
        <div id={detailsID} className='mt-3'>
          <StepDetails operation={operation} />
        </div>
      )}
    </article>
  )
}

function OperationsSkeleton() {
  return (
    <div className='grid gap-3'>
      {Array.from({ length: 6 }).map((_, index) => (
        <Skeleton key={index} className='h-16 w-full' />
      ))}
    </div>
  )
}

export function OperationsPage() {
  const [filter, setFilter] = useState<StatusFilter>('all')
  const [pageTokens, setPageTokens] = useState<string[]>([''])
  const [pageIndex, setPageIndex] = useState(0)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const selectedStatus = STATUS_FILTERS.find(
    (item) => item.key === filter
  )?.value
  const params = useMemo(
    () => ({
      status: selectedStatus,
      page_size: 20,
      page_token: pageTokens[pageIndex] || undefined,
    }),
    [pageIndex, pageTokens, selectedStatus]
  )
  const query = useAgentOperations(params)
  const operations = query.data?.operations ?? []
  const nextToken = query.data?.next_page_token

  function changeFilter(value: string) {
    setFilter(value as StatusFilter)
    setPageTokens([''])
    setPageIndex(0)
    setExpanded(new Set())
  }

  function toggleOperation(id: string) {
    setExpanded((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function nextPage() {
    if (!nextToken) return
    if (pageIndex === pageTokens.length - 1)
      setPageTokens((tokens) => [...tokens, nextToken])
    setPageIndex((index) => index + 1)
    setExpanded(new Set())
  }

  return (
    <Page>
      <PageHeader
        title='Operations'
        subtitle='Durable agent lifecycle activity for this workspace.'
        actions={
          <Button
            type='button'
            size='icon'
            variant='outline'
            aria-label='Refresh operations'
            title='Refresh operations'
            disabled={query.isFetching}
            onClick={() => query.refetch()}
          >
            <RefreshCw
              className={cn(
                'size-4',
                query.isFetching && 'motion-safe:animate-spin'
              )}
            />
          </Button>
        }
      />
      <PageScroll>
        <Tabs value={filter} onValueChange={changeFilter} className='mb-4'>
          <div className='overflow-x-auto pb-1'>
            <TabsList aria-label='Filter operations by status'>
              {STATUS_FILTERS.map((item) => (
                <TabsTrigger key={item.key} value={item.key}>
                  {item.label}
                </TabsTrigger>
              ))}
            </TabsList>
          </div>
        </Tabs>

        {query.isError ? (
          <Alert variant='destructive'>
            <AlertCircle />
            <AlertTitle>Could not load operations</AlertTitle>
            <AlertDescription>
              <p>{query.error.message}</p>
              <Button
                size='sm'
                variant='outline'
                onClick={() => query.refetch()}
              >
                <RefreshCw /> Retry loading
              </Button>
            </AlertDescription>
          </Alert>
        ) : query.isLoading ? (
          <OperationsSkeleton />
        ) : operations.length === 0 ? (
          <div className='border-y py-14 text-center'>
            <CircleDot className='mx-auto size-7 text-muted-foreground' />
            <h2 className='mt-3 text-sm font-semibold'>No operations found</h2>
            <p className='mt-1 text-sm text-muted-foreground'>
              No records match the selected status.
            </p>
          </div>
        ) : (
          <>
            <div className='hidden rounded-lg border md:block'>
              <Table className='table-fixed'>
                <TableHeader>
                  <TableRow>
                    <TableHead className='w-10'>
                      <span className='sr-only'>Steps</span>
                    </TableHead>
                    <TableHead className='w-[20%]'>Agent</TableHead>
                    <TableHead className='w-[16%]'>Operation</TableHead>
                    <TableHead className='w-[14%]'>Status</TableHead>
                    <TableHead className='w-[10%] text-center'>
                      Attempts
                    </TableHead>
                    <TableHead className='w-[20%]'>Updated</TableHead>
                    <TableHead className='w-[20%] text-end'>Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {operations.map((operation) => {
                    const isExpanded = expanded.has(operation.id)
                    const detailsID = `operation-${operation.id}`
                    return (
                      <Fragment key={operation.id}>
                        <TableRow>
                          <TableCell>
                            <Button
                              type='button'
                              size='icon'
                              variant='ghost'
                              className='size-8'
                              aria-label={
                                isExpanded
                                  ? 'Hide operation steps'
                                  : 'Show operation steps'
                              }
                              aria-expanded={isExpanded}
                              aria-controls={detailsID}
                              onClick={() => toggleOperation(operation.id)}
                            >
                              <ChevronDown
                                className={cn(
                                  'size-4 transition-transform',
                                  isExpanded && 'rotate-180'
                                )}
                              />
                            </Button>
                          </TableCell>
                          <TableCell className='overflow-hidden'>
                            <p className='truncate font-medium'>
                              {operation.agentName || operation.agentId}
                            </p>
                            <p
                              className='truncate font-mono text-xs text-muted-foreground'
                              title={operation.id}
                            >
                              {shortID(operation.id)}
                            </p>
                          </TableCell>
                          <TableCell>
                            {OPERATION_LABELS[operation.type]}
                          </TableCell>
                          <TableCell>
                            <OperationStatusBadge status={operation.status} />
                          </TableCell>
                          <TableCell className='text-center tabular-nums'>
                            {operation.attemptCount}
                          </TableCell>
                          <TableCell className='text-xs text-muted-foreground'>
                            {formatTimestamp(operation.updatedAt)}
                          </TableCell>
                          <TableCell>
                            <div className='flex justify-end'>
                              <RetryButton operation={operation} />
                            </div>
                          </TableCell>
                        </TableRow>
                        {isExpanded && (
                          <TableRow
                            id={detailsID}
                            className='hover:bg-transparent'
                          >
                            <TableCell
                              colSpan={7}
                              className='bg-muted/25 p-4 whitespace-normal'
                            >
                              {operation.error && (
                                <p className='mb-3 text-sm break-words text-destructive'>
                                  {operation.error}
                                </p>
                              )}
                              <StepDetails operation={operation} />
                            </TableCell>
                          </TableRow>
                        )}
                      </Fragment>
                    )
                  })}
                </TableBody>
              </Table>
            </div>

            <div className='grid gap-3 md:hidden'>
              {operations.map((operation) => (
                <MobileOperation
                  key={operation.id}
                  operation={operation}
                  expanded={expanded.has(operation.id)}
                  onToggle={() => toggleOperation(operation.id)}
                />
              ))}
            </div>
          </>
        )}

        {!query.isLoading && !query.isError && operations.length > 0 && (
          <div className='mt-4 flex items-center justify-between gap-3 text-xs text-muted-foreground'>
            <span>Page {pageIndex + 1}</span>
            <div className='flex items-center gap-2'>
              <Button
                type='button'
                size='icon'
                variant='outline'
                aria-label='Previous page'
                disabled={pageIndex === 0}
                onClick={() => {
                  setPageIndex((index) => Math.max(0, index - 1))
                  setExpanded(new Set())
                }}
              >
                <ChevronLeft />
              </Button>
              <Button
                type='button'
                size='icon'
                variant='outline'
                aria-label='Next page'
                disabled={!nextToken}
                onClick={nextPage}
              >
                <ChevronRight />
              </Button>
            </div>
          </div>
        )}
      </PageScroll>
    </Page>
  )
}
