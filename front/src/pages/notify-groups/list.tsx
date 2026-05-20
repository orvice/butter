import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Bell, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { useDeleteNotifyGroup, useNotifyGroups } from "@/api/notify-groups";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { NotifyGroup } from "@/types/api";

export default function NotifyGroupListPage() {
  const { data, isLoading } = useNotifyGroups();
  const deleteMutation = useDeleteNotifyGroup();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const groups = data?.notify_groups ?? [];

  const columns: Column<NotifyGroup>[] = [
    {
      header: "Group",
      cell: (row) => (
        <div className="flex items-center gap-2">
          <Bell className="h-4 w-4 text-muted-foreground" />
          <div>
            <div className="font-medium">{row.name}</div>
            <div className="text-xs text-muted-foreground">{row.targets?.length ?? 0} targets</div>
          </div>
        </div>
      ),
    },
    {
      header: "Status",
      cell: (row) => row.enabled ? <Badge>Enabled</Badge> : <Badge variant="secondary">Disabled</Badge>,
    },
    {
      header: "Targets",
      cell: (row) => (
        <div className="flex flex-wrap gap-1">
          {(row.targets ?? []).slice(0, 4).map((target, index) => (
            <Badge key={`${target.name ?? index}:${target.type}`} variant="outline" className="text-[10px]">
              {target.name || target.type || "target"}
            </Badge>
          ))}
          {(row.targets?.length ?? 0) > 4 && (
            <Badge variant="outline" className="text-[10px]">+{(row.targets?.length ?? 0) - 4}</Badge>
          )}
        </div>
      ),
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => navigate(`/notify-groups/${encodeURIComponent(row.name)}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
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
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight sm:text-2xl">Notify Groups</h2>
          <p className="text-sm text-muted-foreground">Outbound notification targets for cron jobs.</p>
        </div>
        <Button className="w-full sm:w-auto" onClick={() => navigate("/notify-groups/create")}>
          <Plus className="mr-2 h-4 w-4" /> Add Group
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={groups}
        isLoading={isLoading}
        emptyMessage="No notify groups configured."
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Notify Group"
        description="Cron jobs using this group will no longer be able to send notify-group delivery."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("Notify group deleted");
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
