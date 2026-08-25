import { useButterBoxStatus } from '@/api/butterboxes'
import { Badge } from '@/components/ui/badge'

/**
 * Reachability badge for one ButterBox. Unreachable is data, not an
 * exception: the probe error rides along as a native tooltip.
 */
export function BoxStatusBadge({ id }: { id: string }) {
  const { data, isLoading } = useButterBoxStatus(id)
  if (isLoading || !data)
    return (
      <Badge variant='outline' className='text-xs'>
        …
      </Badge>
    )
  if (data.reachable) {
    return (
      <Badge className='bg-success-muted text-success-foreground'>
        <span className='h-1.5 w-1.5 rounded-full bg-current' />
        Reachable
      </Badge>
    )
  }
  return (
    <Badge
      className='bg-danger-muted text-danger-foreground'
      title={data.error || undefined}
    >
      <span className='h-1.5 w-1.5 rounded-full bg-current' />
      Unreachable
    </Badge>
  )
}

/** Active pi session count; a dash while the box is unreachable. */
export function BoxSessionsInline({ id }: { id: string }) {
  const { data } = useButterBoxStatus(id)
  if (!data?.reachable)
    return <span className='text-sm text-muted-foreground'>-</span>
  return <span className='text-sm'>{data.activeSessions}</span>
}

/** One-line status summary for the edit page, error text inline. */
export function BoxStatusLine({ id }: { id: string }) {
  const { data, isLoading } = useButterBoxStatus(id)
  if (isLoading || !data)
    return <p className='text-sm text-muted-foreground'>Checking status…</p>
  if (data.reachable) {
    return (
      <p className='text-sm text-muted-foreground'>
        Reachable · {data.activeSessions} active session
        {data.activeSessions === 1 ? '' : 's'}
      </p>
    )
  }
  return (
    <p className='text-sm text-destructive'>
      Unreachable{data.error ? `: ${data.error}` : ''}
    </p>
  )
}
