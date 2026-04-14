import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useRemoteAgents, useDeleteRemoteAgent } from "@/api/remote-agents";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import type { RemoteAgent } from "@/types/api";

export default function RemoteAgentListPage() {
  const { data, isLoading } = useRemoteAgents();
  const deleteMutation = useDeleteRemoteAgent();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<RemoteAgent>[] = [
    { header: "ID", accessorKey: "id" },
    { header: "Name", accessorKey: "name" },
    { header: "URL", accessorKey: "url" },
    {
      header: "Protocol",
      cell: (row) => <Badge variant="secondary">{row.protocol === "REMOTE_AGENT_PROTOCOL_A2A" ? "A2A" : "Unknown"}</Badge>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/remote-agents/${row.id}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.id)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Remote Agents" createLabel="Create Remote Agent" createTo="/remote-agents/create" />
      <DataTable columns={columns} data={data?.remote_agents} isLoading={isLoading} emptyMessage="No remote agents yet." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Remote Agent"
        description="Are you sure? This action cannot be undone."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Remote agent deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
