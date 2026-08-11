import { useRef, useState } from 'react'
import type { Agent } from '@/types/api'
import { ArrowUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AgentAvatar } from '@/components/butter/primitives'
import { agentIconUrl } from '@/features/agents/icon-utils'

interface DraftComposerProps {
  agent: Agent | null
  agentSelector: React.ReactNode
  onSend: (message: string) => void
  busy?: boolean
}

export function DraftComposer({
  agent,
  agentSelector,
  onSend,
  busy,
}: DraftComposerProps) {
  const [draft, setDraft] = useState('')
  const taRef = useRef<HTMLTextAreaElement>(null)

  const canSend = !!agent && draft.trim().length > 0 && !busy

  function handleSend() {
    const text = draft.trim()
    if (!text || !agent || busy) return
    onSend(text)
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (
      e.key === 'Enter' &&
      !e.shiftKey &&
      !e.nativeEvent.isComposing &&
      e.keyCode !== 229
    ) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-1 items-center justify-center px-4 py-8 sm:px-6">
        <div className="w-full max-w-2xl">
          {agent ? (
            <div className="mb-8 flex flex-col items-center text-center">
              <AgentAvatar
                name={agent.name}
                iconUrl={agentIconUrl(agent)}
                size="lg"
              />
              <h2 className="mt-3 text-lg font-semibold">{agent.name}</h2>
              {agent.description && (
                <p className="mt-1 max-w-md text-sm text-muted-foreground text-pretty">
                  {agent.description}
                </p>
              )}
            </div>
          ) : (
            <div className="mb-8 flex flex-col items-center text-center">
              <h1 className="font-manrope text-2xl font-semibold text-balance">
                Start a new chat
              </h1>
              <p className="mt-1.5 text-sm text-muted-foreground">
                Choose an agent and send a message to begin.
              </p>
            </div>
          )}

          <div className="mb-4 flex justify-center">{agentSelector}</div>

          <div
            className={cn(
              'flex items-end gap-1.5 rounded-lg border border-border/70 bg-card p-1.5 shadow-sm transition-[border-color,box-shadow] focus-within:border-foreground/25 focus-within:ring-2 focus-within:ring-ring/10 sm:gap-2 sm:p-2',
              !agent && 'opacity-60',
              busy && 'pointer-events-none opacity-60',
            )}
          >
            <textarea
              ref={taRef}
              rows={1}
              value={draft}
              disabled={!agent || busy}
              onChange={(e) => {
                setDraft(e.target.value)
                e.target.style.height = 'auto'
                e.target.style.height = `${Math.min(e.target.scrollHeight, 160)}px`
              }}
              onKeyDown={handleKeyDown}
              placeholder={
                agent
                  ? `Message ${agent.name}…`
                  : 'Select an agent above to start chatting…'
              }
              data-testid="draft-composer-input"
              className="max-h-40 min-h-10 flex-1 resize-none bg-transparent py-2.5 pl-2 text-[0.9rem] leading-5 outline-none placeholder:text-muted-foreground/75"
            />
            <button
              type="button"
              onClick={handleSend}
              disabled={!canSend}
              data-testid="draft-composer-send"
              aria-label="Send message"
              className="inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-primary text-primary-foreground transition-[background-color,opacity,scale] duration-150 ease-out hover:bg-primary/90 active:scale-[0.96] motion-reduce:active:scale-100 disabled:opacity-35"
            >
              <ArrowUp className="size-4" />
            </button>
          </div>
          <p className="mt-1.5 px-2 text-center text-[0.7rem] leading-4 text-muted-foreground/80">
            Butter can make mistakes. Verify important actions before running
            them.
          </p>
        </div>
      </div>
    </div>
  )
}
