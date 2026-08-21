import {
  type ComponentProps,
  useCallback,
  useEffect,
  useRef,
  useState,
} from 'react'
import type { SessionInfo } from '@/types/api'
import {
  AssistantRuntimeProvider,
  ThreadPrimitive,
  ComposerPrimitive,
  MessagePrimitive,
  ActionBarPrimitive,
  unstable_useComposerInput,
} from '@assistant-ui/react'
import {
  ArrowUp,
  ChevronDown,
  Copy,
  Loader2,
  MoreHorizontal,
  Paperclip,
  Pencil,
  Square,
  Trash2,
  Undo2,
  Wrench,
  X,
} from 'lucide-react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useUpdateSessionTitle } from '@/api/sessions'
import { sessionTitle } from '@/lib/session-title'
import { cn } from '@/lib/utils'
import { useImageAttachments } from '@/hooks/use-image-attachments'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { AgentAvatar } from '@/components/butter/primitives'
import { InlineTitleInput } from '@/components/inline-title-input'
import { useButterRuntime, type TerminalNotice } from './butter-runtime'

const REMARK_PLUGINS = [remarkGfm]

interface AUIChatWindowProps {
  session: SessionInfo | null
  userId: string
  agentName: string | null
  agentId?: string | null
  onDelete?: () => void
  onInvocationAccepted?: (invocationId: string, message: string) => void
  pendingMessage?: string
  initialInvocationId?: string
}

export function AUIChatWindow({
  session,
  userId,
  agentName,
  agentId,
  onDelete,
  onInvocationAccepted,
  pendingMessage,
  initialInvocationId,
}: AUIChatWindowProps) {
  const sessionId = session?.session_id ?? ''
  const {
    attachments,
    previewUrls,
    isDragOver,
    addFiles,
    removeAttachment,
    clearAttachments,
    validateForSend,
    handleDragOver,
    handleDragEnter,
    handleDragLeave,
    handleDrop,
    handlePaste,
    fileInputRef,
    openFilePicker,
    handleFileInputChange,
    fileAccept,
  } = useImageAttachments(sessionId)

  const attachmentsRef = useRef<File[]>([])
  useEffect(() => {
    attachmentsRef.current = attachments
  }, [attachments])

  const { runtime, isRunning, notice, liveQuery, restoreInput } =
    useButterRuntime({
      sessionId,
      userId,
      agentId: agentId ?? null,
      onInvocationAccepted,
      pendingMessage,
      initialInvocationId,
      attachmentsRef,
      clearAttachments,
      addFiles,
    })

  const [pendingRestoreText, setPendingRestoreText] = useState<string | null>(
    null
  )
  const handleRestore = useCallback(async () => {
    const text = await restoreInput()
    if (text) setPendingRestoreText(text)
  }, [restoreInput])

  if (!session) {
    return (
      <div className='flex h-full items-center justify-center text-sm text-muted-foreground'>
        Select a chat in the sidebar or start a new one.
      </div>
    )
  }

  const activeNotice = notice && notice.sessionId === sessionId ? notice : null

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      <div
        className={cn(
          'flex min-h-0 flex-1 flex-col bg-background',
          isDragOver && 'ring-2 ring-ring/50 ring-inset'
        )}
        onDragOver={handleDragOver}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <ChatHeader
          session={session}
          agentName={agentName}
          sessionId={sessionId}
          onDelete={onDelete}
        />
        <ThreadArea
          agentName={agentName}
          isLoading={liveQuery.isLoading}
          isRunning={isRunning}
          notice={activeNotice}
          onRestore={handleRestore}
        />
        <ChatComposer
          agentName={agentName}
          attachments={attachments}
          previewUrls={previewUrls}
          onRemoveAttachment={removeAttachment}
          onOpenFilePicker={openFilePicker}
          onPaste={handlePaste}
          fileInputRef={fileInputRef}
          fileAccept={fileAccept}
          onFileInputChange={handleFileInputChange}
          validateForSend={validateForSend}
          pendingRestoreText={pendingRestoreText}
          onRestoreTextConsumed={() => setPendingRestoreText(null)}
          isRunning={isRunning}
        />
      </div>
    </AssistantRuntimeProvider>
  )
}

function ChatHeader({
  session,
  agentName,
  sessionId,
  onDelete,
}: {
  session: SessionInfo
  agentName: string | null
  sessionId: string
  onDelete?: () => void
}) {
  const renameMutation = useUpdateSessionTitle()
  const [editingTitle, setEditingTitle] = useState(false)

  return (
    <header className='flex h-12 shrink-0 items-center justify-between gap-2 border-b border-border/60 bg-background/95 px-2.5 sm:px-4 md:px-6'>
      <div className='flex min-w-0 flex-1 items-center gap-2.5'>
        <AgentAvatar name={agentName ?? '?'} size='sm' />
        <span className='flex max-w-md min-w-0 flex-1 flex-col'>
          {editingTitle ? (
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
              onClose={() => setEditingTitle(false)}
              className='text-sm font-medium'
            />
          ) : (
            <span className='flex min-w-0 items-center gap-1'>
              <span className='truncate text-sm leading-tight font-semibold'>
                {sessionTitle(session)}
              </span>
              <button
                type='button'
                onClick={() => setEditingTitle(true)}
                aria-label='Rename chat'
                title='Rename chat'
                className='inline-flex size-9 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100'
              >
                <Pencil className='size-3.5' />
              </button>
            </span>
          )}
          <span
            title={sessionId}
            className='truncate font-mono text-[0.65rem] leading-tight text-muted-foreground/80'
          >
            {agentName ?? 'Unknown agent'}
            <span className='hidden sm:inline'> / {sessionId.slice(0, 8)}</span>
          </span>
        </span>
      </div>

      <DropdownMenu>
        <DropdownMenuTrigger
          aria-label='Chat options'
          className='inline-flex size-9 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100'
        >
          <MoreHorizontal className='size-4' />
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' sideOffset={6}>
          <DropdownMenuItem variant='destructive' onClick={onDelete}>
            <Trash2 />
            Delete chat
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}

function ThreadArea({
  agentName,
  isLoading,
  isRunning,
  notice,
  onRestore,
}: {
  agentName: string | null
  isLoading: boolean
  isRunning: boolean
  notice: TerminalNotice | null
  onRestore?: () => void
}) {
  if (isLoading) {
    return (
      <div className='min-h-0 flex-1 overflow-y-auto'>
        <div className='mx-auto w-full max-w-4xl px-3.5 sm:px-5 lg:px-6'>
          <div className='space-y-6 py-8'>
            <div className='flex gap-3'>
              <Skeleton className='size-8 shrink-0 rounded-md' />
              <div className='w-full max-w-xl space-y-2'>
                <Skeleton className='h-3 w-28' />
                <Skeleton className='h-4 w-full' />
                <Skeleton className='h-4 w-4/5' />
              </div>
            </div>
            <Skeleton className='ml-auto h-14 w-1/2 max-w-md rounded-lg' />
          </div>
        </div>
      </div>
    )
  }

  return (
    <ThreadPrimitive.Root className='min-h-0 flex-1 overflow-y-auto'>
      <ThreadPrimitive.Viewport className='mx-auto w-full max-w-4xl px-3.5 sm:px-5 lg:px-6'>
        <ThreadPrimitive.Empty>
          <div className='flex min-h-[50vh] flex-col items-center justify-center text-center'>
            <AgentAvatar name={agentName ?? '?'} size='lg' />
            <h2 className='mt-3 text-lg font-semibold'>
              {agentName ?? 'Unknown agent'}
            </h2>
            <p className='mt-1 max-w-sm text-sm text-pretty text-muted-foreground'>
              Send a message below to start the conversation.
            </p>
          </div>
        </ThreadPrimitive.Empty>
        <div className='py-4 sm:py-6'>
          <ThreadPrimitive.Messages
            components={{
              UserMessage: UserMessageView,
              AssistantMessage: () => (
                <AssistantMessageView agentName={agentName ?? 'agent'} />
              ),
            }}
          />
          {isRunning && (
            <div className='grid grid-cols-[2rem_minmax(0,1fr)] gap-3 py-2'>
              <span />
              <div className='flex items-center gap-2 text-xs text-muted-foreground'>
                <Loader2 className='size-3 animate-spin' /> Thinking…
              </div>
            </div>
          )}
        </div>
        {notice && !isRunning && (
          <InvocationNotice notice={notice} onRestore={onRestore} />
        )}
      </ThreadPrimitive.Viewport>
    </ThreadPrimitive.Root>
  )
}

function UserMessageView() {
  return (
    <MessagePrimitive.Root className='flex flex-col items-end py-3 sm:py-4'>
      <div className='max-w-[92%] rounded-lg rounded-tr-sm bg-secondary px-3.5 py-2.5 text-[0.9rem] leading-relaxed text-secondary-foreground sm:max-w-[min(80%,48rem)]'>
        <MessagePrimitive.Content
          components={{
            Text: ({ text }) => (
              <ReactMarkdown
                remarkPlugins={REMARK_PLUGINS}
                components={MARKDOWN_COMPONENTS}
              >
                {text}
              </ReactMarkdown>
            ),
          }}
        />
      </div>
    </MessagePrimitive.Root>
  )
}

function AssistantMessageView({ agentName }: { agentName: string }) {
  return (
    <MessagePrimitive.Root className='group/message grid grid-cols-[1.75rem_minmax(0,1fr)] gap-2.5 pt-4 pb-2 sm:grid-cols-[2rem_minmax(0,1fr)] sm:gap-3'>
      <div className='pt-0.5'>
        <AgentAvatar name={agentName} size='sm' />
      </div>
      <div className='min-w-0'>
        <div className='mb-1 flex min-h-5 items-center gap-2'>
          <span className='text-sm font-medium'>{agentName}</span>
        </div>
        <div className='text-[0.9rem] leading-6 text-foreground'>
          <MessagePrimitive.Content
            components={{
              Text: MarkdownText,
              tools: { Fallback: ToolCallFallback },
            }}
          />
        </div>
        <MessagePrimitive.If lastOrHover>
          <div className='mt-1 flex min-h-6 items-center gap-0.5 text-muted-foreground opacity-100 transition-opacity sm:opacity-0 sm:group-hover/message:opacity-100 sm:focus-within:opacity-100'>
            <ActionBarPrimitive.Root>
              <ActionBarPrimitive.Copy asChild>
                <button
                  type='button'
                  title='Copy message'
                  aria-label='Copy message'
                  className='inline-flex size-8 items-center justify-center rounded-md transition-colors hover:bg-muted hover:text-foreground'
                >
                  <Copy className='size-3.5' />
                </button>
              </ActionBarPrimitive.Copy>
            </ActionBarPrimitive.Root>
          </div>
        </MessagePrimitive.If>
      </div>
    </MessagePrimitive.Root>
  )
}

function MarkdownText({ text }: { text: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={REMARK_PLUGINS}
      components={MARKDOWN_COMPONENTS}
    >
      {text}
    </ReactMarkdown>
  )
}

function ChatComposer({
  agentName,
  attachments,
  previewUrls,
  onRemoveAttachment,
  onOpenFilePicker,
  onPaste,
  fileInputRef,
  fileAccept,
  onFileInputChange,
  validateForSend: _validateForSend,
  pendingRestoreText,
  onRestoreTextConsumed,
  isRunning,
}: {
  agentName: string | null
  attachments: File[]
  previewUrls: string[]
  onRemoveAttachment: (index: number) => void
  onOpenFilePicker: () => void
  onPaste: (e: React.ClipboardEvent) => void
  fileInputRef: React.RefObject<HTMLInputElement | null>
  fileAccept: string
  onFileInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  validateForSend: () => string[]
  pendingRestoreText: string | null
  onRestoreTextConsumed: () => void
  isRunning: boolean
}) {
  return (
    <div className='shrink-0 border-t border-border/60 bg-background/95 backdrop-blur-sm'>
      <div className='mx-auto w-full max-w-4xl px-3 pt-2.5 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:px-5 md:pb-3.5'>
        {attachments.length > 0 && (
          <div className='mb-2 flex flex-wrap gap-1.5'>
            {attachments.map((file, index) => (
              <span
                key={`${file.name}-${index}`}
                className='group relative inline-flex'
              >
                <img
                  src={previewUrls[index]}
                  alt={file.name}
                  title={file.name}
                  className='h-14 w-14 rounded-md object-cover outline outline-1 -outline-offset-1 outline-black/10 dark:outline-white/10'
                />
                <button
                  type='button'
                  onClick={() => onRemoveAttachment(index)}
                  aria-label={`Remove ${file.name}`}
                  className='absolute -top-2 -right-2 inline-flex size-6 touch-manipulation items-center justify-center rounded-full border border-border bg-background text-muted-foreground shadow-sm transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100'
                >
                  <X className='size-3' />
                </button>
              </span>
            ))}
          </div>
        )}
        <ComposerPrimitive.Root className='flex items-end gap-1.5 rounded-lg border border-border/70 bg-card p-1.5 shadow-sm transition-[border-color,box-shadow] focus-within:border-foreground/25 focus-within:ring-2 focus-within:ring-ring/10 sm:gap-2 sm:p-2'>
          <input
            ref={fileInputRef}
            type='file'
            accept={fileAccept}
            multiple
            className='hidden'
            onChange={onFileInputChange}
          />
          <button
            type='button'
            disabled={!agentName || isRunning}
            onClick={onOpenFilePicker}
            aria-label='Attach images'
            className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] disabled:pointer-events-none motion-reduce:active:scale-100'
          >
            <Paperclip className='size-4' />
          </button>
          <ComposerPrimitive.Input
            autoFocus
            onPaste={onPaste}
            placeholder={
              agentName
                ? `Message ${agentName}...`
                : 'This chat is missing an agent reference; cannot send.'
            }
            rows={1}
            className='max-h-40 min-h-10 flex-1 resize-none bg-transparent py-2.5 text-[0.9rem] leading-5 outline-none placeholder:text-muted-foreground/75'
          />
          <ThreadPrimitive.If running>
            <ComposerPrimitive.Cancel asChild>
              <button
                type='button'
                aria-label='Stop generating'
                className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-secondary text-secondary-foreground transition-[background-color,scale] duration-150 ease-out hover:bg-secondary/80 active:scale-[0.96] motion-reduce:active:scale-100'
              >
                <Square className='size-4 fill-current' />
              </button>
            </ComposerPrimitive.Cancel>
          </ThreadPrimitive.If>
          <ThreadPrimitive.If running={false}>
            <ComposerPrimitive.Send asChild>
              <button
                type='button'
                aria-label='Send message'
                className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-primary text-primary-foreground transition-[background-color,opacity,scale] duration-150 ease-out hover:bg-primary/90 active:scale-[0.96] disabled:opacity-35 motion-reduce:active:scale-100'
              >
                <ArrowUp className='size-4' />
              </button>
            </ComposerPrimitive.Send>
          </ThreadPrimitive.If>
        </ComposerPrimitive.Root>
        <p className='mt-1.5 px-2 text-center text-[0.7rem] leading-4 text-muted-foreground/80'>
          Butter can make mistakes. Verify important actions before running
          them.
        </p>
      </div>
      <ComposerTextSetter
        text={pendingRestoreText}
        onDone={onRestoreTextConsumed}
      />
    </div>
  )
}

function ComposerTextSetter({
  text,
  onDone,
}: {
  text: string | null
  onDone: () => void
}) {
  const { setText } = unstable_useComposerInput()
  const onDoneRef = useRef(onDone)
  useEffect(() => {
    onDoneRef.current = onDone
  }, [onDone])

  useEffect(() => {
    if (text !== null) {
      setText(text)
      onDoneRef.current()
    }
  }, [text, setText])
  return null
}

function InvocationNotice({
  notice,
  onRestore,
}: {
  notice: TerminalNotice
  onRestore?: () => void
}) {
  const failed = notice.status === 'failed'
  return (
    <div
      role={failed ? 'alert' : 'status'}
      className={cn(
        'mb-4 rounded-lg border px-3.5 py-3 text-sm',
        failed
          ? 'border-destructive/35 bg-destructive/5'
          : 'border-border/70 bg-muted/30'
      )}
    >
      {notice.input && (
        <p className='mb-2 truncate border-l-2 border-border pl-2 text-[0.85rem] text-muted-foreground italic'>
          {notice.input}
        </p>
      )}
      <p
        className={cn(
          'font-medium',
          failed ? 'text-destructive' : 'text-foreground'
        )}
      >
        {failed ? 'This run failed' : 'Stopped'}
      </p>
      <p className='mt-1 text-[0.85rem] leading-5 text-muted-foreground'>
        {failed
          ? notice.error || 'The run ended with an error.'
          : 'You stopped this response before it finished.'}
      </p>
      {onRestore && (
        <div className='mt-2 flex flex-wrap items-center gap-x-3 gap-y-1.5'>
          <button
            type='button'
            onClick={onRestore}
            className='inline-flex touch-manipulation items-center gap-1.5 rounded-md border border-border bg-background px-2.5 py-1.5 text-xs font-medium transition-[background-color,scale] hover:bg-muted active:scale-[0.97] motion-reduce:active:scale-100'
          >
            <Undo2 className='size-3.5' />
            Restore input
          </button>
          <span className='text-[0.75rem] leading-4 text-muted-foreground/90'>
            Sending again starts a new run and may repeat external tool actions.
          </span>
        </div>
      )}
    </div>
  )
}

const HUMAN_INPUT_TOOL = 'ask_user'

function ToolCallFallback({
  toolName,
  args,
  result,
  status,
}: {
  toolName: string
  args?: Record<string, unknown>
  result?: unknown
  status?: { type: string }
}) {
  const isHumanInput = toolName === HUMAN_INPUT_TOOL
  const isRunning = status?.type === 'running'
  const hasResult = result !== undefined

  if (isHumanInput) {
    return (
      <HumanInputView
        args={args}
        result={result}
        isRunning={isRunning}
        hasResult={hasResult}
      />
    )
  }

  return (
    <details className='group/tool my-2 max-w-[72ch] overflow-hidden rounded-md border border-border/60 bg-muted/20'>
      <summary className='flex cursor-pointer list-none items-center gap-2 px-2.5 py-1.5 text-left transition-colors hover:bg-muted/45 [&::-webkit-details-marker]:hidden'>
        <Wrench className='size-3.5 shrink-0 text-muted-foreground' />
        <span className='truncate font-mono text-xs text-foreground/85'>
          {toolName}
        </span>
        {isRunning && !hasResult && (
          <Loader2 className='size-3 shrink-0 animate-spin text-muted-foreground' />
        )}
        {hasResult && (
          <span className='shrink-0 text-[0.7rem] text-muted-foreground'>
            done
          </span>
        )}
        <ChevronDown className='ml-auto size-3.5 shrink-0 text-muted-foreground transition-transform group-open/tool:rotate-180' />
      </summary>
      <div className='space-y-2 border-t border-border/60 px-2.5 py-2'>
        {args && Object.keys(args).length > 0 && (
          <div>
            <div className='mb-1 text-[0.68rem] text-muted-foreground'>
              Arguments
            </div>
            <pre className='max-h-64 scrollbar-thin overflow-auto rounded bg-muted/55 p-2 font-mono text-xs leading-5'>
              {formatJson(args)}
            </pre>
          </div>
        )}
        {hasResult && (
          <div>
            <div className='mb-1 text-[0.68rem] text-muted-foreground'>
              Result
            </div>
            <pre className='max-h-64 scrollbar-thin overflow-auto rounded bg-muted/55 p-2 font-mono text-xs leading-5'>
              {formatJson(result)}
            </pre>
          </div>
        )}
      </div>
    </details>
  )
}

function HumanInputView({
  args,
  result,
  isRunning,
  hasResult,
}: {
  args?: Record<string, unknown>
  result?: unknown
  isRunning: boolean
  hasResult: boolean
}) {
  const question =
    typeof args?.question === 'string'
      ? args.question
      : typeof args?.message === 'string'
        ? args.message
        : formatJson(args)

  return (
    <div
      className={cn(
        'my-2 max-w-[72ch] rounded-lg border px-3.5 py-3 text-sm',
        isRunning && !hasResult
          ? 'border-amber-500/35 bg-amber-500/5'
          : 'border-border/70 bg-muted/30'
      )}
    >
      <p className='font-medium text-foreground'>
        {isRunning && !hasResult ? 'Waiting for input' : 'Human Input'}
      </p>
      <p className='mt-1 text-[0.85rem] leading-5 text-muted-foreground'>
        {question}
      </p>
      {hasResult && (
        <p className='mt-2 border-l-2 border-border pl-2 text-[0.85rem] text-foreground'>
          {typeof result === 'string' ? result : formatJson(result)}
        </p>
      )}
      {isRunning && !hasResult && (
        <p className='mt-2 text-[0.75rem] text-amber-600 dark:text-amber-400'>
          Send a message below to answer this question.
        </p>
      )}
    </div>
  )
}

function formatJson(value: unknown): string {
  if (value === null || value === undefined) return ''
  try {
    return typeof value === 'string' ? value : JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

const MARKDOWN_COMPONENTS: Components = {
  a: MarkdownLink,
  code: MarkdownCode,
  pre: MarkdownPre,
  table: MarkdownTable,
  th: MarkdownTableHeader,
  td: MarkdownTableCell,
  p: ({ children }) => (
    <p className='my-2 max-w-[72ch] first:mt-0 last:mb-0'>{children}</p>
  ),
  ul: ({ children }) => (
    <ul className='my-2 max-w-[72ch] list-disc space-y-1 pl-5 first:mt-0 last:mb-0'>
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className='my-2 max-w-[72ch] list-decimal space-y-1 pl-5 first:mt-0 last:mb-0'>
      {children}
    </ol>
  ),
  li: ({ children }) => <li className='pl-1'>{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className='my-2 max-w-[72ch] border-l-2 border-border pl-3 text-muted-foreground italic first:mt-0 last:mb-0'>
      {children}
    </blockquote>
  ),
  hr: () => <hr className='my-3 border-border' />,
  h1: ({ children }) => (
    <h1 className='my-3 max-w-[72ch] text-lg font-semibold first:mt-0 last:mb-0'>
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className='my-3 max-w-[72ch] text-base font-semibold first:mt-0 last:mb-0'>
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className='my-2 max-w-[72ch] text-sm font-semibold first:mt-0 last:mb-0'>
      {children}
    </h3>
  ),
}

function MarkdownLink(props: ComponentProps<'a'>) {
  return (
    <a
      {...props}
      target='_blank'
      rel='noopener noreferrer'
      className='font-medium underline underline-offset-2 hover:opacity-80'
    />
  )
}

function MarkdownCode({ children, className }: ComponentProps<'code'>) {
  const isInline = !className
  if (isInline) {
    return (
      <code className='rounded bg-muted px-1 py-0.5 font-mono text-[0.85em] text-foreground'>
        {children}
      </code>
    )
  }
  return <code className={cn('font-mono text-xs', className)}>{children}</code>
}

function MarkdownPre({ children }: ComponentProps<'pre'>) {
  return (
    <pre className='my-2 max-w-full scrollbar-thin overflow-x-auto rounded-md border border-border/70 bg-muted/35 p-3 text-foreground first:mt-0 last:mb-0'>
      {children}
    </pre>
  )
}

function MarkdownTable({ children }: ComponentProps<'table'>) {
  return (
    <div className='my-3 max-w-full scrollbar-thin overflow-x-auto rounded-md border border-border/70 first:mt-0 last:mb-0'>
      <table className='w-full min-w-[42rem] border-separate border-spacing-0 text-left text-xs'>
        {children}
      </table>
    </div>
  )
}

function MarkdownTableHeader({ children }: ComponentProps<'th'>) {
  return (
    <th className='border-b border-border bg-muted/55 px-3 py-2 font-semibold whitespace-nowrap text-foreground'>
      {children}
    </th>
  )
}

function MarkdownTableCell({ children }: ComponentProps<'td'>) {
  return (
    <td className='border-b border-border/50 px-3 py-2 align-top leading-5 last:border-r-0'>
      {children}
    </td>
  )
}
