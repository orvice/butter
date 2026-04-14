import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useAgents, useDeleteAgent } from "@/api/agents";
import { PageHeader } from "@/components/page-header";
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
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import type { Agent } from "@/types/api";

export default function AgentListPage() {
  const { data, isLoading } = useAgents();
  const deleteMutation = useDeleteAgent();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<Agent>[] = [
    { header: "Name", accessorKey: "name" },
    {
      header: "Type",
      cell: (row) => <Badge variant="secondary">{AGENT_TYPE_LABELS[row.type ?? "AGENT_TYPE_UNSPECIFIED"]}</Badge>,
    },
    { header: "Description", accessorKey: "description" },
    {
      header: "A2A",
      cell: (row) => row.enable_a2a ? <Badge>Enabled</Badge> : <Badge variant="outline">Disabled</Badge>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/agents/${row.name}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
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
      <PageHeader title="Agents" createLabel="Create Agent" createTo="/agents/create" />
      <DataTable columns={columns} data={data?.agents} isLoading={isLoading} emptyMessage="No agents yet. Create your first agent to get started." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Agent"
        description={`Are you sure you want to delete "${deleteTarget}"? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Agent deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
