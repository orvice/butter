import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useCronJobs,
  useCronExecutions,
  useDeleteCronJob,
  useRunCronJobNow,
  useUpdateCronJob,
} from "@/api/cron";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  History,
  Play,
  Webhook,
  MessageCircle,
  ScrollText,
  Bell,
} from "lucide-react";
import type { CronJob, CronDeliveryType, CronExecution } from "@/types/api";

const DELIVERY_META: Record<
  CronDeliveryType,
  { icon: typeof Webhook; label: string }
> = {
  CRON_DELIVERY_TYPE_WEBHOOK: { icon: Webhook, label: "Webhook" },
  CRON_DELIVERY_TYPE_CHANNEL: { icon: MessageCircle, label: "Channel" },
  CRON_DELIVERY_TYPE_NOTIFY_GROUP: { icon: Bell, label: "Notify Group" },
  CRON_DELIVERY_TYPE_LOG: { icon: ScrollText, label: "Log" },
  CRON_DELIVERY_TYPE_UNSPECIFIED: { icon: ScrollText, label: "Log" },
};

function timeAgo(ts?: string): string {
  if (!ts) return "—";
  const d = Date.now() - new Date(ts).getTime();
  if (d < 60_000) return `${Math.max(1, Math.floor(d / 1000))}s ago`;
  if (d < 3600_000) return `${Math.floor(d / 60_000)}m ago`;
  if (d < 86_400_000) return `${Math.floor(d / 3600_000)}h ago`;
  return `${Math.floor(d / 86_400_000)}d ago`;
}

export default function CronJobListPage() {
  const { data, isLoading } = useCronJobs();
  const { data: recentExec } = useCronExecutions(undefined, 200);
  const deleteMutation = useDeleteCronJob();
  const updateMutation = useUpdateCronJob();
  const runNow = useRunCronJobNow();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  function toggleEnabled(job: CronJob) {
    updateMutation.mutate(
      { ...job, enabled: !job.enabled },
      {
        onSuccess: () => toast.success(`Job ${job.enabled ? "disabled" : "enabled"}`),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  // Map job name → most recent execution
  const lastExecByJob = new Map<string, CronExecution>();
  for (const e of recentExec?.executions ?? []) {
    const prev = lastExecByJob.get(e.job_name);
    if (!prev || new Date(e.started_at ?? 0) > new Date(prev.started_at ?? 0)) {
      lastExecByJob.set(e.job_name, e);
    }
  }

  const columns: Column<CronJob>[] = [
    {
      header: "Job",
      cell: (row) => (
        <div>
          <div className="font-medium">{row.name}</div>
          {row.input && (
            <div className="text-xs text-muted-foreground line-clamp-1 max-w-xs">{row.input}</div>
          )}
        </div>
      ),
    },
    {
      header: "Schedule",
      cell: (row) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{row.schedule}</code>
      ),
    },
    { header: "Agent", accessorKey: "agent_name" },
    {
      header: "Delivery",
      cell: (row) => {
        const type = (row.delivery?.type ?? "CRON_DELIVERY_TYPE_UNSPECIFIED") as CronDeliveryType;
        const meta = DELIVERY_META[type];
        const Icon = meta.icon;
        return (
          <Badge variant="outline" className="text-xs">
            <Icon className="mr-1 h-3 w-3" /> {meta.label}
          </Badge>
        );
      },
    },
    {
      header: "Last Execution",
      cell: (row) => {
        const exec = lastExecByJob.get(row.name);
        if (!exec) return <span className="text-xs text-muted-foreground">—</span>;
        const ok = exec.status === "CRON_EXECUTION_STATUS_SUCCESS";
        return (
          <div className="flex items-center gap-2">
            <Badge className={ok ? "bg-emerald-500/10 text-emerald-700" : "bg-rose-500/10 text-rose-700"}>
              {ok ? "Success" : "Error"}
            </Badge>
            <span className="text-xs text-muted-foreground">{timeAgo(exec.started_at)}</span>
          </div>
        );
      },
    },
    {
      header: "Enabled",
      cell: (row) => (
        <Switch checked={row.enabled ?? false} onCheckedChange={() => toggleEnabled(row)} />
      ),
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="ghost"
            size="sm"
            disabled={runNow.isPending}
            onClick={() =>
              runNow.mutate(row.name, {
                onSuccess: () => toast.success(`${row.name} executed`),
                onError: (err) => toast.error(err.message),
              })
            }
          >
            <Play className="mr-1 h-3 w-3" /> Run now
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => navigate(`/cron/${encodeURIComponent(row.name)}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate(`/cron/${encodeURIComponent(row.name)}/executions`)}>
                <History className="mr-2 h-4 w-4" /> Executions
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                onClick={() => setDeleteTarget(row.name)}
              >
                <Trash2 className="mr-2 h-4 w-4" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title="Cron Jobs"
        description="Scheduled agent execution. Run jobs immediately with Run now."
        createLabel="Add Schedule"
        createTo="/cron/create"
      />
      <DataTable
        columns={columns}
        data={data?.cron_jobs}
        isLoading={isLoading}
        emptyMessage="No cron jobs yet"
        emptyDescription="Create a schedule to run an agent automatically and deliver the result."
        emptyAction={<Button onClick={() => navigate("/cron/create")}><Play className="mr-2 h-4 w-4" />Add Schedule</Button>}
      />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Cron Job"
        description={`Delete "${deleteTarget}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("Cron job deleted");
                setDeleteTarget(null);
              },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
