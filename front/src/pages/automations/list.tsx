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
      className="group flex w-full cursor-pointer flex-col gap-3 border-b border-border px-4 py-3.5 text-left transition-colors last:border-b-0 hover:bg-accent/40 md:flex-row md:items-center md:gap-4 md:px-6"
    >
      {/* status + name */}
      <div className="flex min-w-0 flex-1 items-start gap-3">
        <div className="mt-0.5">
          <AgentAvatar name={job.agent_name} size="md" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="truncate text-sm font-medium">{job.name}</span>
            <StatusBadge status={status} />
          </div>
          <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1">
              <CalendarClock className="size-3.5" />
              <code className="font-mono">{job.schedule}</code>
            </span>
            <span>{job.agent_name}</span>
            <span className="inline-flex items-center gap-1">
              <DeliveryIcon className="size-3.5" />
              {delivery.label}
            </span>
            {lastRunAgo && <span>Last run {lastRunAgo}</span>}
          </div>
        </div>
      </div>

      {/* actions */}
      <div
        className="flex items-center gap-1.5 self-end md:self-center"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <span
          role="switch"
          aria-checked={enabled}
          aria-label="Toggle enabled"
          tabIndex={0}
          onClick={toggleEnabled}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              toggleEnabled();
            }
          }}
          className={cn(
            "relative inline-flex h-5 w-9 cursor-pointer items-center rounded-full transition-colors",
            enabled ? "bg-success" : "bg-muted-foreground/30",
          )}
        >
          <span
            className={cn(
              "inline-block size-4 translate-x-0.5 rounded-full bg-background transition-transform",
              enabled && "translate-x-[18px]",
            )}
          />
        </span>
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
          <DropdownMenuTrigger className="rounded-md p-1.5 text-muted-foreground hover:bg-muted">
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
            ? `${needsAttention} need attention · ${enabledCount} active`
            : `${enabledCount} active automations`
        }
        actions={
          <Button size="sm" render={<Link to="/automations/create" />}>
            <Plus />
            Create Automation
          </Button>
        }
      />
      <PageScroll className="max-w-5xl">
        {isLoading ? (
          <Skeleton className="h-64" />
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
            <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="relative w-full sm:max-w-xs">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search automations"
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

            <div className="overflow-hidden rounded-lg border border-border bg-card">
              {sorted.length === 0 ? (
                <p className="py-16 text-center text-sm text-muted-foreground">
                  No automations match your filters.
                </p>
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
