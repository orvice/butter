import { useCallback } from 'react'
import {
  ActionBarPrimitive,
  AuiIf,
  MessagePrimitive,
  ThreadPrimitive,
  type AssistantState,
  type MessageState,
} from '@assistant-ui/react'
import { Copy } from 'lucide-react'
import { AgentAvatar } from '@/components/butter/primitives'
import { useChatAgent } from './chat-agent'
import { MessageErrorBoundary } from './error-boundary'
import { UserMarkdownText } from './markdown'

// PartComponents is how a chat draws the parts of a message: its text, tool
// calls and data parts.
export type PartComponents = MessagePrimitive.Parts.Props['components']

// ThreadMessages draws a thread's messages: what the user sent, and each
// reply with replyParts. replyParts must keep its identity across renders,
// or every message is drawn again.
export function ThreadMessages({ replyParts }: { replyParts: PartComponents }) {
  const render = useCallback(
    ({ message }: { message: MessageState }) => {
      switch (message.role) {
        case 'user':
          return <UserMessage />
        case 'assistant':
          return <AssistantMessage parts={replyParts} />
        default:
          return null
      }
    },
    [replyParts]
  )
  return <ThreadPrimitive.Messages>{render}</ThreadPrimitive.Messages>
}

const USER_PARTS: PartComponents = { Text: UserMarkdownText }

// UserMessage is a message the user sent, drawn as Markdown that keeps the
// lines they typed.
function UserMessage() {
  return (
    <MessagePrimitive.Root
      data-message-role='user'
      className='flex flex-col items-end py-3 sm:py-4'
    >
      <div className='max-w-[92%] rounded-lg rounded-tr-sm bg-secondary px-3.5 py-2.5 text-[0.9rem] leading-relaxed text-secondary-foreground sm:max-w-[min(80%,48rem)]'>
        <MessageErrorBoundary>
          <MessagePrimitive.Parts components={USER_PARTS} />
        </MessageErrorBoundary>
      </div>
    </MessagePrimitive.Root>
  )
}

const lastOrHovered = (s: AssistantState) =>
  s.message.isLast || s.message.isHovering

// AssistantMessage is one reply: the agent's avatar and name, the reply's
// parts as the chat draws them, and a copy action.
function AssistantMessage({ parts }: { parts: PartComponents }) {
  const agent = useChatAgent()
  return (
    <MessagePrimitive.Root
      data-message-role='assistant'
      className='group/message grid grid-cols-[1.75rem_minmax(0,1fr)] gap-2.5 pt-4 pb-2 sm:grid-cols-[2rem_minmax(0,1fr)] sm:gap-3'
    >
      <div className='pt-0.5'>
        <AgentAvatar name={agent.name} iconUrl={agent.iconUrl} size='sm' />
      </div>
      <div className='min-w-0'>
        <div className='mb-1 flex min-h-5 items-center gap-2'>
          <span className='text-sm font-medium'>{agent.name}</span>
        </div>
        <div className='text-[0.9rem] leading-6 text-foreground'>
          <MessageErrorBoundary>
            <MessagePrimitive.Parts components={parts} />
          </MessageErrorBoundary>
        </div>
        <AuiIf condition={lastOrHovered}>
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
        </AuiIf>
      </div>
    </MessagePrimitive.Root>
  )
}
