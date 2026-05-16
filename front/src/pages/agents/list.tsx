import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useAgents,
  useDeleteAgent,
  useReloadAgents,
  useInvokeAgent,
  useAgentRuntimeStatuses,
} from "@/api/agents";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Card } from "@/components/ui/card";
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  Play,
  RefreshCw,
  Plus,
  Bot,
  ListChecks,
} from "lucide-react";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import type { Agent, AgentRuntimeStatus } from "@/types/api";

const TYPE_ICON: Record<string, string> = {
  AGENT_TYPE_SEQUENTIAL: "→",
  AGENT_TYPE_PARALLEL: "⇉",
  AGENT_TYPE_LOOP: "↻",
  AGENT_TYPE_LLM: "✦",
};

const STATE_BADGE: Record<string, { variant: "default" | "secondary" | "outline" | "destructive"; cls?: string; label: string }> = {
  AGENT_RUNTIME_STATE_RUNNING: { variant: "default", cls: "bg-blue-500/10 text-blue-600", label: "Running" },
  AGENT_RUNTIME_STATE_IDLE: { variant: "secondary", label: "Idle" },
  AGENT_RUNTIME_STATE_FAILED: { variant: "destructive", label: "Failed" },
  AGENT_RUNTIME_STATE_UNSPECIFIED: { variant: "outline", label: "Unknown" },
};

function timeAgo(ts?: string): string {
  if (!ts) return "—";
  const d = Date.now() - new Date(ts).getTime();
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`;
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`;
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`;
  return `${Math.floor(d / 86_400_000)}d ago`;
}

export default function AgentListPage() {
  const { data, isLoading } = useAgents();
  const agents = data?.agents ?? [];
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

  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [invokeTarget, setInvokeTarget] = useState<Agent | null>(null);
  const [invokeInput, setInvokeInput] = useState("");
  const [invokeResult, setInvokeResult] = useState<{ session_id: string; response: string } | null>(null);

  const columns: Column<Agent>[] = [
    {
      header: "Name",
      cell: (row) => (
        <div className="flex items-center gap-2">
          <span className="text-base">{TYPE_ICON[row.type ?? ""] ?? "•"}</span>
          <div>
            <div className="font-medium">{row.name}</div>
            {row.description && (
              <div className="text-xs text-muted-foreground line-clamp-1 max-w-md">
                {row.description}
              </div>
            )}
          </div>
        </div>
      ),
    },
    {
      header: "Type",
      cell: (row) => (
        <Badge variant="secondary" className="font-mono text-[10px]">
          {AGENT_TYPE_LABELS[row.type ?? "AGENT_TYPE_UNSPECIFIED"]}
        </Badge>
      ),
    },
    {
      header: "Status",
      cell: (row) => {
        const rt = runtimeMap.get(row.name);
        const state = rt?.state ?? "AGENT_RUNTIME_STATE_UNSPECIFIED";
        const badge = STATE_BADGE[state];
        return (
          <div className="flex items-center gap-2">
            <Badge variant={badge.variant} className={badge.cls}>
              {badge.label}
            </Badge>
            {(rt?.in_flight ?? 0) > 0 && (
              <span className="text-xs text-muted-foreground">×{rt!.in_flight}</span>
            )}
          </div>
        );
      },
    },
    {
      header: "Last Run",
      cell: (row) => (
        <span className="text-xs text-muted-foreground">
          {timeAgo(runtimeMap.get(row.name)?.last_run_at)}
        </span>
      ),
    },
    {
      header: "A2A",
      cell: (row) =>
        row.enable_a2a ? <Badge variant="outline" className="text-xs">A2A</Badge> : null,
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setInvokeTarget(row);
              setInvokeResult(null);
              setInvokeInput("");
            }}
          >
            <Play className="mr-1 h-3 w-3" /> Run
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => navigate(`/agents/${row.name}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
                <Trash2 className="mr-2 h-4 w-4" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      ),
    },
  ];

  const summary = useMemo(() => {
    let running = 0,
      idle = 0,
      failed = 0;
    for (const a of agents) {
      const rt = runtimeMap.get(a.name);
      if (rt?.state === "AGENT_RUNTIME_STATE_RUNNING") running++;
      else if (rt?.state === "AGENT_RUNTIME_STATE_FAILED") failed++;
      else idle++;
    }
    return { running, idle, failed };
  }, [agents, runtimeMap]);

  return (
    <>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Agents</h2>
          <p className="text-sm text-muted-foreground">Manage and monitor orchestration nodes.</p>
        </div>
        <div className="flex items-center gap-2">
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
            <RefreshCw className={`mr-2 h-3 w-3 ${reloadMutation.isPending ? "animate-spin" : ""}`} />
            Hot-reload
          </Button>
          <Button onClick={() => navigate("/agents/create")}>
            <Plus className="mr-2 h-4 w-4" />
            Deploy Agent
          </Button>
        </div>
      </div>

      {/* Summary strip */}
      <div className="mb-4 grid grid-cols-4 gap-3">
        <Card className="p-3">
          <div className="flex items-center gap-2">
            <Bot className="h-4 w-4 text-muted-foreground" />
            <div className="text-xs text-muted-foreground">Total</div>
          </div>
          <div className="mt-1 text-2xl font-bold">{agents.length}</div>
        </Card>
        <Card className="p-3">
          <div className="text-xs text-muted-foreground">Running</div>
          <div className="mt-1 text-2xl font-bold text-blue-600">{summary.running}</div>
        </Card>
        <Card className="p-3">
          <div className="text-xs text-muted-foreground">Idle</div>
          <div className="mt-1 text-2xl font-bold">{summary.idle}</div>
        </Card>
        <Card className="p-3">
          <div className="text-xs text-muted-foreground">Failed</div>
          <div className="mt-1 text-2xl font-bold text-red-600">{summary.failed}</div>
        </Card>
      </div>

      <DataTable
        columns={columns}
        data={agents}
        isLoading={isLoading}
        emptyMessage="No agents yet. Deploy your first agent to get started."
      />

      <div className="mt-2 text-xs text-muted-foreground">
        Showing {agents.length} of {data?.total ?? agents.length} agents
      </div>

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
              <Play className="h-4 w-4" /> Run {invokeTarget?.name}
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
              <div className="rounded-md border bg-muted p-3 text-sm whitespace-pre-wrap">
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

      {/* Discoverability hint */}
      <div className="mt-6 flex items-start gap-2 rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        <ListChecks className="mt-0.5 h-3.5 w-3.5" />
        <span>
          Tip: use the <strong>Run</strong> button to test an agent with a one-off input. Recent
          invocations are visible on the Overview page.
        </span>
      </div>
    </>
  );
}
