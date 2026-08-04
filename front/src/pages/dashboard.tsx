import { useMemo, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { useOverview, useActivityFeed, useActivityMetrics } from "@/api/dashboard";
import { useAgents } from "@/api/agents";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { AgentAvatar } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import {
  AlertTriangle,
  ArrowRight,
  Check,
  CheckCircle2,
  CircleDashed,
  Clock3,
  Database,
  HardDrive,
  Link2,
  Loader2,
  Plus,
  RefreshCw,
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

function relTime(iso?: string): string {
  if (!iso) return "";
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return "";
  const s = Math.max(0, Math.round((Date.now() - t) / 1000));
  if (s < 60) return `${s}s ago`;
  const m = Math.round(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.round(h / 24)}d ago`;
}

type Tone = "danger" | "warning" | "running" | "success" | "muted";

const toneIcon: Record<Tone, string> = {
  danger: "bg-danger-muted text-danger-foreground",
  warning: "bg-warning-muted text-warning-foreground",
  running: "bg-running-muted text-running-foreground",
  success: "bg-success-muted text-success-foreground",
  muted: "bg-muted text-muted-foreground",
};

function SectionHeader({
  title,
  description,
  aside,
}: {
  title: string;
  description?: string;
  aside?: ReactNode;
}) {
  return (
    <div className="mb-3 flex min-h-8 flex-col items-start justify-between gap-2 sm:flex-row sm:items-end sm:gap-4">
      <div>
        <h2 className="text-sm font-semibold">{title}</h2>
        {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
      </div>
      {aside && <div className="shrink-0 text-xs text-muted-foreground sm:text-end">{aside}</div>}
    </div>
  );
}

function healthLabel(status?: ComponentHealthStatus) {
  switch (status) {
    case "STATUS_HEALTHY":
      return "Healthy";
    case "STATUS_DEGRADED":
      return "Degraded";
    case "STATUS_DOWN":
      return "Down";
    default:
      return "Unknown";
  }
}

function healthTone(status?: ComponentHealthStatus): Tone {
  if (status === "STATUS_HEALTHY") return "success";
  if (status === "STATUS_DEGRADED") return "warning";
  if (status === "STATUS_DOWN") return "danger";
  return "muted";
}

function HealthItem({
  label,
  health,
  icon,
}: {
  label: string;
  health?: ComponentHealth;
  icon: ReactNode;
}) {
  const tone = healthTone(health?.status);
  return (
    <div className="flex min-w-0 items-center gap-2.5 px-3 py-2.5 sm:px-4">
      <span className={cn("flex size-7 shrink-0 items-center justify-center rounded-md [&_svg]:size-3.5", toneIcon[tone])}>
        {icon}
      </span>
      <div className="min-w-0">
        <p className="truncate text-xs text-muted-foreground">{label}</p>
        <div className="flex items-baseline gap-1.5">
          <p className="text-sm font-medium">{healthLabel(health?.status)}</p>
          {typeof health?.latency_ms === "number" && health.latency_ms > 0 && (
            <span className="font-mono text-[0.68rem] text-muted-foreground">
              {health.latency_ms}ms
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

function SystemHealth({ health }: { health?: { mongodb?: ComponentHealth; redis?: ComponentHealth; runner?: ComponentHealth } }) {
  const items = [health?.mongodb, health?.redis, health?.runner];
  const problemCount = items.filter(
    (item) => item?.status === "STATUS_DEGRADED" || item?.status === "STATUS_DOWN",
  ).length;

  return (
    <section className="overflow-hidden rounded-lg border border-transparent bg-card shadow-card">
      <div className="grid lg:grid-cols-[minmax(220px,1.35fr)_repeat(3,minmax(128px,1fr))]">
        <div
          className={cn(
            "flex items-center gap-3 border-b border-border px-4 py-3 lg:border-b-0 lg:border-e",
            problemCount > 0 ? "bg-warning-muted/35" : "bg-success-muted/30",
          )}
        >
          <span
            className={cn(
              "flex size-8 shrink-0 items-center justify-center rounded-md",
              problemCount > 0
                ? "bg-warning-muted text-warning-foreground"
                : "bg-success-muted text-success-foreground",
            )}
          >
            {problemCount > 0 ? <AlertTriangle className="size-4" /> : <CheckCircle2 className="size-4" />}
          </span>
          <div className="min-w-0">
            <p className="text-sm font-semibold">
              {problemCount > 0 ? `${problemCount} system issue${problemCount === 1 ? "" : "s"}` : "Systems operational"}
            </p>
            <p className="truncate text-xs text-muted-foreground">
              {problemCount > 0 ? "Review infrastructure health" : "All core services are responding"}
            </p>
          </div>
        </div>
        <div className="grid grid-cols-1 divide-y divide-border min-[480px]:grid-cols-3 min-[480px]:divide-x min-[480px]:divide-y-0 lg:contents">
          <HealthItem label="Database" health={health?.mongodb} icon={<Database />} />
          <HealthItem label="Cache" health={health?.redis} icon={<HardDrive />} />
          <HealthItem label="Runner" health={health?.runner} icon={<Server />} />
        </div>
      </div>
    </section>
  );
}

function Metric({
  label,
  value,
  detail,
  emphasis,
}: {
  label: string;
  value: string;
  detail: string;
  emphasis?: boolean;
}) {
  return (
    <div className={cn("min-w-0 px-4 py-4 sm:px-5", emphasis && "bg-accent/65")}>
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <p
        className={cn(
          "mt-1.5 font-mono text-2xl font-semibold tabular-nums tracking-normal sm:text-3xl",
          emphasis && "text-accent-foreground",
        )}
      >
        {value}
      </p>
      <p className="mt-1 truncate text-xs text-muted-foreground">{detail}</p>
    </div>
  );
}

type Attention = {
  id: string;
  tone: Tone;
  icon: ReactNode;
  title: string;
  meta: string;
  detail?: string;
  time?: string;
  action?: { label: string; to: string };
};

function AttentionRow({ item }: { item: Attention }) {
  return (
    <div className="flex items-start gap-3 border-b border-border px-4 py-3.5 last:border-b-0">
      <span className={cn("mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md [&_svg]:size-3.5", toneIcon[item.tone])}>
        {item.icon}
      </span>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="truncate text-sm font-medium">{item.title}</p>
          {item.time && <span className="shrink-0 text-xs text-muted-foreground">{item.time}</span>}
        </div>
        <p className="mt-0.5 text-xs text-muted-foreground">{item.meta}</p>
        {item.detail && <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{item.detail}</p>}
      </div>
      {item.action && (
        <Button
          variant="ghost"
          size="sm"
          className="shrink-0"
          nativeButton={false}
          render={<Link to={item.action.to} />}
        >
          {item.action.label}
          <ArrowRight className="size-3.5" />
        </Button>
      )}
    </div>
  );
}

function ActiveRow({ event }: { event: ActivityEvent }) {
  const title = event.message?.trim() || event.actor || "Agent run";
  return (
    <div className="flex items-center gap-3 border-b border-border px-4 py-3.5 last:border-b-0">
      <span className="flex size-7 shrink-0 items-center justify-center rounded-md bg-running-muted text-running-foreground">
        <Loader2 className="size-3.5 animate-spin" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{title}</p>
        <p className="mt-0.5 truncate text-xs text-muted-foreground">
          {event.actor ?? "Unknown agent"}
          {event.timestamp && ` · started ${relTime(event.timestamp)}`}
        </p>
      </div>
      <span className="hidden items-center gap-1.5 text-xs font-medium text-running-foreground sm:flex">
        <span className="size-1.5 rounded-full bg-running animate-blink" />
        Running
      </span>
    </div>
  );
}

function activityTone(kind: string): { tone: Tone; icon: ReactNode } {
  switch (kind) {
    case "error":
      return { tone: "danger", icon: <AlertTriangle /> };
    case "execution_completed":
      return { tone: "success", icon: <Check /> };
    case "invocation":
      return { tone: "running", icon: <Loader2 className="animate-spin" /> };
    case "channel":
    case "integration":
      return { tone: "muted", icon: <Link2 /> };
    default:
      return { tone: "muted", icon: <Terminal /> };
  }
}

function ActivityRow({ event }: { event: ActivityEvent }) {
  const { tone, icon } = activityTone(event.kind ?? "");
  const title = event.message?.trim() || event.actor || "Activity";
  return (
    <li className="group flex gap-3 px-1 py-2.5">
      <span className={cn("mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md [&_svg]:size-3.5", toneIcon[tone])}>
        {icon}
      </span>
      <div className="min-w-0 flex-1 border-b border-border pb-2.5 group-last:border-b-0 group-last:pb-0">
        <p className="line-clamp-2 text-sm leading-snug">{title}</p>
        <p className="mt-1 flex items-center gap-1.5 truncate text-xs text-muted-foreground">
          <span className="truncate">{event.actor ?? "Unknown"}</span>
          {event.timestamp && (
            <>
              <span aria-hidden>·</span>
              <span className="shrink-0">{relTime(event.timestamp)}</span>
            </>
          )}
        </p>
      </div>
    </li>
  );
}

function EmptyRow({ icon, children }: { icon: ReactNode; children: ReactNode }) {
  return (
    <div className="flex min-h-24 items-center justify-center gap-2.5 px-4 py-5 text-sm text-muted-foreground">
      <span className="[&_svg]:size-4">{icon}</span>
      <span>{children}</span>
    </div>
  );
}

export default function DashboardPage() {
  const [range, setRange] = useState<CronTimeseriesRange>("RANGE_7D");
  const { data: overviewData, isLoading: loadingOverview } = useOverview("production");
  const { data: metrics } = useActivityMetrics(range);
  const activityQuery = useActivityFeed(50);
  const { data: agentsData } = useAgents({ page_size: 10 });

  const health = overviewData?.health;
  const agentRuns = metrics?.agent_runs ?? 0;
  const automationRuns = metrics?.automation_runs ?? 0;
  const failedRuns = (metrics?.agent_runs_failed ?? 0) + (metrics?.automation_runs_failed ?? 0);
  const totalRuns = agentRuns + automationRuns;
  const successRate = totalRuns > 0 ? Math.round(((totalRuns - failedRuns) / totalRuns) * 100) : 100;

  const ranges: { key: CronTimeseriesRange; label: string }[] = [
    { key: "RANGE_7D", label: "7 days" },
    { key: "RANGE_30D", label: "30 days" },
  ];
  const events = useMemo(() => activityQuery.data?.events ?? [], [activityQuery.data]);
  const quickStartAgents = (agentsData?.agents ?? []).slice(0, 4);

  const infraAttention = useMemo<Attention[]>(() => {
    const components: { label: string; health?: ComponentHealth }[] = [
      { label: "MongoDB", health: health?.mongodb },
      { label: "Redis cache", health: health?.redis },
      { label: "Runner", health: health?.runner },
    ];
    return components
      .filter((c) => c.health?.status === "STATUS_DEGRADED" || c.health?.status === "STATUS_DOWN")
      .map((c) => {
        const down = c.health?.status === "STATUS_DOWN";
        return {
          id: `infra-${c.label}`,
          tone: down ? "danger" : "warning",
          icon: <AlertTriangle />,
          title: `${c.label} ${down ? "unavailable" : "degraded"}`,
          meta: "Infrastructure",
          detail: c.health?.detail,
        };
      });
  }, [health]);

  const errorAttention = useMemo<Attention[]>(
    () =>
      events
        .filter((e) => e.kind === "error")
        .slice(0, 5)
        .map((e) => ({
          id: e.id,
          tone: "danger" as const,
          icon: <AlertTriangle />,
          title: e.actor ? `${e.actor} run failed` : "Agent run failed",
          meta: `Agent run${e.actor ? ` · ${e.actor}` : ""}`,
          detail: e.message,
          time: relTime(e.timestamp),
          action: { label: "Sessions", to: "/sessions" },
        })),
    [events],
  );

  const needsAttention = [...infraAttention, ...errorAttention];
  const attentionShown = needsAttention.slice(0, 6);
  const activeNow = useMemo(() => events.filter((e) => e.kind === "invocation").slice(0, 6), [events]);
  const recent = events.slice(0, 5);

  return (
    <Page>
      <PageHeader
        className="max-w-[1320px]"
        title="Overview"
        subtitle="Workspace health and recent agent activity"
        actions={
          <>
            <Button
              variant="outline"
              size="sm"
              nativeButton={false}
              render={<Link to="/automations" />}
            >
              <Workflow className="size-4" />
              Automations
            </Button>
            <Button size="sm" nativeButton={false} render={<Link to="/chat?new=1" />}>
              <Plus className="size-4" />
              New chat
            </Button>
          </>
        }
      />
      <PageScroll className="max-w-[1320px] py-4 md:py-5">
        {loadingOverview ? (
          <div className="space-y-5">
            <Skeleton className="h-20" />
            <Skeleton className="h-32" />
            <div className="grid gap-5 lg:grid-cols-[minmax(0,1.6fr)_minmax(280px,0.8fr)]">
              <Skeleton className="h-72" />
              <Skeleton className="h-96" />
            </div>
          </div>
        ) : (
          <div className="space-y-5">
            <SystemHealth health={health} />

            <section>
              <SectionHeader
                title="Run activity"
                description="Execution volume across agents and automations"
                aside={
                  <div className="inline-flex rounded-md border border-border bg-card p-0.5">
                    {ranges.map((item) => (
                      <button
                        key={item.key}
                        type="button"
                        onClick={() => setRange(item.key)}
                        aria-pressed={range === item.key}
                        className={cn(
                          "rounded px-2.5 py-1 text-xs font-medium transition-[color,background-color,box-shadow,scale] duration-150 ease-out active:scale-[0.96] motion-reduce:active:scale-100",
                          range === item.key
                            ? "bg-accent text-accent-foreground shadow-sm"
                            : "text-muted-foreground hover:bg-muted hover:text-foreground",
                        )}
                      >
                        {item.label}
                      </button>
                    ))}
                  </div>
                }
              />
              <div className="grid grid-cols-2 overflow-hidden rounded-lg border border-transparent bg-card shadow-card [&>*:nth-child(even)]:border-l [&>*:nth-child(n+3)]:border-t sm:grid-cols-4 sm:divide-x sm:divide-border sm:[&>*:nth-child(even)]:border-l-0 sm:[&>*:nth-child(n+3)]:border-t-0">
                <Metric label="Agent runs" value={agentRuns.toLocaleString()} detail="Agent invocations" emphasis />
                <Metric label="Automation runs" value={automationRuns.toLocaleString()} detail="Scheduled executions" />
                <Metric label="Failed runs" value={failedRuns.toLocaleString()} detail={failedRuns > 0 ? "Requires review" : "No failures"} />
                <Metric label="Success rate" value={`${successRate}%`} detail={`${totalRuns.toLocaleString()} total runs`} />
              </div>
            </section>

            <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1.6fr)_minmax(300px,0.8fr)]">
              <div className="space-y-5">
                <section>
                  <SectionHeader
                    title="Needs attention"
                    description="Failures and degraded infrastructure"
                    aside={needsAttention.length > 0 ? `${needsAttention.length} open` : "Clear"}
                  />
                  <div className="overflow-hidden rounded-lg border border-transparent bg-card shadow-card">
                    {attentionShown.length === 0 ? (
                      <EmptyRow icon={<CheckCircle2 className="text-success-foreground" />}>
                        Nothing needs attention right now
                      </EmptyRow>
                    ) : (
                      attentionShown.map((item) => <AttentionRow key={item.id} item={item} />)
                    )}
                  </div>
                </section>

                <section>
                  <SectionHeader
                    title="Active now"
                    description="Agent runs currently in progress"
                    aside={activeNow.length > 0 ? `${activeNow.length} running` : "Idle"}
                  />
                  <div className="overflow-hidden rounded-lg border border-transparent bg-card shadow-card">
                    {activeNow.length === 0 ? (
                      <EmptyRow icon={<CircleDashed />}>No agent runs are active</EmptyRow>
                    ) : (
                      activeNow.map((event) => <ActiveRow key={event.id} event={event} />)
                    )}
                  </div>
                </section>
              </div>

              <aside className="space-y-5">
                <section>
                  <SectionHeader
                    title="Recent activity"
                    description="Latest workspace events"
                    aside={
                      <button
                        type="button"
                        onClick={() => activityQuery.refetch()}
                        aria-label="Refresh recent activity"
                        className="inline-flex size-7 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100"
                      >
                        <RefreshCw className={cn("size-3.5", activityQuery.isFetching && "animate-spin")} />
                      </button>
                    }
                  />
                  {recent.length === 0 ? (
                    <div className="rounded-lg border border-transparent bg-card shadow-card">
                      <EmptyRow icon={<Clock3 />}>No recent activity</EmptyRow>
                    </div>
                  ) : (
                    <ol className="rounded-lg border border-transparent bg-card px-3 py-1 shadow-card">
                      {recent.map((event) => <ActivityRow key={event.id} event={event} />)}
                    </ol>
                  )}
                </section>

                {quickStartAgents.length > 0 && (
                  <section>
                    <SectionHeader title="Start a chat" description="Jump into an available agent" />
                    <div className="grid grid-cols-2 gap-2">
                      {quickStartAgents.map((agent) => (
                        <Link
                          key={agent.name}
                          to={`/chat?new=1&agent=${encodeURIComponent(agent.name)}`}
                          className="group flex min-w-0 touch-manipulation items-center gap-2.5 rounded-lg border border-transparent bg-card px-3 py-2.5 text-left shadow-card transition-[background-color,box-shadow,scale] duration-150 ease-out hover:bg-accent hover:shadow-card-hover active:scale-[0.96] motion-reduce:active:scale-100"
                        >
                          <AgentAvatar name={agent.name} size="sm" />
                          <span className="min-w-0 flex-1 truncate text-sm font-medium">{agent.name}</span>
                          <ArrowRight className="size-3.5 shrink-0 text-muted-foreground motion-safe:transition-transform motion-safe:duration-150 motion-safe:ease-out motion-safe:group-hover:translate-x-0.5" />
                        </Link>
                      ))}
                    </div>
                  </section>
                )}
              </aside>
            </div>
          </div>
        )}
      </PageScroll>
    </Page>
  );
}
