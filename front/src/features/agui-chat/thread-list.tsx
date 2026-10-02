import { useState } from 'react'
import type { SessionInfo } from '@/types/api'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { InlineTitleInput } from '@/components/inline-title-input'
import { threadIdOf, threadTitle } from './threads'

// ThreadList shows the caller's threads with the selected agent, newest
// first. A new thread has no session until its first run, so it appears
// here only after that.
export function ThreadList({
  threads,
  activeThreadId,
  isLoading,
  onSelect,
  onRename,
  onDelete,
}: {
  threads: SessionInfo[]
  activeThreadId: string
  isLoading: boolean
  onSelect: (threadId: string) => void
  onRename: (session: SessionInfo, title: string) => Promise<void>
  onDelete: (session: SessionInfo) => void
}) {
  const [renamingId, setRenamingId] = useState<string | null>(null)

  if (threads.length === 0) {
    return (
      <p className='px-3 py-4 text-center text-xs text-muted-foreground'>
        {isLoading ? 'Loading threads…' : 'No threads with this agent yet.'}
      </p>
    )
  }

  return (
    <ul className='flex flex-col gap-0.5 p-1.5'>
      {threads.map((s) => {
        const threadId = threadIdOf(s)
        if (!threadId) return null
        const active = threadId === activeThreadId
        if (renamingId === s.session_id) {
          return (
            <li
              key={s.session_id}
              className='flex h-9 items-center rounded-md bg-muted/60 px-2'
            >
              <InlineTitleInput
                initial={s.title?.trim() ?? ''}
                onSave={(title) => onRename(s, title)}
                onClose={() => setRenamingId(null)}
                className='text-sm'
              />
            </li>
          )
        }
        return (
          <li key={s.session_id} className='group/row relative'>
            <button
              type='button'
              onClick={() => onSelect(threadId)}
              aria-current={active ? 'true' : undefined}
              title={threadTitle(s)}
              className={cn(
                'flex h-9 w-full items-center rounded-md ps-2.5 pe-9 text-start text-sm transition-colors hover:bg-muted',
                active
                  ? 'bg-muted font-medium text-foreground'
                  : 'text-muted-foreground'
              )}
            >
              <span className='truncate'>{threadTitle(s)}</span>
            </button>
            <DropdownMenu>
              <DropdownMenuTrigger
                aria-label='Thread actions'
                className='absolute end-0.5 top-0.5 inline-flex size-8 items-center justify-center rounded-md text-muted-foreground opacity-100 hover:bg-background hover:text-foreground md:opacity-0 md:group-focus-within/row:opacity-100 md:group-hover/row:opacity-100 md:data-[state=open]:opacity-100'
              >
                <MoreHorizontal className='size-4' />
              </DropdownMenuTrigger>
              <DropdownMenuContent align='start' sideOffset={4}>
                <DropdownMenuItem onClick={() => setRenamingId(s.session_id)}>
                  <Pencil />
                  Rename
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  variant='destructive'
                  onClick={() => onDelete(s)}
                >
                  <Trash2 />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </li>
        )
      })}
    </ul>
  )
}
