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

export default function CronExecutionsPage() {
  const { name } = useParams<{ name: string }>();
  const { data, isLoading } = useCronExecutions(name);

  const columns: Column<CronExecution>[] = [
    { header: "ID", cell: (row) => <span className="max-w-[100px] truncate text-xs">{row.id}</span> },
    { header: "Agent", accessorKey: "agent_name" },
    {
      header: "Status",
      cell: (row) =>
        row.status === "CRON_EXECUTION_STATUS_SUCCESS"
          ? <Badge className="bg-emerald-500/10 text-emerald-700">Success</Badge>
          : <Badge className="bg-rose-500/10 text-rose-700">Error</Badge>,
    },
    { header: "Duration", cell: (row) => formatDuration(row.started_at, row.finished_at) },
    { header: "Started", cell: (row) => row.started_at ? new Date(row.started_at).toLocaleString() : "-" },
    {
      header: "Output",
      cell: (row) => <span className="block max-w-xs truncate text-xs text-muted-foreground">{row.output ?? "-"}</span>,
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
