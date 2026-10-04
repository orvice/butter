import { cn } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'
import { AgentAvatar } from '@/components/butter/primitives'

// ThreadSkeleton stands in for a thread's messages while they load.
export function ThreadSkeleton({ className }: { className?: string }) {
  return (
    <div
      role='status'
      aria-label='Loading conversation'
      className={cn('space-y-6 py-8', className)}
    >
      <div className='flex gap-3'>
        <Skeleton className='size-8 shrink-0 rounded-md' />
        <div className='w-full max-w-xl space-y-2'>
          <Skeleton className='h-3 w-28' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-4/5' />
        </div>
      </div>
      <Skeleton className='ml-auto h-14 w-1/2 max-w-md rounded-lg' />
    </div>
  )
}

// AgentHero introduces the agent of a thread with no messages yet. A thread
// whose agent is not known shows "Unknown agent".
export function AgentHero({
  name,
  iconUrl,
  className,
}: {
  name: string | null
  iconUrl?: string
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex min-h-[50vh] flex-col items-center justify-center text-center',
        className
      )}
    >
      <AgentAvatar name={name ?? '?'} iconUrl={iconUrl} size='lg' />
      <h2 className='mt-3 text-lg font-semibold'>{name ?? 'Unknown agent'}</h2>
      <p className='mt-1 max-w-sm text-sm text-pretty text-muted-foreground'>
        Send a message below to start the conversation.
      </p>
    </div>
  )
}

// ChatDisclaimer goes under a chat's composer.
export function ChatDisclaimer({ className }: { className?: string }) {
  return (
    <p
      className={cn(
        'mt-1.5 px-2 text-center text-[0.7rem] leading-4 text-muted-foreground/80',
        className
      )}
    >
      Butter can make mistakes. Verify important actions before running them.
    </p>
  )
}
