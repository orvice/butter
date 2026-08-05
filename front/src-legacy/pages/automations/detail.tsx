import { useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useCronJob, useCronExecutions, useRunCronJobNow } from "@/api/cron";
import { Page } from "@/components/butter/page-parts";
import { AgentAvatar, StatusBadge, type RunStatus } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import type { CronExecution } from "@/types/api";
import {
  ArrowLeft,
  ChevronDown,
  ChevronRight,
  Clock,
  Pencil,
  Play,
  TriangleAlert,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { cronExecStatus } from "./run-status";

function formatTime(ts?: string) {
  return ts ? new Date(ts).toLocaleString() : "—";
}

function formatDuration(ms?: number) {
  if (ms === undefined) return "—";
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function Stat({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card px-3 py-2.5">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 truncate text-sm font-medium">{value}</p>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4 py-2">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

function ExecutionPanel({ execution }: { execution: CronExecution }) {
  const [outputOpen, setOutputOpen] = useState(false);
  const status = cronExecStatus(execution.status);

  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
        <div className="flex items-center gap-3">
          <StatusBadge status={status} />
          <span className="text-sm text-muted-foreground">{formatTime(execution.started_at)}</span>
          {(execution.attempt_count ?? 0) > 1 && (
            <span className="text-xs text-muted-foreground">×{execution.attempt_count} attempts</span>
          )}
        </div>
        <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
          <Clock className="size-3.5" />
          {status === "running" ? "In progress" : formatDuration(execution.duration_ms)}
        </span>
      </div>

      {execution.error && (
        <div className="border-b border-border bg-danger-muted/40 px-4 py-3">
          <div className="flex items-start gap-2 text-sm text-danger-foreground">
            <TriangleAlert className="mt-0.5 size-4 shrink-0" />
            <div>
              <p className="font-medium">Error</p>
              <p className="mt-0.5 text-danger-foreground/90">{execution.error}</p>
            </div>
          </div>
        </div>
      )}

      {status === "waiting" && (
        <div className="border-b border-border bg-warning-muted/40 px-4 py-3">
          <div className="flex items-start gap-2 text-sm">
            <TriangleAlert className="mt-0.5 size-4 shrink-0 text-warning-foreground" />
            <div>
              <p className="font-medium text-warning-foreground">Waiting for human input</p>
              <p className="mt-0.5 text-foreground">
                This run paused at a human-input step. Reply in the linked chat session to resume it.
              </p>
            </div>
          </div>
        </div>
      )}

      {execution.skipped_reason && (
        <div className="border-b border-border px-4 py-2.5 text-xs text-muted-foreground">
          Skipped: {execution.skipped_reason}
        </div>
      )}

      {execution.output && (
        <div>
          <button
            type="button"
            onClick={() => setOutputOpen((v) => !v)}
            className="flex w-full items-center gap-2 px-4 py-2.5 text-left text-sm hover:bg-muted/40"
          >
            <span className="text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
              Output{execution.truncated ? " (truncated)" : ""}
            </span>
            <ChevronDown
              className={cn(
                "ml-auto size-4 text-muted-foreground transition-transform",
                outputOpen && "rotate-180",
              )}
            />
          </button>
          {outputOpen && (
            <pre className="scrollbar-thin max-h-72 overflow-auto border-t border-border bg-muted/40 px-4 py-3 font-mono text-xs leading-5 whitespace-pre-wrap">
              {execution.output}
            </pre>
          )}
        </div>
      )}
    </div>
  );
}

export default function AutomationDetailPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const { data: execData, isLoading: loadingExecs } = useCronExecutions(name, 100);
  const runNow = useRunCronJobNow();
  const [tab, setTab] = useState<"overview" | "runs" | "trigger">("overview");

  const job = data?.cron_job;
  const executions = useMemo(() => execData?.executions ?? [], [execData?.executions]);
  const latest = executions[0];

  if (isLoading) {
    return (
      <div className="p-6">
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (!job) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-muted-foreground">Automation not found.</p>
      </div>
    );
  }

  const enabled = job.enabled ?? false;
  const status: RunStatus = enabled ? cronExecStatus(latest?.status ?? "") : "disabled";
  const delivery = job.delivery?.type?.replace("CRON_DELIVERY_TYPE_", "").toLowerCase() ?? "log";

  const tabs = [
    { key: "overview" as const, label: "Overview" },
    { key: "runs" as const, label: `Recent runs (${executions.length})` },
    { key: "trigger" as const, label: "Trigger" },
  ];

  return (
    <Page>
      {/* header */}
      <div className="border-b border-border px-4 py-4 md:px-6">
        <Link
          to="/automations"
          className="mb-3 inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-3.5" />
          Automations
          <ChevronRight className="size-3" />
          <span className="text-foreground">{job.name}</span>
        </Link>
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div className="flex min-w-0 items-start gap-3">
            <AgentAvatar name={job.agent_name} size="lg" />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-lg font-semibold tracking-tight">{job.name}</h1>
                <StatusBadge status={status} />
              </div>
              <p className="mt-0.5 text-sm text-muted-foreground">
                Runs <span className="text-foreground">{job.agent_name}</span> on{" "}
                <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs">{job.schedule}</code>
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate(`/automations/${encodeURIComponent(job.name)}/edit`)}
            >
              <Pencil />
              Edit
            </Button>
            <Button
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
          </div>
        </div>

        {/* tabs */}
        <div className="mt-4 flex gap-1">
          {tabs.map((t) => (
            <button
              key={t.key}
              type="button"
              onClick={() => setTab(t.key)}
              className={cn(
                "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                tab === t.key
                  ? "bg-secondary text-secondary-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      <div className="scrollbar-thin min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-4xl px-4 py-5 md:px-6">
          {tab === "overview" && (
            <div className="space-y-4">
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                <Stat
                  label="Schedule"
                  value={<code className="font-mono text-xs">{job.schedule}</code>}
                />
                <Stat label="Agent" value={job.agent_name} />
                <Stat label="Last run" value={latest ? formatTime(latest.started_at) : "Never"} />
                <Stat label="Delivery" value={delivery} />
              </div>
              {loadingExecs ? (
                <Skeleton className="h-40" />
              ) : latest ? (
                <div>
                  <h2 className="mb-2 text-sm font-medium">Latest run</h2>
                  <ExecutionPanel execution={latest} />
                </div>
              ) : (
                <div className="rounded-lg border border-dashed border-border bg-card/40 px-6 py-10 text-center">
                  <p className="text-sm font-medium">This automation has never run</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Run it now to see execution details here.
                  </p>
                  <Button
                    size="sm"
                    className="mt-4"
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
                </div>
              )}
            </div>
          )}

          {tab === "runs" && (
            <div className="space-y-4">
              {loadingExecs ? (
                <Skeleton className="h-40" />
              ) : executions.length === 0 ? (
                <p className="py-16 text-center text-sm text-muted-foreground">
                  No runs recorded yet.
                </p>
              ) : (
                executions.map((e) => <ExecutionPanel key={e.id} execution={e} />)
              )}
            </div>
          )}

          {tab === "trigger" && (
            <div className="space-y-4">
              <div className="rounded-lg border border-border bg-card p-4">
                <h2 className="text-sm font-medium">Trigger configuration</h2>
                <dl className="mt-3 divide-y divide-border text-sm">
                  <Row
                    label="Schedule (cron)"
                    value={
                      <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                        {job.schedule}
                      </code>
                    }
                  />
                  <Row label="Timezone" value={job.timezone || "UTC"} />
                  <Row label="Enabled" value={enabled ? "Yes" : "No"} />
                  <Row
                    label="Concurrency"
                    value={(job.concurrency_policy ?? "CRON_CONCURRENCY_POLICY_SKIP")
                      .replace("CRON_CONCURRENCY_POLICY_", "")
                      .toLowerCase()}
                  />
                  {job.timeout_seconds ? (
                    <Row label="Timeout" value={`${job.timeout_seconds}s`} />
                  ) : null}
                  {job.retry?.max_attempts ? (
                    <Row
                      label="Retry"
                      value={`${job.retry.max_attempts} attempts, ${job.retry.backoff_seconds ?? 0}s backoff`}
                    />
                  ) : null}
                </dl>
              </div>
              <div className="rounded-lg border border-border bg-card p-4">
                <h2 className="text-sm font-medium">Input &amp; delivery</h2>
                <dl className="mt-3 divide-y divide-border text-sm">
                  <Row label="Delivery" value={delivery} />
                  {job.delivery?.webhook_url && (
                    <Row
                      label="Webhook"
                      value={<span className="break-all font-mono text-xs">{job.delivery.webhook_url}</span>}
                    />
                  )}
                  {job.delivery?.channel_name && (
                    <Row label="Channel" value={job.delivery.channel_name} />
                  )}
                  {job.delivery?.notify_group_name && (
                    <Row label="Notify group" value={job.delivery.notify_group_name} />
                  )}
                </dl>
                {job.input && (
                  <div className="mt-3">
                    <p className="mb-1 text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
                      Input message
                    </p>
                    <pre className="scrollbar-thin overflow-auto rounded-md bg-muted/60 p-2.5 font-mono text-xs leading-5 whitespace-pre-wrap">
                      {job.input}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </Page>
  );
}
