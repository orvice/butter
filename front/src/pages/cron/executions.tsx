import { useParams } from "react-router-dom";
import { useCronExecutions } from "@/api/cron";
import { DataTable, type Column } from "@/components/data-table";
import { Badge } from "@/components/ui/badge";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import type { CronExecution } from "@/types/api";

function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "-";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  return `${(ms / 1000).toFixed(1)}s`;
}

function statusBadge(row: CronExecution) {
  const cls =
    row.status === "CRON_EXECUTION_STATUS_SUCCESS"
      ? "bg-emerald-500/10 text-emerald-700"
      : row.status === "CRON_EXECUTION_STATUS_ERROR" || row.status === "CRON_EXECUTION_STATUS_CANCELLED"
        ? "bg-rose-500/10 text-rose-700"
        : "bg-amber-500/10 text-amber-700";
  return <Badge className={cls}>{row.status.replace("CRON_EXECUTION_STATUS_", "")}</Badge>;
}

export default function CronExecutionsPage() {
  const { name } = useParams<{ name: string }>();
  const { data, isLoading } = useCronExecutions(name);

  const columns: Column<CronExecution>[] = [
    { header: "ID", cell: (row) => <span className="max-w-[100px] truncate text-xs">{row.id}</span> },
    { header: "Agent", accessorKey: "agent_name" },
    {
      header: "Status",
      cell: statusBadge,
    },
    { header: "Duration", cell: (row) => row.duration_ms != null ? `${row.duration_ms}ms` : formatDuration(row.started_at, row.finished_at) },
    { header: "Attempts", cell: (row) => row.attempt_count ?? "-" },
    { header: "Trigger", cell: (row) => row.trigger_type?.replace("CRON_EXECUTION_TRIGGER_TYPE_", "") ?? "-" },
    { header: "Started", cell: (row) => row.started_at ? new Date(row.started_at).toLocaleString() : "-" },
    {
      header: "Output",
      cell: (row) => (
        <span className="block max-w-xs truncate text-xs text-muted-foreground">
          {row.output || row.error || row.skipped_reason || "-"}
          {row.truncated ? " [truncated]" : ""}
        </span>
      ),
    },
  ];

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Executions</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Executions: {name}</h2>
      <DataTable columns={columns} data={data?.executions} isLoading={isLoading} emptyMessage="No executions yet." />
    </>
  );
}
