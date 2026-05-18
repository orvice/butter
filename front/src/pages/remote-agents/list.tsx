import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useDeleteRemoteAgent, useRemoteAgents, useRemoteAgentStatus } from "@/api/remote-agents";
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
import {
  MoreVertical,
  Pencil,
  Trash2,
  Link as LinkIcon,
  Cpu,
  ShieldCheck,
  Plus,
} from "lucide-react";
import type { RemoteAgent, RemoteAgentState } from "@/types/api";

const STATE_LABEL: Record<RemoteAgentState, { cls: string; label: string }> = {
  STATE_UNSPECIFIED: { cls: "bg-muted text-muted-foreground", label: "Unknown" },
  STATE_CONFIGURED: { cls: "bg-muted text-muted-foreground", label: "Configured" },
  STATE_ACTIVE: { cls: "bg-sky-500/10 text-sky-700", label: "Active" },
  STATE_IDLE: { cls: "bg-emerald-500/10 text-emerald-700", label: "Idle" },
  STATE_UNREACHABLE: { cls: "bg-rose-500/10 text-rose-700", label: "Unreachable" },
  STATE_ERROR: { cls: "bg-rose-500/10 text-rose-700", label: "Error" },
};

function StatusBadge({ id }: { id: string }) {
  const { data, isLoading } = useRemoteAgentStatus(id);
  if (isLoading || !data) return <Badge variant="outline" className="text-xs">…</Badge>;
  const state = (data.status.state ?? "STATE_UNSPECIFIED") as RemoteAgentState;
  const p = STATE_LABEL[state];
  return (
    <Badge className={p.cls}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {p.label}
    </Badge>
  );
}

export default function RemoteAgentListPage() {
  const { data, isLoading } = useRemoteAgents();
  const deleteMutation = useDeleteRemoteAgent();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<RemoteAgent>[] = [
    {
      header: "Remote Agent",
      cell: (row) => {
        const isDaemon = row.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON";
        const Icon = isDaemon ? Cpu : LinkIcon;
        return (
          <div className="flex items-center gap-2">
            <Icon className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="font-medium">{row.name}</div>
              <div className="text-xs text-muted-foreground line-clamp-1 max-w-md">
                {isDaemon ? `Cap: ${row.daemon_capability ?? "—"}` : row.url}
              </div>
            </div>
          </div>
        );
      },
    },
    {
      header: "Protocol",
      cell: (row) => (
        <Badge variant="outline" className="font-mono text-[10px]">
          {row.protocol === "REMOTE_AGENT_PROTOCOL_A2A"
            ? "A2A"
            : row.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON"
              ? "Daemon"
              : "Unknown"}
        </Badge>
      ),
    },
    {
      header: "Status",
      cell: (row) => <StatusBadge id={row.id} />,
    },
    {
      header: "Verified",
      cell: (row) =>
        row.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? (
          <ShieldCheck className="h-4 w-4 text-emerald-700" />
        ) : (
          <span className="text-xs text-muted-foreground">—</span>
        ),
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreVertical className="h-4 w-4" />
              </Button>
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
        </div>
      ),
    },
  ];

  return (
    <>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight sm:text-2xl">Remote Agents</h2>
          <p className="text-sm text-muted-foreground">
            External orchestrators and autonomous daemon instances.
          </p>
        </div>
        <Button className="w-full sm:w-auto" onClick={() => navigate("/remote-agents/create")}>
          <Plus className="mr-2 h-4 w-4" /> Register Agent
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={data?.remote_agents}
        isLoading={isLoading}
        emptyMessage="No remote agents registered."
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Remote Agent"
        description="Delete this remote agent? This action cannot be undone."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("Remote agent deleted");
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
