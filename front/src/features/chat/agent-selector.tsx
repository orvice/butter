import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import type { Agent } from '@/types/api'
import { Search } from 'lucide-react'
import { useAgents } from '@/api/agents'
import { cn } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'
import { AgentAvatar } from '@/components/butter/primitives'
import { agentIconUrl } from '@/features/agents/icon-utils'

export function AgentSelector({
  onPick,
  busy,
}: {
  onPick: (agentName: string) => void
  busy?: boolean
}) {
  const { data, isLoading } = useAgents({ page_size: 200 })
  const [query, setQuery] = useState('')

  const agents = useMemo(() => data?.agents ?? [], [data])

  const filtered = useMemo(
    () =>
      agents.filter(
        (a) =>
          a.name.toLowerCase().includes(query.toLowerCase()) ||
          (a.description ?? '').toLowerCase().includes(query.toLowerCase())
      ),
    [agents, query]
  )

  if (isLoading) {
    return (
      <div className='mx-auto w-full max-w-3xl'>
        <div className='space-y-2'>
          <Skeleton className='h-7 w-44' />
          <Skeleton className='h-4 w-64 max-w-full' />
        </div>
        <Skeleton className='mt-6 h-10 w-full max-w-md' />
        <div className='mt-6 grid gap-2.5 sm:grid-cols-2'>
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className='h-[76px] rounded-lg' />
          ))}
        </div>
      </div>
    )
  }

  if (agents.length === 0) {
    return (
      <div className='mx-auto max-w-md rounded-lg border border-dashed border-border bg-card/50 px-6 py-10 text-center'>
        <h2 className='text-base font-semibold'>No agents yet</h2>
        <p className='mt-1 text-sm text-muted-foreground'>
          Create your first agent to start chatting.
        </p>
        <Link
          to='/agents/create'
          className='mt-4 inline-block rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90'
        >
          Create Agent
        </Link>
      </div>
    )
  }

  return (
    <div className='mx-auto w-full max-w-3xl'>
      <div>
        <h1 className='font-manrope text-2xl font-semibold text-balance'>
          Start a new chat
        </h1>
        <p className='mt-1.5 text-sm text-muted-foreground'>
          Choose an agent to begin.
        </p>
      </div>

      <div className='relative mt-6 max-w-md'>
        <Search className='pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground' />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder='Search agents'
          className='h-10 w-full rounded-md border border-border bg-card pr-3 pl-9 text-sm transition-[border-color,box-shadow] outline-none focus:border-ring focus:ring-2 focus:ring-ring/10'
        />
      </div>

      <div className='mt-6'>
        {filtered.length === 0 ? (
          <p className='py-8 text-center text-sm text-muted-foreground'>
            No agents match “{query}”.
          </p>
        ) : (
          <div className='grid gap-2.5 sm:grid-cols-2'>
            {filtered.map((a: Agent) => (
              <button
                key={a.name}
                type='button'
                disabled={busy}
                onClick={() => onPick(a.name)}
                className={cn(
                  'flex min-h-[76px] items-start gap-3 rounded-lg border border-border bg-card p-3.5 text-left transition-[border-color,background-color,scale]',
                  busy
                    ? 'cursor-wait opacity-60'
                    : 'hover:border-ring/60 hover:bg-accent/50 active:scale-[0.99] motion-reduce:active:scale-100'
                )}
              >
                <AgentAvatar
                  name={a.name}
                  iconUrl={agentIconUrl(a)}
                  size='md'
                />
                <div className='min-w-0 flex-1'>
                  <span className='block truncate text-sm font-medium'>
                    {a.name}
                  </span>
                  <p className='mt-0.5 line-clamp-2 text-xs text-muted-foreground'>
                    {a.description || 'No description.'}
                  </p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
