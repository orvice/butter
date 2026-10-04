import { useState } from 'react'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { AgentAvatar } from '@/components/butter/primitives'
import { InlineTitleInput } from '@/components/inline-title-input'

const ICON_BUTTON =
  'inline-flex size-9 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100'

// ThreadHeader names the open thread and the agent it runs with: the title
// (the agent's name until the thread has one), renamed in place, and a menu
// to delete the thread.
export function ThreadHeader({
  title,
  agentName,
  agentIconUrl,
  threadId,
  onRename,
  onDelete,
}: {
  title: string
  agentName: string
  agentIconUrl?: string
  threadId: string
  onRename: (title: string) => Promise<void>
  onDelete: () => void
}) {
  const [editing, setEditing] = useState(false)

  return (
    <header className='flex h-12 shrink-0 items-center justify-between gap-2 border-b border-border/60 bg-background/95 px-2.5 sm:px-4'>
      <div className='flex min-w-0 flex-1 items-center gap-2.5'>
        <AgentAvatar name={agentName} iconUrl={agentIconUrl} size='sm' />
        <span className='flex max-w-md min-w-0 flex-1 flex-col'>
          {editing ? (
            <InlineTitleInput
              initial={title}
              onSave={onRename}
              onClose={() => setEditing(false)}
              className='text-sm font-medium'
            />
          ) : (
            <span className='flex min-w-0 items-center gap-1'>
              <h1 className='truncate text-sm leading-tight font-semibold'>
                {title}
              </h1>
              <button
                type='button'
                onClick={() => setEditing(true)}
                aria-label='Rename thread'
                title='Rename thread'
                className={ICON_BUTTON}
              >
                <Pencil className='size-3.5' />
              </button>
            </span>
          )}
          <span
            title={threadId}
            className='truncate font-mono text-[0.65rem] leading-tight text-muted-foreground/80'
          >
            {agentName}
            <span className='hidden sm:inline'> / {threadId.slice(0, 8)}</span>
          </span>
        </span>
      </div>

      <DropdownMenu>
        <DropdownMenuTrigger
          aria-label='Thread options'
          className={ICON_BUTTON}
        >
          <MoreHorizontal className='size-4' />
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' sideOffset={6}>
          <DropdownMenuItem variant='destructive' onClick={onDelete}>
            <Trash2 />
            Delete thread
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
