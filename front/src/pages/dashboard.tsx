import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useOverview, useActivityFeed, useCronTimeseries } from "@/api/dashboard";
import { useAgents } from "@/api/agents";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { StatusDot } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import {
  Activity as ActivityIcon,
  AlertTriangle,
  ArrowRight,
  Bot,
  CheckCircle2,
  Check,
  Cpu,
  Database,
  HardDrive,
  MessageSquare,
  Plus,
  Server,
  Terminal,
  Workflow,
} from "lucide-react";
import type {
  ActivityEvent,
  ComponentHealth,
  ComponentHealthStatus,
  CronTimeseriesRange,
} from "@/types/api";

const HEALTH_TO_STATUS: Record<ComponentHealthStatus, "success" | "waiting" | "failed" | "never"> = {
  STATUS_HEALTHY: "success",
  STATUS_DEGRADED: "waiting",
  STATUS_DOWN: "failed",
  STATUS_UNSPECIFIED: "never",
};

function MetricCard({
  label,
  value,
  icon,
}: {
  label: string;
  value: number;
  icon: React.ReactNode;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">{label}</span>
        <span className="text-muted-foreground [&_svg]:size-4">{icon}</span>
      </div>
      <p className="mt-2 text-2xl font-semibold tabular-nums tracking-tight">
        {value.toLocaleString()}
      </p>
    </div>
  );
}

function HealthRow({
  label,
  icon,
  health,
  first,
}: {
  label: string;
  icon: React.ReactNode;
  health?: ComponentHealth;
  first?: boolean;
}) {
  const status = HEALTH_TO_STATUS[health?.status ?? "STATUS_UNSPECIFIED"];
  return (
    <div className={cn("flex items-center gap-3 px-3 py-2.5", !first && "border-t border-border")}>
      <span className="text-muted-foreground [&_svg]:size-4">{icon}</span>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium">{label}</p>
          <StatusDot status={status} />
        </div>
        {health?.detail && (
          <p className="truncate text-xs text-muted-foreground">{health.detail}</p>
        )}
      </div>
      {health?.latency_ms !== undefined && health.latency_ms > 0 && (
        <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
          {health.latency_ms}ms
        </span>
      )}
    </div>
  );
}

function ActivityRow({ event, last }: { event: ActivityEvent; last: boolean }) {
  const tone =
    event.kind === "error"
      ? "bg-danger-muted text-danger-foreground"
      : event.kind === "execution_completed"
        ? "bg-success-muted text-success-foreground"
        : "bg-muted text-muted-foreground";
  const icon =
    event.kind === "error" ? (
      <AlertTriangle className="size-3.5" />
    ) : event.kind === "execution_completed" ? (
      <Check className="size-3.5" />
    ) : (
      <Terminal className="size-3.5" />
    );
  return (
    <li className="flex gap-3 px-2 py-2.5">
      <div className="flex flex-col items-center">
        <span className={cn("flex size-6 shrink-0 items-center justify-center rounded-full", tone)}>
          {icon}
        </span>
        {!last && <span className="mt-1 w-px flex-1 bg-border" />}
      </div>
      <div className="min-w-0 flex-1 pb-1">
        <p className="text-sm leading-snug">
          <span className="font-medium">{event.actor ?? "unknown"}</span>
          <span className="text-muted-foreground"> {event.message ?? ""}</span>
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {event.timestamp ? new Date(event.timestamp).toLocaleString() : ""}
        </p>
      </div>
    </li>
  );
}

export default function DashboardPage() {
  const [range, setRange] = useState<CronTimeseriesRange>("RANGE_7D");
  const { data: overviewData, isLoading: loadingOverview } = useOverview("production");
  const { data: activity } = useActivityFeed(20);
  const { data: timeseries } = useCronTimeseries(range);
  const { data: agentsData } = useAgents({ page_size: 10 });

  const counts = overviewData?.counts;
  const health = overviewData?.health;
  const events = activity?.events ?? [];
  const quickStartAgents = (agentsData?.agents ?? []).slice(0, 3);

  const unhealthy = useMemo(() => {
    const entries: { label: string; health?: ComponentHealth }[] = [
      { label: "MongoDB", health: health?.mongodb },
      { label: "Redis", health: health?.redis },
      { label: "Runner", health: health?.runner },
    ];
    return entries.filter(
      (e) => e.health?.status === "STATUS_DEGRADED" || e.health?.status === "STATUS_DOWN",
    );
  }, [health]);

  const chartData = (timeseries?.buckets ?? []).map((b) => ({
    label: b.start
      ? new Date(b.start).toLocaleString(
          undefined,
          range === "RANGE_1D" ? { hour: "2-digit" } : { month: "short", day: "numeric" },
        )
      : "",
    success: b.success ?? 0,
    error: b.error ?? 0,
  }));

  const ranges: { key: CronTimeseriesRange; label: string }[] = [
    { key: "RANGE_1D", label: "1 day" },
    { key: "RANGE_7D", label: "7 days" },
    { key: "RANGE_30D", label: "30 days" },
  ];

  return (
    <Page>
      <PageHeader
        title="Overview"
        subtitle="Everything happening across this workspace"
        actions={
          <>
            <Button variant="outline" size="sm" render={<Link to="/automations" />}>
              <Workflow className="size-4" />
              Automations
            </Button>
            <Button size="sm" render={<Link to="/chat?new=1" />}>
              <Plus className="size-4" />
              New chat
            </Button>
          </>
        }
      />
      <PageScroll className="max-w-7xl">
        {loadingOverview ? (
          <div className="space-y-5">
            <Skeleton className="h-12" />
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-24" />
              ))}
            </div>
            <Skeleton className="h-72" />
          </div>
        ) : (
          <>
            {/* System status banner */}
            <div
              className={cn(
                "mb-5 flex items-center justify-between gap-3 rounded-lg border px-4 py-3",
                unhealthy.length > 0
                  ? "border-warning/40 bg-warning-muted/40"
                  : "border-success/40 bg-success-muted/40",
              )}
            >
              <div className="flex items-center gap-2.5">
                {unhealthy.length > 0 ? (
                  <AlertTriangle className="size-4 text-warning-foreground" />
                ) : (
                  <CheckCircle2 className="size-4 text-success-foreground" />
                )}
                <div className="text-sm">
                  <span className="font-medium">
                    {unhealthy.length > 0
                      ? `${unhealthy.length} ${unhealthy.length === 1 ? "component needs" : "components need"} attention`
                      : "All systems operational"}
                  </span>
                  <span className="ml-2 text-muted-foreground">
                    {unhealthy.length > 0
                      ? unhealthy.map((u) => u.label).join(", ")
                      : "Database, cache, and runner are healthy."}
                  </span>
                </div>
              </div>
            </div>

            {/* Metrics */}
            <div className="mb-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
              <MetricCard label="Active agents" value={counts?.active_agents ?? 0} icon={<Bot />} />
              <MetricCard label="MCP servers" value={counts?.mcp_servers ?? 0} icon={<Server />} />
              <MetricCard label="Connected daemons" value={counts?.connected_daemons ?? 0} icon={<Cpu />} />
              <MetricCard label="Active sessions" value={counts?.active_sessions ?? 0} icon={<MessageSquare />} />
            </div>

            <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
              {/* Left column: executions chart + health */}
              <div className="flex flex-col gap-5 lg:col-span-2">
                <section>
                  <div className="mb-2 flex items-center justify-between">
                    <h2 className="text-sm font-medium">Cron executions</h2>
                    <div className="inline-flex rounded-md border border-border p-0.5">
                      {ranges.map((r) => (
                        <button
                          key={r.key}
                          onClick={() => setRange(r.key)}
                          className={cn(
                            "rounded px-2.5 py-1 text-xs font-medium transition-colors",
                            range === r.key
                              ? "bg-secondary text-secondary-foreground"
                              : "text-muted-foreground hover:text-foreground",
                          )}
                        >
                          {r.label}
                        </button>
                      ))}
                    </div>
                  </div>
                  <div className="rounded-lg border border-border bg-card p-4">
                    <ResponsiveContainer width="100%" height={260}>
                      <BarChart data={chartData}>
                        <XAxis
                          dataKey="label"
                          tick={{ fontSize: 10, fill: "var(--muted-foreground)" }}
                          axisLine={false}
                          tickLine={false}
                        />
                        <YAxis
                          tick={{ fontSize: 10, fill: "var(--muted-foreground)" }}
                          axisLine={false}
                          tickLine={false}
                        />
                        <Tooltip
                          cursor={{ fill: "color-mix(in srgb, var(--foreground) 6%, transparent)" }}
                          contentStyle={{
                            backgroundColor: "var(--popover)",
                            border: "1px solid var(--border)",
                            borderRadius: "0.5rem",
                            color: "var(--popover-foreground)",
                            fontSize: "0.75rem",
                          }}
                        />
                        <Bar dataKey="success" stackId="a" fill="var(--success)" name="Success" radius={[3, 3, 0, 0]} />
                        <Bar dataKey="error" stackId="a" fill="var(--danger)" name="Error" radius={[3, 3, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </section>

                <section>
                  <h2 className="mb-2 text-sm font-medium">System health</h2>
                  <div className="overflow-hidden rounded-lg border border-border bg-card">
                    <HealthRow first label="MongoDB" icon={<Database />} health={health?.mongodb} />
                    <HealthRow label="Redis cache" icon={<HardDrive />} health={health?.redis} />
                    <HealthRow label="Runner" icon={<Cpu />} health={health?.runner} />
                  </div>
                </section>
              </div>

              {/* Right column: activity feed + quick start */}
              <div className="lg:col-span-1">
                <section>
                  <div className="mb-2 flex items-center gap-2">
                    <ActivityIcon className="size-4 text-muted-foreground" />
                    <h2 className="text-sm font-medium">Recent activity</h2>
                  </div>
                  <div className="rounded-lg border border-border bg-card p-1">
                    {events.length === 0 ? (
                      <p className="px-3 py-6 text-center text-sm text-muted-foreground">
                        No recent activity.
                      </p>
                    ) : (
                      <ol className="relative">
                        {events.slice(0, 8).map((e, i, arr) => (
                          <ActivityRow key={e.id} event={e} last={i === arr.length - 1} />
                        ))}
                      </ol>
                    )}
                  </div>
                </section>

                {quickStartAgents.length > 0 && (
                  <section className="mt-5">
                    <h2 className="mb-2 text-sm font-medium">Quick start</h2>
                    <div className="flex flex-col gap-2">
                      {quickStartAgents.map((a) => (
                        <Link
                          key={a.name}
                          to={`/chat?new=1&agent=${encodeURIComponent(a.name)}`}
                          className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                        >
                          <Bot className="size-4 text-muted-foreground" />
                          <span className="flex-1 truncate">{a.name}</span>
                          <ArrowRight className="size-3.5 text-muted-foreground" />
                        </Link>
                      ))}
                    </div>
                  </section>
                )}
              </div>
            </div>
          </>
        )}
      </PageScroll>
    </Page>
  );
}
