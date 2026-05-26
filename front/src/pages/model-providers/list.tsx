import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { Bot, MoreHorizontal, Pencil, Plus, Trash2, Cpu } from "lucide-react";
import { useDeleteModelProvider, useModelProviders } from "@/api/model-providers";
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
import type { ModelProvider } from "@/types/api";

export default function ModelProviderListPage() {
  const { data, isLoading } = useModelProviders();
  const deleteMutation = useDeleteModelProvider();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const providers = data?.model_providers ?? [];

  const columns: Column<ModelProvider>[] = [
    {
      header: "Provider",
      cell: (row) => (
        <div className="flex items-center gap-2">
          <Bot className="h-4 w-4 text-muted-foreground" />
          <div>
            <div className="font-medium">{row.name}</div>
            <div className="text-xs text-muted-foreground line-clamp-1 max-w-xs">
              {row.base_url || "Default provider endpoint"}
            </div>
          </div>
        </div>
      ),
    },
    {
      header: "Type",
      cell: (row) => <Badge variant="outline" className="font-mono text-[10px]">{row.type || "-"}</Badge>,
    },
    {
      header: "Models",
      cell: (row) => (
        <div className="flex flex-wrap gap-1">
          {(row.models ?? []).slice(0, 4).map((m) => (
            <Badge key={`${m.name}:${m.alias}`} variant="secondary" className="text-[10px]">
              <Cpu className="mr-1 h-3 w-3" /> {m.alias || m.name}
            </Badge>
          ))}
          {(row.models?.length ?? 0) > 4 && (
            <Badge variant="outline" className="text-[10px]">+{(row.models?.length ?? 0) - 4}</Badge>
          )}
          {(row.models?.length ?? 0) === 0 && <span className="text-sm text-muted-foreground">-</span>}
        </div>
      ),
    },
    {
      header: "API Key",
      cell: (row) => row.api_key ? <Badge variant="outline" className="text-[10px]">Configured</Badge> : <span className="text-sm text-muted-foreground">-</span>,
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
              <DropdownMenuItem onClick={() => navigate(`/model-providers/${encodeURIComponent(row.name)}/edit`)}>
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
          <h2 className="text-xl font-bold tracking-tight sm:text-2xl">Model Providers</h2>
          <p className="text-sm text-muted-foreground">
            DB-backed LLM provider configuration used by agents and channels.
          </p>
        </div>
        <Button className="w-full sm:w-auto" onClick={() => navigate("/model-providers/create")}>
          <Plus className="mr-2 h-4 w-4" /> Add Provider
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={providers}
        isLoading={isLoading}
        emptyMessage="No model providers configured. Add one to make LLM models available."
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Model Provider"
        description="Are you sure you want to delete this model provider? Agents or channels using its aliases may stop working."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("Model provider deleted");
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
