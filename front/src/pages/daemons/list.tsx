import { useDaemons, useDaemonTasks, useBridgeDiagnostics, useCancelDaemonTask } from "@/api/daemons";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import type { DaemonStatus, DaemonTaskInFlight } from "@/types/api";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip } from "recharts";
import { Cpu, MemoryStick, Activity, X } from "lucide-react";

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
  return <Badge variant={variant[key] ?? "outline"}>{label[key] ?? "Unknown"}</Badge>;
}

export default function DaemonListPage() {
  const { data: daemonData, isLoading: loadingDaemons } = useDaemons();
  const { data: taskData } = useDaemonTasks();
  const { data: bridgeData } = useBridgeDiagnostics();
  const cancelTask = useCancelDaemonTask();

  const daemons = daemonData?.daemons ?? [];
  const tasks = taskData?.tasks ?? [];
  const diag = bridgeData?.diagnostics;

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
          {(r.executors ?? r.capabilities ?? []).map((e) => (
            <Badge key={e} variant="outline" className="text-xs">{e}</Badge>
          ))}
        </div>
      ),
    },
    { header: "In-flight", cell: (r) => <Badge variant="secondary">{r.active_tasks ?? 0}</Badge> },
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
          onClick={() => {
            cancelTask.mutate(
              { taskId: r.task_id, daemonId: r.daemon_id },
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

  return (
    <div className="space-y-6">
      <PageHeader title="Daemon Monitor" />

      {/* Top stats */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Active Daemons</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {onlineCount}
              <span className="text-base text-muted-foreground"> / {daemons.length}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground flex items-center gap-2"><Activity className="h-4 w-4" /> Active Tasks</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{tasks.length}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground flex items-center gap-2"><Cpu className="h-4 w-4" /> Router CPU</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{(diag?.cpu_percent ?? 0).toFixed(1)}%</div>
            <p className="text-xs text-muted-foreground">{diag?.goroutines ?? 0} goroutines</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground flex items-center gap-2"><MemoryStick className="h-4 w-4" /> Router Memory</CardTitle></CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{fmtBytes(diag?.memory_used_bytes)}</div>
            {diag?.memory_limit_bytes ? (
              <p className="text-xs text-muted-foreground">of {fmtBytes(diag.memory_limit_bytes)}</p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      {/* Daemons */}
      <Card>
        <CardHeader><CardTitle>Connected Daemons</CardTitle></CardHeader>
        <CardContent>
          <DataTable columns={daemonCols} data={daemons} isLoading={loadingDaemons} emptyMessage="No daemons connected." />
        </CardContent>
      </Card>

      {/* Active tasks */}
      <Card>
        <CardHeader><CardTitle>Active Tasks</CardTitle></CardHeader>
        <CardContent>
          <DataTable columns={taskCols} data={tasks} isLoading={false} emptyMessage="No tasks running." />
        </CardContent>
      </Card>

      {/* Bridge latency */}
      <Card>
        <CardHeader><CardTitle>Bridge Latency (recent samples)</CardTitle></CardHeader>
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
    </div>
  );
}
