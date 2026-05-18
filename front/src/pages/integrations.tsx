import { useNavigate } from "react-router-dom";
import { useMCPServers, useMCPTools } from "@/api/mcp-servers";
import { useRemoteAgents, useRemoteAgentStatus } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  CheckCircle2,
  Cloud,
  Cpu,
  ExternalLink,
  Filter,
  Globe,
  Plus,
  Radio,
  Server,
  ShieldCheck,
  SlidersHorizontal,
  Wrench,
  XCircle,
} from "lucide-react";
import type { MCPServer, MCPServerTransport, MCPTool, RemoteAgent, RemoteAgentState } from "@/types/api";
import { MCP_TRANSPORT_LABELS } from "@/lib/constants";
import { ServerStatusBadge, ServerStatusInline } from "./mcp-servers/status-cell";

const TRANSPORT_ICON: Record<MCPServerTransport, typeof Server> = {
  MCP_SERVER_TRANSPORT_STDIO: Server,
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: Cloud,
  MCP_SERVER_TRANSPORT_SSE: Radio,
  MCP_SERVER_TRANSPORT_UNSPECIFIED: Server,
};

const REMOTE_STATE: Record<RemoteAgentState, { cls: string; label: string }> = {
  STATE_UNSPECIFIED: { cls: "bg-muted text-muted-foreground", label: "Unknown" },
  STATE_CONFIGURED: { cls: "bg-muted text-muted-foreground", label: "Configured" },
  STATE_ACTIVE: { cls: "bg-sky-500/10 text-sky-700", label: "Active" },
  STATE_IDLE: { cls: "bg-emerald-500/10 text-emerald-700", label: "Idle" },
  STATE_UNREACHABLE: { cls: "bg-rose-500/10 text-rose-700", label: "Unreachable" },
  STATE_ERROR: { cls: "bg-rose-500/10 text-rose-700", label: "Error" },
};

function RemoteStatusBadge({ id }: { id: string }) {
  const { data, isLoading } = useRemoteAgentStatus(id);
  if (isLoading || !data) return <Badge variant="outline">...</Badge>;
  const state = (data.status.state ?? "STATE_UNSPECIFIED") as RemoteAgentState;
  const palette = REMOTE_STATE[state];
  return <Badge className={palette.cls}>{palette.label}</Badge>;
}

function ServerRow({ server }: { server: MCPServer }) {
  const navigate = useNavigate();
  const transport = (server.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED") as MCPServerTransport;
  const Icon = TRANSPORT_ICON[transport];
  return (
    <div className="grid gap-3 border-b px-5 py-4 last:border-b-0 md:grid-cols-[1.7fr_0.8fr_0.8fr_0.5fr_auto] md:items-center">
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border bg-muted">
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
        <div className="min-w-0">
          <div className="truncate font-medium">{server.name}</div>
          <div className="truncate font-mono text-xs text-muted-foreground">
            {server.url || server.command || "No endpoint configured"}
          </div>
        </div>
      </div>
      <Badge variant="outline" className="w-fit font-mono text-[10px] uppercase">
        {MCP_TRANSPORT_LABELS[transport]}
      </Badge>
      <ServerStatusBadge id={server.id ?? ""} />
      <div className="text-sm text-muted-foreground">
        <ServerStatusInline id={server.id ?? ""} />
      </div>
      <Button size="sm" variant="ghost" onClick={() => navigate(`/mcp-servers/${server.id}/edit`)}>
        <SlidersHorizontal className="mr-1 h-3.5 w-3.5" />
        Settings
      </Button>
    </div>
  );
}

function RemoteAgentCard({ agent }: { agent: RemoteAgent }) {
  const navigate = useNavigate();
  const isDaemon = agent.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON";
  const Icon = isDaemon ? Cpu : Globe;
  return (
    <div className="rounded-lg border bg-card p-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border bg-muted">
            <Icon className="h-5 w-5 text-primary" />
          </div>
          <div className="min-w-0">
            <h3 className="truncate font-semibold">{agent.name}</h3>
            <p className="truncate font-mono text-xs text-muted-foreground">
              {isDaemon ? agent.daemon_capability || "daemon bridge" : agent.url}
            </p>
          </div>
        </div>
        <RemoteStatusBadge id={agent.id} />
      </div>
      <div className="flex items-center justify-between border-t pt-3">
        <Badge variant="outline" className="font-mono text-[10px] uppercase">
          {isDaemon ? "Daemon" : "A2A"}
        </Badge>
        <Button size="sm" variant="ghost" onClick={() => navigate(`/remote-agents/${agent.id}/edit`)}>
          Open
          <ExternalLink className="ml-1 h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}

function ToolRow({ tool }: { tool: MCPTool }) {
  return (
    <div className="flex items-start gap-3 rounded-md px-2 py-2 hover:bg-muted/50">
      <Wrench className="mt-0.5 h-4 w-4 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-mono text-sm">{tool.name}</span>
          {tool.allowed ? (
            <Badge variant="outline" className="border-emerald-500/30 text-emerald-700">
              <CheckCircle2 className="mr-1 h-3 w-3" />
              allowed
            </Badge>
          ) : (
            <Badge variant="outline" className="text-muted-foreground">filtered</Badge>
          )}
        </div>
        {tool.description ? (
          <p className="line-clamp-2 text-xs text-muted-foreground">{tool.description}</p>
        ) : null}
      </div>
      <span className="hidden shrink-0 text-xs text-muted-foreground sm:block">
        {tool.server_name ?? tool.server_id}
      </span>
    </div>
  );
}

export default function IntegrationsPage() {
  const navigate = useNavigate();
  const { data: serverData, isLoading: loadingServers } = useMCPServers();
  const { data: remoteData, isLoading: loadingRemote } = useRemoteAgents();
  const { data: toolsData } = useMCPTools();

  const servers = serverData?.mcp_servers ?? [];
  const remoteAgents = remoteData?.remote_agents ?? [];
  const tools = toolsData?.tools ?? [];
  const toolErrors = toolsData?.errors ?? {};

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Integrations & External Tools</h2>
          <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
            Manage MCP servers, configure tool whitelists, and monitor remote agent endpoints across the orchestration network.
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row">
          <Button variant="outline" onClick={() => navigate("/remote-agents/create")}>
            <Plus className="mr-2 h-4 w-4" />
            Register Agent
          </Button>
          <Button onClick={() => navigate("/mcp-servers/create")}>
            <Server className="mr-2 h-4 w-4" />
            Add Server
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <div className="space-y-6 xl:col-span-2">
          <Card>
            <CardHeader className="flex flex-row items-start justify-between border-b pb-4">
              <div>
                <CardTitle>MCP Toolsets</CardTitle>
                <CardDescription>Connected servers exposing tools and resources.</CardDescription>
              </div>
              <Badge variant="outline">{servers.length} servers</Badge>
            </CardHeader>
            <CardContent className="p-0">
              {loadingServers ? (
                <div className="space-y-3 p-5">
                  {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-16" />)}
                </div>
              ) : servers.length === 0 ? (
                <div className="p-8 text-center text-sm text-muted-foreground">No MCP servers configured.</div>
              ) : (
                <div>
                  <div className="hidden grid-cols-[1.7fr_0.8fr_0.8fr_0.5fr_auto] border-b bg-muted/60 px-5 py-3 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground md:grid">
                    <span>Server</span>
                    <span>Transport</span>
                    <span>Status</span>
                    <span>Tools</span>
                    <span className="text-right">Actions</span>
                  </div>
                  {servers.map((server) => <ServerRow key={server.id ?? server.name} server={server} />)}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-start justify-between border-b pb-4">
              <div>
                <CardTitle>Remote Agents</CardTitle>
                <CardDescription>A2A endpoints and daemon bridge targets.</CardDescription>
              </div>
              <Badge variant="outline">{remoteAgents.length} endpoints</Badge>
            </CardHeader>
            <CardContent>
              {loadingRemote ? (
                <div className="grid gap-4 md:grid-cols-2">
                  {Array.from({ length: 2 }).map((_, i) => <Skeleton key={i} className="h-32" />)}
                </div>
              ) : remoteAgents.length === 0 ? (
                <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                  No remote agents registered.
                </div>
              ) : (
                <div className="grid gap-4 md:grid-cols-2">
                  {remoteAgents.map((agent) => <RemoteAgentCard key={agent.id} agent={agent} />)}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <Card className="xl:sticky xl:top-24 xl:self-start">
          <CardHeader className="border-b pb-4">
            <CardTitle className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-primary" />
              Tool Filter
            </CardTitle>
            <CardDescription>Global execution whitelist across connected servers.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 pt-4">
            {Object.entries(toolErrors).map(([id, err]) => (
              <div
                key={id}
                className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-700"
              >
                <XCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                <div>
                  <div className="font-mono">{id}</div>
                  <div>{err}</div>
                </div>
              </div>
            ))}
            {tools.length === 0 ? (
              <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                No tools enumerated yet.
              </div>
            ) : (
              <div className="space-y-1">
                {tools.slice(0, 12).map((tool) => <ToolRow key={`${tool.server_id}-${tool.name}`} tool={tool} />)}
              </div>
            )}
            <div className="border-t pt-4">
              <Button variant="outline" className="w-full" onClick={() => navigate("/mcp-servers")}>
                <ShieldCheck className="mr-2 h-4 w-4" />
                Manage whitelist
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
