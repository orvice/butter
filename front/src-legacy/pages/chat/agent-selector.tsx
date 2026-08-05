import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useAgents } from "@/api/agents";
import { AgentAvatar } from "@/components/butter/primitives";
import { agentIconUrl } from "@/pages/agents/icon-utils";
import { cn } from "@/lib/utils";
import { Loader2, Search } from "lucide-react";
import type { Agent } from "@/types/api";

export function AgentSelector({
  onPick,
  busy,
}: {
  onPick: (agentName: string) => void;
  busy?: boolean;
}) {
  const { data, isLoading } = useAgents({ page_size: 200 });
  const [query, setQuery] = useState("");

  const agents = useMemo(() => data?.agents ?? [], [data]);

  const filtered = useMemo(
    () =>
      agents.filter(
        (a) =>
          a.name.toLowerCase().includes(query.toLowerCase()) ||
          (a.description ?? "").toLowerCase().includes(query.toLowerCase()),
      ),
    [agents, query],
  );

  if (isLoading) {
    return (
      <div className="flex flex-col items-center gap-3 text-sm text-muted-foreground">
        <Loader2 className="size-5 animate-spin" />
        Loading agents…
      </div>
    );
  }

  if (agents.length === 0) {
    return (
      <div className="mx-auto max-w-md rounded-lg border border-dashed border-border bg-card/50 px-6 py-10 text-center">
        <h2 className="text-base font-semibold">No agents yet</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Create your first agent to start chatting.
        </p>
        <Link
          to="/agents/create"
          className="mt-4 inline-block rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Create Agent
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-2xl">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-balance">
          What can we get done today?
        </h1>
        <p className="mt-2 text-sm text-muted-foreground text-pretty">
          Pick an agent to start a new conversation. Each chat stays with one agent.
        </p>
      </div>

      <div className="relative mx-auto mt-6 max-w-md">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search agents"
          className="w-full rounded-md border border-border bg-card py-2 pl-9 pr-3 text-sm outline-none focus:border-ring"
        />
      </div>

      <div className="mt-6">
        {filtered.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">
            No agents match “{query}”.
          </p>
        ) : (
          <div className="grid gap-2 sm:grid-cols-2">
            {filtered.map((a: Agent) => (
              <button
                key={a.name}
                type="button"
                disabled={busy}
                onClick={() => onPick(a.name)}
                className={cn(
                  "flex items-start gap-3 rounded-lg border border-border bg-card p-3 text-left transition-colors",
                  busy
                    ? "cursor-wait opacity-60"
                    : "hover:border-ring/60 hover:bg-accent/50",
                )}
              >
                <AgentAvatar name={a.name} iconUrl={agentIconUrl(a)} size="md" />
                <div className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">{a.name}</span>
                  <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
                    {a.description || "No description."}
                  </p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
