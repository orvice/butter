import { useMemo, useState } from 'react'
import { Link, useLocation } from '@tanstack/react-router'
import type { SessionInfo } from '@/types/api'
import { MoreHorizontal, Pencil, Search, SquarePen } from 'lucide-react'
import { useSessions, useUpdateSessionTitle } from '@/api/sessions'
import { useAuthStore } from '@/stores/auth-store'
import { CHAT_APP_NAME } from '@/lib/constants'
import { sessionAgentName, sessionTitle } from '@/lib/session-title'
import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { AgentAvatar } from '@/components/butter/primitives'
import { InlineTitleInput } from '@/components/inline-title-input'

type SessionGroupKey = 'today' | 'week' | 'older'

function sessionGroup(session: SessionInfo): SessionGroupKey {
  if (!session.last_update_time) return 'older'
  const updated = new Date(session.last_update_time)
  const now = new Date()
  const startOfToday = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate()
  )
  if (updated >= startOfToday) return 'today'
  if (now.getTime() - updated.getTime() < 7 * 24 * 60 * 60 * 1000) return 'week'
  return 'older'
}

const GROUP_TITLES: Record<SessionGroupKey, string> = {
  today: 'Today',
  week: 'Previous 7 days',
  older: 'Older',
}

export function NavChatHistory() {
  const user = useAuthStore((state) => state.auth.user)
  const location = useLocation()
  const [query, setQuery] = useState('')
  const [renamingSessionId, setRenamingSessionId] = useState<string | null>(
    null
  )

  const userId = user?.id ?? ''
  const sessionsQuery = useSessions(
    { app_name: CHAT_APP_NAME, user_id: userId || undefined, page_size: 100 },
    { enabled: !!userId }
  )

  const activeSessionId =
    location.pathname === '/chat'
      ? ((location.search as { session?: string }).session ?? null)
      : null

  const filtered = useMemo(() => {
    const sessions = sessionsQuery.data?.sessions ?? []
    const q = query.trim().toLowerCase()
    if (!q) return sessions
    return sessions.filter((s) => sessionTitle(s).toLowerCase().includes(q))
  }, [sessionsQuery.data, query])

  const groups = useMemo(
    () =>
      (['today', 'week', 'older'] as const).map((key) => ({
        key,
        items: filtered.filter((s) => sessionGroup(s) === key),
      })),
    [filtered]
  )

  return (
    <SidebarGroup className='py-1 group-data-[collapsible=icon]:hidden'>
      <SidebarGroupLabel>Chats</SidebarGroupLabel>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            className='h-9 border border-sidebar-border bg-background/60 font-medium shadow-none'
          >
            <Link to='/chat' search={{}}>
              <SquarePen />
              <span>New Chat</span>
            </Link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <div className='relative px-1 py-1.5'>
        <Search className='pointer-events-none absolute start-3.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground' />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder='Search chats'
          className='h-9 w-full rounded-md border border-sidebar-border bg-background/60 py-0 ps-8 pe-2 text-sm outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-sidebar-ring/10'
        />
      </div>
      {filtered.length === 0 ? (
        <p className='px-3 py-4 text-center text-xs text-muted-foreground'>
          {sessionsQuery.isLoading
            ? 'Loading chats…'
            : 'No conversations found.'}
        </p>
      ) : (
        groups.map(
          ({ key, items }) =>
            items.length > 0 && (
              <div key={key}>
                <div className='px-2.5 pt-2.5 pb-1 text-[0.7rem] font-medium text-muted-foreground'>
                  {GROUP_TITLES[key]}
                </div>
                <SidebarMenu>
                  {items.map((s) => (
                    <ConversationRow
                      key={s.session_id}
                      session={s}
                      active={s.session_id === activeSessionId}
                      renaming={s.session_id === renamingSessionId}
                      onRenameStart={() => setRenamingSessionId(s.session_id)}
                      onRenameEnd={() => setRenamingSessionId(null)}
                    />
                  ))}
                </SidebarMenu>
              </div>
            )
        )
      )}
    </SidebarGroup>
  )
}

function ConversationRow({
  session,
  active,
  renaming,
  onRenameStart,
  onRenameEnd,
}: {
  session: SessionInfo
  active: boolean
  renaming: boolean
  onRenameStart: () => void
  onRenameEnd: () => void
}) {
  const renameMutation = useUpdateSessionTitle()
  const agent = sessionAgentName(session.state)

  if (renaming) {
    return (
      <SidebarMenuItem>
        <div className='flex h-10 items-center gap-2 rounded-md bg-sidebar-accent/60 px-2.5'>
          {agent && (
            <AgentAvatar
              name={agent}
              size='sm'
              className='size-4 shrink-0 rounded text-[0.6rem]'
            />
          )}
          <InlineTitleInput
            initial={sessionTitle(session)}
            onSave={async (title) => {
              await renameMutation.mutateAsync({
                app_name: session.app_name,
                user_id: session.user_id,
                session_id: session.session_id,
                title,
              })
            }}
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
        <Link to='/chat' search={{ session: session.session_id }}>
          {agent && (
            <AgentAvatar
              name={agent}
              size='sm'
              className='size-4 shrink-0 rounded text-[0.6rem]'
            />
          )}
          <span className='truncate'>{sessionTitle(session)}</span>
        </Link>
      </SidebarMenuButton>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <SidebarMenuAction
            aria-label='Chat actions'
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
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  )
}
