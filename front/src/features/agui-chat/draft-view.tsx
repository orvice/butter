import { useState, type KeyboardEvent, type ReactNode } from 'react'
import type { Agent } from '@/types/api'
import { ArrowUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AgentHero, ChatDisclaimer } from '@/components/chat/thread-states'
import { agentIconUrl } from '@/features/agents/icon-utils'

// DraftView is a new chat before its first message: the agent it will run
// with, the selector to change that agent, and the composer. Nothing exists
// on the server yet; the first message starts the thread.
export function DraftView({
  agent,
  agentSelector,
  onSend,
}: {
  agent: Agent | null
  agentSelector: ReactNode
  onSend: (message: string) => void
}) {
  const [draft, setDraft] = useState('')
  const canSend = !!agent && draft.trim().length > 0

  function send() {
    const text = draft.trim()
    if (!text || !agent) return
    onSend(text)
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (
      e.key === 'Enter' &&
      !e.shiftKey &&
      !e.nativeEvent.isComposing &&
      e.keyCode !== 229
    ) {
      e.preventDefault()
      send()
    }
  }

  return (
    <div className='relative flex min-h-0 flex-1 flex-col'>
      <div className='flex flex-1 items-center justify-center overflow-y-auto px-4 py-8 sm:px-6'>
        <div className='w-full max-w-2xl'>
          {agent ? (
            <AgentHero
              name={agent.name}
              iconUrl={agentIconUrl(agent) || undefined}
              className='mb-8 min-h-0'
            />
          ) : (
            <div className='mb-8 flex flex-col items-center text-center'>
              <h1 className='font-manrope text-2xl font-semibold text-balance'>
                Start a new chat
              </h1>
              <p className='mt-1.5 text-sm text-muted-foreground'>
                Choose an agent and send a message to begin.
              </p>
            </div>
          )}

          <div className='mb-4 flex justify-center'>{agentSelector}</div>

          <div
            className={cn(
              'flex items-end gap-1.5 rounded-lg border border-border/70 bg-card p-1.5 shadow-sm transition-[border-color,box-shadow] focus-within:border-foreground/25 focus-within:ring-2 focus-within:ring-ring/10 sm:gap-2 sm:p-2',
              !agent && 'opacity-60'
            )}
          >
            <textarea
              rows={1}
              value={draft}
              disabled={!agent}
              onChange={(e) => {
                setDraft(e.target.value)
                e.target.style.height = 'auto'
                e.target.style.height = `${Math.min(e.target.scrollHeight, 160)}px`
              }}
              onKeyDown={handleKeyDown}
              aria-label='Message'
              placeholder={
                agent
                  ? `Message ${agent.name}…`
                  : 'Choose an agent above to start chatting…'
              }
              className='max-h-40 min-h-10 flex-1 resize-none bg-transparent py-2.5 pl-2 text-[0.9rem] leading-5 outline-none placeholder:text-muted-foreground/75'
            />
            <button
              type='button'
              onClick={send}
              disabled={!canSend}
              aria-label='Send message'
              className='inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-primary text-primary-foreground transition-[background-color,opacity,scale] duration-150 ease-out hover:bg-primary/90 active:scale-[0.96] disabled:opacity-35 motion-reduce:active:scale-100'
            >
              <ArrowUp className='size-4' />
            </button>
          </div>
          <ChatDisclaimer />
        </div>
      </div>
    </div>
  )
}
