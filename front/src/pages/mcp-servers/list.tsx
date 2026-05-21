import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import {
  useDisconnectMCPServerOAuth,
  useDeleteMCPServer,
  useMCPServerOAuthStatus,
  useMCPServers,
  useMCPTools,
  useStartMCPServerOAuth,
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
  KeyRound,
  LogIn,
  Unplug,
} from "lucide-react";
import type { MCPOAuthConnectionState, MCPServer, MCPServerAuthType, MCPServerTransport, MCPTool } from "@/types/api";
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
  const [searchParams, setSearchParams] = useSearchParams();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  useEffect(() => {
    const result = searchParams.get("mcp_oauth");
    if (!result) return;
    if (result === "success") {
      toast.success("OAuth connection completed");
    } else {
      toast.error("OAuth connection failed");
    }
    const next = new URLSearchParams(searchParams);
    next.delete("mcp_oauth");
    next.delete("server_id");
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

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
      header: "Auth",
      cell: (row) => <AuthStatusCell server={row} />,
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
              <OAuthMenuItems server={row} />
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
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight sm:text-2xl">MCP Toolsets</h2>
          <p className="text-sm text-muted-foreground">
            Connected servers exposing tools and resources.
          </p>
        </div>
        <Button className="w-full sm:w-auto" onClick={() => navigate("/mcp-servers/create")}>
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
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
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
                  className="flex flex-wrap items-center gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-1.5 text-xs text-amber-700 dark:text-amber-400"
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

function OAuthMenuItems({ server }: { server: MCPServer }) {
  const startMutation = useStartMCPServerOAuth();
  const disconnectMutation = useDisconnectMCPServerOAuth();
  if (authType(server) !== "MCP_SERVER_AUTH_TYPE_OAUTH2" || !server.id) return null;
  return (
    <>
      <DropdownMenuItem
        disabled={startMutation.isPending}
        onClick={() => {
          startMutation.mutate(
            { serverId: server.id ?? "", returnUrl: `${window.location.origin}/mcp-servers` },
            {
              onSuccess: (data) => {
                window.location.assign(data.authorization_url);
              },
              onError: (err) => toast.error(err.message),
            },
          );
        }}
      >
        <LogIn className="mr-2 h-4 w-4" /> Connect OAuth
      </DropdownMenuItem>
      <DropdownMenuItem
        disabled={disconnectMutation.isPending}
        onClick={() => {
          disconnectMutation.mutate(server.id ?? "", {
            onSuccess: () => toast.success("OAuth connection disconnected"),
            onError: (err) => toast.error(err.message),
          });
        }}
      >
        <Unplug className="mr-2 h-4 w-4" /> Disconnect OAuth
      </DropdownMenuItem>
    </>
  );
}

const OAUTH_PALETTE: Record<MCPOAuthConnectionState, { cls: string; label: string }> = {
  MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED: { cls: "bg-muted text-muted-foreground", label: "OAuth" },
  MCPO_AUTH_CONNECTION_STATE_DISCONNECTED: { cls: "bg-muted text-muted-foreground", label: "Disconnected" },
  MCPO_AUTH_CONNECTION_STATE_CONNECTED: { cls: "bg-emerald-500/10 text-emerald-700", label: "Connected" },
  MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED: { cls: "bg-amber-500/10 text-amber-700", label: "Reconnect" },
  MCPO_AUTH_CONNECTION_STATE_ERROR: { cls: "bg-rose-500/10 text-rose-700", label: "Error" },
};

function AuthStatusCell({ server }: { server: MCPServer }) {
  const type = authType(server);
  const isOAuth = type === "MCP_SERVER_AUTH_TYPE_OAUTH2";
  const { data, isLoading } = useMCPServerOAuthStatus(server.id ?? "", isOAuth);
  if (type === "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS") {
    return <Badge variant="outline" className="text-xs"><KeyRound className="mr-1 h-3 w-3" /> Headers</Badge>;
  }
  if (!isOAuth) {
    return <span className="text-sm text-muted-foreground">None</span>;
  }
  if (isLoading || !data?.status) {
    return <Badge variant="outline" className="text-xs">OAuth</Badge>;
  }
  const state = (data.status.state ?? "MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED") as MCPOAuthConnectionState;
  const palette = OAUTH_PALETTE[state];
  return (
    <Badge className={palette.cls}>
      <KeyRound className="mr-1 h-3 w-3" />
      {palette.label}
    </Badge>
  );
}

function authType(server: MCPServer): MCPServerAuthType {
  if (server.auth?.type && server.auth.type !== "MCP_SERVER_AUTH_TYPE_UNSPECIFIED") {
    return server.auth.type;
  }
  return server.headers && Object.keys(server.headers).length > 0
    ? "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS"
    : "MCP_SERVER_AUTH_TYPE_NONE";
}

function ToolRow({ tool }: { tool: MCPTool }) {
  return (
    <div className="flex flex-col gap-2 rounded-md px-2 py-1.5 hover:bg-muted/40 sm:flex-row sm:items-center sm:gap-3">
      <Wrench className="h-3.5 w-3.5 text-muted-foreground" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm">{tool.name}</span>
          {tool.allowed ? (
            <Badge variant="outline" className="border-emerald-500/30 text-[10px] text-emerald-700">
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
