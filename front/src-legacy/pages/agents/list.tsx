import { useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useAgents,
  useDeleteAgent,
  useReloadAgents,
  useInvokeAgent,
  useAgentRuntimeStatuses,
} from "@/api/agents";
import { DeleteDialog } from "@/components/delete-dialog";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { AgentAvatar, StatusBadge, type RunStatus } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Bot,
  MessageSquarePlus,
  MoreVertical,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Search,
  Trash2,
} from "lucide-react";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import { agentIconUrl } from "./icon-utils";
import { cn } from "@/lib/utils";
import type { Agent, AgentRuntimeStatus } from "@/types/api";

type RuntimeState = "running" | "idle" | "failed";
type RuntimeFilter = "all" | RuntimeState;

function runtimeStatusOf(rt?: AgentRuntimeStatus): RuntimeState {
  switch (rt?.state) {
    case "AGENT_RUNTIME_STATE_RUNNING":
      return "running";
    case "AGENT_RUNTIME_STATE_FAILED":
      return "failed";
    default:
      return "idle";
  }
}

const RUNTIME_TO_BADGE: Record<RuntimeState, { status: RunStatus; label: string }> = {
  running: { status: "running", label: "Running" },
  idle: { status: "success", label: "Available" },
  failed: { status: "failed", label: "Failed" },
};

function timeAgo(ts?: string): string | null {
  if (!ts) return null;
  const d = Date.now() - new Date(ts).getTime();
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`;
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`;
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`;
  return `${Math.floor(d / 86_400_000)}d ago`;
}

function AgentCard({
  agent,
  runtime,
  onDelete,
  onRun,
}: {
  agent: Agent;
  runtime?: AgentRuntimeStatus;
  onDelete: () => void;
  onRun: () => void;
}) {
  const navigate = useNavigate();
  const status = runtimeStatusOf(runtime);
  const badge = RUNTIME_TO_BADGE[status];
  const lastRun = timeAgo(runtime?.last_run_at);

  return (
    <div className="group relative flex flex-col rounded-lg border border-transparent bg-card p-4 shadow-card transition-[background-color,border-color,box-shadow] duration-150 ease-out hover:border-ring/40 hover:shadow-card-hover">
      <div className="flex items-start gap-3">
        <AgentAvatar name={agent.name} iconUrl={agentIconUrl(agent)} size="lg" />
        <div className="min-w-0 flex-1">
          <h3 className="truncate text-sm font-semibold">{agent.name}</h3>
          <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-muted-foreground">
            {agent.description || "No description."}
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger className="rounded-md p-1.5 text-muted-foreground hover:bg-muted">
            <MoreVertical className="size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" sideOffset={6}>
            <DropdownMenuItem onClick={() => navigate(`/agents/${encodeURIComponent(agent.name)}/edit`)}>
              <Pencil /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem onClick={onRun}>
              <Play /> Run once
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={onDelete}>
              <Trash2 /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="mt-3 flex items-center gap-2 text-xs">
        <StatusBadge status={badge.status} label={badge.label} />
        {(runtime?.in_flight ?? 0) > 0 && (
          <span className="text-muted-foreground">×{runtime!.in_flight} in flight</span>
        )}
      </div>

      <div className="mt-3 flex flex-wrap gap-1">
        <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 font-mono text-[0.7rem] text-muted-foreground">
          {AGENT_TYPE_LABELS[agent.type ?? "AGENT_TYPE_UNSPECIFIED"]}
        </span>
        {agent.enable_a2a && (
          <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[0.7rem] text-muted-foreground">
            A2A
          </span>
        )}
        {agent.enable_openai_api && (
          <span className="rounded border border-border bg-muted/50 px-1.5 py-0.5 text-[0.7rem] text-muted-foreground">
            OpenAI API
          </span>
        )}
      </div>

      <div className="mt-3 flex items-center justify-between border-t border-border pt-3">
        <span className="text-xs text-muted-foreground">
          {lastRun ? `Last run ${lastRun}` : "Not run yet"}
        </span>
        <Button size="sm" render={<Link to={`/chat?new=1&agent=${encodeURIComponent(agent.name)}`} />}>
          <MessageSquarePlus />
          Start Chat
        </Button>
      </div>
    </div>
  );
}

export default function AgentListPage() {
  const { data, isLoading } = useAgents();
  const agents = useMemo(() => data?.agents ?? [], [data?.agents]);
  const names = useMemo(() => agents.map((a) => a.name), [agents]);
  const { data: runtimeData } = useAgentRuntimeStatuses(names);

  const runtimeMap = useMemo(() => {
    const m = new Map<string, AgentRuntimeStatus>();
    for (const s of runtimeData?.statuses ?? []) m.set(s.name, s);
    return m;
  }, [runtimeData]);

  const deleteMutation = useDeleteAgent();
  const reloadMutation = useReloadAgents();
  const invokeMutation = useInvokeAgent();

  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<RuntimeFilter>("all");
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [invokeTarget, setInvokeTarget] = useState<Agent | null>(null);
  const [invokeInput, setInvokeInput] = useState("");
  const [invokeResult, setInvokeResult] = useState<{ session_id: string; response: string } | null>(null);

  const filtered = useMemo(
    () =>
      agents.filter(
        (a) =>
          (filter === "all" || runtimeStatusOf(runtimeMap.get(a.name)) === filter) &&
          (a.name.toLowerCase().includes(query.toLowerCase()) ||
            (a.description ?? "").toLowerCase().includes(query.toLowerCase())),
      ),
    [agents, query, filter, runtimeMap],
  );

  const filters: { key: RuntimeFilter; label: string }[] = [
    { key: "all", label: "All" },
    { key: "running", label: "Running" },
    { key: "idle", label: "Idle" },
    { key: "failed", label: "Failed" },
  ];

  return (
    <Page>
      <PageHeader
        title="Agents"
        subtitle="Browse agents and start a conversation, or configure how they work."
        actions={
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                reloadMutation.mutate(undefined, {
                  onSuccess: () => toast.success("Agents reloaded"),
                  onError: (err) => toast.error(err.message),
                })
              }
              disabled={reloadMutation.isPending}
            >
              <RefreshCw className={cn("size-4", reloadMutation.isPending && "animate-spin")} />
              Hot-reload
            </Button>
            <Button size="sm" render={<Link to="/agents/create" />}>
              <Plus className="size-4" />
              Create Agent
            </Button>
          </>
        }
      />
      <PageScroll>
        {isLoading ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-48" />
            ))}
          </div>
        ) : agents.length === 0 ? (
          <div className="mx-auto max-w-md rounded-lg border border-dashed border-border bg-card/40 px-6 py-14 text-center">
            <div className="mx-auto flex size-11 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <Bot className="size-5" />
            </div>
            <h2 className="mt-4 text-base font-semibold">No agents yet</h2>
            <p className="mx-auto mt-1 max-w-xs text-sm text-muted-foreground text-pretty">
              Agents are configurable assistants with their own model, tools, and
              instructions. Create one to start chatting and automating.
            </p>
            <Button className="mt-5" render={<Link to="/agents/create" />}>
              <Plus />
              Create your first Agent
            </Button>
          </div>
        ) : (
          <>
            <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="relative w-full sm:max-w-xs">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search agents"
                  className="w-full rounded-md border border-border bg-card py-2 pl-9 pr-3 text-sm outline-none focus:border-ring"
                />
              </div>
              <div className="flex flex-wrap items-center gap-1 rounded-md border border-border bg-card p-0.5">
                {filters.map((f) => (
                  <button
                    key={f.key}
                    type="button"
                    onClick={() => setFilter(f.key)}
                    className={cn(
                      "rounded px-2.5 py-1 text-xs font-medium transition-colors",
                      filter === f.key
                        ? "bg-secondary text-secondary-foreground"
                        : "text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {f.label}
                  </button>
                ))}
              </div>
            </div>

            {filtered.length === 0 ? (
              <p className="py-16 text-center text-sm text-muted-foreground">
                No agents match your filters.
              </p>
            ) : (
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {filtered.map((a) => (
                  <AgentCard
                    key={a.name}
                    agent={a}
                    runtime={runtimeMap.get(a.name)}
                    onDelete={() => setDeleteTarget(a.name)}
                    onRun={() => {
                      setInvokeTarget(a);
                      setInvokeResult(null);
                      setInvokeInput("");
                    }}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Agent"
        description={`Delete "${deleteTarget}"? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("Agent deleted");
                setDeleteTarget(null);
              },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />

      {/* Invoke dialog */}
      <Dialog
        open={!!invokeTarget}
        onOpenChange={(o) => {
          if (!o) {
            setInvokeTarget(null);
            setInvokeResult(null);
          }
        }}
      >
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Play className="size-4" /> Run {invokeTarget?.name}
            </DialogTitle>
            <DialogDescription>
              Sends a one-off invocation via the API. Creates an ephemeral session.
            </DialogDescription>
          </DialogHeader>

          {!invokeResult ? (
            <div className="space-y-2">
              <Label htmlFor="invoke-input">Input</Label>
              <Textarea
                id="invoke-input"
                rows={5}
                placeholder="What should the agent do?"
                value={invokeInput}
                onChange={(e) => setInvokeInput(e.target.value)}
              />
            </div>
          ) : (
            <div className="space-y-2">
              <div className="text-xs text-muted-foreground">
                Session: <span className="font-mono">{invokeResult.session_id}</span>
              </div>
              <div className="whitespace-pre-wrap rounded-md border bg-muted p-3 text-sm">
                {invokeResult.response || <span className="italic text-muted-foreground">(empty response)</span>}
              </div>
            </div>
          )}

          <DialogFooter>
            {!invokeResult ? (
              <>
                <Button variant="outline" onClick={() => setInvokeTarget(null)}>
                  Cancel
                </Button>
                <Button
                  disabled={!invokeInput.trim() || invokeMutation.isPending}
                  onClick={() =>
                    invokeTarget &&
                    invokeMutation.mutate(
                      { agent_name: invokeTarget.name, input: invokeInput.trim() },
                      {
                        onSuccess: (res) => setInvokeResult(res),
                        onError: (err) => toast.error(err.message),
                      },
                    )
                  }
                >
                  {invokeMutation.isPending ? "Running…" : "Run"}
                </Button>
              </>
            ) : (
              <Button onClick={() => setInvokeTarget(null)}>Done</Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Page>
  );
}
