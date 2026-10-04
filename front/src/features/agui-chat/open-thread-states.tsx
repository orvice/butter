import type { ReactNode } from 'react'
import { Loader2, MessageSquareOff, TriangleAlert } from 'lucide-react'
import { Button } from '@/components/ui/button'

// The states of opening the thread the URL names, shown in place of the
// conversation: it is being looked up, it could not be read, or it is not
// one this caller can open here. The conversation's own states (its loading
// skeleton, the empty-thread hero) are in components/chat/thread-states.

function ThreadState({
  icon,
  title,
  detail,
  action,
  role,
}: {
  icon: ReactNode
  title: string
  detail?: string
  action?: ReactNode
  role?: 'status' | 'alert'
}) {
  return (
    <div className='flex flex-1 items-center justify-center px-6 py-10'>
      <div
        role={role}
        className='flex max-w-sm flex-col items-center gap-2 text-center'
      >
        <div className='mb-1 flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground'>
          {icon}
        </div>
        <p className='text-sm font-medium'>{title}</p>
        {detail && (
          <p className='text-sm text-pretty text-muted-foreground'>{detail}</p>
        )}
        {action && <div className='mt-2'>{action}</div>}
      </div>
    </div>
  )
}

export function ThreadLoading() {
  return (
    <ThreadState
      role='status'
      icon={<Loader2 className='size-5 animate-spin' />}
      title='Loading thread…'
    />
  )
}

export function ThreadLoadFailed({
  error,
  onRetry,
}: {
  error: unknown
  onRetry: () => void
}) {
  return (
    <ThreadState
      role='alert'
      icon={<TriangleAlert className='size-5' />}
      title='Failed to load this thread.'
      detail={
        error instanceof Error && error.message ? error.message : undefined
      }
      action={
        <Button size='sm' variant='outline' onClick={onRetry}>
          Retry
        </Button>
      }
    />
  )
}

// ThreadNotFound covers a thread that does not exist and one this caller
// cannot open here: another user's, another workspace's, or one bound to an
// agent this workspace cannot run. detail says why when the server refused
// the thread.
export function ThreadNotFound({
  detail,
  onStartNew,
}: {
  detail?: string
  onStartNew: () => void
}) {
  return (
    <ThreadState
      icon={<MessageSquareOff className='size-5' />}
      title='Thread not found.'
      detail={
        detail ??
        'It may have been deleted, or it belongs to another workspace or agent.'
      }
      action={
        <Button size='sm' onClick={onStartNew}>
          Start a new chat
        </Button>
      }
    />
  )
}
