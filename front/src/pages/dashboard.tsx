import { useState } from "react";
import { useOverview, useActivityFeed, useCronTimeseries } from "@/api/dashboard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
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
} from "lucide-react";
import type {
  ActivityEvent,
  ComponentHealth,
  CronTimeseriesRange,
} from "@/types/api";

const STATUS_COLORS: Record<string, string> = {
  STATUS_HEALTHY: "bg-green-500/10 text-green-600",
  STATUS_DEGRADED: "bg-amber-500/10 text-amber-600",
  STATUS_DOWN: "bg-red-500/10 text-red-600",
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
    <div className="flex items-center justify-between gap-2 py-2">
      <div className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-muted-foreground" />
        <div>
          <div className="text-sm font-medium">{label}</div>
          <div className="text-xs text-muted-foreground">{health?.detail ?? ""}</div>
        </div>
      </div>
      <div className="flex items-center gap-2">
        {health?.latency_ms !== undefined && health.latency_ms > 0 && (
          <span className="text-xs text-muted-foreground">{health.latency_ms}ms</span>
        )}
        <Badge className={STATUS_COLORS[status] ?? STATUS_COLORS.STATUS_UNSPECIFIED}>{STATUS_LABEL[status]}</Badge>
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
    ? "text-red-500"
    : event.kind === "execution_completed"
      ? "text-green-500"
      : "text-muted-foreground";
  return (
    <div className="flex items-start gap-3 py-2">
      <Icon className={`mt-0.5 h-4 w-4 ${color}`} />
      <div className="flex-1 min-w-0">
        <div className="text-sm">
          <span className="font-medium">{event.actor ?? "unknown"}</span>
          <span className="text-muted-foreground"> — {event.message ?? ""}</span>
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
  const { data: overviewData, isLoading: loadingOverview } = useOverview();
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

  if (loadingOverview) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold">Overview</h2>
        <div className="grid grid-cols-4 gap-4">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-28" />)}</div>
        <Skeleton className="h-72" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Overview</h2>
          <p className="text-sm text-muted-foreground">Platform metrics and system health at a glance.</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard label="Active Agents" value={counts?.active_agents ?? 0} icon={Bot} />
        <StatCard label="MCP Servers" value={counts?.mcp_servers ?? 0} icon={Server} />
        <StatCard label="Connected Daemons" value={counts?.connected_daemons ?? 0} icon={Cpu} />
        <StatCard label="ADK Sessions" value={counts?.active_sessions ?? 0} icon={HardDrive} />
      </div>

      {/* Cron timeseries */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle>Cron Executions</CardTitle>
          <div className="flex gap-1">
            {(["RANGE_1D", "RANGE_7D", "RANGE_30D"] as const).map((r) => (
              <Button
                key={r}
                size="sm"
                variant={range === r ? "default" : "outline"}
                onClick={() => setRange(r)}
              >
                {r.replace("RANGE_", "")}
              </Button>
            ))}
          </div>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={chartData}>
              <XAxis dataKey="label" tick={{ fontSize: 10 }} />
              <YAxis tick={{ fontSize: 10 }} />
              <Tooltip />
              <Bar dataKey="success" stackId="a" fill="#4ade80" name="Success" />
              <Bar dataKey="error" stackId="a" fill="#ef4444" name="Error" />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Health */}
        <Card>
          <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2"><Heart className="h-4 w-4" /> System Health</CardTitle></CardHeader>
          <CardContent>
            <HealthRow label="MongoDB" icon={Database} health={health?.mongodb} />
            <HealthRow label="Redis" icon={HardDrive} health={health?.redis} />
            <HealthRow label="Runner" icon={Cpu} health={health?.runner} />
          </CardContent>
        </Card>

        {/* Latest daemon handshake */}
        <Card>
          <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2"><Terminal className="h-4 w-4" /> Latest Daemon Handshake</CardTitle></CardHeader>
          <CardContent>
            {handshake?.daemon_id ? (
              <pre className="rounded bg-muted p-3 text-xs overflow-x-auto">
                {JSON.stringify(
                  {
                    daemon_id: handshake.daemon_id,
                    name: handshake.name,
                    os: handshake.os,
                    capabilities: handshake.capabilities,
                    connected_at: handshake.connected_at,
                  },
                  null,
                  2,
                )}
              </pre>
            ) : (
              <p className="text-sm text-muted-foreground">No daemons have connected yet.</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Activity feed */}
      <Card>
        <CardHeader className="pb-2"><CardTitle className="flex items-center gap-2"><Activity className="h-4 w-4" /> Activity Feed</CardTitle></CardHeader>
        <CardContent>
          {events.length === 0 ? (
            <p className="text-sm text-muted-foreground py-4">No recent activity.</p>
          ) : (
            <div className="divide-y">
              {events.map((e) => <ActivityRow key={e.id} event={e} />)}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({ label, value, icon: Icon }: { label: string; value: number; icon: typeof Bot }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold">{value.toLocaleString()}</div>
      </CardContent>
    </Card>
  );
}
