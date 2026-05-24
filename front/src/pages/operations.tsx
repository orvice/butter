import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCronExecutions, useCronJobs, useRunCronJobNow } from "@/api/cron";
import { useChannels } from "@/api/channels";
import { useSession, useSessions } from "@/api/sessions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  CalendarDays,
  Clock,
  Copy,
  ExternalLink,
  Filter,
  History,
  Play,
  Plus,
  ScrollText,
  Search,
  Webhook,
  MessageCircle,
  Bell,
} from "lucide-react";
import type { CronDeliveryType, CronExecution, CronJob, SessionInfo } from "@/types/api";

const DELIVERY_META: Record<CronDeliveryType, { icon: typeof Webhook; label: string }> = {
  CRON_DELIVERY_TYPE_WEBHOOK: { icon: Webhook, label: "Webhook" },
  CRON_DELIVERY_TYPE_CHANNEL: { icon: MessageCircle, label: "Channel" },
  CRON_DELIVERY_TYPE_NOTIFY_GROUP: { icon: Bell, label: "Notify Group" },
  CRON_DELIVERY_TYPE_LOG: { icon: ScrollText, label: "Log" },
  CRON_DELIVERY_TYPE_UNSPECIFIED: { icon: ScrollText, label: "Log" },
};

function timeAgo(ts?: string): string {
  if (!ts) return "-";
  const d = Date.now() - new Date(ts).getTime();
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`;
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`;
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`;
  return `${Math.floor(d / 86_400_000)}d ago`;
}

function fmtDuration(d?: string): string {
  if (!d) return "-";
  const secs = Number.parseFloat(d);
  if (!Number.isFinite(secs)) return d;
  const min = Math.floor(secs / 60);
  const s = Math.floor(secs % 60);
  return min > 0 ? `${min}m ${s}s` : `${s}s`;
}

function lastExecMap(executions: CronExecution[]) {
  const map = new Map<string, CronExecution>();
  for (const execution of executions) {
    const previous = map.get(execution.job_name);
    if (!previous || new Date(execution.started_at ?? 0) > new Date(previous.started_at ?? 0)) {
      map.set(execution.job_name, execution);
    }
  }
  return map;
}

async function copyText(text: string, successMessage: string) {
  try {
    await navigator.clipboard.writeText(text);
    toast.success(successMessage);
  } catch {
    toast.error("Copy failed");
  }
}

function CronRow({ job, execution }: { job: CronJob; execution?: CronExecution }) {
  const navigate = useNavigate();
  const runNow = useRunCronJobNow();
  const delivery = (job.delivery?.type ?? "CRON_DELIVERY_TYPE_UNSPECIFIED") as CronDeliveryType;
  const DeliveryIcon = DELIVERY_META[delivery].icon;
  const ok = execution?.status === "CRON_EXECUTION_STATUS_SUCCESS";

  return (
    <div className="grid gap-3 border-b px-5 py-4 last:border-b-0 lg:grid-cols-[1.2fr_1fr_1fr_0.8fr_1fr_auto] lg:items-center">
      <div>
        <div className="font-medium">{job.name}</div>
        {job.input ? <div className="line-clamp-1 text-xs text-muted-foreground">{job.input}</div> : null}
      </div>
      <code className="w-fit rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">{job.schedule}</code>
      <span className="text-sm">{job.agent_name}</span>
      <Badge variant="outline" className="w-fit">
        <DeliveryIcon className="mr-1 h-3.5 w-3.5" />
        {DELIVERY_META[delivery].label}
      </Badge>
      {execution ? (
        <div className="flex items-center gap-2">
          <Badge className={ok ? "bg-emerald-500/10 text-emerald-700" : "bg-rose-500/10 text-rose-700"}>
            <span className="h-1.5 w-1.5 rounded-full bg-current" />
            {ok ? "Success" : "Failed"}
          </Badge>
          <span className="text-xs text-muted-foreground">{timeAgo(execution.started_at)}</span>
        </div>
      ) : (
        <Badge variant="outline" className="w-fit text-muted-foreground">Pending</Badge>
      )}
      <div className="flex justify-end gap-1">
        <Button
          size="icon-sm"
          variant="ghost"
          disabled={runNow.isPending}
          onClick={() =>
            runNow.mutate(job.name, {
              onSuccess: () => toast.success(`${job.name} executed`),
              onError: (err) => toast.error(err.message),
            })
          }
          aria-label={`Run ${job.name}`}
        >
          <Play className="h-4 w-4" />
        </Button>
        <Button size="icon-sm" variant="ghost" onClick={() => navigate(`/cron/${encodeURIComponent(job.name)}/edit`)} aria-label={`Edit ${job.name}`}>
          <ExternalLink className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function SessionRow({
  session,
  selected,
  onSelect,
}: {
  session: SessionInfo;
  selected: boolean;
  onSelect: (session: SessionInfo) => void;
}) {
  return (
    <button
      className={`grid w-full gap-3 border-b px-5 py-3 text-left text-sm transition-colors last:border-b-0 md:grid-cols-[1.4fr_1fr_1fr_0.5fr_0.8fr] md:items-center ${
        selected ? "border-l-2 border-l-primary bg-primary/5" : "border-l-2 border-l-transparent hover:bg-muted/50"
      }`}
      onClick={() => onSelect(session)}
    >
      <code className="truncate text-xs">{session.session_id}</code>
      <span className="truncate">{session.app_name || "API"}</span>
      <span className="text-muted-foreground">{session.last_update_time ? new Date(session.last_update_time).toLocaleTimeString() : "-"}</span>
      <span>{session.turn_count ?? 0}</span>
      <span className="justify-self-start text-primary md:justify-self-end">Langfuse</span>
    </button>
  );
}

function SessionDetailPanel({ session }: { session?: SessionInfo }) {
  const { data, isLoading } = useSession(session?.app_name ?? "", session?.user_id ?? "", session?.session_id ?? "");
  const detail = data?.session_detail;
  const memory = detail?.session.state ?? session?.state ?? {};
  const events = detail?.events ?? [];
  const sessionJson = JSON.stringify(detail ?? session ?? {}, null, 2);
  const memoryJson = JSON.stringify(memory, null, 2);

  return (
    <Card className="border-t-4 border-t-primary xl:sticky xl:top-24">
      <CardHeader className="flex flex-row items-start justify-between border-b bg-muted/30 pb-4">
        <div>
          <CardTitle>Session Detail</CardTitle>
          <CardDescription className="font-mono">{session?.session_id ?? "No session selected"}</CardDescription>
        </div>
        <Button
          size="icon-sm"
          variant="ghost"
          aria-label="Copy session JSON"
          disabled={!session}
          onClick={() => void copyText(sessionJson, "Session JSON copied")}
        >
          <Copy className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent className="space-y-6 pt-5">
        {!session ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
            Select a session to inspect its memory and event log.
          </div>
        ) : isLoading ? (
          <div className="space-y-3">
            <Skeleton className="h-16" />
            <Skeleton className="h-44" />
            <Skeleton className="h-32" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="block text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">User</span>
                <span className="font-medium">{session.user_id}</span>
              </div>
              <div>
                <span className="block text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">Duration</span>
                <span className="font-medium">{fmtDuration(detail?.duration)}</span>
              </div>
            </div>
            <div>
              <div className="mb-2 flex items-center justify-between">
                <h4 className="text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">Memory Context</h4>
                <Button size="sm" variant="ghost" onClick={() => void copyText(memoryJson, "Memory JSON copied")}>Copy JSON</Button>
              </div>
              <pre className="max-h-64 overflow-auto rounded-lg bg-[#111827] p-4 text-[13px] leading-6 text-gray-200">
                {JSON.stringify(memory, null, 2)}
              </pre>
            </div>
            <div>
              <h4 className="mb-3 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">Event Log</h4>
              {events.length === 0 ? (
                <p className="text-sm text-muted-foreground">No events returned for this session.</p>
              ) : (
                <div className="relative space-y-4 before:absolute before:bottom-2 before:left-1 before:top-2 before:w-px before:bg-border">
                  {events.slice(0, 6).map((event) => (
                    <div key={event.event_id} className="relative pl-6">
                      <span className="absolute left-0 top-1.5 h-2 w-2 rounded-full bg-primary ring-4 ring-card" />
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">{event.author || "Event"}</span>
                        <span className="text-[10px] text-muted-foreground">{event.timestamp ? new Date(event.timestamp).toLocaleTimeString() : ""}</span>
                      </div>
                      <p className="line-clamp-2 text-xs text-muted-foreground">{event.content_json || event.invocation_id || event.event_id}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

export default function OperationsPage() {
  const navigate = useNavigate();
  const [selectedSession, setSelectedSession] = useState<SessionInfo | undefined>();
  const [sessionFilterDraft, setSessionFilterDraft] = useState({ appName: "", userId: "" });
  const [sessionFilters, setSessionFilters] = useState({ appName: "", userId: "" });
  const { data: jobsData, isLoading: loadingJobs } = useCronJobs();
  const { data: execData } = useCronExecutions(undefined, 200);
  const { data: channelData } = useChannels();
  const { data: sessionData, isLoading: loadingSessions } = useSessions({
    app_name: sessionFilters.appName || undefined,
    user_id: sessionFilters.userId || undefined,
    page_size: 50,
  });

  const jobs = jobsData?.cron_jobs ?? [];
  const sessions = sessionData?.sessions ?? [];
  const channelNames = Array.from(
    new Set(
      (channelData?.channels ?? [])
        .map((channel) => channel.name)
        .filter((name) => name && !["api", "telegram", "web-chat"].includes(name)),
    ),
  );
  const executionByJob = useMemo(() => lastExecMap(execData?.executions ?? []), [execData?.executions]);
  const currentSession = selectedSession ?? sessions[0];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Operations Center</h2>
          <p className="mt-1 text-sm text-muted-foreground">Cron schedules, ADK sessions, memory state, and tracing links.</p>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Clock className="h-4 w-4" />
          Last synced: Just now
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between border-b pb-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              <CalendarDays className="h-4 w-4 text-primary" />
              Cron Job Manager
            </CardTitle>
            <CardDescription>Scheduled agent execution with delivery status.</CardDescription>
          </div>
          <Button onClick={() => navigate("/cron/create")}>
            <Plus className="mr-2 h-4 w-4" />
            Add Schedule
          </Button>
        </CardHeader>
        <CardContent className="p-0">
          {loadingJobs ? (
            <div className="space-y-3 p-5">
              {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-14" />)}
            </div>
          ) : jobs.length === 0 ? (
            <div className="p-8 text-center text-sm text-muted-foreground">No cron jobs yet.</div>
          ) : (
            <div>
              <div className="hidden grid-cols-[1.2fr_1fr_1fr_0.8fr_1fr_auto] border-b bg-muted/60 px-5 py-3 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground lg:grid">
                <span>Job Name</span>
                <span>Schedule</span>
                <span>Target Agent</span>
                <span>Delivery</span>
                <span>Last Execution</span>
                <span className="text-right">Actions</span>
              </div>
              {jobs.slice(0, 6).map((job) => (
                <CronRow key={job.name} job={job} execution={executionByJob.get(job.name)} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(360px,1fr)]">
        <Card className="min-h-[520px]">
          <CardHeader className="border-b bg-muted/30 pb-4">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Search className="h-4 w-4 text-primary" />
                  Session Explorer
                  <span className="text-sm font-normal text-muted-foreground">(MongoDB ADK)</span>
                </CardTitle>
                <CardDescription>Inspect session turns, memory, and trace context.</CardDescription>
              </div>
              <div className="grid gap-2 sm:grid-cols-3">
                <Select
                  value={sessionFilterDraft.appName || "__all__"}
                  onValueChange={(value) =>
                    setSessionFilterDraft((current) => ({
                      ...current,
                      appName: !value || value === "__all__" ? "" : value,
                    }))
                  }
                >
                  <SelectTrigger className="w-full sm:w-36">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__all__">All Channels</SelectItem>
                    <SelectItem value="api">API</SelectItem>
                    <SelectItem value="web-chat">Web Chat</SelectItem>
                    <SelectItem value="telegram">Telegram</SelectItem>
                    {channelNames.map((name) => (
                      <SelectItem key={name} value={name}>{name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Input
                  placeholder="User ID..."
                  value={sessionFilterDraft.userId}
                  onChange={(event) => setSessionFilterDraft((current) => ({ ...current, userId: event.target.value }))}
                />
                <Button
                  variant="outline"
                  onClick={() => {
                    setSelectedSession(undefined);
                    setSessionFilters({
                      appName: sessionFilterDraft.appName,
                      userId: sessionFilterDraft.userId.trim(),
                    });
                  }}
                >
                  <Filter className="mr-2 h-4 w-4" />
                  Apply
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {loadingSessions ? (
              <div className="space-y-3 p-5">
                {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-12" />)}
              </div>
            ) : sessions.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">No sessions found.</div>
            ) : (
              <div>
                <div className="hidden grid-cols-[1.4fr_1fr_1fr_0.5fr_0.8fr] border-b bg-muted/60 px-5 py-3 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground md:grid">
                  <span>Session ID</span>
                  <span>Channel</span>
                  <span>Start Time</span>
                  <span>Turns</span>
                  <span className="text-right">Tracing</span>
                </div>
                {sessions.slice(0, 12).map((session) => (
                  <SessionRow
                    key={`${session.app_name}-${session.user_id}-${session.session_id}`}
                    session={session}
                    selected={currentSession?.session_id === session.session_id}
                    onSelect={setSelectedSession}
                  />
                ))}
                <div className="flex items-center justify-between border-t bg-muted/30 px-5 py-3 text-sm text-muted-foreground">
                  <span>Showing {sessions.length} of {sessionData?.total ?? sessions.length} sessions</span>
                  <Button size="sm" variant="ghost" onClick={() => navigate("/sessions")}>
                    <History className="mr-2 h-4 w-4" />
                    Full explorer
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <SessionDetailPanel session={currentSession} />
      </div>
    </div>
  );
}
