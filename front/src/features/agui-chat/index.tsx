import {
  createContext,
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type SyntheticEvent,
} from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useSearch } from '@tanstack/react-router'
import type { Agent, SessionInfo } from '@/types/api'
import {
  AssistantRuntimeProvider,
  AuiIf,
  ThreadPrimitive,
  ComposerPrimitive,
  useAui,
  useAuiState,
  type ThreadHistoryAdapter,
} from '@assistant-ui/react'
import {
  useAgUiRuntime,
  useAgUiInterrupts,
  useAgUiSteerAway,
  useAgUiState,
} from '@assistant-ui/react-ag-ui'
import {
  ChevronDown,
  History,
  MessageSquarePlus,
  PanelLeft,
  PlugZap,
  Reply,
  Send,
  Square,
} from 'lucide-react'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { BASE_URL, authHeaders } from '@/api/client'
import {
  useAllSessions,
  useDeleteSession,
  useUpdateSessionTitle,
} from '@/api/sessions'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { useWorkspace } from '@/context/workspace-provider'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Textarea } from '@/components/ui/textarea'
import { ChatAgentContext, type ChatAgent } from '@/components/chat/chat-agent'
import { ThreadErrorBoundary } from '@/components/chat/error-boundary'
import { MarkdownText } from '@/components/chat/markdown'
import { ThreadMessages, type PartComponents } from '@/components/chat/messages'
import {
  AgentHero,
  ChatDisclaimer,
  ThreadSkeleton,
} from '@/components/chat/thread-states'
import { ToolCallView } from '@/components/chat/tool-views'
import { chatAuiConfig } from '@/components/chat/toolkit'
import { DeleteDialog } from '@/components/delete-dialog'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { agentIconUrl } from '@/features/agents/icon-utils'
import { ButterAGUIAgent } from './a2ui/agent'
import { readableReply, submissionPayload } from './a2ui/form'
import { EVENT_NAME, envelopeOp, isA2UIEventValue } from './a2ui/protocol'
import { A2UIStore, useA2UIStore } from './a2ui/store'
import { A2UISurfaceView } from './a2ui/surface-view'
import { errorReporter } from './errors'
import { loadThread, threadRepository } from './history'
import { ThreadList } from './thread-list'
import {
  currentThread,
  newThreadId,
  threadPointerKey,
  writeThreadPointer,
} from './thread-pointer'
import {
  AGUI_APP_NAME,
  THREAD_PAGE_SIZE,
  TITLE_REFRESH_DELAYS_MS,
  agentThreads,
  threadIdOf,
  threadTitle,
} from './threads'

// Every runnable agent can be opened here. enable_agui is not needed: it only
// gates programmatic AG-UI access (API and root tokens), not signed-in users.
function isSelectableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  const runnable =
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  return runnable && !!a.agent_id
}

function makeHttpAgent(agentId: string, threadId: string): ButterAGUIAgent {
  return new ButterAGUIAgent({
    url: `${BASE_URL}/api/agui/${encodeURIComponent(agentId)}`,
    threadId,
    fetch: (url, init) => {
      const headers = {
        ...authHeaders(),
        ...(init.headers as Record<string, string> | undefined),
      }
      return globalThis.fetch(url, { ...init, headers })
    },
  })
}

export function AGUIChatPage() {
  const search = useSearch({ from: '/_authenticated/agui-chat' })
  const agentsQuery = useAgents({ page_size: 200 })
  const agents = useMemo(
    () => (agentsQuery.data?.agents ?? []).filter(isSelectableAgent),
    [agentsQuery.data]
  )

  const [pickedAgentId, setPickedAgentId] = useState<string | null>(
    search.agent ?? null
  )
  const agentId = pickedAgentId ?? agents[0]?.agent_id ?? null
  const agent = useMemo(
    () => agents.find((a) => a.agent_id === agentId) ?? null,
    [agents, agentId]
  )
  const { selectedWorkspaceId } = useWorkspace()
  const userId = useAuthStore((state) => state.auth.user?.id ?? '')

  // The current thread is remembered per workspace, user and agent, so a
  // refresh returns to it; switching any of them selects that context's own
  // thread and discards everything shown for the previous one.
  const pointerKey =
    agentId && userId && selectedWorkspaceId
      ? threadPointerKey(selectedWorkspaceId, userId, agentId)
      : null
  const [threads, setThreads] = useState<Record<string, string>>({})
  const threadId = pointerKey
    ? (threads[pointerKey] ?? currentThread(pointerKey))
    : null
  const selectThread = (id: string) => {
    if (!pointerKey) return
    writeThreadPointer(pointerKey, id)
    setThreads((t) => ({ ...t, [pointerKey]: id }))
  }
  const startNewThread = () => selectThread(newThreadId())

  const httpAgent = useMemo(
    () => (agentId && threadId ? makeHttpAgent(agentId, threadId) : null),
    [agentId, threadId]
  )

  const queryClient = useQueryClient()
  // The server keeps the listing to the caller's own `agui` sessions in this
  // workspace; every page is read, so no thread is dropped.
  const sessionsQuery = useAllSessions(
    {
      app_name: AGUI_APP_NAME,
      user_id: userId || undefined,
      workspace_scoped: true,
      page_size: THREAD_PAGE_SIZE,
    },
    { enabled: !!userId && !!selectedWorkspaceId }
  )
  const agentThreadList = useMemo(
    () =>
      agentId && selectedWorkspaceId
        ? agentThreads(
            sessionsQuery.data?.sessions ?? [],
            selectedWorkspaceId,
            agentId
          )
        : [],
    [sessionsQuery.data, selectedWorkspaceId, agentId]
  )
  const renameMutation = useUpdateSessionTitle()
  const deleteMutation = useDeleteSession()
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null)

  // After every run: a thread's first run creates its session, so refresh
  // the list. The server titles a thread that has none once a run on it
  // succeeds, and that title lands after the run ends, so for an untitled
  // thread the list is read again later to show it.
  const handleRunSettled = () => {
    void queryClient.invalidateQueries({ queryKey: ['sessions'] })
    if (!threadId) return
    const listed = agentThreadList.find((s) => threadIdOf(s) === threadId)
    if (listed?.title?.trim()) return
    for (const delay of TITLE_REFRESH_DELAYS_MS) {
      setTimeout(() => {
        void queryClient.invalidateQueries({ queryKey: ['sessions'] })
      }, delay)
    }
  }

  const handleRename = async (session: SessionInfo, title: string) => {
    await renameMutation.mutateAsync({
      app_name: session.app_name,
      user_id: session.user_id,
      session_id: session.session_id,
      title,
    })
  }

  const handleDeleteConfirm = () => {
    if (!deleteTarget) return
    const target = deleteTarget
    const isCurrent = threadIdOf(target) === threadId
    // Stop a run on the thread before its session goes away.
    if (isCurrent) httpAgent?.abortRun()
    deleteMutation.mutate(
      {
        app_name: target.app_name,
        user_id: target.user_id,
        session_id: target.session_id,
      },
      {
        onSuccess: () => {
          toast.success('Thread deleted')
          setDeleteTarget(null)
          if (isCurrent) startNewThread()
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  if (!httpAgent || !agentId || !threadId) {
    return (
      <>
        <PageHeader />
        <Main fixed fluid className='flex flex-col px-0 py-0'>
          <AgentBar
            agents={agents}
            agentId={agentId}
            onAgentChange={setPickedAgentId}
            isLoading={agentsQuery.isLoading}
          />
          <div className='flex flex-1 items-center justify-center px-6 text-center'>
            <p className='text-sm text-muted-foreground'>
              {agentsQuery.isLoading
                ? 'Loading agents…'
                : 'No runnable agents in this workspace. Create an agent to chat with it here.'}
            </p>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <AGUIChatWithRuntime
        key={`${selectedWorkspaceId}:${userId}:${agentId}:${threadId}`}
        httpAgent={httpAgent}
        agents={agents}
        agentId={agentId}
        threadId={threadId}
        agentName={agent?.name ?? 'Agent'}
        onAgentChange={(id) => setPickedAgentId(id)}
        onNewThread={startNewThread}
        threads={agentThreadList}
        threadsLoading={sessionsQuery.isLoading}
        onSelectThread={selectThread}
        onRenameThread={handleRename}
        onDeleteThread={setDeleteTarget}
        onRunSettled={handleRunSettled}
      />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete thread'
        description={`Delete thread "${deleteTarget ? threadTitle(deleteTarget) : ''}"? Its messages, cards and forms are removed. This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={handleDeleteConfirm}
      />
    </>
  )
}

// A2UIStoreContext gives message parts access to the thread's surfaces.
const A2UIStoreContext = createContext<A2UIStore | null>(null)

function useA2UIContext(): A2UIStore {
  const store = useContext(A2UIStoreContext)
  if (!store) throw new Error('A2UI store missing')
  return store
}

// useOwnedA2UIStore creates and owns one thread's surfaces and feeds every
// butter.a2ui CUSTOM event to it in arrival order. What the thread already
// holds arrives with its history (useThreadHistory).
function useOwnedA2UIStore(httpAgent: ButterAGUIAgent) {
  const [store] = useState(() => new A2UIStore())

  useEffect(() => {
    const sub = httpAgent.subscribe({
      onCustomEvent: ({ event }) => {
        if (event.name === EVENT_NAME) store.apply(event.value)
      },
    })
    return () => sub.unsubscribe()
  }, [httpAgent, store])

  useEffect(() => () => store.dispose(), [store])
  return store
}

// useThreadHistory hydrates the runtime with the thread's conversation. Its
// history and UI snapshot load together: the surfaces land in the store and
// each reply shows the ones it produced. The runtime loads once per thread,
// so StrictMode's rehearsal unmount must not cancel the load; leaving the
// thread only stops retrying.
function useThreadHistory(
  agentId: string,
  threadId: string,
  store: A2UIStore
): ThreadHistoryAdapter {
  const mounted = useRef(true)
  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])
  return useMemo<ThreadHistoryAdapter>(
    () => ({
      async load() {
        try {
          const { history, snapshot } = await loadThread(
            agentId,
            threadId,
            () => mounted.current
          )
          const { repository, placed } = threadRepository(history, snapshot)
          store.markPlaced(placed)
          store.applySnapshot(snapshot)
          return repository
        } catch (err) {
          if (mounted.current) {
            toast.error(
              `Could not restore this thread: ${
                err instanceof Error ? err.message : String(err)
              }`
            )
          }
          return { messages: [] }
        }
      },
      // The server keeps the conversation; nothing to write back.
      async append() {},
    }),
    [agentId, threadId, store]
  )
}

function AGUIChatWithRuntime({
  httpAgent,
  agents,
  agentId,
  threadId,
  onAgentChange,
  onNewThread,
  threads,
  threadsLoading,
  onSelectThread,
  onRenameThread,
  onDeleteThread,
  onRunSettled,
}: {
  httpAgent: ButterAGUIAgent
  agents: Agent[]
  agentId: string
  threadId: string
  agentName: string
  onAgentChange: (id: string) => void
  onNewThread: () => void
  threads: SessionInfo[]
  threadsLoading: boolean
  onSelectThread: (threadId: string) => void
  onRenameThread: (session: SessionInfo, title: string) => Promise<void>
  onDeleteThread: (session: SessionInfo) => void
  onRunSettled: () => void
}) {
  const store = useOwnedA2UIStore(httpAgent)
  const history = useThreadHistory(agentId, threadId, store)
  const [reportError] = useState(() =>
    errorReporter((message) => toast.error(message))
  )
  const runtime = useAgUiRuntime({
    agent: httpAgent,
    onError: reportError,
    adapters: { history },
  })
  const [threadsOpen, setThreadsOpen] = useState(false)

  const runSettledRef = useRef(onRunSettled)
  useEffect(() => {
    runSettledRef.current = onRunSettled
  })
  useEffect(() => {
    const sub = httpAgent.subscribe({
      onRunFinalized: () => {
        runSettledRef.current()
      },
    })
    return () => sub.unsubscribe()
  }, [httpAgent])

  const threadList = (
    <ThreadList
      threads={threads}
      activeThreadId={threadId}
      isLoading={threadsLoading}
      onSelect={(id) => {
        if (id === threadId) {
          setThreadsOpen(false)
          return
        }
        httpAgent.abortRun()
        onSelectThread(id)
      }}
      onRename={onRenameThread}
      onDelete={onDeleteThread}
    />
  )

  return (
    <AssistantRuntimeProvider runtime={runtime} config={chatAuiConfig}>
      <A2UIStoreContext.Provider value={store}>
        <FormSubmitBridge store={store} httpAgent={httpAgent} />
        <PageHeader />
        <Main fixed fluid className='flex flex-col px-0 py-0'>
          <AgentBar
            agents={agents}
            agentId={agentId}
            onAgentChange={(id) => {
              httpAgent.abortRun()
              onAgentChange(id)
            }}
            onNewThread={() => {
              httpAgent.abortRun()
              onNewThread()
            }}
            onOpenThreads={() => setThreadsOpen(true)}
          />
          <div className='flex min-h-0 flex-1'>
            <aside
              aria-label='Threads'
              className='hidden w-60 shrink-0 overflow-y-auto border-e border-border/60 md:block'
            >
              {threadList}
            </aside>
            <div className='flex min-w-0 flex-1 flex-col'>
              <ThreadArea
                agent={agents.find((a) => a.agent_id === agentId)}
                httpAgent={httpAgent}
                onSendError={reportError}
              />
              <SharedStatePanel />
              <ComposerArea httpAgent={httpAgent} onSendError={reportError} />
            </div>
          </div>
        </Main>
        <Sheet open={threadsOpen} onOpenChange={setThreadsOpen}>
          <SheetContent side='left' className='w-72 gap-0 p-0'>
            <SheetHeader className='border-b border-border/60'>
              <SheetTitle>Threads</SheetTitle>
            </SheetHeader>
            <div className='overflow-y-auto'>{threadList}</div>
          </SheetContent>
        </Sheet>
      </A2UIStoreContext.Provider>
    </AssistantRuntimeProvider>
  )
}

// FormSubmitBridge sends form submissions through the chat runtime so the
// resumed run streams into this conversation. The submission shows up as a
// readable user reply; the request itself carries only the form's own
// addressed resume entry (ButterAGUIAgent.resumeNextRunWith).
function FormSubmitBridge({
  store,
  httpAgent,
}: {
  store: A2UIStore
  httpAgent: ButterAGUIAgent
}) {
  const steerAway = useAgUiSteerAway()
  useEffect(() => {
    store.setSubmitter(async (entry, values) => {
      const form = entry.form
      if (!form) throw new Error('not a form')
      httpAgent.resumeNextRunWith(
        form.interruptId,
        submissionPayload(entry.id, form, values)
      )
      try {
        await steerAway(readableReply(form, values))
      } finally {
        httpAgent.clearNextRun()
      }
    })
    return () => store.setSubmitter(undefined)
  }, [store, httpAgent, steerAway])
  return null
}

// useReplies answers open Interrupts the way a form does: the runtime's
// steer-away path shows the reply as the user's message and starts the run
// although Interrupts are open, and the agent decides what that one run's
// resume carries. A failure is reported instead of lost.
function useReplies(
  httpAgent: ButterAGUIAgent,
  onError: (err: unknown) => void
) {
  const steerAway = useAgUiSteerAway()
  const steer = async (reply: string) => {
    try {
      await steerAway(reply)
    } catch (err) {
      onError(err)
    } finally {
      httpAgent.clearNextRun()
    }
  }
  return {
    // answer sends text as the answer to exactly this Interrupt.
    answer(interruptId: string, text: string) {
      httpAgent.resumeNextRunWith(interruptId, text)
      return steer(text)
    },
    // answerOldest sends text as a plain message, which the server takes as
    // the answer to its oldest open Interrupt (ADR-0002).
    answerOldest(text: string) {
      httpAgent.sendNextRunAsMessage()
      return steer(text)
    },
  }
}

function PageHeader() {
  return (
    <Header fixed className='h-14 border-b border-border/60 bg-background/95'>
      <Search className='sm:w-44 md:w-52 lg:w-60 xl:w-72' />
      <div className='ms-auto flex items-center gap-1 sm:gap-2'>
        <ThemeSwitch />
        <ProfileDropdown />
      </div>
    </Header>
  )
}

function AgentBar({
  agents,
  agentId,
  onAgentChange,
  onNewThread,
  onOpenThreads,
  isLoading,
}: {
  agents: Agent[]
  agentId: string | null
  onAgentChange: (id: string) => void
  onNewThread?: () => void
  onOpenThreads?: () => void
  isLoading?: boolean
}) {
  return (
    <div className='flex items-center gap-2 border-b border-border/60 px-4 py-2'>
      {onOpenThreads && (
        <Button
          variant='ghost'
          size='icon'
          className='size-8 md:hidden'
          aria-label='Show threads'
          onClick={onOpenThreads}
        >
          <PanelLeft className='size-4' />
        </Button>
      )}
      <PlugZap className='size-4 text-muted-foreground' />
      <span className='hidden text-sm font-medium whitespace-nowrap sm:inline'>
        AG-UI Chat
      </span>
      <Select
        value={agentId ?? undefined}
        onValueChange={onAgentChange}
        disabled={isLoading}
      >
        <SelectTrigger className='ml-2 h-8 w-56' size='sm'>
          <SelectValue placeholder='Pick an agent' />
        </SelectTrigger>
        <SelectContent>
          {agents.map((a) => (
            <SelectItem key={a.agent_id} value={a.agent_id ?? ''}>
              {a.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {onNewThread && (
        <Button
          variant='ghost'
          size='sm'
          className='ms-auto h-8'
          onClick={onNewThread}
        >
          <MessageSquarePlus className='size-4' />
          New thread
        </Button>
      )}
    </div>
  )
}

// A2UIDataPart places a surface where its create event arrived in the
// message; later updates of the same surface re-render it in place.
function A2UIDataPart({ data }: { data: unknown }) {
  const store = useA2UIContext()
  const locked = useAuiState((s) => s.thread.isRunning)
  if (
    !isA2UIEventValue(data) ||
    envelopeOp(data.envelope) !== 'createSurface'
  ) {
    return null
  }
  return (
    <A2UISurfaceView
      store={store}
      surfaceId={data.surfaceId}
      locked={locked}
      className='my-2 max-w-full'
    />
  )
}

// The parts of a reply: Markdown text, tool calls (the shared toolkit draws
// render_ui and adk_request_input, ToolCallView the rest), and each A2UI
// surface where its create event landed.
const REPLY_PARTS: PartComponents = {
  Text: MarkdownText,
  tools: { Fallback: ToolCallView },
  data: { by_name: { [EVENT_NAME]: A2UIDataPart } },
}

// restoredSurfaces lists what the UI snapshot brought back that no reply of
// the restored history shows.
function restoredSurfaces(store: A2UIStore) {
  return store
    .list()
    .filter(
      (e) => e.origin === 'snapshot' && !e.deleted && !store.isPlaced(e.id)
    )
}

// RestoredSurfaces shows the restored surfaces no reply shows, so each
// surface says where it came from.
function RestoredSurfaces() {
  const store = useA2UIStore(useA2UIContext())
  const locked = useAuiState((s) => s.thread.isRunning)
  const restored = restoredSurfaces(store)
  if (restored.length === 0) return null
  return (
    <section
      aria-label='Restored from this conversation'
      className='flex flex-col gap-3'
    >
      <p className='flex items-center gap-1.5 text-xs text-muted-foreground'>
        <History className='size-3.5' />
        Restored from earlier in this conversation
      </p>
      {restored.map((entry) => (
        <div key={entry.id} className='flex flex-col gap-1'>
          <p className='text-xs text-muted-foreground'>
            {entry.kind === 'form'
              ? 'Waiting for your answer'
              : 'Result card from an earlier reply'}
          </p>
          <A2UISurfaceView store={store} surfaceId={entry.id} locked={locked} />
        </div>
      ))}
    </section>
  )
}

// The runtime starts loading a thread's history only after it mounts, so
// for a moment every thread looks empty, and a quick load shows its skeleton
// only briefly. The skeleton and the hero fade in after a short delay, so
// neither flashes.
const FADE_IN_LATE = 'animate-in fade-in fill-mode-both delay-150 duration-300'

// ThreadArea shows the conversation. Each reply carries the agent's avatar
// and name; a skeleton stands in while the history loads, and a thread
// without messages introduces the agent.
function ThreadArea({
  agent,
  httpAgent,
  onSendError,
}: {
  agent: Agent | undefined
  httpAgent: ButterAGUIAgent
  onSendError: (err: unknown) => void
}) {
  const chatAgent = useMemo<ChatAgent>(
    () => ({
      name: agent?.name || 'Agent',
      iconUrl: (agent && agentIconUrl(agent)) || undefined,
    }),
    [agent]
  )
  return (
    <ChatAgentContext.Provider value={chatAgent}>
      <ThreadErrorBoundary>
        <ThreadPrimitive.Root className='flex-1 overflow-y-auto px-4 py-4'>
          <ThreadPrimitive.Viewport className='mx-auto flex max-w-3xl flex-col gap-3'>
            <AuiIf condition={(s) => s.thread.isLoading}>
              <ThreadSkeleton className={FADE_IN_LATE} />
            </AuiIf>
            <AuiIf condition={(s) => s.thread.isEmpty}>
              <EmptyThread agent={chatAgent} />
            </AuiIf>
            <RestoredSurfaces />
            <div>
              <ThreadMessages replyParts={REPLY_PARTS} />
            </div>
            <InterruptPrompts httpAgent={httpAgent} onSendError={onSendError} />
            <AuiIf condition={(s) => s.thread.isRunning}>
              <p className='text-xs text-muted-foreground'>Running…</p>
            </AuiIf>
          </ThreadPrimitive.Viewport>
        </ThreadPrimitive.Root>
      </ThreadErrorBoundary>
    </ChatAgentContext.Provider>
  )
}

// EmptyThread introduces the agent of a thread without messages, unless
// cards restored from the thread are shown instead.
function EmptyThread({ agent }: { agent: ChatAgent }) {
  const store = useA2UIStore(useA2UIContext())
  if (restoredSurfaces(store).length > 0) return null
  return (
    <AgentHero
      name={agent.name}
      iconUrl={agent.iconUrl}
      className={FADE_IN_LATE}
    />
  )
}

// InterruptPrompts asks each open question that has no form. An answer goes
// to its own Interrupt alone; the others stay open.
function InterruptPrompts({
  httpAgent,
  onSendError,
}: {
  httpAgent: ButterAGUIAgent
  onSendError: (err: unknown) => void
}) {
  const interrupts = useAgUiInterrupts()
  const replies = useReplies(httpAgent, onSendError)
  const store = useA2UIStore(useA2UIContext())

  // An Interrupt with a form is answered through the form; the text prompt
  // is only for the rest.
  const textOnly = interrupts.filter((i) => !store.formFor(i.id))
  if (textOnly.length === 0) return null

  return (
    <>
      {textOnly.map((interrupt) => (
        <InterruptPrompt
          key={interrupt.id}
          interrupt={interrupt}
          onResolve={(answer) => void replies.answer(interrupt.id, answer)}
        />
      ))}
    </>
  )
}

function InterruptPrompt({
  interrupt,
  onResolve,
}: {
  interrupt: { id: string; message?: string }
  onResolve: (answer: string) => void
}) {
  const [answer, setAnswer] = useState('')
  const questionId = useId()
  return (
    <div
      role='group'
      aria-labelledby={questionId}
      className='max-w-[85%] rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm'
    >
      <p id={questionId} className='font-medium'>
        {interrupt.message || 'The workflow needs your input.'}
      </p>
      <div className='mt-2 flex items-end gap-2'>
        <Textarea
          value={answer}
          onChange={(e) => setAnswer(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              if (answer.trim() !== '') onResolve(answer.trim())
            }
          }}
          rows={1}
          className='min-h-0 resize-none text-sm'
          placeholder='Type your answer…'
        />
        <Button
          size='sm'
          disabled={answer.trim() === ''}
          onClick={() => onResolve(answer.trim())}
        >
          Answer
        </Button>
      </div>
    </div>
  )
}

function SharedStatePanel() {
  const state = useAgUiState()
  const [stateOpen, setStateOpen] = useState(false)

  if (state === undefined) return null

  return (
    <div className='border-t border-border/60 px-4 py-2'>
      <div className='mx-auto max-w-3xl'>
        <button
          type='button'
          className='flex items-center gap-1 text-xs text-muted-foreground'
          onClick={() => setStateOpen((v) => !v)}
        >
          <ChevronDown
            className={cn(
              'size-3 transition-transform',
              !stateOpen && '-rotate-90'
            )}
          />
          Shared state
        </button>
        {stateOpen && (
          <pre className='mt-1 max-h-40 overflow-auto rounded border border-border bg-muted/40 p-2 text-xs'>
            {JSON.stringify(state, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

// ComposerArea sends the user's messages. While questions are open the
// runtime refuses an ordinary send, so the text goes out as a plain message
// instead, and the server takes it as the answer to its earliest open
// question (ADR-0002). A hint says so.
function ComposerArea({
  httpAgent,
  onSendError,
}: {
  httpAgent: ButterAGUIAgent
  onSendError: (err: unknown) => void
}) {
  const answering = useAgUiInterrupts().length > 0
  const replies = useReplies(httpAgent, onSendError)
  const aui = useAui()
  const hintId = useId()

  // Takes over Enter (the form's submit) and the Send button.
  const sendAsAnswer = (e: SyntheticEvent) => {
    if (!answering) return
    e.preventDefault()
    const text = aui.composer.getState().text.trim()
    if (text === '') return
    aui.composer.setText('')
    void replies.answerOldest(text)
  }

  return (
    <div className='border-t border-border/60 px-4 py-3'>
      {answering && (
        <p
          id={hintId}
          className='mx-auto mb-2 flex max-w-3xl items-center gap-1.5 text-xs text-muted-foreground'
        >
          <Reply className='size-3.5' />
          Sending answers the earliest open question.
        </p>
      )}
      <ComposerPrimitive.Root
        onSubmit={sendAsAnswer}
        className='mx-auto flex max-w-3xl items-end gap-2'
      >
        <ComposerPrimitive.Input
          autoFocus
          placeholder='Message the agent over AG-UI…'
          aria-describedby={answering ? hintId : undefined}
          rows={2}
          className='min-h-0 flex-1 resize-none rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring'
        />
        <AuiIf condition={(s) => s.thread.isRunning}>
          <ComposerPrimitive.Cancel asChild>
            <Button variant='outline' size='icon' aria-label='Stop'>
              <Square className='size-4' />
            </Button>
          </ComposerPrimitive.Cancel>
        </AuiIf>
        <AuiIf condition={(s) => !s.thread.isRunning}>
          <ComposerPrimitive.Send asChild onClick={sendAsAnswer}>
            <Button size='icon' aria-label='Send'>
              <Send className='size-4' />
            </Button>
          </ComposerPrimitive.Send>
        </AuiIf>
      </ComposerPrimitive.Root>
      <ChatDisclaimer className='mx-auto max-w-3xl' />
    </div>
  )
}
