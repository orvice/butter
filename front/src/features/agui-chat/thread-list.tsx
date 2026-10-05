import { useId, useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { MoreHorizontal, Pencil, Search, SquarePen, Trash2 } from 'lucide-react'
import { useAgents } from '@/api/agents'
import { useUpdateSessionTitle } from '@/api/sessions'
import { groupByRecency } from '@/lib/recency-groups'
import { cn } from '@/lib/utils'
import { useWorkspace } from '@/context/workspace-provider'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { AgentAvatar } from '@/components/butter/primitives'
import { InlineTitleInput } from '@/components/inline-title-input'
import { agentIconUrl } from '@/features/agents/icon-utils'
import { useThreadDelete } from './thread-delete'
import { searchThreads, threadRows, type ThreadRow } from './threads'
import { useChatSearch, useThreadSessions } from './use-threads'

// NavThreads is Chat's history in the sidebar: the caller's threads in this
// workspace, whatever their agent, grouped by when they were last updated,
// with title search. Each row shows its agent's avatar and title and links to
// its thread; the thread Chat has open is highlighted. On narrow screens it
// lives in the sidebar's sheet, which a row or New thread closes.
export function NavThreads() {
  const { selectedWorkspaceId } = useWorkspace()
  const { setOpenMobile } = useSidebar()
  const chat = useChatSearch()
  const sessionsQuery = useThreadSessions()
  // The agents give each row its avatar and the name an untitled thread
  // shows. Chat reads the same list.
  const agentsQuery = useAgents(
    { page_size: 200 },
    { enabled: !!selectedWorkspaceId }
  )
  const renameMutation = useUpdateSessionTitle()
  const { requestDelete } = useThreadDelete()
  const [query, setQuery] = useState('')
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const labelId = useId()

  const rows = useMemo(
    () =>
      threadRows(
        sessionsQuery.data?.sessions ?? [],
        selectedWorkspaceId,
        agentsQuery.data?.agents ?? []
      ),
    [sessionsQuery.data, selectedWorkspaceId, agentsQuery.data]
  )
  const groups = useMemo(
    () =>
      groupByRecency(
        searchThreads(rows, query),
        (row) => row.session.last_update_time,
        new Date()
      ),
    [rows, query]
  )
  const loading = sessionsQuery.isLoading || agentsQuery.isLoading

  const activeThreadId = chat?.thread ?? null
  // New thread keeps the agent in use: the open thread's, or the draft's.
  const newThreadAgent = !chat
    ? undefined
    : chat.thread
      ? rows.find((row) => row.threadId === chat.thread)?.agentId
      : chat.agent
  const closeSheet = () => setOpenMobile(false)

  let note = 'No threads found.'
  if (loading) note = 'Loading threads…'
  else if (sessionsQuery.isError) note = 'Failed to load threads.'

  return (
    <SidebarGroup
      role='navigation'
      aria-labelledby={labelId}
      className='py-1 group-data-[collapsible=icon]:hidden'
    >
      <SidebarGroupLabel id={labelId}>Threads</SidebarGroupLabel>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            className='h-9 border border-sidebar-border bg-background/60 font-medium shadow-none'
          >
            <Link
              to='/chat'
              search={newThreadAgent ? { agent: newThreadAgent } : {}}
              onClick={closeSheet}
            >
              <SquarePen />
              <span>New thread</span>
            </Link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <div className='relative px-1 py-1.5'>
        <Search className='pointer-events-none absolute start-3.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground' />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder='Search threads'
          aria-label='Search threads'
          className='h-9 w-full rounded-md border border-sidebar-border bg-background/60 py-0 ps-8 pe-2 text-sm outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-sidebar-ring/10'
        />
      </div>
      {loading || groups.length === 0 ? (
        <p className='px-3 py-4 text-center text-xs text-muted-foreground'>
          {note}
        </p>
      ) : (
        groups.map(({ key, title, items }) => (
          <div key={key} role='group' aria-labelledby={`${labelId}-${key}`}>
            <div
              id={`${labelId}-${key}`}
              className='px-2.5 pt-2.5 pb-1 text-[0.7rem] font-medium text-muted-foreground'
            >
              {title}
            </div>
            <SidebarMenu>
              {items.map((row) => (
                <ThreadRowItem
                  key={row.session.session_id}
                  row={row}
                  active={row.threadId === activeThreadId}
                  renaming={row.session.session_id === renamingId}
                  onNavigate={closeSheet}
                  onRenameStart={() => setRenamingId(row.session.session_id)}
                  onRenameEnd={() => setRenamingId(null)}
                  onRename={async (title) => {
                    await renameMutation.mutateAsync({
                      app_name: row.session.app_name,
                      user_id: row.session.user_id,
                      session_id: row.session.session_id,
                      title,
                    })
                  }}
                  onDelete={() =>
                    requestDelete({
                      session: {
                        app_name: row.session.app_name,
                        user_id: row.session.user_id,
                        session_id: row.session.session_id,
                      },
                      title: row.title,
                      agentId: row.agentId,
                    })
                  }
                />
              ))}
            </SidebarMenu>
          </div>
        ))
      )}
    </SidebarGroup>
  )
}

function ThreadRowItem({
  row,
  active,
  renaming,
  onNavigate,
  onRenameStart,
  onRenameEnd,
  onRename,
  onDelete,
}: {
  row: ThreadRow
  active: boolean
  renaming: boolean
  onNavigate: () => void
  onRenameStart: () => void
  onRenameEnd: () => void
  onRename: (title: string) => Promise<void>
  onDelete: () => void
}) {
  const avatar = (
    <AgentAvatar
      name={row.agent?.name ?? ''}
      iconUrl={(row.agent && agentIconUrl(row.agent)) || undefined}
      size='sm'
      className='size-4 shrink-0 rounded text-[0.6rem]'
    />
  )

  if (renaming) {
    return (
      <SidebarMenuItem>
        <div className='flex h-10 items-center gap-2 rounded-md bg-sidebar-accent/60 px-2.5'>
          {avatar}
          <InlineTitleInput
            initial={row.title}
            onSave={onRename}
            onClose={onRenameEnd}
          />
        </div>
      </SidebarMenuItem>
    )
  }

  return (
    <SidebarMenuItem className='group/row'>
      <SidebarMenuButton
        asChild
        isActive={active}
        className={cn(
          'h-10 ps-2.5 pe-10',
          !active && 'text-sidebar-foreground/75'
        )}
      >
        <Link
          to='/chat'
          search={{ thread: row.threadId }}
          onClick={onNavigate}
        >
          {avatar}
          <span className='truncate'>{row.title}</span>
        </Link>
      </SidebarMenuButton>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <SidebarMenuAction
            aria-label='Thread actions'
            className='end-0.5 top-0.5 size-9 opacity-100 md:opacity-0 md:group-focus-within/row:opacity-100 md:group-hover/row:opacity-100 md:data-[state=open]:opacity-100'
          >
            <MoreHorizontal className='size-4' />
          </SidebarMenuAction>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='start' sideOffset={4}>
          <DropdownMenuItem onClick={onRenameStart}>
            <Pencil />
            Rename
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem variant='destructive' onClick={onDelete}>
            <Trash2 />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  )
}
