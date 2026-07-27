import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Bell, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { useDeleteNotifyGroup, useNotifyGroups } from "@/api/notify-groups";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { StatusBadge } from "@/components/butter/primitives";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { enumLabel } from "@/lib/constants";
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
      cell: (row) =>
        row.enabled ? (
          <StatusBadge status="success" label="Enabled" />
        ) : (
          <StatusBadge status="disabled" />
        ),
    },
    {
      header: "Targets",
      cell: (row) => (
        <div className="flex flex-wrap gap-1">
          {(row.targets ?? []).slice(0, 4).map((target, index) => (
            <Badge key={`${target.name ?? index}:${target.type}`} variant="outline" className="text-[10px]">
              {target.name || enumLabel(target.type, "Target")}
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
    <Page>
      <PageHeader
        title="Notify Groups"
        subtitle="Outbound notification targets for cron jobs."
        actions={
          <Button size="sm" render={<Link to="/notify-groups/create" />}>
            <Plus className="size-4" />
            Add Group
          </Button>
        }
      />
      <PageScroll>
        <DataTable
          columns={columns}
          data={groups}
          isLoading={isLoading}
          emptyMessage="No notify groups configured."
        />
      </PageScroll>

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
    </Page>
  );
}
