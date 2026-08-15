import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSearch } from '@tanstack/react-router'
import type { Agent } from '@/types/api'
import {
  ChevronDown,
  CircleAlert,
  PlugZap,
  RotateCcw,
  Send,
  Square,
  Wrench,
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import {
  applyAGUIStateDelta,
  runAGUIAgent,
  type AGUIEvent,
  type AGUIInterrupt,
  type AGUIMessage,
  type AGUIRunInput,
} from '@/api/agui'
import { newClientID } from '@/lib/client-id'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

// The dashboard AG-UI client (issue #291 item 4): drives an enable_agui agent
// over POST /api/agui/:agent_id and renders the protocol stream — text,
// tool calls/results, shared state, run errors, and Workflow Interrupts with
// addressed resume. The classic (non-AG-UI) chat page is untouched.

type ThreadItem =
  | { kind: 'user'; id: string; text: string }
  | { kind: 'assistant'; id: string; text: string; streaming: boolean }
  | { kind: 'tool'; id: string; name: string; args: string; result?: string }
  | { kind: 'error'; id: string; message: string }

function isSelectableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  const runnable =
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  return runnable && !!a.enable_agui && !!a.agent_id
}

export function AGUIChatPage() {
  const search = useSearch({ from: '/_authenticated/agui-chat' })
  const agentsQuery = useAgents({ page_size: 200 })
  const agents = useMemo(
    () => (agentsQuery.data?.agents ?? []).filter(isSelectableAgent),
    [agentsQuery.data]
  )

  // Explicit selection (URL or picker); falls back to the first AG-UI agent.
  const [pickedAgentId, setPickedAgentId] = useState<string | null>(
    search.agent ?? null
  )
  const agentId = pickedAgentId ?? agents[0]?.agent_id ?? null
  const agent = useMemo(
    () => agents.find((a) => a.agent_id === agentId) ?? null,
    [agents, agentId]
  )

  const [threadId, setThreadId] = useState(() => newClientID())
  const [items, setItems] = useState<ThreadItem[]>([])
  const [interrupts, setInterrupts] = useState<AGUIInterrupt[]>([])
  const [stateMirror, setStateMirror] = useState<Record<string, unknown>>({})
  const [hasState, setHasState] = useState(false)
  const [stateOpen, setStateOpen] = useState(false)
  const [running, setRunning] = useState(false)
  const [draft, setDraft] = useState('')
  const abortRef = useRef<AbortController | null>(null)
  const historyRef = useRef<AGUIMessage[]>([])
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [items, interrupts])

  useEffect(() => () => abortRef.current?.abort(), [])

  const upsertItem = useCallback(
    (
      id: string,
      make: () => ThreadItem,
      update: (item: ThreadItem) => ThreadItem
    ) => {
      setItems((prev) => {
        const at = prev.findIndex((it) => it.id === id)
        if (at < 0) return [...prev, make()]
        const next = [...prev]
        next[at] = update(next[at])
        return next
      })
    },
    []
  )

  const handleEvent = useCallback(
    (event: AGUIEvent) => {
      switch (event.type) {
        case 'TEXT_MESSAGE_START': {
          const id = String(event.messageId ?? newClientID())
          upsertItem(
            id,
            () => ({ kind: 'assistant', id, text: '', streaming: true }),
            (it) => it
          )
          break
        }
        case 'TEXT_MESSAGE_CONTENT': {
          const id = String(event.messageId ?? '')
          const delta = String(event.delta ?? '')
          upsertItem(
            id,
            () => ({ kind: 'assistant', id, text: delta, streaming: true }),
            (it) =>
              it.kind === 'assistant' ? { ...it, text: it.text + delta } : it
          )
          break
        }
        case 'TEXT_MESSAGE_END': {
          const id = String(event.messageId ?? '')
          setItems((prev) => {
            const finished = prev.map((it) =>
              it.id === id && it.kind === 'assistant'
                ? { ...it, streaming: false }
                : it
            )
            const done = finished.find((it) => it.id === id)
            if (done && done.kind === 'assistant' && done.text !== '') {
              historyRef.current = [
                ...historyRef.current,
                { id, role: 'assistant', content: done.text },
              ]
            }
            return finished
          })
          break
        }
        case 'TOOL_CALL_START': {
          const id = String(event.toolCallId ?? newClientID())
          const name = String(event.toolCallName ?? 'tool')
          upsertItem(
            id,
            () => ({ kind: 'tool', id, name, args: '' }),
            (it) => it
          )
          break
        }
        case 'TOOL_CALL_ARGS': {
          const id = String(event.toolCallId ?? '')
          const delta = String(event.delta ?? '')
          upsertItem(
            id,
            () => ({ kind: 'tool', id, name: 'tool', args: delta }),
            (it) => (it.kind === 'tool' ? { ...it, args: it.args + delta } : it)
          )
          break
        }
        case 'TOOL_CALL_RESULT': {
          const id = String(event.toolCallId ?? '')
          upsertItem(
            id,
            () => ({
              kind: 'tool',
              id,
              name: 'tool',
              args: '',
              result: String(event.content ?? ''),
            }),
            (it) =>
              it.kind === 'tool'
                ? { ...it, result: String(event.content ?? '') }
                : it
          )
          break
        }
        case 'STATE_SNAPSHOT': {
          const snapshot = event.snapshot
          if (
            snapshot &&
            typeof snapshot === 'object' &&
            !Array.isArray(snapshot)
          ) {
            setStateMirror(snapshot as Record<string, unknown>)
            setHasState(true)
          }
          break
        }
        case 'STATE_DELTA': {
          const ops = event.delta
          if (Array.isArray(ops)) {
            setHasState(true)
            setStateMirror((prev) => {
              const next = applyAGUIStateDelta(
                prev,
                ops as Array<{ op: string; path: string; value?: unknown }>
              )
              // An unsupported patch keeps the stale mirror; sending it back
              // on the next run makes the server re-baseline with a snapshot.
              return next ?? prev
            })
          }
          break
        }
        case 'RUN_FINISHED': {
          const outcome = event.outcome as
            { type?: string; interrupts?: AGUIInterrupt[] } | undefined
          if (
            outcome?.type === 'interrupt' &&
            Array.isArray(outcome.interrupts)
          ) {
            setInterrupts(outcome.interrupts)
          }
          break
        }
        case 'RUN_ERROR': {
          const message = String(event.message ?? 'run failed')
          setItems((prev) => [
            ...prev,
            { kind: 'error', id: newClientID(), message },
          ])
          break
        }
        default:
          break
      }
    },
    [upsertItem]
  )

  const startRun = useCallback(
    async (input: Omit<AGUIRunInput, 'threadId' | 'runId' | 'state'>) => {
      if (!agent?.agent_id) return
      setRunning(true)
      const controller = new AbortController()
      abortRef.current = controller
      try {
        await runAGUIAgent(
          agent.agent_id,
          {
            threadId,
            runId: newClientID(),
            state: hasState ? stateMirror : null,
            ...input,
          },
          { signal: controller.signal, onEvent: handleEvent }
        )
      } catch (err) {
        if (controller.signal.aborted) return
        const message =
          err instanceof Error ? err.message : 'AG-UI request failed'
        setItems((prev) => [
          ...prev,
          { kind: 'error', id: newClientID(), message },
        ])
        toast.error(message)
      } finally {
        setRunning(false)
      }
    },
    [agent, threadId, hasState, stateMirror, handleEvent]
  )

  function handleSend() {
    const text = draft.trim()
    if (!text || running || !agent) return
    setDraft('')
    const userMessage: AGUIMessage = {
      id: newClientID(),
      role: 'user',
      content: text,
    }
    historyRef.current = [...historyRef.current, userMessage]
    setItems((prev) => [...prev, { kind: 'user', id: userMessage.id, text }])
    void startRun({ messages: historyRef.current })
  }

  function handleResume(interrupt: AGUIInterrupt, answer: string) {
    if (running) return
    setInterrupts((prev) => prev.filter((it) => it.id !== interrupt.id))
    setItems((prev) => [
      ...prev,
      { kind: 'user', id: newClientID(), text: answer },
    ])
    void startRun({
      messages: [],
      resume: [
        { interruptId: interrupt.id, status: 'resolved', payload: answer },
      ],
    })
  }

  function handleStop() {
    abortRef.current?.abort()
    setRunning(false)
  }

  function handleNewThread() {
    abortRef.current?.abort()
    setThreadId(newClientID())
    setItems([])
    setInterrupts([])
    setStateMirror({})
    setHasState(false)
    setRunning(false)
    historyRef.current = []
  }

  return (
    <>
      <Header fixed className='h-14 border-b border-border/60 bg-background/95'>
        <Search className='sm:w-44 md:w-52 lg:w-60 xl:w-72' />
        <div className='ms-auto flex items-center gap-1 sm:gap-2'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main fixed fluid className='flex flex-col px-0 py-0'>
        <div className='flex items-center gap-2 border-b border-border/60 px-4 py-2'>
          <PlugZap className='size-4 text-muted-foreground' />
          <span className='text-sm font-medium'>AG-UI Chat</span>
          <Select
            value={agentId ?? undefined}
            onValueChange={(value) => {
              setPickedAgentId(value)
              handleNewThread()
            }}
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
          <Button
            variant='ghost'
            size='sm'
            className='ml-auto'
            onClick={handleNewThread}
            disabled={items.length === 0 && !running}
          >
            <RotateCcw className='size-4' /> New thread
          </Button>
        </div>

        {agents.length === 0 && !agentsQuery.isLoading ? (
          <div className='flex flex-1 items-center justify-center px-6 text-center'>
            <p className='text-sm text-muted-foreground'>
              No AG-UI-enabled agents in this workspace. Enable “AG-UI” on an
              agent to chat with it here.
            </p>
          </div>
        ) : (
          <div ref={scrollRef} className='flex-1 overflow-y-auto px-4 py-4'>
            <div className='mx-auto flex max-w-3xl flex-col gap-3'>
              {items.map((item) => (
                <ThreadItemView key={item.id} item={item} />
              ))}
              {interrupts.map((interrupt) => (
                <InterruptPrompt
                  key={interrupt.id}
                  interrupt={interrupt}
                  disabled={running}
                  onResolve={(answer) => handleResume(interrupt, answer)}
                />
              ))}
              {running && (
                <p className='text-xs text-muted-foreground'>Running…</p>
              )}
            </div>
          </div>
        )}

        {hasState && (
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
                  {JSON.stringify(stateMirror, null, 2)}
                </pre>
              )}
            </div>
          </div>
        )}

        <div className='border-t border-border/60 px-4 py-3'>
          <div className='mx-auto flex max-w-3xl items-end gap-2'>
            <Textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  handleSend()
                }
              }}
              placeholder={
                interrupts.length > 0
                  ? 'Answer the pending question above, or send a new message'
                  : 'Message the agent over AG-UI…'
              }
              rows={2}
              className='min-h-0 resize-none'
              disabled={!agent}
            />
            {running ? (
              <Button
                variant='outline'
                size='icon'
                onClick={handleStop}
                aria-label='Stop'
              >
                <Square className='size-4' />
              </Button>
            ) : (
              <Button
                size='icon'
                onClick={handleSend}
                disabled={!agent || draft.trim() === ''}
                aria-label='Send'
              >
                <Send className='size-4' />
              </Button>
            )}
          </div>
        </div>
      </Main>
    </>
  )
}

function ThreadItemView({ item }: { item: ThreadItem }) {
  switch (item.kind) {
    case 'user':
      return (
        <div className='ml-auto max-w-[85%] rounded-lg bg-primary px-3 py-2 text-sm whitespace-pre-wrap text-primary-foreground'>
          {item.text}
        </div>
      )
    case 'assistant':
      return (
        <div className='max-w-[85%] rounded-lg border border-border bg-card px-3 py-2 text-sm'>
          <div className='space-y-2 [&_a]:text-primary [&_a]:underline [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:font-mono [&_code]:text-xs [&_h1]:text-base [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:text-sm [&_h3]:font-medium [&_ol]:list-decimal [&_ol]:pl-5 [&_pre]:overflow-x-auto [&_pre]:rounded [&_pre]:bg-muted [&_pre]:p-2 [&_pre_code]:bg-transparent [&_pre_code]:px-0 [&_table]:text-xs [&_td]:border [&_td]:border-border [&_td]:px-2 [&_td]:py-1 [&_th]:border [&_th]:border-border [&_th]:px-2 [&_th]:py-1 [&_ul]:list-disc [&_ul]:pl-5'>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {item.text}
            </ReactMarkdown>
          </div>
          {item.streaming && <span className='text-muted-foreground'>…</span>}
        </div>
      )
    case 'tool':
      return <ToolCallView item={item} />
    case 'error':
      return (
        <div className='flex max-w-[85%] items-start gap-2 rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm'>
          <CircleAlert className='mt-0.5 size-4 shrink-0 text-destructive' />
          <span>{item.message}</span>
        </div>
      )
  }
}

function ToolCallView({
  item,
}: {
  item: Extract<ThreadItem, { kind: 'tool' }>
}) {
  const [open, setOpen] = useState(false)
  return (
    <div className='max-w-[85%] rounded-lg border border-border bg-muted/40 px-3 py-2 text-xs'>
      <button
        type='button'
        className='flex items-center gap-1.5 font-mono'
        onClick={() => setOpen((v) => !v)}
      >
        <Wrench className='size-3.5 text-muted-foreground' />
        {item.name}
        {item.result === undefined && (
          <span className='text-muted-foreground'>(pending)</span>
        )}
        <ChevronDown
          className={cn('size-3 transition-transform', !open && '-rotate-90')}
        />
      </button>
      {open && (
        <div className='mt-1 space-y-1'>
          {item.args !== '' && (
            <pre className='overflow-x-auto rounded bg-background/60 p-1.5'>
              {item.args}
            </pre>
          )}
          {item.result !== undefined && (
            <pre className='overflow-x-auto rounded bg-background/60 p-1.5'>
              {item.result}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}

function InterruptPrompt({
  interrupt,
  disabled,
  onResolve,
}: {
  interrupt: AGUIInterrupt
  disabled: boolean
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
          disabled={disabled}
        />
        <Button
          size='sm'
          disabled={disabled || answer.trim() === ''}
          onClick={() => onResolve(answer.trim())}
        >
          Answer
        </Button>
      </div>
    </div>
  )
}
