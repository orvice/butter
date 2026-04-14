import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCronJobs, useDeleteCronJob, useUpdateCronJob } from "@/api/cron";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2, History } from "lucide-react";
import type { CronJob } from "@/types/api";

export default function CronJobListPage() {
  const { data, isLoading } = useCronJobs();
  const deleteMutation = useDeleteCronJob();
  const updateMutation = useUpdateCronJob();
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

  const columns: Column<CronJob>[] = [
    { header: "Name", accessorKey: "name" },
    { header: "Schedule", cell: (row) => <code className="text-xs">{row.schedule}</code> },
    { header: "Agent", accessorKey: "agent_name" },
    { header: "Timezone", cell: (row) => row.timezone || "UTC" },
    {
      header: "Enabled",
      cell: (row) => <Switch checked={row.enabled ?? false} onCheckedChange={() => toggleEnabled(row)} />,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/cron/${row.name}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate(`/cron/${row.name}/executions`)}>
              <History className="mr-2 h-4 w-4" /> Executions
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Cron Jobs" createLabel="Create Cron Job" createTo="/cron/create" />
      <DataTable columns={columns} data={data?.cron_jobs} isLoading={isLoading} emptyMessage="No cron jobs yet." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Cron Job"
        description={`Delete "${deleteTarget}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Cron job deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
