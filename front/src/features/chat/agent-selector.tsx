import { useMemo, useRef, useState } from 'react'
import { Link } from '@tanstack/react-router'
import type { Agent } from '@/types/api'
import { Check, ChevronDown, Search } from 'lucide-react'
import { useAgents } from '@/api/agents'
import { cn } from '@/lib/utils'
import { AgentAvatar } from '@/components/butter/primitives'
import { agentIconUrl } from '@/features/agents/icon-utils'

function isRunnableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  return (
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  )
}

export function AgentSelector({
  selected,
  onPick,
}: {
  selected: Agent | null
  onPick: (agent: Agent) => void
}) {
  const { data, isLoading } = useAgents({ page_size: 200 })
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const agents = useMemo(
    () => (data?.agents ?? []).filter(isRunnableAgent),
    [data],
  )

  const filtered = useMemo(
    () =>
      agents.filter(
        (a) =>
          a.name.toLowerCase().includes(query.toLowerCase()) ||
          (a.description ?? '').toLowerCase().includes(query.toLowerCase()),
      ),
    [agents, query],
  )

  function handleSelect(agent: Agent) {
    onPick(agent)
    setOpen(false)
    setQuery('')
  }

  function handleToggle() {
    if (open) {
      setOpen(false)
      setQuery('')
    } else {
      setOpen(true)
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }

  function handleBlur(e: React.FocusEvent) {
    if (containerRef.current?.contains(e.relatedTarget as Node)) return
    setOpen(false)
    setQuery('')
  }

  if (isLoading) {
    return (
      <div className="h-10 w-64 animate-pulse rounded-lg border border-border bg-card" />
    )
  }

  if (agents.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-border bg-card/50 px-4 py-3 text-center">
        <p className="text-sm text-muted-foreground">No agents available.</p>
        <Link
          to="/agents/create"
          className="mt-2 inline-block text-sm font-medium text-primary hover:underline"
        >
          Create an agent
        </Link>
      </div>
    )
  }

  return (
    <div ref={containerRef} className="relative w-full max-w-xs" onBlur={handleBlur}>
      <button
        type="button"
        onClick={handleToggle}
        data-testid="agent-selector-trigger"
        className={cn(
          'flex h-10 w-full items-center gap-2.5 rounded-lg border border-border bg-card px-3 text-left text-sm transition-[border-color,box-shadow] hover:border-ring/60 focus:border-ring focus:ring-2 focus:ring-ring/10 focus:outline-none',
          open && 'border-ring ring-2 ring-ring/10',
        )}
      >
        {selected ? (
          <>
            <AgentAvatar
              name={selected.name}
              iconUrl={agentIconUrl(selected)}
              size="sm"
              className="size-5 shrink-0 rounded text-[0.55rem]"
            />
            <span className="min-w-0 flex-1 truncate font-medium">
              {selected.name}
            </span>
          </>
        ) : (
          <span className="min-w-0 flex-1 truncate text-muted-foreground">
            Choose an agent…
          </span>
        )}
        <ChevronDown
          className={cn(
            'size-4 shrink-0 text-muted-foreground transition-transform',
            open && 'rotate-180',
          )}
        />
      </button>

      {open && (
        <div className="absolute top-full left-0 z-50 mt-1.5 w-72 rounded-lg border border-border bg-popover shadow-lg">
          <div className="relative border-b border-border/60 px-2.5 py-2">
            <Search className="pointer-events-none absolute top-1/2 left-5 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <input
              ref={inputRef}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search agents…"
              className="h-8 w-full rounded-md bg-transparent pl-7 pr-2 text-sm outline-none placeholder:text-muted-foreground/75"
            />
          </div>
          <div className="scrollbar-thin max-h-64 overflow-y-auto py-1">
            {filtered.length === 0 ? (
              <p className="px-3 py-4 text-center text-xs text-muted-foreground">
                No agents match &ldquo;{query}&rdquo;.
              </p>
            ) : (
              filtered.map((a) => {
                const isSelected =
                  selected &&
                  ((a.agent_id && a.agent_id === selected.agent_id) ||
                    a.name === selected.name)
                return (
                  <button
                    key={a.agent_id || a.name}
                    type="button"
                    data-testid={`agent-option-${a.agent_id || a.name}`}
                    onClick={() => handleSelect(a)}
                    className={cn(
                      'flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm transition-colors hover:bg-accent/50',
                      isSelected && 'bg-accent/30',
                    )}
                  >
                    <AgentAvatar
                      name={a.name}
                      iconUrl={agentIconUrl(a)}
                      size="sm"
                      className="size-6 shrink-0 rounded text-[0.55rem]"
                    />
                    <div className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium">
                        {a.name}
                      </span>
                      {a.description && (
                        <span className="block truncate text-xs text-muted-foreground">
                          {a.description}
                        </span>
                      )}
                    </div>
                    {isSelected && (
                      <Check className="size-3.5 shrink-0 text-primary" />
                    )}
                  </button>
                )
              })
            )}
          </div>
        </div>
      )}
    </div>
  )
}
