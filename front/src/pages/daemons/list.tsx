import { useState } from "react";
import {
  useBridgeDiagnostics,
  useCancelDaemonTask,
  useCreateDaemonRuntime,
  useCreateDaemonRuntimeToken,
  useDaemons,
  useDaemonRuntimes,
  useDaemonTasks,
} from "@/api/daemons";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import type { DaemonRuntime, DaemonStatus, DaemonTaskInFlight } from "@/types/api";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip } from "recharts";
import { AlertCircle, Cpu, MemoryStick, Activity, X, Terminal, Router, Copy, KeyRound, Plus } from "lucide-react";

const DAEMON_URL = `${window.location.hostname || "localhost"}:9090`;

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "Unknown error";
}

function fmtBytes(n: number | undefined): string {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(1)} ${units[i]}`;
}

function fmtDuration(d?: string): string {
  if (!d) return "-";
  // Protobuf Duration over JSON is "1.234s"
  const m = /^(\d+)(\.\d+)?s$/.exec(d);
  if (!m) return d;
  const secs = parseFloat(d);
  const h = Math.floor(secs / 3600);
  const min = Math.floor((secs % 3600) / 60);
  const s = Math.floor(secs % 60);
  if (h > 0) return `${h}h ${min}m`;
  if (min > 0) return `${min}m ${s}s`;
  return `${s}s`;
}

function DaemonStateBadge({ state }: { state?: DaemonStatus["state"] }) {
  const variant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
    STATE_ONLINE: "default",
    STATE_IDLE: "secondary",
    STATE_OFFLINE: "destructive",
  };
  const label: Record<string, string> = {
    STATE_ONLINE: "Online",
    STATE_IDLE: "Idle",
    STATE_OFFLINE: "Offline",
  };
  const key = state ?? "STATE_UNSPECIFIED";
  const cls: Record<string, string> = {
    STATE_ONLINE: "bg-emerald-500/10 text-emerald-700",
    STATE_IDLE: "bg-muted text-muted-foreground",
    STATE_OFFLINE: "bg-rose-500/10 text-rose-700",
  };
  return (
    <Badge variant={variant[key] ?? "outline"} className={cls[key]}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {label[key] ?? "Unknown"}
    </Badge>
  );
}

export default function DaemonListPage() {
  const [runtimeOpen, setRuntimeOpen] = useState(false);
  const [tokenTarget, setTokenTarget] = useState<DaemonRuntime | null>(null);
  const [newRuntime, setNewRuntime] = useState({ id: "", name: "", description: "" });
  const [tokenName, setTokenName] = useState("");
  const [runtimeSecret, setRuntimeSecret] = useState<{ runtime: DaemonRuntime; secret: string } | null>(null);
  const { data: runtimeData, isLoading: loadingRuntimes, error: runtimeError } = useDaemonRuntimes();
  const { data: daemonData, isLoading: loadingDaemons, error: daemonError } = useDaemons();
  const { data: taskData, error: taskError } = useDaemonTasks();
  const { data: bridgeData, error: bridgeError } = useBridgeDiagnostics();
  const createRuntime = useCreateDaemonRuntime();
  const createToken = useCreateDaemonRuntimeToken();
  const cancelTask = useCancelDaemonTask();

  const runtimes = runtimeData?.runtimes ?? [];
  const daemons = daemonData?.daemons ?? [];
  const tasks = taskData?.tasks ?? [];
  const diag = bridgeData?.diagnostics;
  const connectedRuntimeIds = new Set(daemons.map((d) => d.daemon_runtime_id));

  const onlineCount = daemons.filter((d) => d.state === "STATE_ONLINE" || d.state === "STATE_IDLE").length;

  const daemonCols: Column<DaemonStatus>[] = [
    { header: "Daemon", accessorKey: "name" },
    { header: "Version", cell: (r) => <span className="text-xs text-muted-foreground">{r.version ?? "-"}</span> },
    { header: "OS", cell: (r) => <span className="text-xs">{r.os ?? "-"}</span> },
    { header: "Address", cell: (r) => <span className="text-xs font-mono">{r.remote_addr ?? "-"}</span> },
    { header: "Status", cell: (r) => <DaemonStateBadge state={r.state} /> },
    { header: "Uptime", cell: (r) => <span className="text-xs">{fmtDuration(r.uptime)}</span> },
    {
      header: "Executors",
      cell: (r) => (
        <div className="flex flex-wrap gap-1">
          {(r.executors ?? r.acp_runtimes ?? []).map((e: string) => (
            <Badge key={e} variant="outline" className="text-xs">{e}</Badge>
          ))}
        </div>
      ),
    },
    { header: "In-flight", cell: (r) => <Badge variant="secondary">{r.active_tasks ?? 0}</Badge> },
  ];

  const runtimeCols: Column<DaemonRuntime>[] = [
    {
      header: "Runtime",
      cell: (r) => (
        <div>
          <div className="font-medium">{r.name || r.id}</div>
          <div className="font-mono text-xs text-muted-foreground">{r.id}</div>
        </div>
      ),
    },
    {
      header: "Status",
      cell: (r) => (
        <Badge variant={connectedRuntimeIds.has(r.id) ? "default" : "outline"}>
          {connectedRuntimeIds.has(r.id) ? "Connected" : "Waiting"}
        </Badge>
      ),
    },
    {
      header: "Description",
      cell: (r) => <span className="text-xs text-muted-foreground">{r.description || "-"}</span>,
    },
    {
      header: "Actions",
      cell: (r) => (
        <Button
          size="sm"
          variant="outline"
          onClick={() => {
            setTokenTarget(r);
            setTokenName(`${r.name || r.id} daemon`);
          }}
        >
          <KeyRound className="mr-1 h-3.5 w-3.5" />
          Token
        </Button>
      ),
    },
  ];

  const taskCols: Column<DaemonTaskInFlight>[] = [
    { header: "Task", cell: (r) => <span className="font-mono text-xs">{r.task_id?.slice(0, 12) ?? "-"}</span> },
    { header: "Agent", accessorKey: "agent_name" },
    { header: "Daemon", accessorKey: "daemon_name" },
    { header: "Step", cell: (r) => <span className="text-xs">{r.current_step ?? "-"}</span> },
    {
      header: "Progress",
      cell: (r) => {
        const pct = Math.min(Math.max(r.progress ?? 0, 0), 100);
        return (
          <div className="flex items-center gap-2">
            <div className="h-2 w-24 overflow-hidden rounded bg-muted">
              <div className="h-full bg-primary" style={{ width: `${pct}%` }} />
            </div>
            <span className="text-xs text-muted-foreground">{pct}%</span>
          </div>
        );
      },
    },
    { header: "Elapsed", cell: (r) => <span className="text-xs">{fmtDuration(r.elapsed)}</span> },
    {
      header: "",
      cell: (r) => (
        <Button
          size="icon"
          variant="ghost"
          disabled={!r.task_id}
          onClick={() => {
            if (!r.task_id) return;
            cancelTask.mutate(
              { taskId: r.task_id, daemonRuntimeId: r.daemon_runtime_id },
              {
                onSuccess: () => toast.success("Cancel signal sent"),
                onError: (e) => toast.error(e.message),
              },
            );
          }}
        >
          <X className="h-4 w-4" />
        </Button>
      ),
    },
  ];

  const latencyData = (diag?.latency ?? []).map((p) => ({
    ts: p.timestamp ? new Date(p.timestamp).toLocaleTimeString() : "",
    ms: p.latency_ms ?? 0,
  }));

  function createRuntimeSubmit() {
    const id = newRuntime.id.trim();
    const name = newRuntime.name.trim();
    if (!id || !name) {
      toast.error("Runtime ID and name are required");
      return;
    }
    createRuntime.mutate(
      { id, name, description: newRuntime.description.trim() },
      {
        onSuccess: () => {
          toast.success("Daemon runtime created");
          setRuntimeOpen(false);
          setNewRuntime({ id: "", name: "", description: "" });
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function createTokenSubmit() {
    if (!tokenTarget) return;
    createToken.mutate(
      {
        daemon_runtime_id: tokenTarget.id,
        name: tokenName.trim() || `${tokenTarget.name || tokenTarget.id} daemon`,
      },
      {
        onSuccess: (res) => {
          setRuntimeSecret({ runtime: tokenTarget, secret: res.secret });
          setTokenTarget(null);
          setTokenName("");
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  async function copyText(text: string, label: string) {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(`${label} copied`);
    } catch {
      toast.error("Copy failed");
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Daemon Monitor"
        description="Real-time telemetry and execution state for connected butter-daemons."
      />

      {[runtimeError, daemonError, taskError, bridgeError].filter(Boolean).length > 0 ? (
        <Card className="border-destructive/30 bg-destructive/5">
          <CardContent className="flex items-start gap-3 pt-6 text-sm text-destructive">
            <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
            <div className="space-y-1">
              <p className="font-medium">Some daemon data could not be loaded.</p>
              <ul className="list-disc pl-4 text-xs">
                {runtimeError ? <li>Runtimes: {errorMessage(runtimeError)}</li> : null}
                {daemonError ? <li>Daemons: {errorMessage(daemonError)}</li> : null}
                {taskError ? <li>Tasks: {errorMessage(taskError)}</li> : null}
                {bridgeError ? <li>Diagnostics: {errorMessage(bridgeError)}</li> : null}
              </ul>
            </div>
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader className="flex flex-col gap-3 border-b bg-muted/30 pb-4 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="flex items-center gap-2">
            <Router className="h-4 w-4 text-muted-foreground" />
            Daemon Runtimes
          </CardTitle>
          <Button className="w-full sm:w-auto" onClick={() => setRuntimeOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Runtime
          </Button>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={runtimeCols}
            data={runtimes}
            isLoading={loadingRuntimes}
            emptyMessage="No daemon runtimes configured."
          />
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">Active Daemons</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {onlineCount}
              <span className="text-base text-muted-foreground"> / {daemons.length}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground"><Activity className="h-4 w-4" /> Active Tasks</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{tasks.length}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground"><Cpu className="h-4 w-4" /> Router CPU</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{(diag?.cpu_percent ?? 0).toFixed(1)}%</div>
            <p className="text-xs text-muted-foreground">{diag?.goroutines ?? 0} goroutines</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground"><MemoryStick className="h-4 w-4" /> Router Memory</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{fmtBytes(diag?.memory_used_bytes)}</div>
            {diag?.memory_limit_bytes ? (
              <p className="text-xs text-muted-foreground">of {fmtBytes(diag.memory_limit_bytes)}</p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="border-b bg-muted/30 pb-4">
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-muted-foreground" />
            Connected Daemons
          </CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable columns={daemonCols} data={daemons} isLoading={loadingDaemons} emptyMessage="No daemons connected." />
        </CardContent>
      </Card>

      <Card className="border-[#374151] bg-[#111827] text-gray-200">
        <CardHeader className="border-b border-gray-800 bg-gray-900 pb-4">
          <CardTitle className="flex items-center gap-2 text-gray-200">
            <Activity className="h-4 w-4 text-primary" />
            Active Tasks
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-5">
          <DataTable columns={taskCols} data={tasks} isLoading={false} emptyMessage="No tasks running." />
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="border-b pb-4">
          <CardTitle className="flex items-center gap-2">
            <Router className="h-4 w-4 text-muted-foreground" />
            Bridge Latency
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={latencyData}>
              <XAxis dataKey="ts" tick={{ fontSize: 10 }} />
              <YAxis tick={{ fontSize: 10 }} unit="ms" />
              <Tooltip />
              <Line type="monotone" dataKey="ms" stroke="#3b82f6" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      <Dialog open={runtimeOpen} onOpenChange={setRuntimeOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Daemon Runtime</DialogTitle>
            <DialogDescription>
              Create a workspace runtime before starting butter-daemon with a runtime token.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="runtime-id">Runtime ID</Label>
              <Input
                id="runtime-id"
                placeholder="dev-machine-1"
                value={newRuntime.id}
                onChange={(e) => setNewRuntime((v) => ({ ...v, id: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="runtime-name">Name</Label>
              <Input
                id="runtime-name"
                placeholder="Dev machine"
                value={newRuntime.name}
                onChange={(e) => setNewRuntime((v) => ({ ...v, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="runtime-description">Description</Label>
              <Input
                id="runtime-description"
                placeholder="Local daemon runtime"
                value={newRuntime.description}
                onChange={(e) => setNewRuntime((v) => ({ ...v, description: e.target.value }))}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRuntimeOpen(false)}>Cancel</Button>
            <Button onClick={createRuntimeSubmit} disabled={createRuntime.isPending}>
              {createRuntime.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!tokenTarget} onOpenChange={(open) => !open && setTokenTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Runtime Token</DialogTitle>
            <DialogDescription>
              Use this token when starting butter-daemon for {tokenTarget?.name || tokenTarget?.id}.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="runtime-token-name">Token name</Label>
            <Input
              id="runtime-token-name"
              value={tokenName}
              onChange={(e) => setTokenName(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTokenTarget(null)}>Cancel</Button>
            <Button onClick={createTokenSubmit} disabled={createToken.isPending}>
              {createToken.isPending ? "Creating..." : "Create Token"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!runtimeSecret} onOpenChange={(open) => !open && setRuntimeSecret(null)}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>Runtime Token Created</DialogTitle>
            <DialogDescription>
              Copy the secret now. It is only shown once.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div className="rounded-md border bg-muted p-3">
              <div className="mb-1 text-xs font-medium text-muted-foreground">Secret</div>
              <div className="flex items-center gap-2">
                <code className="min-w-0 flex-1 break-all text-xs">{runtimeSecret?.secret}</code>
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={() => runtimeSecret?.secret && void copyText(runtimeSecret.secret, "Secret")}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="rounded-md border bg-muted p-3">
              <div className="mb-1 text-xs font-medium text-muted-foreground">Start command</div>
              <div className="flex items-center gap-2">
                <code className="min-w-0 flex-1 break-all text-xs">
                  {`butter-daemon --url ${DAEMON_URL} --token ${runtimeSecret?.secret ?? ""}`}
                </code>
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={() =>
                    runtimeSecret?.secret &&
                    void copyText(`butter-daemon --url ${DAEMON_URL} --token ${runtimeSecret.secret}`, "Command")
                  }
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setRuntimeSecret(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
