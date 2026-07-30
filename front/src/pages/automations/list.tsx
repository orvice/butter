import { useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useCronJobs,
  useCronExecutions,
  useDeleteCronJob,
  useRunCronJobNow,
  useUpdateCronJob,
} from "@/api/cron";
import { DeleteDialog } from "@/components/delete-dialog";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { AgentAvatar, StatusBadge, type RunStatus } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { CronExecution, CronJob } from "@/types/api";
import {
  Bell,
  CalendarClock,
  MessageCircle,
  MoreVertical,
  Pencil,
  Play,
  Plus,
  ScrollText,
  Search,
  Trash2,
  Webhook,
  Workflow,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { cronExecStatus } from "./run-status";

const DELIVERY_META: Record<string, { icon: typeof Webhook; label: string }> = {
  CRON_DELIVERY_TYPE_WEBHOOK: { icon: Webhook, label: "Webhook" },
  CRON_DELIVERY_TYPE_CHANNEL: { icon: MessageCircle, label: "Channel" },
  CRON_DELIVERY_TYPE_NOTIFY_GROUP: { icon: Bell, label: "Notify Group" },
  CRON_DELIVERY_TYPE_LOG: { icon: ScrollText, label: "Log" },
  CRON_DELIVERY_TYPE_UNSPECIFIED: { icon: ScrollText, label: "Log" },
};

function timeAgo(ts?: string): string | null {
  if (!ts) return null;
  const d = Date.now() - new Date(ts).getTime();
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`;
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`;
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`;
  return `${Math.floor(d / 86_400_000)}d ago`;
}

function lastExecMap(executions: CronExecution[]) {
  const map = new Map<string, CronExecution>();
  for (const e of executions) {
    const prev = map.get(e.job_name);
    if (!prev || new Date(e.started_at ?? 0) > new Date(prev.started_at ?? 0)) {
      map.set(e.job_name, e);
    }
  }
  return map;
}

function jobStatus(job: CronJob, lastExec?: CronExecution): RunStatus {
  if (!(job.enabled ?? false)) return "disabled";
  return cronExecStatus(lastExec?.status ?? "");
}

// priority ordering: failed, waiting, running, success, never, disabled
const priority: Record<RunStatus, number> = {
  failed: 0,
  waiting: 1,
  running: 2,
  success: 3,
  never: 4,
  disabled: 5,
};

function AutomationRow({
  job,
  lastExec,
  onDelete,
}: {
  job: CronJob;
  lastExec?: CronExecution;
  onDelete: () => void;
}) {
  const navigate = useNavigate();
  const updateMutation = useUpdateCronJob();
  const runNow = useRunCronJobNow();
  const enabled = job.enabled ?? false;
  const status = jobStatus(job, lastExec);
  const lastRunAgo = timeAgo(lastExec?.started_at);
  const delivery = DELIVERY_META[job.delivery?.type ?? "CRON_DELIVERY_TYPE_LOG"] ?? DELIVERY_META.CRON_DELIVERY_TYPE_LOG;
  const DeliveryIcon = delivery.icon;

  function toggleEnabled() {
    updateMutation.mutate(
      { ...job, enabled: !enabled },
      {
        onSuccess: () => toast.success(`Automation ${enabled ? "disabled" : "enabled"}`),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => navigate(`/automations/${encodeURIComponent(job.name)}`)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          navigate(`/automations/${encodeURIComponent(job.name)}`);
        }
      }}
      className="group grid w-full cursor-pointer grid-cols-2 gap-x-4 gap-y-3 border-b border-border px-4 py-4 text-left transition-colors last:border-b-0 hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring md:grid-cols-3 md:px-5 lg:grid-cols-[minmax(190px,1.5fr)_minmax(100px,.72fr)_minmax(110px,.75fr)_90px_125px_190px] lg:items-center lg:gap-4 lg:px-5 lg:py-3.5"
    >
      <div className="col-span-2 flex min-w-0 items-center gap-3 md:col-span-3 lg:col-span-1">
        <AgentAvatar name={job.agent_name} size="md" />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold">{job.name}</div>
          <div className="mt-0.5 truncate text-xs text-muted-foreground">{job.agent_name}</div>
        </div>
      </div>

      <div className="min-w-0">
        <div className="mb-1 text-[0.68rem] font-medium uppercase text-muted-foreground lg:hidden">
          Schedule
        </div>
        <div className="flex min-w-0 items-center gap-1.5 text-xs text-foreground">
          <CalendarClock className="size-3.5 shrink-0 text-muted-foreground lg:hidden" />
          <code className="truncate font-mono">{job.schedule}</code>
        </div>
      </div>

      <div className="min-w-0">
        <div className="mb-1 text-[0.68rem] font-medium uppercase text-muted-foreground lg:hidden">
          Delivery
        </div>
        <div className="flex min-w-0 items-center gap-1.5 text-xs">
          <DeliveryIcon className="size-3.5 shrink-0 text-muted-foreground" />
          <span className="truncate">{delivery.label}</span>
        </div>
      </div>

      <div className="min-w-0">
        <div className="mb-1 text-[0.68rem] font-medium uppercase text-muted-foreground lg:hidden">
          Last run
        </div>
        <div className="truncate text-xs text-muted-foreground">{lastRunAgo ?? "Not run"}</div>
      </div>

      <div className="min-w-0">
        <div className="mb-1 text-[0.68rem] font-medium uppercase text-muted-foreground lg:hidden">
          Status
        </div>
        <StatusBadge status={status} className="max-w-full" />
      </div>

      <div
        className="col-span-2 flex min-w-0 items-center justify-end gap-2 border-t border-border/70 pt-3 md:col-span-2 md:border-t-0 md:pt-0 lg:col-span-1"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <Switch
          checked={enabled}
          disabled={updateMutation.isPending}
          onCheckedChange={toggleEnabled}
          aria-label={`${enabled ? "Disable" : "Enable"} ${job.name}`}
          className="mr-auto lg:mr-1"
        />
        <Button
          variant="outline"
          size="sm"
          disabled={runNow.isPending}
          onClick={() =>
            runNow.mutate(job.name, {
              onSuccess: () => toast.success(`${job.name} triggered`),
              onError: (err) => toast.error(err.message),
            })
          }
        >
          <Play />
          Run now
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger
            aria-label={`More actions for ${job.name}`}
            title="More actions"
            className="rounded-md p-1.5 text-muted-foreground outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring"
          >
            <MoreVertical className="size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" sideOffset={6}>
            <DropdownMenuItem
              onClick={() => navigate(`/automations/${encodeURIComponent(job.name)}/edit`)}
            >
              <Pencil /> Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={onDelete}>
              <Trash2 /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

type StatusFilter = "all" | "failed" | "waiting" | "success" | "disabled";

export default function AutomationListPage() {
  const { data, isLoading } = useCronJobs();
  const { data: execData } = useCronExecutions(undefined, 200);
  const deleteMutation = useDeleteCronJob();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<StatusFilter>("all");

  const jobs = useMemo(() => data?.cron_jobs ?? [], [data?.cron_jobs]);
  const lastByJob = useMemo(() => lastExecMap(execData?.executions ?? []), [execData?.executions]);

  const sorted = useMemo(() => {
    return jobs
      .filter((j) => {
        const status = jobStatus(j, lastByJob.get(j.name));
        return (
          (filter === "all" || status === filter) &&
          (j.name.toLowerCase().includes(query.toLowerCase()) ||
            j.agent_name.toLowerCase().includes(query.toLowerCase()))
        );
      })
      .slice()
      .sort(
        (a, b) =>
          priority[jobStatus(a, lastByJob.get(a.name))] -
          priority[jobStatus(b, lastByJob.get(b.name))],
      );
  }, [jobs, query, filter, lastByJob]);

  const enabledCount = jobs.filter((j) => j.enabled).length;
  const needsAttention = jobs.filter((j) => {
    const s = jobStatus(j, lastByJob.get(j.name));
    return s === "failed" || s === "waiting";
  }).length;

  const statusCounts = useMemo(() => {
    const counts: Record<StatusFilter, number> = {
      all: jobs.length,
      failed: 0,
      waiting: 0,
      success: 0,
      disabled: 0,
    };
    for (const job of jobs) {
      const status = jobStatus(job, lastByJob.get(job.name));
      if (
        status === "failed" ||
        status === "waiting" ||
        status === "success" ||
        status === "disabled"
      ) {
        counts[status] += 1;
      }
    }
    return counts;
  }, [jobs, lastByJob]);

  const filters: { key: StatusFilter; label: string }[] = [
    { key: "all", label: "All" },
    { key: "failed", label: "Failed" },
    { key: "waiting", label: "Waiting" },
    { key: "success", label: "Success" },
    { key: "disabled", label: "Disabled" },
  ];

  return (
    <Page>
      <PageHeader
        title="Automations"
        subtitle={
          needsAttention > 0
            ? `${jobs.length} automations, ${enabledCount} active, ${needsAttention} need attention`
            : `${jobs.length} automations, ${enabledCount} active`
        }
        actions={
          <Button size="sm" render={<Link to="/automations/create" />}>
            <Plus />
            Create Automation
          </Button>
        }
      />
      <PageScroll className="max-w-6xl">
        {isLoading ? (
          <div className="overflow-hidden rounded-lg border border-border bg-card">
            <div className="hidden grid-cols-[minmax(190px,1.5fr)_minmax(100px,.72fr)_minmax(110px,.75fr)_90px_125px_190px] gap-4 border-b border-border bg-muted/35 px-5 py-2.5 lg:grid">
              {Array.from({ length: 6 }).map((_, index) => (
                <Skeleton key={index} className="h-3 w-16" />
              ))}
            </div>
            {Array.from({ length: 4 }).map((_, index) => (
              <div
                key={index}
                className="grid grid-cols-2 gap-4 border-b border-border px-4 py-4 last:border-b-0 md:grid-cols-3 lg:grid-cols-[minmax(190px,1.5fr)_minmax(100px,.72fr)_minmax(110px,.75fr)_90px_125px_190px] lg:px-5"
              >
                <div className="col-span-2 flex items-center gap-3 md:col-span-3 lg:col-span-1">
                  <Skeleton className="size-8 shrink-0 rounded-md" />
                  <div className="min-w-0 flex-1 space-y-1.5">
                    <Skeleton className="h-3.5 w-28" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                </div>
                {Array.from({ length: 5 }).map((__, cellIndex) => (
                  <Skeleton key={cellIndex} className="h-5 w-full max-w-24" />
                ))}
              </div>
            ))}
          </div>
        ) : jobs.length === 0 ? (
          <div className="mx-auto max-w-md rounded-lg border border-dashed border-border bg-card/40 px-6 py-14 text-center">
            <div className="mx-auto flex size-11 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <Workflow className="size-5" />
            </div>
            <h2 className="mt-4 text-base font-semibold">No automations yet</h2>
            <p className="mx-auto mt-1 max-w-xs text-sm text-muted-foreground text-pretty">
              Automations run an agent on a schedule and deliver the result to a
              webhook, channel, or notify group.
            </p>
            <Button className="mt-5" render={<Link to="/automations/create" />}>
              <Plus />
              Create your first Automation
            </Button>
          </div>
        ) : (
          <>
            <div className="mb-4 flex flex-col gap-2 rounded-lg border border-border bg-card p-2 lg:flex-row lg:items-center">
              <div className="relative w-full lg:max-w-[17rem] lg:shrink-0">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search automations"
                  className="h-9 w-full rounded-md border border-transparent bg-muted/55 pl-9 pr-3 text-sm outline-none transition-colors placeholder:text-muted-foreground focus:border-ring focus:bg-background"
                />
              </div>
              <div className="flex flex-wrap items-center gap-1 lg:min-w-0">
                {filters.map((f) => (
                  <button
                    key={f.key}
                    type="button"
                    onClick={() => setFilter(f.key)}
                    className={cn(
                      "inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors",
                      filter === f.key
                        ? "bg-secondary text-secondary-foreground shadow-sm"
                        : "text-muted-foreground hover:bg-muted/60 hover:text-foreground",
                    )}
                  >
                    <span>{f.label}</span>
                    <span
                      className={cn(
                        "tabular-nums text-[0.68rem]",
                        filter === f.key ? "text-secondary-foreground/70" : "text-muted-foreground/70",
                      )}
                    >
                      {statusCounts[f.key]}
                    </span>
                  </button>
                ))}
              </div>
            </div>

            <div className="overflow-hidden rounded-lg border border-border bg-card">
              <div className="hidden grid-cols-[minmax(190px,1.5fr)_minmax(100px,.72fr)_minmax(110px,.75fr)_90px_125px_190px] gap-4 border-b border-border bg-muted/35 px-5 py-2.5 text-[0.68rem] font-medium uppercase text-muted-foreground lg:grid">
                <span>Automation</span>
                <span>Schedule</span>
                <span>Delivery</span>
                <span>Last run</span>
                <span>Status</span>
                <span className="text-right">Controls</span>
              </div>
              {sorted.length === 0 ? (
                <div className="px-6 py-14 text-center">
                  <Search className="mx-auto size-5 text-muted-foreground" />
                  <p className="mt-3 text-sm font-medium">No matching automations</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Try a different search or status filter.
                  </p>
                </div>
              ) : (
                sorted.map((j) => (
                  <AutomationRow
                    key={j.name}
                    job={j}
                    lastExec={lastByJob.get(j.name)}
                    onDelete={() => setDeleteTarget(j.name)}
                  />
                ))
              )}
            </div>
          </>
        )}
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Automation"
        description={`Delete "${deleteTarget}"? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget, {
            onSuccess: () => {
              toast.success("Automation deleted");
              setDeleteTarget(null);
            },
            onError: (err) => toast.error(err.message),
          });
        }}
      />
    </Page>
  );
}
