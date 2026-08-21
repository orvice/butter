import { useMemo, useState } from 'react'
import { useSearch } from '@tanstack/react-router'
import type { Agent } from '@/types/api'
import { HttpAgent } from '@ag-ui/client'
import {
  AssistantRuntimeProvider,
  ThreadPrimitive,
  ComposerPrimitive,
  MessagePrimitive,
  ActionBarPrimitive,
  useAssistantToolUI,
} from '@assistant-ui/react'
import {
  useAgUiRuntime,
  useAgUiInterrupts,
  useAgUiSubmitInterruptResponses,
  useAgUiState,
} from '@assistant-ui/react-ag-ui'
import { ChevronDown, Copy, PlugZap, Send, Square, Wrench } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { BASE_URL, authHeaders } from '@/api/client'
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

function isSelectableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  const runnable =
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  return runnable && !!a.enable_agui && !!a.agent_id
}

function makeHttpAgent(agentId: string): HttpAgent {
  return new HttpAgent({
    url: `${BASE_URL}/api/agui/${encodeURIComponent(agentId)}`,
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

  const httpAgent = useMemo(
    () => (agentId ? makeHttpAgent(agentId) : null),
    [agentId]
  )

  if (!httpAgent) {
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
    <AGUIChatWithRuntime
      httpAgent={httpAgent}
      agents={agents}
      agentId={agentId}
      agentName={agent?.name ?? 'Agent'}
      onAgentChange={(id) => setPickedAgentId(id)}
    />
  )
}

function AGUIChatWithRuntime({
  httpAgent,
  agents,
  agentId,
  onAgentChange,
}: {
  httpAgent: HttpAgent
  agents: Agent[]
  agentId: string | null
  agentName: string
  onAgentChange: (id: string) => void
}) {
  const runtime = useAgUiRuntime({
    agent: httpAgent,
    onError: (err) => toast.error(err.message || 'AG-UI request failed'),
  })

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      <PageHeader />
      <Main fixed fluid className='flex flex-col px-0 py-0'>
        <AgentBar
          agents={agents}
          agentId={agentId}
          onAgentChange={(id) => {
            httpAgent.abortRun()
            onAgentChange(id)
          }}
        />
        <ToolFallbackUI />
        <ThreadArea />
        <SharedStatePanel />
        <ComposerArea />
      </Main>
    </AssistantRuntimeProvider>
  )
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
  isLoading,
}: {
  agents: Agent[]
  agentId: string | null
  onAgentChange: (id: string) => void
  isLoading?: boolean
}) {
  return (
    <div className='flex items-center gap-2 border-b border-border/60 px-4 py-2'>
      <PlugZap className='size-4 text-muted-foreground' />
      <span className='text-sm font-medium'>AG-UI Chat</span>
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
    </div>
  )
}

function ToolFallbackUI() {
  useAssistantToolUI({
    toolName: '*',
    render: (props) => <GenericToolCallView {...props} />,
  })
  return null
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

function ThreadArea() {
  return (
    <ThreadPrimitive.Root className='flex-1 overflow-y-auto px-4 py-4'>
      <ThreadPrimitive.Viewport className='mx-auto flex max-w-3xl flex-col gap-3'>
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

  if (interrupts.length === 0) return null

  return (
    <>
      {interrupts.map((interrupt) => (
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
