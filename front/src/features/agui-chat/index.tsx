import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type ReactNode,
  type RefObject,
  type SyntheticEvent,
} from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate, useSearch } from '@tanstack/react-router'
import type { Agent } from '@/types/api'
import {
  AssistantRuntimeProvider,
  AuiIf,
  ThreadPrimitive,
  ComposerPrimitive,
  useAui,
  useAuiEvent,
  useAuiState,
  type AssistantRuntime,
  type CreateAppendMessage,
  type ThreadHistoryAdapter,
} from '@assistant-ui/react'
import {
  useAgUiRuntime,
  useAgUiInterrupts,
  useAgUiSteerAway,
  useAgUiState,
} from '@assistant-ui/react-ag-ui'
import { ChevronDown, History, Reply, Send, Square } from 'lucide-react'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { BASE_URL, authHeaders } from '@/api/client'
import { useSessionInfo, useUpdateSessionTitle } from '@/api/sessions'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { useWorkspace } from '@/context/workspace-provider'
import { Button } from '@/components/ui/button'
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
import { AgentSelector } from './agent-selector'
import {
  draftAgent,
  isSelectableAgent,
  readLastAgent,
  rememberLastAgent,
} from './agents'
import {
  ImageAttachmentAdapter,
  attachmentRefusal,
  imageAttachment,
  sendAll,
} from './attachments'
import { AttachImagesButton, ComposerImages } from './composer-attachments'
import { DraftView, type DraftMessage } from './draft-view'
import { errorReporter, threadRefusal } from './errors'
import { loadThread, threadRepository } from './history'
import {
  ThreadLoadFailed,
  ThreadLoading,
  ThreadNotFound,
} from './open-thread-states'
import { useThreadDelete, type SessionAddress } from './thread-delete'
import { ThreadHeader } from './thread-header'
import {
  AGUI_APP_NAME,
  TITLE_REFRESH_DELAYS_MS,
  newThreadId,
  resolveThreadView,
  sessionIdOf,
  threadIdOf,
} from './threads'
import { useThreadSessions } from './use-threads'

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

// StartedThread is a thread this page started from the new-chat draft. It
// opens before its session exists, and its first message is sent once its
// runtime is up.
interface StartedThread {
  workspaceId: string
  userId: string
  agentId: string
  firstMessage?: DraftMessage
}

// AGUIChatPage is AG-UI Chat. The URL is its source of truth: ?thread=<id>
// opens that thread with the agent its binding names; without it the page is
// a new-chat draft, whose agent ?agent=<agent_id> preselects. The first
// message of a draft starts a thread and puts it in the URL.
export function AGUIChatPage() {
  const search = useSearch({ from: '/_authenticated/agui-chat' })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { selectedWorkspaceId, workspaces } = useWorkspace()
  const userId = useAuthStore((state) => state.auth.user?.id ?? '')
  const agentsQuery = useAgents({ page_size: 200 })
  const agents = useMemo(
    () => (agentsQuery.data?.agents ?? []).filter(isSelectableAgent),
    [agentsQuery.data]
  )
  const threadId = search.thread || null

  // The caller's threads, as the sidebar lists them.
  const sessionsQuery = useThreadSessions()

  // A new chat starts with the agent the URL names, else the one picked
  // last in this workspace. A pick is remembered and goes into the URL.
  const draft = draftAgent(
    agents,
    search.agent,
    readLastAgent(selectedWorkspaceId)
  )
  const pickAgent = (agent: Agent) => {
    const id = agent.agent_id
    if (!id || !selectedWorkspaceId) return
    rememberLastAgent(selectedWorkspaceId, id)
    void navigate({ to: '/agui-chat', search: { agent: id }, replace: true })
  }

  // The thread the URL names: read by its address, which is authoritative,
  // with the thread list's copy standing in until then. Threads started here
  // open before their session exists.
  const [started, setStarted] = useState<Record<string, StartedThread>>({})
  const startedHere = threadId ? started[threadId] : undefined
  const startedThread =
    startedHere?.workspaceId === selectedWorkspaceId &&
    startedHere.userId === userId
      ? startedHere
      : undefined
  const sessionQuery = useSessionInfo(
    AGUI_APP_NAME,
    userId,
    threadId ? sessionIdOf(threadId) : '',
    { enabled: !!selectedWorkspaceId }
  )
  const view =
    threadId && selectedWorkspaceId
      ? resolveThreadView({
          threadId,
          workspaceId: selectedWorkspaceId,
          requestedAgentId: search.agent,
          startedAgentId: startedThread?.agentId,
          session: sessionQuery.data,
          // A read being retried shows as loading again.
          sessionError: sessionQuery.isFetching ? null : sessionQuery.error,
          listed: sessionsQuery.data?.sessions?.find(
            (s) => threadIdOf(s) === threadId
          ),
          agentIds: agentsQuery.data
            ? agents.map((a) => a.agent_id ?? '')
            : undefined,
          agentsError: agentsQuery.isFetching ? null : agentsQuery.error,
        })
      : null
  const open = view?.kind === 'open' ? view : null
  const openAgentId = open?.agentId ?? null
  const openAgent = agents.find((a) => a.agent_id === openAgentId) ?? null
  const openSession = open?.session ?? null
  // Threads are deleted from the header and from the sidebar alike; the
  // open thread's AG-UI client is held there, so its run stops first.
  const { requestDelete, openClient } = useThreadDelete()

  // A thread the server refused to run, and one whose conversation could
  // not be read, show that instead of the conversation. A retry reads the
  // conversation again with a new runtime.
  const [refused, setRefused] = useState<{
    threadId: string
    message: string
  } | null>(null)
  const [attempt, setAttempt] = useState(0)
  const runtimeKey =
    openAgentId && threadId
      ? `${selectedWorkspaceId}:${userId}:${openAgentId}:${threadId}:${attempt}`
      : null
  const [historyFailure, setHistoryFailure] = useState<{
    runtimeKey: string
    error: unknown
  } | null>(null)
  const handleRefused = useCallback(
    (message: string) => {
      if (threadId) setRefused({ threadId, message })
    },
    [threadId]
  )
  const handleHistoryFailed = useCallback(
    (error: unknown) => {
      if (runtimeKey) setHistoryFailure({ runtimeKey, error })
    },
    [runtimeKey]
  )
  const handleFirstMessageSent = useCallback(() => {
    if (!threadId) return
    setStarted((s) =>
      s[threadId]
        ? { ...s, [threadId]: { ...s[threadId], firstMessage: undefined } }
        : s
    )
  }, [threadId])

  // Switching workspace leaves the previous workspace's thread or draft for
  // a new chat. A stored workspace that turned out invalid and was replaced
  // was never shown, so a link opened meanwhile stays.
  const shownWorkspace = useRef(selectedWorkspaceId)
  useEffect(() => {
    const previous = shownWorkspace.current
    shownWorkspace.current = selectedWorkspaceId
    if (
      previous &&
      selectedWorkspaceId &&
      previous !== selectedWorkspaceId &&
      workspaces.some((w) => w.id === previous) &&
      (search.thread || search.agent)
    ) {
      void navigate({ to: '/agui-chat', search: {}, replace: true })
    }
  }, [selectedWorkspaceId, workspaces, search.thread, search.agent, navigate])

  // The first message of a draft names a new thread and moves the page to
  // it; the thread's runtime sends the message.
  const startThread = (message: DraftMessage) => {
    const agentId = draft?.agent_id
    if (!agentId || !selectedWorkspaceId || !userId) return
    const id = newThreadId()
    setStarted((s) => ({
      ...s,
      [id]: {
        workspaceId: selectedWorkspaceId,
        userId,
        agentId,
        firstMessage: message,
      },
    }))
    void navigate({ to: '/agui-chat', search: { thread: id }, replace: true })
  }

  const startNewChat = (agentId?: string | null) =>
    void navigate({
      to: '/agui-chat',
      search: agentId ? { agent: agentId } : {},
      replace: true,
    })

  // After every run: a thread's first run creates its session, so refresh
  // the threads. The server titles a thread that has none once a run on it
  // succeeds, and that title lands after the run ends, so for an untitled
  // thread they are read again later to show it.
  const handleRunSettled = () => {
    void queryClient.invalidateQueries({ queryKey: ['sessions'] })
    if (openSession?.title?.trim()) return
    for (const delay of TITLE_REFRESH_DELAYS_MS) {
      setTimeout(() => {
        void queryClient.invalidateQueries({ queryKey: ['sessions'] })
      }, delay)
    }
  }

  const renameMutation = useUpdateSessionTitle()
  const handleRename = async (session: SessionAddress, title: string) => {
    await renameMutation.mutateAsync({
      app_name: session.app_name,
      user_id: session.user_id,
      session_id: session.session_id,
      title,
    })
  }

  let header: ReactNode = null
  let content: ReactNode
  if (!threadId) {
    content = (
      <DraftView
        agent={draft}
        agentSelector={<AgentSelector selected={draft} onPick={pickAgent} />}
        onSend={startThread}
      />
    )
  } else if (refused?.threadId === threadId) {
    content = (
      <ThreadNotFound
        detail={refused.message}
        onStartNew={() => startNewChat(openAgentId)}
      />
    )
  } else if (!view || view.kind === 'loading') {
    content = <ThreadLoading />
  } else if (view.kind === 'failed') {
    content = (
      <ThreadLoadFailed
        error={view.error}
        onRetry={() => {
          if (sessionQuery.isError) void sessionQuery.refetch()
          if (agentsQuery.isError) void agentsQuery.refetch()
        }}
      />
    )
  } else if (view.kind === 'not-found') {
    content = <ThreadNotFound onStartNew={() => startNewChat()} />
  } else if (historyFailure && historyFailure.runtimeKey === runtimeKey) {
    content = (
      <ThreadLoadFailed
        error={historyFailure.error}
        onRetry={() => setAttempt((a) => a + 1)}
      />
    )
  } else if (!runtimeKey) {
    content = <ThreadLoading />
  } else {
    const agentName = openAgent?.name ?? 'Agent'
    const title = openSession?.title?.trim() || agentName
    const address: SessionAddress = {
      app_name: AGUI_APP_NAME,
      user_id: userId,
      session_id: sessionIdOf(threadId),
    }
    header = (
      <ThreadHeader
        title={title}
        agentName={agentName}
        agentIconUrl={openAgent ? agentIconUrl(openAgent) : undefined}
        threadId={threadId}
        onRename={(t) => handleRename(address, t)}
        onDelete={() =>
          requestDelete({ session: address, title, agentId: openAgentId })
        }
      />
    )
    content = (
      <AGUIChatWithRuntime
        key={runtimeKey}
        clientRef={openClient}
        agentId={view.agentId}
        agent={openAgent ?? undefined}
        threadId={threadId}
        fresh={!!startedThread?.firstMessage}
        firstMessage={startedThread?.firstMessage}
        onFirstMessageSent={handleFirstMessageSent}
        onRunSettled={handleRunSettled}
        onThreadRefused={handleRefused}
        onHistoryFailed={handleHistoryFailed}
      />
    )
  }

  return (
    <>
      <PageHeader />
      <Main fixed fluid className='flex flex-col px-0 py-0'>
        {header}
        {content}
      </Main>
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
// thread only stops retrying. A failed load goes to onFailed, and the page
// offers Retry.
function useThreadHistory(
  agentId: string,
  threadId: string,
  store: A2UIStore,
  onFailed: (err: unknown) => void
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
          if (mounted.current) onFailed(err)
          return { messages: [] }
        }
      },
      // The server keeps the conversation; nothing to write back.
      async append() {},
    }),
    [agentId, threadId, store, onFailed]
  )
}

// useRunErrorReporter shows each failed run once, as a toast. A run the
// server refused because the thread cannot be used from here goes to
// onRefused instead, and the page shows the thread as not found.
function useRunErrorReporter(onRefused: (message: string) => void) {
  const refusedRef = useRef(onRefused)
  useEffect(() => {
    refusedRef.current = onRefused
  })
  const [report] = useState(() => {
    const toastOnce = errorReporter((message) => toast.error(message))
    return (err: unknown) => {
      const refusal = threadRefusal(err)
      if (refusal === null) toastOnce(err)
      else refusedRef.current(refusal)
    }
  })
  return report
}

// useFirstMessage sends the message that started the thread from the draft
// once its runtime is up, as the composer sends one: its text, and its
// images read into attachments. A failure is reported instead of lost.
function useFirstMessage(
  runtime: AssistantRuntime,
  message: DraftMessage | undefined,
  onSent: () => void,
  onError: (err: unknown) => void
) {
  const sent = useRef(false)
  useEffect(() => {
    if (!message || sent.current) return
    sent.current = true
    onSent()
    void Promise.all(message.images.map(imageAttachment))
      .then((attachments) =>
        runtime.thread.append({
          role: 'user',
          content: message.text ? [{ type: 'text', text: message.text }] : [],
          attachments,
        })
      )
      .catch(onError)
  }, [runtime, message, onSent, onError])
}

// AGUIChatWithRuntime is one thread's conversation and composer, with its own
// AG-UI client. The page keys it by thread, so it lives exactly as long as
// that thread is open; clientRef holds its client meanwhile.
function AGUIChatWithRuntime({
  clientRef,
  agentId,
  agent,
  threadId,
  fresh,
  firstMessage,
  onFirstMessageSent,
  onRunSettled,
  onThreadRefused,
  onHistoryFailed,
}: {
  clientRef: RefObject<ButterAGUIAgent | null>
  agentId: string
  // agent is the thread's agent as the agent list has it, for its name and
  // avatar.
  agent: Agent | undefined
  threadId: string
  // fresh is a thread started on this page: it has no conversation to read.
  fresh: boolean
  firstMessage?: DraftMessage
  onFirstMessageSent: () => void
  onRunSettled: () => void
  onThreadRefused: (message: string) => void
  onHistoryFailed: (err: unknown) => void
}) {
  const [httpAgent] = useState(() => makeHttpAgent(agentId, threadId))
  useEffect(() => {
    clientRef.current = httpAgent
    return () => {
      if (clientRef.current === httpAgent) clientRef.current = null
    }
  }, [clientRef, httpAgent])
  // The runtime reads the thread once, when it starts.
  const [readsHistory] = useState(!fresh)
  const store = useOwnedA2UIStore(httpAgent)
  const history = useThreadHistory(agentId, threadId, store, onHistoryFailed)
  const reportError = useRunErrorReporter(onThreadRefused)
  // The composer takes images through the adapter, whose limits count the
  // images the composer holds once the runtime is up.
  const [imageAdapter] = useState(() => new ImageAttachmentAdapter())
  const adapters = useMemo(
    () =>
      readsHistory
        ? { history, attachments: imageAdapter }
        : { attachments: imageAdapter },
    [readsHistory, history, imageAdapter]
  )
  const runtime = useAgUiRuntime({
    agent: httpAgent,
    onError: reportError,
    adapters,
  })
  useEffect(() => {
    imageAdapter.serve(() => runtime.thread.composer.getState().attachments)
  }, [imageAdapter, runtime])
  useFirstMessage(runtime, firstMessage, onFirstMessageSent, reportError)

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

  return (
    <AssistantRuntimeProvider runtime={runtime} config={chatAuiConfig}>
      <A2UIStoreContext.Provider value={store}>
        <FormSubmitBridge store={store} httpAgent={httpAgent} />
        {/* Images dropped anywhere on the conversation go to the composer. */}
        <ComposerPrimitive.AttachmentDropzone className='flex min-h-0 flex-1 flex-col data-[dragging=true]:ring-2 data-[dragging=true]:ring-ring/50 data-[dragging=true]:ring-inset'>
          <ThreadArea
            agent={agent}
            httpAgent={httpAgent}
            onSendError={reportError}
          />
          <SharedStatePanel />
          <ComposerArea
            httpAgent={httpAgent}
            imageAdapter={imageAdapter}
            onSendError={reportError}
          />
        </ComposerPrimitive.AttachmentDropzone>
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
  const steer = async (reply: CreateAppendMessage) => {
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
    // answerOldest sends a plain message, which the server takes as the
    // answer to its oldest open Interrupt (ADR-0002): its text is the
    // answer, and its images go with it.
    answerOldest(reply: CreateAppendMessage) {
      httpAgent.sendNextRunAsMessage()
      return steer(reply)
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

// ComposerArea sends the user's messages, with the images attached to them:
// picked, pasted, or dropped on the conversation. A file the adapter refuses
// is reported with why. While questions are open the runtime refuses an
// ordinary send, so the message goes out as a plain message instead, and the
// server takes its text as the answer to its earliest open question
// (ADR-0002), its images going with it. A hint says so.
function ComposerArea({
  httpAgent,
  imageAdapter,
  onSendError,
}: {
  httpAgent: ButterAGUIAgent
  imageAdapter: ImageAttachmentAdapter
  onSendError: (err: unknown) => void
}) {
  const answering = useAgUiInterrupts().length > 0
  const replies = useReplies(httpAgent, onSendError)
  const aui = useAui()
  const hintId = useId()

  useAuiEvent('composer.attachmentAddError', (event) => {
    toast.error(attachmentRefusal(event))
  })

  // Takes over Enter (the form's submit) and the Send button.
  const sendAsAnswer = (e: SyntheticEvent) => {
    if (!answering) return
    e.preventDefault()
    const { text, attachments } = aui.composer.getState()
    const answer = text.trim()
    if (answer === '') {
      // The server takes a message's text as the answer, so images alone
      // answer nothing.
      if (attachments.length > 0) {
        toast.error('Type your answer; the images are sent with it.')
      }
      return
    }
    aui.composer.setText('')
    void aui.composer.clearAttachments()
    void sendAll(imageAdapter, attachments).then(
      (sent) =>
        replies.answerOldest({
          role: 'user',
          content: [{ type: 'text', text: answer }],
          attachments: sent,
        }),
      onSendError
    )
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
      <ComposerImages className='mx-auto mb-3 max-w-3xl' />
      <ComposerPrimitive.Root
        onSubmit={sendAsAnswer}
        className='mx-auto flex max-w-3xl items-end gap-2'
      >
        <AttachImagesButton className='size-9' />
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
