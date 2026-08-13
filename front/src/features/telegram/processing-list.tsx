import { useState } from 'react'
import { toast } from 'sonner'
import { RefreshCw, Send } from 'lucide-react'
import {
  useResendTelegramReply,
  useTelegramProcessingRecords,
} from '@/api/telegram'
import {
  TelegramProcessingStatus,
  type TelegramProcessingRecord,
} from '@/gen/agents/v1/telegram_pb'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
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
import { PROCESSING_STATUS_LABELS } from './labels'

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: 'all', label: 'All statuses' },
  { value: String(TelegramProcessingStatus.SUCCEEDED), label: 'Succeeded' },
  { value: String(TelegramProcessingStatus.FAILED), label: 'Failed' },
  { value: String(TelegramProcessingStatus.FAILED_UNCERTAIN), label: 'Needs review' },
  { value: String(TelegramProcessingStatus.PROCESSING), label: 'Processing' },
]

/** A resend is only meaningful when a complete response exists to resend. */
function canResend(record: TelegramProcessingRecord): boolean {
  return (
    record.segments.length > 0 &&
    record.segments.some((segment) => segment.status !== 'sent')
  )
}

export function TelegramProcessingList() {
  const [status, setStatus] = useState('all')
  const { data: records, isLoading } = useTelegramProcessingRecords(
    status === 'all'
      ? {}
      : { status: Number(status) as TelegramProcessingStatus }
  )
  const resend = useResendTelegramReply()

  return (
    <Page>
      <PageHeader
        title='Telegram updates'
        subtitle='Processing history for accepted Telegram updates in this workspace.'
        actions={
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className='w-56' aria-label='Status filter'>
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
        }
      />
      <PageScroll>
        {isLoading ? (
          <Skeleton className='h-40' />
        ) : !records?.length ? (
          <Card>
            <CardContent className='py-10 text-center text-sm text-muted-foreground'>
              No Telegram updates have been processed yet.
            </CardContent>
          </Card>
        ) : (
          <div className='space-y-3'>
            {records.map((record) => (
              <Card key={record.id} data-testid={`processing-${record.id}`}>
                <CardHeader className='flex flex-row items-start justify-between gap-2 pb-2'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <CardTitle className='text-sm font-medium'>
                      update {String(record.updateId)}
                    </CardTitle>
                    <Badge
                      variant='outline'
                      className={
                        record.status === TelegramProcessingStatus.FAILED_UNCERTAIN
                          ? 'border-destructive text-destructive'
                          : undefined
                      }
                    >
                      {PROCESSING_STATUS_LABELS[record.status] ?? 'Unknown'}
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
                    aria-label={`Resend reply for update ${String(record.updateId)}`}
                    disabled={!canResend(record) || resend.isPending}
                    onClick={async () => {
                      try {
                        await resend.mutateAsync(record.id)
                        toast.success('Reply resent')
                      } catch (err) {
                        toast.error(err instanceof Error ? err.message : 'Resend failed')
                      }
                    }}
                  >
                    <Send className='h-4 w-4' />
                    Resend
                  </Button>
                </CardHeader>
                <CardContent className='space-y-2 text-sm text-muted-foreground'>
                  <div className='font-mono text-xs'>
                    attempts {record.attempts} · invocation {record.invocationId || '—'}
                  </div>
                  {record.error && (
                    <p className='text-destructive' data-testid={`processing-error-${record.id}`}>
                      {record.error}
                    </p>
                  )}
                  {record.segments.length > 0 && (
                    <div className='flex flex-wrap items-center gap-1'>
                      <RefreshCw className='h-3 w-3' />
                      {record.segments.map((segment) => (
                        <Badge
                          key={segment.index}
                          variant='outline'
                          className='text-[10px]'
                        >
                          {segment.index + 1}: {segment.status}
                        </Badge>
                      ))}
                    </div>
                  )}
                  {record.status === TelegramProcessingStatus.FAILED_UNCERTAIN && (
                    <p className='text-xs'>
                      The agent may already have taken action, so this is not retried
                      automatically. Resend the reply if one was produced.
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
