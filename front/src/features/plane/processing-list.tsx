import { PlaneProcessingStatus } from '@/gen/agents/v1/plane_pb'
import type { PlaneProcessingRecord } from '@/gen/agents/v1/plane_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { usePlaneProcessingRecords } from '@/api/plane'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const labels: Record<number, string> = {
  [PlaneProcessingStatus.RECEIVED]: 'Received',
  [PlaneProcessingStatus.PROCESSING]: 'Processing',
  [PlaneProcessingStatus.SUCCEEDED]: 'Succeeded',
  [PlaneProcessingStatus.FAILED]: 'Failed',
  [PlaneProcessingStatus.CANCELLED]: 'Cancelled',
  [PlaneProcessingStatus.QUEUED]: 'Queued',
  [PlaneProcessingStatus.FAILED_UNCERTAIN]: 'Uncertain — review required',
}

function recordTime(record: PlaneProcessingRecord) {
  const value = record.updatedAt ?? record.receivedAt
  return value ? timestampDate(value).toLocaleString() : '—'
}

export function PlaneProcessingList() {
  const { data = [], isLoading, error } = usePlaneProcessingRecords()
  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>Plane processing</CardTitle>
        <p className='text-sm text-muted-foreground'>
          Prompt batches, actor provenance, execution and delivery outcomes.
        </p>
      </CardHeader>
      <CardContent>
        {isLoading && <p className='text-sm text-muted-foreground'>Loading…</p>}
        {error && (
          <p className='text-sm text-destructive'>
            Unable to load Plane processing records.
          </p>
        )}
        {!isLoading && !error && data.length === 0 && (
          <p className='text-sm text-muted-foreground'>
            No Plane activity yet.
          </p>
        )}
        <div className='divide-y'>
          {data.slice(0, 8).map((record) => (
            <div
              key={record.id}
              className='grid grid-cols-[1fr_auto] gap-3 py-3 text-sm sm:grid-cols-[1fr_auto_auto]'
            >
              <div className='min-w-0'>
                <p className='truncate font-medium'>
                  {record.eventAction || record.eventType || 'Activity'}
                </p>
                <p className='truncate text-xs text-muted-foreground'>
                  {record.runId} · {record.activityId}
                </p>
                <p className='text-xs break-all text-muted-foreground'>
                  Actor: {record.actorUserId || 'unknown'} · Agent:{' '}
                  {record.agentId || '—'}
                </p>
                {record.invocationId && (
                  <p className='text-xs break-all text-muted-foreground'>
                    Invocation: {record.invocationId}
                  </p>
                )}
                {record.batchActivityIds.length > 1 && (
                  <p className='text-xs break-all text-muted-foreground'>
                    Batch: {record.batchActivityIds.join(', ')}
                  </p>
                )}
                {(record.queueConfirmationStatus ||
                  record.replyDeliveryStatus) && (
                  <p className='text-xs text-muted-foreground'>
                    Queue confirmation: {record.queueConfirmationStatus || '—'}{' '}
                    · Reply delivery: {record.replyDeliveryStatus || '—'}
                  </p>
                )}
                {record.memoryCaptureDisabledReason && (
                  <p className='text-xs text-muted-foreground'>
                    Memory capture disabled:{' '}
                    {record.memoryCaptureDisabledReason}
                  </p>
                )}
                {record.failureReason && (
                  <p className='truncate text-xs text-destructive'>
                    {record.failureReason}
                  </p>
                )}
              </div>
              <Badge
                variant={
                  record.status === PlaneProcessingStatus.FAILED ||
                  record.status === PlaneProcessingStatus.FAILED_UNCERTAIN
                    ? 'destructive'
                    : 'secondary'
                }
              >
                {labels[record.status] ?? 'Unknown'}
              </Badge>
              <time className='hidden text-xs text-muted-foreground sm:block'>
                {recordTime(record)}
              </time>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
