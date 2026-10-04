import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import type { SessionInfo } from '@/types/api'
import { MessageSquarePlus, MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { InlineTitleInput } from '@/components/inline-title-input'
import { threadIdOf, threadTitle } from './threads'

interface ThreadListProps {
  threads: SessionInfo[]
  activeThreadId: string | null
  isLoading: boolean
  // agentId is the agent whose threads are listed, if one is chosen.
  agentId: string | null
  onRename: (session: SessionInfo, title: string) => Promise<void>
  onDelete: (session: SessionInfo) => void
}

// ThreadsPanel holds the in-page thread list until the sidebar lists threads
// (#398): a column beside the chat on wide screens, a drawer on narrow ones.
// Its New thread link opens a new-chat draft with the listed agent.
export function ThreadsPanel({
  drawerOpen,
  onDrawerOpenChange,
  ...props
}: ThreadListProps & {
  drawerOpen: boolean
  onDrawerOpenChange: (open: boolean) => void
}) {
  const close = () => onDrawerOpenChange(false)
  const content = (
    <>
      <div className='px-1.5 pt-1.5'>
        <Link
          to='/agui-chat'
          search={props.agentId ? { agent: props.agentId } : {}}
          onClick={close}
          className='flex h-9 w-full items-center gap-2 rounded-md border border-border/70 px-2.5 text-sm font-medium transition-colors hover:bg-muted'
        >
          <MessageSquarePlus className='size-4' />
          New thread
        </Link>
      </div>
      <ThreadList {...props} onNavigate={close} />
    </>
  )
  return (
    <>
      <aside
        aria-label='Threads'
        className='hidden w-60 shrink-0 overflow-y-auto border-e border-border/60 md:block'
      >
        {content}
      </aside>
      <Sheet open={drawerOpen} onOpenChange={onDrawerOpenChange}>
        <SheetContent side='left' className='w-72 gap-0 p-0'>
          <SheetHeader className='border-b border-border/60'>
            <SheetTitle>Threads</SheetTitle>
          </SheetHeader>
          <div className='overflow-y-auto'>{content}</div>
        </SheetContent>
      </Sheet>
    </>
  )
}

// ThreadList shows the caller's threads with one agent, newest first. Each
// row links to its thread (?thread=). A new thread has no session until its
// first run, so it appears here only after that.
export function ThreadList({
  threads,
  activeThreadId,
  isLoading,
  agentId,
  onNavigate,
  onRename,
  onDelete,
}: ThreadListProps & { onNavigate?: () => void }) {
  const [renamingId, setRenamingId] = useState<string | null>(null)

  if (threads.length === 0) {
    let note = 'No threads with this agent yet.'
    if (!agentId) note = 'Choose an agent to see its threads.'
    else if (isLoading) note = 'Loading threads…'
    return (
      <p className='px-3 py-4 text-center text-xs text-muted-foreground'>
        {note}
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
            <Link
              to='/agui-chat'
              search={{ thread: threadId }}
              onClick={onNavigate}
              title={threadTitle(s)}
              className={cn(
                'flex h-9 w-full items-center rounded-md ps-2.5 pe-9 text-start text-sm transition-colors hover:bg-muted',
                active
                  ? 'bg-muted font-medium text-foreground'
                  : 'text-muted-foreground'
              )}
            >
              <span className='truncate'>{threadTitle(s)}</span>
            </Link>
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
