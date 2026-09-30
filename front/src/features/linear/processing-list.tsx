import { useState } from 'react'
import {
  LinearProcessingStatus,
  type LinearProcessingRecord,
} from '@/gen/agents/v1/linear_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Link } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight, Send } from 'lucide-react'
import { toast } from 'sonner'
import {
  useLinearApps,
  useLinearProcessingRecords,
  useResendLinearReply,
} from '@/api/linear'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'

const STATUS_LABELS: Record<number, string> = {
  [LinearProcessingStatus.QUEUED]: 'Queued',
  [LinearProcessingStatus.RECEIVED]: 'Received',
  [LinearProcessingStatus.PROCESSING]: 'Processing',
  [LinearProcessingStatus.READY_TO_DELIVER]: 'Ready to post',
  [LinearProcessingStatus.SUCCEEDED]: 'Succeeded',
  [LinearProcessingStatus.FAILED]: 'Failed',
  [LinearProcessingStatus.FAILED_UNCERTAIN]: 'Needs review',
  [LinearProcessingStatus.CANCELLED]: 'Cancelled',
}

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: 'all', label: 'All statuses' },
  { value: String(LinearProcessingStatus.SUCCEEDED), label: 'Succeeded' },
  { value: String(LinearProcessingStatus.FAILED), label: 'Failed' },
  {
    value: String(LinearProcessingStatus.FAILED_UNCERTAIN),
    label: 'Needs review',
  },
  { value: String(LinearProcessingStatus.PROCESSING), label: 'Processing' },
  { value: String(LinearProcessingStatus.QUEUED), label: 'Queued' },
  { value: String(LinearProcessingStatus.CANCELLED), label: 'Cancelled' },
]

/** Only a reply that was produced but never posted can be resent. */
function canResend(record: LinearProcessingRecord): boolean {
  return (
    record.status === LinearProcessingStatus.FAILED &&
    Boolean(record.output) &&
    !record.delivered
  )
}

export function LinearProcessingList() {
  const [status, setStatus] = useState('all')
  const [appId, setAppId] = useState('all')
  const { data: apps } = useLinearApps()
  const { data: records, isLoading } = useLinearProcessingRecords({
    appId: appId === 'all' ? undefined : appId,
    status:
      status === 'all' ? undefined : (Number(status) as LinearProcessingStatus),
  })
  const resend = useResendLinearReply()
  const appName = (id: string) =>
    apps?.find((app) => app.id === id)?.displayName ?? id

  return (
    <Page>
      <PageHeader
        title='Linear deliveries'
        breadcrumb={
          <Link
            to='/linear-apps'
            className='inline-flex items-center gap-1.5 hover:text-foreground'
          >
            <ArrowLeft className='size-3.5' />
            Linear Apps
            <ChevronRight className='size-3' />
            <span className='text-foreground'>Deliveries</span>
          </Link>
        }
        subtitle='Processing history for Linear Agent Session events in this workspace.'
        actions={
          <div className='flex gap-2'>
            <Select value={appId} onValueChange={setAppId}>
              <SelectTrigger className='w-48' aria-label='Linear App filter'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='all'>All Linear Apps</SelectItem>
                {(apps ?? []).map((app) => (
                  <SelectItem key={app.id} value={app.id}>
                    {app.displayName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className='w-48' aria-label='Status filter'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {STATUS_FILTERS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        }
      />
      <PageScroll>
        {isLoading ? (
          <Skeleton className='h-40' />
        ) : !records?.length ? (
          <Card>
            <CardContent className='py-10 text-center text-sm text-muted-foreground'>
              No Linear deliveries have been processed yet.
            </CardContent>
          </Card>
        ) : (
          <div className='space-y-3'>
            {records.map((record) => (
              <Card
                key={record.id}
                data-testid={`linear-processing-${record.id}`}
              >
                <CardHeader className='flex flex-row items-start justify-between gap-2 pb-2'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <CardTitle className='text-sm font-medium'>
                      {record.issueIdentifier || record.agentSessionId} ·{' '}
                      {record.action}
                    </CardTitle>
                    <Badge
                      variant='outline'
                      className={
                        record.status ===
                        LinearProcessingStatus.FAILED_UNCERTAIN
                          ? 'border-destructive text-destructive'
                          : undefined
                      }
                    >
                      {STATUS_LABELS[record.status] ?? 'Unknown'}
                    </Badge>
                    {record.deadLettered && (
                      <Badge className='bg-destructive/10 text-destructive'>
                        Needs review
                      </Badge>
                    )}
                  </div>
                  <Button
                    variant='outline'
                    size='sm'
                    aria-label={`Resend reply for ${record.issueIdentifier || record.id}`}
                    disabled={!canResend(record) || resend.isPending}
                    onClick={async () => {
                      try {
                        await resend.mutateAsync(record.id)
                        toast.success('Reply resent')
                      } catch (err) {
                        toast.error(
                          err instanceof Error ? err.message : 'Resend failed'
                        )
                      }
                    }}
                  >
                    <Send className='h-4 w-4' />
                    Resend
                  </Button>
                </CardHeader>
                <CardContent className='space-y-2 text-sm text-muted-foreground'>
                  <div className='font-mono text-xs'>
                    {appName(record.appId)} · attempts {record.attempts} ·
                    invocation {record.invocationId || '—'}
                    {record.createdAt &&
                      ` · ${timestampDate(record.createdAt).toLocaleString()}`}
                  </div>
                  {record.error && (
                    <p className='text-destructive'>{record.error}</p>
                  )}
                  {record.status ===
                    LinearProcessingStatus.FAILED_UNCERTAIN && (
                    <p className='text-xs'>
                      The agent may already have taken action, so this is not
                      retried automatically.
                    </p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </PageScroll>
    </Page>
  )
}
