import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useSearch } from '@tanstack/react-router'
import type { Agent, SessionInfo } from '@/types/api'
import {
  AssistantRuntimeProvider,
  ThreadPrimitive,
  ComposerPrimitive,
  MessagePrimitive,
  ActionBarPrimitive,
  useAuiState,
} from '@assistant-ui/react'
import {
  useAgUiRuntime,
  useAgUiInterrupts,
  useAgUiSteerAway,
  useAgUiSubmitInterruptResponses,
  useAgUiState,
} from '@assistant-ui/react-ag-ui'
import {
  ChevronDown,
  Copy,
  History,
  MessageSquarePlus,
  PanelLeft,
  PlugZap,
  Send,
  Square,
  Wrench,
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { fetchAGUIUISnapshot } from '@/api/agui'
import { BASE_URL, authHeaders } from '@/api/client'
import {
  useDeleteSession,
  useGenerateSessionTitle,
  useSessions,
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
import { DeleteDialog } from '@/components/delete-dialog'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ButterAGUIAgent } from './a2ui/agent'
import { readableReply, submissionPayload } from './a2ui/form'
import {
  EVENT_NAME,
  envelopeOp,
  isA2UIEventValue,
  type UISnapshot,
} from './a2ui/protocol'
import { A2UIStore, useA2UIStore } from './a2ui/store'
import { A2UISurfaceView } from './a2ui/surface-view'
import { ThreadList } from './thread-list'
import {
  currentThread,
  newThreadId,
  threadPointerKey,
  writeThreadPointer,
} from './thread-pointer'
import { AGUI_APP_NAME, agentThreads, threadIdOf, threadTitle } from './threads'

function isSelectableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  const runnable =
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  return runnable && !!a.enable_agui && !!a.agent_id
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
  const sessionsQuery = useSessions(
    { app_name: AGUI_APP_NAME, user_id: userId || undefined, page_size: 100 },
    { enabled: !!userId }
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
  const generateTitleMutation = useGenerateSessionTitle()
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null)

  // After every run: a thread's first run creates its session, so refresh
  // the list, and title a thread that has none yet (the server keeps any
  // title that already exists, manual ones included).
  const handleRunSettled = () => {
    void queryClient.invalidateQueries({ queryKey: ['sessions'] })
    if (!threadId) return
    const listed = agentThreadList.find((s) => threadIdOf(s) === threadId)
    if (listed?.title?.trim()) return
    generateTitleMutation.mutate(
      {
        app_name: AGUI_APP_NAME,
        user_id: userId,
        session_id: `agui-${threadId}`,
      },
      { onError: () => {} }
    )
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
                : 'No AG-UI-enabled agents in this workspace. Enable "AG-UI" on an agent to chat with it here.'}
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

// useOwnedA2UIStore creates and owns one thread's surfaces: it feeds every butter.a2ui CUSTOM
// event to the store in arrival order and restores the thread's persisted
// cards and unanswered forms from the UI snapshot on mount.
function useOwnedA2UIStore(
  httpAgent: ButterAGUIAgent,
  agentId: string,
  threadId: string
) {
  const [store] = useState(() => new A2UIStore())

  useEffect(() => {
    const sub = httpAgent.subscribe({
      onCustomEvent: ({ event }) => {
        if (event.name === EVENT_NAME) store.apply(event.value)
      },
    })
    return () => sub.unsubscribe()
  }, [httpAgent, store])

  useEffect(() => {
    const controller = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined
    const load = async (attempt: number) => {
      try {
        const snap = await fetchAGUIUISnapshot<UISnapshot>(
          agentId,
          threadId,
          controller.signal
        )
        store.applySnapshot(snap)
      } catch (err) {
        if (controller.signal.aborted) return
        // A run holds the thread (409) or the read failed: retry a few
        // times; live events keep arriving meanwhile.
        if (attempt < 5) {
          timer = setTimeout(() => void load(attempt + 1), 500 * 2 ** attempt)
        } else {
          toast.error(
            `Could not restore this thread's cards and forms: ${
              err instanceof Error ? err.message : String(err)
            }`
          )
        }
      }
    }
    void load(0)
    return () => {
      controller.abort()
      if (timer) clearTimeout(timer)
    }
  }, [agentId, threadId, store])

  useEffect(() => () => store.dispose(), [store])
  return store
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
  const runtime = useAgUiRuntime({
    agent: httpAgent,
    onError: (err) => toast.error(err.message || 'AG-UI request failed'),
  })
  const store = useOwnedA2UIStore(httpAgent, agentId, threadId)
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
    <AssistantRuntimeProvider runtime={runtime}>
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
              <ThreadArea />
              <SharedStatePanel />
              <ComposerArea />
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
        httpAgent.clearNextResume()
      }
    })
    return () => store.setSubmitter(undefined)
  }, [store, httpAgent, steerAway])
  return null
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
          <SelectValue placeholder='Pick an AG-UI agent' />
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

function GenericToolCallView(props: {
  toolName: string
  args: Record<string, unknown>
  result?: unknown
  status: { type: string }
}) {
  const [open, setOpen] = useState(false)
  const argsStr = JSON.stringify(props.args, null, 2)
  const resultStr =
    props.result !== undefined ? JSON.stringify(props.result, null, 2) : null
  const pending =
    props.status.type === 'running' || props.status.type === 'requires-action'

  return (
    <div className='max-w-[85%] rounded-lg border border-border bg-muted/40 px-3 py-2 text-xs'>
      <button
        type='button'
        className='flex items-center gap-1.5 font-mono'
        onClick={() => setOpen((v) => !v)}
      >
        <Wrench className='size-3.5 text-muted-foreground' />
        {props.toolName}
        {pending && <span className='text-muted-foreground'>(pending)</span>}
        <ChevronDown
          className={cn('size-3 transition-transform', !open && '-rotate-90')}
        />
      </button>
      {open && (
        <div className='mt-1 space-y-1'>
          {argsStr !== '{}' && (
            <pre className='overflow-x-auto rounded bg-background/60 p-1.5'>
              {argsStr}
            </pre>
          )}
          {resultStr !== null && (
            <pre className='overflow-x-auto rounded bg-background/60 p-1.5'>
              {resultStr}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}

// RenderUIToolView keeps the render_ui call out of the way: its card is the
// visible result. A rejected call still shows the tool error.
function RenderUIToolView(props: Parameters<typeof GenericToolCallView>[0]) {
  const result = props.result as { error?: unknown } | undefined
  if (result && typeof result === 'object' && 'error' in result) {
    return <GenericToolCallView {...props} />
  }
  return null
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

// RestoredSurfaces shows what the UI snapshot brought back after a refresh:
// the conversation text itself is not restored, so each surface says where
// it came from.
function RestoredSurfaces() {
  const store = useA2UIStore(useA2UIContext())
  const locked = useAuiState((s) => s.thread.isRunning)
  const restored = store
    .list()
    .filter((e) => e.origin === 'snapshot' && !e.deleted)
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

function ThreadArea() {
  return (
    <ThreadPrimitive.Root className='flex-1 overflow-y-auto px-4 py-4'>
      <ThreadPrimitive.Viewport className='mx-auto flex max-w-3xl flex-col gap-3'>
        <RestoredSurfaces />
        <ThreadPrimitive.Messages
          components={{
            UserMessage: UserMessageView,
            AssistantMessage: AssistantMessageView,
          }}
        />
        <InterruptPrompts />
        <ThreadPrimitive.If running>
          <p className='text-xs text-muted-foreground'>Running…</p>
        </ThreadPrimitive.If>
      </ThreadPrimitive.Viewport>
    </ThreadPrimitive.Root>
  )
}

function UserMessageView() {
  return (
    <MessagePrimitive.Root className='ml-auto max-w-[85%] rounded-lg bg-primary px-3 py-2 text-sm whitespace-pre-wrap text-primary-foreground'>
      <MessagePrimitive.Content
        components={{
          Text: ({ text }) => <>{text}</>,
        }}
      />
    </MessagePrimitive.Root>
  )
}

const REMARK_PLUGINS = [remarkGfm]

function AssistantMessageView() {
  return (
    <MessagePrimitive.Root className='max-w-[85%] rounded-lg border border-border bg-card px-3 py-2 text-sm'>
      <MessagePrimitive.Content
        components={{
          Text: MarkdownText,
          tools: {
            by_name: { render_ui: RenderUIToolView },
            Fallback: GenericToolCallView,
          },
          data: { by_name: { [EVENT_NAME]: A2UIDataPart } },
        }}
      />
      <MessagePrimitive.If lastOrHover>
        <div className='mt-1 flex items-center gap-0.5'>
          <ActionBarPrimitive.Root>
            <ActionBarPrimitive.Copy asChild>
              <button
                type='button'
                className='inline-flex size-7 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-muted hover:text-foreground'
                aria-label='Copy message'
              >
                <Copy className='size-3.5' />
              </button>
            </ActionBarPrimitive.Copy>
          </ActionBarPrimitive.Root>
        </div>
      </MessagePrimitive.If>
    </MessagePrimitive.Root>
  )
}

function MarkdownText({ text }: { text: string }) {
  return (
    <div className='space-y-2 [&_a]:text-primary [&_a]:underline [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:font-mono [&_code]:text-xs [&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:text-sm [&_h3]:font-medium [&_ol]:list-decimal [&_ol]:pl-5 [&_pre]:overflow-x-auto [&_pre]:rounded [&_pre]:bg-muted [&_pre]:p-2 [&_pre_code]:bg-transparent [&_pre_code]:px-0 [&_table]:text-xs [&_td]:border [&_td]:border-border [&_td]:px-2 [&_td]:py-1 [&_th]:border [&_th]:border-border [&_th]:px-2 [&_th]:py-1 [&_ul]:list-disc [&_ul]:pl-5'>
      <ReactMarkdown remarkPlugins={REMARK_PLUGINS}>{text}</ReactMarkdown>
    </div>
  )
}

function InterruptPrompts() {
  const interrupts = useAgUiInterrupts()
  const submitResponses = useAgUiSubmitInterruptResponses()
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
          onResolve={(answer) =>
            void submitResponses([
              {
                interruptId: interrupt.id,
                status: 'resolved',
                payload: answer,
              },
            ])
          }
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
  return (
    <div className='max-w-[85%] rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm'>
      <p className='font-medium'>
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

function ComposerArea() {
  return (
    <div className='border-t border-border/60 px-4 py-3'>
      <ComposerPrimitive.Root className='mx-auto flex max-w-3xl items-end gap-2'>
        <ComposerPrimitive.Input
          autoFocus
          placeholder='Message the agent over AG-UI…'
          rows={2}
          className='min-h-0 flex-1 resize-none rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring'
        />
        <ThreadPrimitive.If running>
          <ComposerPrimitive.Cancel asChild>
            <Button variant='outline' size='icon' aria-label='Stop'>
              <Square className='size-4' />
            </Button>
          </ComposerPrimitive.Cancel>
        </ThreadPrimitive.If>
        <ThreadPrimitive.If running={false}>
          <ComposerPrimitive.Send asChild>
            <Button size='icon' aria-label='Send'>
              <Send className='size-4' />
            </Button>
          </ComposerPrimitive.Send>
        </ThreadPrimitive.If>
      </ComposerPrimitive.Root>
    </div>
  )
}
