import { useMemo, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { useOverview, useActivityFeed } from "@/api/dashboard";
import { useAgents } from "@/api/agents";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import {
  Activity as ActivityIcon,
  AlertTriangle,
  ArrowRight,
  Bot,
  Check,
  CheckCircle2,
  Link2,
  Loader2,
  MessageSquare,
  Plus,
  RefreshCw,
  Server,
  Terminal,
  Workflow,
} from "lucide-react";
import type { ActivityEvent, ComponentHealth } from "@/types/api";

/* --------------------------------- helpers -------------------------------- */

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

const toneBar: Record<Tone, string> = {
  danger: "bg-danger",
  warning: "bg-warning",
  running: "bg-running",
  success: "bg-success",
  muted: "bg-muted-foreground/40",
};

const toneChip: Record<Tone, string> = {
  danger: "bg-danger-muted text-danger-foreground",
  warning: "bg-warning-muted text-warning-foreground",
  running: "bg-running-muted text-running-foreground",
  success: "bg-success-muted text-success-foreground",
  muted: "bg-muted text-muted-foreground",
};

/* --------------------------------- metrics -------------------------------- */

function MetricCard({
  label,
  value,
  icon,
}: {
  label: string;
  value: number;
  icon: ReactNode;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{label}</span>
        <span className="text-muted-foreground/70 [&_svg]:size-4">{icon}</span>
      </div>
      <p className="mt-3 text-3xl font-semibold tabular-nums tracking-tight">
        {value.toLocaleString()}
      </p>
    </div>
  );
}

/* ----------------------------- needs attention ---------------------------- */

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

function AttentionCard({ item }: { item: Attention }) {
  return (
    <div className="relative flex items-start gap-3 overflow-hidden rounded-lg border border-border bg-card p-4">
      <span className={cn("absolute inset-y-0 left-0 w-[3px]", toneBar[item.tone])} />
      <span
        className={cn(
          "mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full [&_svg]:size-4",
          toneChip[item.tone],
        )}
      >
        {item.icon}
      </span>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="truncate text-sm font-medium">{item.title}</p>
          {item.time && (
            <span className="shrink-0 text-xs text-muted-foreground">{item.time}</span>
          )}
        </div>
        <p className="mt-0.5 text-xs text-muted-foreground">{item.meta}</p>
        {item.detail && (
          <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{item.detail}</p>
        )}
      </div>
      {item.action && (
        <Button
          variant="outline"
          size="sm"
          className="shrink-0"
          render={<Link to={item.action.to} />}
        >
          {item.action.label}
        </Button>
      )}
    </div>
  );
}

/* -------------------------------- active now ------------------------------ */

function ActiveRow({ event }: { event: ActivityEvent }) {
  const title = event.message?.trim() || event.actor || "Agent run";
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start gap-3">
        <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-running-muted text-running-foreground">
          <Loader2 className="size-4 animate-spin" />
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{title}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {event.actor ?? "unknown"}
            {event.timestamp && ` · started ${relTime(event.timestamp)}`}
          </p>
        </div>
      </div>
      <div className="mt-3 h-1 w-full overflow-hidden rounded-full bg-running-muted">
        <div className="h-full w-full animate-pulse bg-running" />
      </div>
    </div>
  );
}

/* ----------------------------- recent activity ---------------------------- */

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
    <li className="flex gap-3 px-2 py-2.5">
      <span
        className={cn(
          "mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full [&_svg]:size-3.5",
          toneChip[tone],
        )}
      >
        {icon}
      </span>
      <div className="min-w-0 flex-1">
        <p className="line-clamp-2 text-sm leading-snug">{title}</p>
        <p className="mt-0.5 truncate text-xs text-muted-foreground">
          {event.actor ?? "unknown"}
          {event.timestamp && ` · ${relTime(event.timestamp)}`}
        </p>
      </div>
    </li>
  );
}

/* --------------------------------- section -------------------------------- */

function SectionHeader({
  title,
  aside,
  icon,
}: {
  title: string;
  aside?: ReactNode;
  icon?: ReactNode;
}) {
  return (
    <div className="mb-2 flex items-center justify-between">
      <div className="flex items-center gap-2">
        {icon && <span className="text-muted-foreground [&_svg]:size-4">{icon}</span>}
        <h2 className="text-sm font-medium">{title}</h2>
      </div>
      {aside && <div className="text-xs text-muted-foreground">{aside}</div>}
    </div>
  );
}

/* --------------------------------- page ----------------------------------- */

export default function DashboardPage() {
  const { data: overviewData, isLoading: loadingOverview } = useOverview("production");
  const activityQuery = useActivityFeed(50);
  const { data: agentsData } = useAgents({ page_size: 10 });

  const counts = overviewData?.counts;
  const health = overviewData?.health;
  const events = useMemo(() => activityQuery.data?.events ?? [], [activityQuery.data]);
  const quickStartAgents = (agentsData?.agents ?? []).slice(0, 4);

  const infraAttention = useMemo<Attention[]>(() => {
    const components: { label: string; health?: ComponentHealth }[] = [
      { label: "MongoDB", health: health?.mongodb },
      { label: "Redis cache", health: health?.redis },
      { label: "Runner", health: health?.runner },
    ];
    return components
      .filter(
        (c) => c.health?.status === "STATUS_DEGRADED" || c.health?.status === "STATUS_DOWN",
      )
      .map((c) => {
        const down = c.health?.status === "STATUS_DOWN";
        return {
          id: `infra-${c.label}`,
          tone: down ? "danger" : "warning",
          icon: <AlertTriangle />,
          title: `${c.label} ${down ? "unavailable" : "degraded"}`,
          meta: "Infrastructure",
          detail: c.health?.detail,
          action: { label: "Details", to: "/operations" },
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
          action: { label: "View", to: "/operations" },
        })),
    [events],
  );

  const needsAttention = [...infraAttention, ...errorAttention];
  const attentionCount = needsAttention.length;
  const attentionShown = needsAttention.slice(0, 6);

  const activeNow = useMemo(
    () => events.filter((e) => e.kind === "invocation").slice(0, 6),
    [events],
  );
  const recent = events.slice(0, 8);

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
            <Skeleton className="h-14" />
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-28" />
              ))}
            </div>
            <Skeleton className="h-72" />
          </div>
        ) : (
          <>
            {/* Attention banner */}
            <div
              className={cn(
                "mb-6 flex items-center gap-2.5 rounded-lg border px-4 py-3",
                attentionCount > 0
                  ? "border-warning/40 bg-warning-muted/40"
                  : "border-success/40 bg-success-muted/40",
              )}
            >
              {attentionCount > 0 ? (
                <AlertTriangle className="size-4 shrink-0 text-warning-foreground" />
              ) : (
                <CheckCircle2 className="size-4 shrink-0 text-success-foreground" />
              )}
              <div className="text-sm">
                <span className="font-medium">
                  {attentionCount > 0
                    ? `${attentionCount} ${attentionCount === 1 ? "item needs" : "items need"} attention`
                    : "All systems operational"}
                </span>
                <span className="ml-2 text-muted-foreground">
                  {attentionCount > 0
                    ? "Review the items below to keep things running."
                    : "Database, cache, and runner are healthy."}
                </span>
              </div>
            </div>

            {/* Metrics */}
            <section className="mb-6">
              <SectionHeader title="At a glance" />
              <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
                <MetricCard label="Active agents" value={counts?.active_agents ?? 0} icon={<Bot />} />
                <MetricCard label="Active sessions" value={counts?.active_sessions ?? 0} icon={<MessageSquare />} />
                <MetricCard label="Automations" value={counts?.cron_jobs ?? 0} icon={<Workflow />} />
                <MetricCard label="MCP servers" value={counts?.mcp_servers ?? 0} icon={<Server />} />
              </div>
            </section>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
              {/* Left: needs attention + active now */}
              <div className="flex flex-col gap-6 lg:col-span-2">
                <section>
                  <SectionHeader
                    title="Needs attention"
                    aside={attentionCount > 0 ? `${attentionCount} open` : undefined}
                  />
                  {attentionShown.length === 0 ? (
                    <div className="rounded-lg border border-border bg-card px-4 py-8 text-center text-sm text-muted-foreground">
                      Nothing needs attention right now.
                    </div>
                  ) : (
                    <div className="flex flex-col gap-2.5">
                      {attentionShown.map((item) => (
                        <AttentionCard key={item.id} item={item} />
                      ))}
                    </div>
                  )}
                </section>

                <section>
                  <SectionHeader
                    title="Active now"
                    aside={
                      activeNow.length > 0 ? (
                        <span className="flex items-center gap-1.5">
                          <span className="size-1.5 rounded-full bg-running animate-blink" />
                          {activeNow.length} running
                        </span>
                      ) : undefined
                    }
                  />
                  {activeNow.length === 0 ? (
                    <div className="rounded-lg border border-border bg-card px-4 py-8 text-center text-sm text-muted-foreground">
                      Nothing running right now.
                    </div>
                  ) : (
                    <div className="flex flex-col gap-2.5">
                      {activeNow.map((e) => (
                        <ActiveRow key={e.id} event={e} />
                      ))}
                    </div>
                  )}
                </section>
              </div>

              {/* Right: recent activity + quick start */}
              <div className="flex flex-col gap-6 lg:col-span-1">
                <section>
                  <SectionHeader title="Recent activity" icon={<ActivityIcon />} />
                  <div className="rounded-lg border border-border bg-card p-1">
                    {recent.length === 0 ? (
                      <p className="px-3 py-6 text-center text-sm text-muted-foreground">
                        No recent activity.
                      </p>
                    ) : (
                      <ol>
                        {recent.map((e) => (
                          <ActivityRow key={e.id} event={e} />
                        ))}
                      </ol>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={() => activityQuery.refetch()}
                    className="mt-2 inline-flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                  >
                    <RefreshCw className={cn("size-3", activityQuery.isFetching && "animate-spin")} />
                    Refresh
                  </button>
                </section>

                {quickStartAgents.length > 0 && (
                  <section>
                    <SectionHeader title="Quick start" />
                    <div className="flex flex-col gap-2">
                      {quickStartAgents.map((a) => (
                        <Link
                          key={a.name}
                          to={`/chat?new=1&agent=${encodeURIComponent(a.name)}`}
                          className="flex items-center gap-2.5 rounded-lg border border-border bg-card px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent"
                        >
                          <Bot className="size-4 text-muted-foreground" />
                          <span className="flex-1 truncate font-medium">{a.name}</span>
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
