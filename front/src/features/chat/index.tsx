import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import type { Agent, SessionInfo } from '@/types/api'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { submitAgentInvocation } from '@/api/chat'
import { useDeleteSession, useSessions } from '@/api/sessions'
import { CHAT_APP_NAME, CHAT_LAST_AGENT_PREFIX } from '@/lib/constants'
import { newClientID } from '@/lib/client-id'
import {
  sessionAgentID,
  sessionAgentName,
  sessionTitle,
} from '@/lib/session-title'
import { useAuth } from '@/hooks/use-auth'
import { useWorkspace } from '@/hooks/use-workspace'
import { DeleteDialog } from '@/components/delete-dialog'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { AgentSelector } from './agent-selector'
import { ChatWindow } from './chat-window'
import { DraftComposer } from './draft-composer'

function isRunnableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  return (
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  )
}

function lastAgentKey(workspaceId: string): string {
  return `${CHAT_LAST_AGENT_PREFIX}${workspaceId}`
}

export function ChatPage() {
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuth()
  const { selectedWorkspaceId } = useWorkspace()
  const navigate = useNavigate()
  const search = useSearch({ from: '/_authenticated/chat' })
  const userId = user?.id ?? ''

  const sessionsQuery = useSessions(
    {
      app_name: CHAT_APP_NAME,
      user_id: userId || undefined,
      page_size: 100,
      workspace_scoped: true,
    },
    { enabled: !!userId }
  )
  const deleteMutation = useDeleteSession()

  const sessions = useMemo(
    () => sessionsQuery.data?.sessions ?? [],
    [sessionsQuery.data]
  )

  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null)
  const [draftSubmitting, setDraftSubmitting] = useState(false)

  // Normalize legacy ?new=1 to plain /chat by stripping the param.
  const wantsLegacyNew = search.new === 1
  const normalizedRef = useRef(false)
  useEffect(() => {
    if (wantsLegacyNew && !normalizedRef.current) {
      normalizedRef.current = true
      navigate({
        to: '/chat',
        search: search.agent ? { agent: search.agent } : {},
        replace: true,
      })
    }
  }, [wantsLegacyNew, search.agent, navigate])

  const requestedSessionId = search.session ?? null
  const requestedAgent = search.agent ?? null

  const agentsQuery = useAgents({ page_size: 200 }, { enabled: !!userId })
  const allAgents = useMemo(
    () => (agentsQuery.data?.agents ?? []).filter(isRunnableAgent),
    [agentsQuery.data],
  )

  // Resolve the selected draft agent from URL, localStorage, or default.
  const resolveInitialAgent = useCallback((): Agent | null => {
    if (requestedAgent && allAgents.length > 0) {
      const match = allAgents.find((a) => a.agent_id === requestedAgent)
      if (match) return match
    }
    if (selectedWorkspaceId) {
      const saved = localStorage.getItem(lastAgentKey(selectedWorkspaceId))
      if (saved) {
        const match = allAgents.find((a) => a.agent_id === saved)
        if (match) return match
        localStorage.removeItem(lastAgentKey(selectedWorkspaceId))
      }
    }
    return null
  }, [requestedAgent, allAgents, selectedWorkspaceId])

  const [draftAgent, setDraftAgent] = useState<Agent | null>(null)
  const agentsLoadedRef = useRef(false)
  useEffect(() => {
    if (allAgents.length > 0 && !agentsLoadedRef.current) {
      agentsLoadedRef.current = true
      setDraftAgent(resolveInitialAgent())
    }
  }, [allAgents, resolveInitialAgent])

  // Re-resolve when ?agent= param changes.
  const prevRequestedAgent = useRef(requestedAgent)
  useEffect(() => {
    if (requestedAgent !== prevRequestedAgent.current) {
      prevRequestedAgent.current = requestedAgent
      if (requestedAgent && allAgents.length > 0) {
        const match = allAgents.find((a) => a.agent_id === requestedAgent)
        if (match) {
          const frame = globalThis.requestAnimationFrame(() => setDraftAgent(match))
          return () => globalThis.cancelAnimationFrame(frame)
        }
      }
    }
  }, [requestedAgent, allAgents])

  function handlePickAgent(agent: Agent) {
    setDraftAgent(agent)
    if (selectedWorkspaceId && agent.agent_id) {
      localStorage.setItem(lastAgentKey(selectedWorkspaceId), agent.agent_id)
    }
  }

  // Determine if we're viewing an existing session.
  // Plain /chat (no ?session=) → show the draft view, never auto-activate.
  const activeSession = useMemo(() => {
    if (!requestedSessionId) return null
    return sessions.find((s) => s.session_id === requestedSessionId) ?? null
  }, [requestedSessionId, sessions])

  const activeAgent = activeSession
    ? (sessionAgentName(activeSession.state) ?? null)
    : null
  const activeAgentId = activeSession
    ? (sessionAgentID(activeSession.state) ?? null)
    : null

  async function handleDraftSend(
    message: string,
    agent: Agent,
  ): Promise<void> {
    if (!userId) {
      toast.error('Missing user context; please re-login.')
      return
    }
    if (draftSubmitting) return
    setDraftSubmitting(true)
    try {
      const resp = await submitAgentInvocation({
        request_id: newClientID(),
        agent_id: agent.agent_id ?? '',
        message,
      })
      await sessionsQuery.refetch()
      navigate({
        to: '/chat',
        search: {
          session: resp.session_id,
          pending_message: message,
          invocation: resp.invocation_id,
        },
        replace: true,
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create chat')
    } finally {
      setDraftSubmitting(false)
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
          navigate({ to: '/chat', search: {}, replace: true })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  const isDraftView = !requestedSessionId

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
  } else if (isDraftView) {
    content = (
      <DraftComposer
        agent={draftAgent}
        agentSelector={
          <AgentSelector selected={draftAgent} onPick={handlePickAgent} />
        }
        onSend={(message) => {
          if (!draftAgent) return
          void handleDraftSend(message, draftAgent)
        }}
        busy={draftSubmitting}
      />
    )
  } else if (
    !activeSession &&
    sessionsQuery.isLoading
  ) {
    content = (
      <div className='flex flex-1 items-center justify-center'>
        <p className='text-sm text-muted-foreground'>Loading chat…</p>
      </div>
    )
  } else if (!activeSession) {
    content = (
      <div className='flex flex-1 items-center justify-center'>
        <p className='text-sm text-muted-foreground'>
          Session not found. <button type="button" className="text-primary underline" onClick={() => navigate({ to: '/chat', search: {}, replace: true })}>Start a new chat</button>
        </p>
      </div>
    )
  } else {
    content = (
      <>
        <ChatWindow
          session={activeSession}
          userId={userId}
          agentName={activeAgent}
          agentId={activeAgentId}
          onDelete={() => setDeleteTarget(activeSession)}
          onInvocationAccepted={(invocationId, message) => {
            navigate({
              to: '/chat',
              search: {
                session: activeSession.session_id,
                invocation: invocationId,
                pending_message: message,
              },
              replace: true,
            })
          }}
          pendingMessage={search.pending_message ?? undefined}
          initialInvocationId={search.invocation ?? undefined}
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
      <Header
        fixed
        className='h-14 border-b border-border/60 bg-background/95 [&_[data-slot=sidebar-trigger]]:size-10 [&_[data-slot=sidebar-trigger]]:scale-100'
      >
        <Search className='max-sm:size-10 max-sm:flex-none max-sm:justify-center max-sm:px-0 sm:w-44 md:w-52 lg:w-60 xl:w-72 max-sm:[&_svg]:static max-sm:[&_svg]:translate-y-0 max-sm:[&>span]:sr-only' />
        <div className='ms-auto flex items-center gap-1 sm:gap-2 [&>button]:size-10 [&>button]:scale-100'>
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
