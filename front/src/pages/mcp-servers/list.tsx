import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useMCPServers, useDeleteMCPServer } from "@/api/mcp-servers";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { MCP_TRANSPORT_LABELS } from "@/lib/constants";
import type { MCPServer } from "@/types/api";

export default function MCPServerListPage() {
  const { data, isLoading } = useMCPServers();
  const deleteMutation = useDeleteMCPServer();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<MCPServer>[] = [
    { header: "ID", accessorKey: "id" },
    { header: "Name", accessorKey: "name" },
    {
      header: "Transport",
      cell: (row) => <Badge variant="secondary">{MCP_TRANSPORT_LABELS[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"]}</Badge>,
    },
    {
      header: "URL / Command",
      cell: (row) => <span className="max-w-xs truncate text-sm text-muted-foreground">{row.url || row.command || "-"}</span>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/mcp-servers/${row.id}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.id ?? null)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="MCP Servers" createLabel="Create MCP Server" createTo="/mcp-servers/create" />
      <DataTable columns={columns} data={data?.mcp_servers} isLoading={isLoading} emptyMessage="No MCP servers yet. Create your first server to get started." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete MCP Server"
        description={`Are you sure you want to delete this MCP server? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("MCP server deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
