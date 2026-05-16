import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useDeleteMCPServer,
  useMCPServers,
  useMCPTools,
} from "@/api/mcp-servers";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Server,
  Cloud,
  Radio,
  Settings,
  MoreHorizontal,
  Pencil,
  Trash2,
  CheckCircle2,
  XCircle,
  Wrench,
  Filter,
} from "lucide-react";
import type { MCPServer, MCPServerTransport, MCPTool } from "@/types/api";
import { MCP_TRANSPORT_LABELS } from "@/lib/constants";
import { ServerStatusBadge, ServerStatusInline } from "./status-cell";

const TRANSPORT_ICON: Record<MCPServerTransport, typeof Server> = {
  MCP_SERVER_TRANSPORT_STDIO: Server,
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: Cloud,
  MCP_SERVER_TRANSPORT_SSE: Radio,
  MCP_SERVER_TRANSPORT_UNSPECIFIED: Server,
};

export default function MCPServerListPage() {
  const { data, isLoading } = useMCPServers();
  const { data: toolsData } = useMCPTools();
  const deleteMutation = useDeleteMCPServer();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const servers = data?.mcp_servers ?? [];
  const tools = toolsData?.tools ?? [];
  const toolErrors = toolsData?.errors ?? {};

  const columns: Column<MCPServer>[] = [
    {
      header: "Server",
      cell: (row) => {
        const Icon = TRANSPORT_ICON[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"];
        return (
          <div className="flex items-center gap-2">
            <Icon className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="font-medium">{row.name}</div>
              <div className="text-xs text-muted-foreground line-clamp-1 max-w-xs">
                {row.url || row.command || "—"}
              </div>
            </div>
          </div>
        );
      },
    },
    {
      header: "Transport",
      cell: (row) => (
        <Badge variant="outline" className="font-mono text-[10px]">
          {MCP_TRANSPORT_LABELS[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"]}
        </Badge>
      ),
    },
    {
      header: "Status",
      cell: (row) => <ServerStatusBadge id={row.id ?? ""} />,
    },
    {
      header: "Tools",
      cell: (row) => <ServerStatusInline id={row.id ?? ""} />,
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => navigate(`/mcp-servers/${row.id}/edit`)}>
                <Settings className="mr-2 h-4 w-4" /> Settings
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate(`/mcp-servers/${row.id}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                onClick={() => setDeleteTarget(row.id ?? null)}
              >
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
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">MCP Toolsets</h2>
          <p className="text-sm text-muted-foreground">
            Connected servers exposing tools and resources.
          </p>
        </div>
        <Button onClick={() => navigate("/mcp-servers/create")}>
          <Server className="mr-2 h-4 w-4" /> Add Server
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={servers}
        isLoading={isLoading}
        emptyMessage="No MCP servers configured. Add a server to expose its tools to agents."
      />

      {/* Tool filter / aggregated view */}
      <Card className="mt-6">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Filter className="h-4 w-4" /> Tool Filter
            </CardTitle>
            <CardDescription>Global execution whitelist across all connected servers.</CardDescription>
          </div>
          <Badge variant="outline" className="text-xs">
            {tools.length} tools
          </Badge>
        </CardHeader>
        <CardContent className="space-y-1">
          {Object.entries(toolErrors).length > 0 && (
            <div className="mb-3 space-y-1">
              {Object.entries(toolErrors).map(([id, err]) => (
                <div
                  key={id}
                  className="flex items-center gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-1.5 text-xs text-amber-700 dark:text-amber-400"
                >
                  <XCircle className="h-3.5 w-3.5" />
                  <span className="font-mono">{id}</span>
                  <span>—</span>
                  <span>{err}</span>
                </div>
              ))}
            </div>
          )}
          {tools.length === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              No tools enumerated yet. Probing happens when MCP servers report status.
            </p>
          ) : (
            tools.map((t) => <ToolRow key={`${t.server_id}-${t.name}`} tool={t} />)
          )}
        </CardContent>
      </Card>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete MCP Server"
        description="Are you sure you want to delete this MCP server? This action cannot be undone."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success("MCP server deleted");
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

function ToolRow({ tool }: { tool: MCPTool }) {
  return (
    <div className="flex items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted/40">
      <Wrench className="h-3.5 w-3.5 text-muted-foreground" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm">{tool.name}</span>
          {tool.allowed ? (
            <Badge variant="outline" className="text-[10px] border-green-500/30 text-green-600">
              <CheckCircle2 className="mr-1 h-3 w-3" /> allowed
            </Badge>
          ) : (
            <Badge variant="outline" className="text-[10px] border-muted text-muted-foreground">
              filtered
            </Badge>
          )}
        </div>
        {tool.description && (
          <div className="text-xs text-muted-foreground line-clamp-1">{tool.description}</div>
        )}
      </div>
      <span className="text-xs text-muted-foreground">from: {tool.server_name ?? tool.server_id}</span>
    </div>
  );
}
