import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { toast } from 'sonner'
import {
  useCreateSession,
  useDeleteSession,
  useSessions,
} from '@/api/sessions'
import { useAuth } from '@/hooks/use-auth'
import { CHAT_APP_NAME } from '@/lib/constants'
import { sessionAgentName, sessionTitle } from '@/lib/session-title'
import { DeleteDialog } from '@/components/delete-dialog'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { AgentSelector } from './agent-selector'
import { ChatWindow } from './chat-window'
import type { SessionInfo } from '@/types/api'

export function ChatPage() {
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuth()
  const navigate = useNavigate()
  const search = useSearch({ from: '/_authenticated/chat' })
  const userId = user?.id ?? ''

  const sessionsQuery = useSessions(
    {
      app_name: CHAT_APP_NAME,
      user_id: userId || undefined,
      page_size: 100,
    },
    { enabled: !!userId }
  )
  const createMutation = useCreateSession()
  const deleteMutation = useDeleteSession()

  const sessions = useMemo(
    () => sessionsQuery.data?.sessions ?? [],
    [sessionsQuery.data]
  )

  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null)

  const wantsNewChat = search.new === 1
  const requestedSessionId = search.session ?? null
  const requestedAgent = search.agent ?? null

  // Quick-start links (/chat?new=1&agent=x) create the session immediately.
  // The guard ref makes this fire once; the redirect replaces the URL so a
  // refresh lands on the created session instead of creating another one.
  const autoCreatedRef = useRef(false)
  useEffect(() => {
    if (!wantsNewChat || !requestedAgent || !userId || autoCreatedRef.current)
      return
    autoCreatedRef.current = true
    void handleCreate(requestedAgent)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wantsNewChat, requestedAgent, userId])

  const activeSession = useMemo(() => {
    if (wantsNewChat) return null
    if (requestedSessionId) {
      return sessions.find((s) => s.session_id === requestedSessionId) ?? null
    }
    return sessions[0] ?? null
  }, [wantsNewChat, requestedSessionId, sessions])

  const activeAgent = activeSession
    ? (sessionAgentName(activeSession.state) ?? null)
    : null

  async function handleCreate(agentName: string) {
    if (!userId) {
      toast.error('Missing user context; please re-login.')
      return
    }
    try {
      const resp = await createMutation.mutateAsync({
        app_name: CHAT_APP_NAME,
        user_id: userId,
        state: { agent_name: agentName },
      })
      navigate({
        to: '/chat',
        search: { session: resp.session.session_id },
        replace: true,
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create chat')
    }
  }

  function handleDeleteConfirm() {
    if (!deleteTarget) return
    deleteMutation.mutate(
      {
        app_name: deleteTarget.app_name,
        user_id: deleteTarget.user_id,
        session_id: deleteTarget.session_id,
      },
      {
        onSuccess: () => {
          toast.success('Chat deleted')
          setDeleteTarget(null)
          navigate({ to: '/chat', search: { new: 1 }, replace: true })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  let content: React.ReactNode
  if (!userId) {
    content = (
      <div className='flex flex-1 items-center justify-center'>
        <p className='text-sm text-muted-foreground'>
          {isAuthenticated || isAuthLoading
            ? 'Loading chat…'
            : 'Sign-in required to use chat.'}
        </p>
      </div>
    )
  } else if (
    !activeSession &&
    (sessionsQuery.isLoading || (wantsNewChat && requestedAgent))
  ) {
    // A specific session was requested but the list is still loading — avoid
    // flashing the agent selector before we know whether it exists.
    content = (
      <div className='flex flex-1 items-center justify-center'>
        <p className='text-sm text-muted-foreground'>Loading chat…</p>
      </div>
    )
  } else if (!activeSession) {
    // New-chat / empty state — centered agent selector
    content = (
      <div className='flex min-h-0 flex-1 flex-col'>
        <div className='scrollbar-thin flex flex-1 items-center justify-center overflow-y-auto p-4'>
          <AgentSelector
            onPick={(name) => void handleCreate(name)}
            busy={createMutation.isPending}
          />
        </div>
      </div>
    )
  } else {
    content = (
      <>
        <ChatWindow
          session={activeSession}
          userId={userId}
          agentName={activeAgent}
          onDelete={() => setDeleteTarget(activeSession)}
        />
        <DeleteDialog
          open={!!deleteTarget}
          onOpenChange={(open) => !open && setDeleteTarget(null)}
          title='Delete chat'
          description={`Delete chat "${deleteTarget ? sessionTitle(deleteTarget) : ''}"? This cannot be undone.`}
          loading={deleteMutation.isPending}
          onConfirm={handleDeleteConfirm}
        />
      </>
    )
  }

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main fixed fluid className='px-0 py-0'>
        {content}
      </Main>
    </>
  )
}
