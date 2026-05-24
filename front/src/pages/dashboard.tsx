import { useState } from "react";
import { useOverview, useActivityFeed, useCronTimeseries } from "@/api/dashboard";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import {
  Activity,
  Bot,
  Check,
  Database,
  HardDrive,
  Server,
  Cpu,
  AlertTriangle,
  Heart,
  Terminal,
  TrendingDown,
  TrendingUp,
  Minus,
  Copy,
} from "lucide-react";
import type {
  ActivityEvent,
  ComponentHealth,
  CronTimeseriesRange,
} from "@/types/api";

type DashboardEnvironment = "production" | "staging" | "development";

const STATUS_COLORS: Record<string, string> = {
  STATUS_HEALTHY: "bg-emerald-500/10 text-emerald-700 border-emerald-500/20",
  STATUS_DEGRADED: "bg-amber-500/10 text-amber-700 border-amber-500/20",
  STATUS_DOWN: "bg-rose-500/10 text-rose-700 border-rose-500/20",
  STATUS_UNSPECIFIED: "bg-muted text-muted-foreground",
};

const STATUS_LABEL: Record<string, string> = {
  STATUS_HEALTHY: "Healthy",
  STATUS_DEGRADED: "Degraded",
  STATUS_DOWN: "Down",
  STATUS_UNSPECIFIED: "Unknown",
};

function HealthRow({ label, icon: Icon, health }: { label: string; icon: typeof Database; health?: ComponentHealth }) {
  const status = health?.status ?? "STATUS_UNSPECIFIED";
  return (
    <div className="flex items-center justify-between gap-3 py-3">
      <div className="flex items-center gap-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
        <div>
          <div className="text-sm font-medium">{label}</div>
          <div className="font-mono text-[11px] text-muted-foreground">{health?.detail ?? "Primary cluster"}</div>
        </div>
      </div>
      <div className="flex items-center gap-2">
        {health?.latency_ms !== undefined && health.latency_ms > 0 && (
          <span className="text-xs text-muted-foreground">{health.latency_ms}ms</span>
        )}
        <Badge className={STATUS_COLORS[status] ?? STATUS_COLORS.STATUS_UNSPECIFIED}>
          <span className="h-1.5 w-1.5 rounded-full bg-current" />
          {STATUS_LABEL[status]}
        </Badge>
      </div>
    </div>
  );
}

function ActivityRow({ event }: { event: ActivityEvent }) {
  const Icon = event.kind === "error"
    ? AlertTriangle
    : event.kind === "execution_completed"
      ? Check
      : Terminal;
  const color = event.kind === "error"
    ? "text-rose-500"
    : event.kind === "execution_completed"
      ? "text-emerald-500"
      : "text-muted-foreground";
  return (
    <div className="relative flex items-start gap-4 py-2">
      <div className="z-10 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 border-card bg-card">
        <Icon className={`h-3.5 w-3.5 ${color}`} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="text-[13px] leading-5">
          <span className="font-medium">{event.actor ?? "unknown"}</span>
          <span className="text-muted-foreground"> {event.message ?? ""}</span>
        </div>
        <div className="text-xs text-muted-foreground">
          {event.timestamp ? new Date(event.timestamp).toLocaleString() : ""}
        </div>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const [range, setRange] = useState<CronTimeseriesRange>("RANGE_1D");
  const [environment, setEnvironment] = useState<DashboardEnvironment>("production");
  const { data: overviewData, isLoading: loadingOverview } = useOverview(environment);
  const { data: activity } = useActivityFeed(20);
  const { data: timeseries } = useCronTimeseries(range);

  const counts = overviewData?.counts;
  const health = overviewData?.health;
  const handshake = overviewData?.latest_daemon_handshake;
  const events = activity?.events ?? [];

  const chartData = (timeseries?.buckets ?? []).map((b) => ({
    label: b.start ? new Date(b.start).toLocaleString(undefined, range === "RANGE_1D" ? { hour: "2-digit" } : { month: "short", day: "numeric" }) : "",
    success: b.success ?? 0,
    error: b.error ?? 0,
  }));
  const handshakeJson = handshake?.daemon_id
    ? formatJsonWithLineNumbers({
        event: "daemon.connect",
        daemon_id: handshake.daemon_id,
        name: handshake.name,
        os: handshake.os,
        capabilities: handshake.capabilities,
        connected_at: handshake.connected_at,
      })
    : "";

  async function copyHandshake() {
    if (!handshakeJson) return;
    try {
      await navigator.clipboard.writeText(handshakeJson);
      toast.success("Handshake JSON copied");
    } catch {
      toast.error("Copy failed");
    }
  }

  if (loadingOverview) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold tracking-tight">Overview</h2>
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-28" />)}</div>
        <Skeleton className="h-72" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-4xl font-bold tracking-tight text-foreground">Overview</h2>
          <p className="text-sm text-muted-foreground">Platform metrics and system health at a glance.</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">Environment</span>
          <Select value={environment} onValueChange={(value) => setEnvironment(value as DashboardEnvironment)}>
            <SelectTrigger className="w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent align="end">
              <SelectItem value="production">Production</SelectItem>
              <SelectItem value="staging">Staging</SelectItem>
              <SelectItem value="development">Development</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Active Agents" value={counts?.active_agents ?? 0} icon={Bot} trend="up" accent="primary" meta="+12%" />
        <StatCard label="MCP Servers" value={counts?.mcp_servers ?? 0} icon={Server} trend="stable" accent="green" meta="Stable" />
        <StatCard label="Connected Daemons" value={counts?.connected_daemons ?? 0} icon={Cpu} trend="up" accent="orange" meta="+5%" />
        <StatCard label="ADK Sessions" value={counts?.active_sessions ?? 0} icon={HardDrive} trend="down" accent="primary" meta="-2" />
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <Card className="xl:col-span-2">
          <CardHeader className="flex flex-col gap-3 border-b bg-card pb-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle>Cron Executions</CardTitle>
              <p className="text-[13px] text-muted-foreground">Recent scheduled agent execution volume.</p>
            </div>
            <div className="grid grid-cols-3 gap-1 sm:flex">
              {(["RANGE_1D", "RANGE_7D", "RANGE_30D"] as const).map((r) => (
                <Button
                  key={r}
                  size="sm"
                  variant={range === r ? "default" : "ghost"}
                  onClick={() => setRange(r)}
                >
                  {r.replace("RANGE_", "")}
                </Button>
              ))}
            </div>
          </CardHeader>
          <CardContent className="pt-5">
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={chartData}>
                <XAxis dataKey="label" tick={{ fontSize: 10, fill: "var(--muted-foreground)" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 10, fill: "var(--muted-foreground)" }} axisLine={false} tickLine={false} />
                <Tooltip cursor={{ fill: "rgba(246, 195, 67, 0.14)" }} />
                <Bar dataKey="success" stackId="a" fill="var(--chart-2)" name="Success" radius={[3, 3, 0, 0]} />
                <Bar dataKey="error" stackId="a" fill="var(--chart-5)" name="Error" radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader className="border-b pb-4">
              <CardTitle className="flex items-center gap-2"><Heart className="h-4 w-4 text-primary" /> System Health</CardTitle>
            </CardHeader>
            <CardContent className="divide-y">
              <HealthRow label="MongoDB" icon={Database} health={health?.mongodb} />
              <HealthRow label="Redis Cache" icon={HardDrive} health={health?.redis} />
              <HealthRow label="Runner" icon={Cpu} health={health?.runner} />
            </CardContent>
          </Card>

          <Card className="min-h-0">
            <CardHeader className="border-b bg-muted/30 pb-4">
              <CardTitle className="flex items-center gap-2"><Activity className="h-4 w-4 text-muted-foreground" /> Activity Feed</CardTitle>
            </CardHeader>
            <CardContent className="relative max-h-64 overflow-y-auto pt-4 before:absolute before:bottom-5 before:left-[31px] before:top-5 before:w-px before:bg-border">
              {events.length === 0 ? (
                <p className="py-4 text-sm text-muted-foreground">No recent activity.</p>
              ) : (
                <div className="space-y-3">
                  {events.slice(0, 6).map((e) => <ActivityRow key={e.id} event={e} />)}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <Card className="border-[#374151] bg-[#111827] text-gray-200">
        <CardHeader className="flex flex-row items-center justify-between border-b border-gray-800 bg-gray-900 pb-4">
          <CardTitle className="flex items-center gap-2 text-sm text-gray-200">
            <Terminal className="h-4 w-4 text-gray-400" />
            Latest Daemon Handshake
          </CardTitle>
          <Button
            size="sm"
            variant="ghost"
            className="text-gray-400 hover:bg-gray-800 hover:text-white"
            disabled={!handshakeJson}
            onClick={() => void copyHandshake()}
          >
            <Copy className="mr-1 h-3 w-3" />
            Copy
          </Button>
        </CardHeader>
        <CardContent className="pt-5">
          {handshake?.daemon_id ? (
            <pre className="overflow-x-auto text-[13px] leading-6 text-gray-300">
              {handshakeJson}
            </pre>
          ) : (
            <p className="text-sm text-gray-400">No daemons have connected yet.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function formatJsonWithLineNumbers(value: unknown) {
  return JSON.stringify(value, null, 2)
    .split("\n")
    .map((line, index) => `${String(index + 1).padStart(2, " ")}  ${line}`)
    .join("\n");
}

function StatCard({
  label,
  value,
  icon: Icon,
  trend,
  accent,
  meta,
}: {
  label: string;
  value: number;
  icon: typeof Bot;
  trend: "up" | "down" | "stable";
  accent: "primary" | "green" | "orange";
  meta: string;
}) {
  const accentClass = {
    primary: "text-primary bg-primary",
    green: "text-secondary bg-secondary",
    orange: "text-orange-600 bg-orange-500",
  }[accent];
  const TrendIcon = trend === "up" ? TrendingUp : trend === "down" ? TrendingDown : Minus;
  const trendClass = trend === "up"
    ? "bg-emerald-500/10 text-emerald-700"
    : trend === "down"
      ? "bg-rose-500/10 text-rose-700"
      : "bg-muted text-muted-foreground";
  return (
    <Card className="relative h-32 justify-between transition-colors hover:border-primary/40">
      <CardHeader className="flex flex-row items-start justify-between pb-0">
        <CardTitle className="text-xs font-medium uppercase tracking-[0.08em] text-muted-foreground">{label}</CardTitle>
        <Icon className={`h-5 w-5 ${accentClass.split(" ")[0]}`} />
      </CardHeader>
      <CardContent className="flex items-end justify-between">
        <div className="text-4xl font-bold tracking-tight">{value.toLocaleString()}</div>
        <div className={`flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${trendClass}`}>
          <TrendIcon className="h-3.5 w-3.5" />
          {meta}
        </div>
        <div className={`absolute bottom-0 left-0 h-1 w-1/3 ${accentClass.split(" ")[1]}`} />
      </CardContent>
    </Card>
  );
}
