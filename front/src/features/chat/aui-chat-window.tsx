import {
  Component,
  type ErrorInfo,
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import type { SessionInfo } from '@/types/api'
import {
  AssistantRuntimeProvider,
  AuiIf,
  ThreadPrimitive,
  ComposerPrimitive,
  unstable_useComposerInput,
} from '@assistant-ui/react'
import {
  ArrowUp,
  Loader2,
  MoreHorizontal,
  Paperclip,
  Pencil,
  Square,
  Trash2,
  X,
} from 'lucide-react'
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
import { AgentAvatar } from '@/components/butter/primitives'
import { ChatAgentContext, type ChatAgent } from '@/components/chat/chat-agent'
import { MarkdownText } from '@/components/chat/markdown'
import { ThreadMessages, type PartComponents } from '@/components/chat/messages'
import { RunNotice } from '@/components/chat/run-notice'
import {
  AgentHero,
  ChatDisclaimer,
  ThreadSkeleton,
} from '@/components/chat/thread-states'
import { ToolCallView } from '@/components/chat/tool-views'
import { chatAuiConfig } from '@/components/chat/toolkit'
import { InlineTitleInput } from '@/components/inline-title-input'
import { useButterRuntime, type TerminalNotice } from './butter-runtime'

// A reply's parts: Markdown text, and tool calls (the shared toolkit draws
// adk_request_input, ToolCallView the rest).
const REPLY_PARTS: PartComponents = {
  Text: MarkdownText,
  tools: { Fallback: ToolCallView },
}

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

type AUIChatWindowInnerProps = Omit<AUIChatWindowProps, 'session'> & {
  session: SessionInfo
}

export function AUIChatWindow(props: AUIChatWindowProps) {
  if (!props.session) {
    return (
      <div className='flex h-full items-center justify-center text-sm text-muted-foreground'>
        Select a chat in the sidebar or start a new one.
      </div>
    )
  }

  return (
    <RuntimeErrorBoundary key={props.session.session_id}>
      <AUIChatWindowInner {...props} session={props.session} />
    </RuntimeErrorBoundary>
  )
}

function AUIChatWindowInner({
  session,
  userId,
  agentName,
  agentId,
  onDelete,
  onInvocationAccepted,
  pendingMessage,
  initialInvocationId,
}: AUIChatWindowInnerProps) {
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
      validateForSend,
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

  const activeNotice = notice && notice.sessionId === sessionId ? notice : null

  return (
    <AssistantRuntimeProvider runtime={runtime} config={chatAuiConfig}>
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
  const chatAgent = useMemo<ChatAgent>(
    () => ({ name: agentName ?? 'agent' }),
    [agentName]
  )

  if (isLoading) {
    return (
      <div className='min-h-0 flex-1 overflow-y-auto'>
        <div className='mx-auto w-full max-w-4xl px-3.5 sm:px-5 lg:px-6'>
          <ThreadSkeleton />
        </div>
      </div>
    )
  }

  return (
    <ChatAgentContext.Provider value={chatAgent}>
      <ThreadPrimitive.Root className='min-h-0 flex-1 overflow-y-auto'>
        <ThreadPrimitive.Viewport className='mx-auto w-full max-w-4xl px-3.5 sm:px-5 lg:px-6'>
          <AuiIf condition={(s) => s.thread.isEmpty}>
            <AgentHero name={agentName} />
          </AuiIf>
          <div className='py-4 sm:py-6'>
            <ThreadMessages replyParts={REPLY_PARTS} />
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
            <RunNotice notice={notice} onRestore={onRestore} />
          )}
        </ThreadPrimitive.Viewport>
      </ThreadPrimitive.Root>
    </ChatAgentContext.Provider>
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
          <AuiIf condition={(s) => s.thread.isRunning}>
            <ComposerPrimitive.Cancel asChild>
              <button
                type='button'
                aria-label='Stop generating'
                className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-secondary text-secondary-foreground transition-[background-color,scale] duration-150 ease-out hover:bg-secondary/80 active:scale-[0.96] motion-reduce:active:scale-100'
              >
                <Square className='size-4 fill-current' />
              </button>
            </ComposerPrimitive.Cancel>
          </AuiIf>
          <AuiIf condition={(s) => !s.thread.isRunning}>
            <ComposerPrimitive.Send asChild>
              <button
                type='button'
                aria-label='Send message'
                className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-primary text-primary-foreground transition-[background-color,opacity,scale] duration-150 ease-out hover:bg-primary/90 active:scale-[0.96] disabled:opacity-35 motion-reduce:active:scale-100'
              >
                <ArrowUp className='size-4' />
              </button>
            </ComposerPrimitive.Send>
          </AuiIf>
        </ComposerPrimitive.Root>
        <ChatDisclaimer />
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

class RuntimeErrorBoundary extends Component<
  { children: ReactNode },
  { error: Error | null }
> {
  state: { error: Error | null } = { error: null }

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    if (
      error.message?.includes('unstable_state') ||
      error.message?.includes('isOptimistic')
    ) {
      setTimeout(() => this.setState({ error: null }), 0)
      return
    }
    void error
    void info
  }

  render() {
    if (this.state.error) return null
    return this.props.children
  }
}
