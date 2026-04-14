import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useSessions, useDeleteSession } from "@/api/sessions";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Eye, Trash2 } from "lucide-react";
import type { SessionInfo } from "@/types/api";

export default function SessionListPage() {
  const [appName, setAppName] = useState("");
  const [userId, setUserId] = useState("");
  const { data, isLoading } = useSessions(appName, userId);
  const deleteMutation = useDeleteSession();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null);

  const columns: Column<SessionInfo>[] = [
    { header: "Session ID", accessorKey: "session_id" },
    { header: "App Name", accessorKey: "app_name" },
    { header: "User ID", accessorKey: "user_id" },
    {
      header: "Last Update",
      cell: (row) => row.last_update_time ? new Date(row.last_update_time).toLocaleString() : "-",
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/sessions/detail?app=${row.app_name}&user=${row.user_id}&session=${row.session_id}`)}>
              <Eye className="mr-2 h-4 w-4" /> View
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Sessions" />
      <div className="mb-4 flex gap-3">
        <Input placeholder="App Name" value={appName} onChange={(e) => setAppName(e.target.value)} className="max-w-xs" />
        <Input placeholder="User ID" value={userId} onChange={(e) => setUserId(e.target.value)} className="max-w-xs" />
      </div>
      {appName && userId ? (
        <DataTable columns={columns} data={data?.sessions} isLoading={isLoading} emptyMessage="No sessions found." />
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
          <p className="text-muted-foreground">Enter an App Name and User ID to search sessions.</p>
        </div>
      )}
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Session"
        description={`Delete session "${deleteTarget?.session_id}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(
              { app_name: deleteTarget.app_name, user_id: deleteTarget.user_id, session_id: deleteTarget.session_id },
              {
                onSuccess: () => { toast.success("Session deleted"); setDeleteTarget(null); },
                onError: (err) => toast.error(err.message),
              },
            );
          }
        }}
      />
    </>
  );
}
