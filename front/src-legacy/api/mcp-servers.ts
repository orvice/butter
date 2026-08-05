import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  MCPServerAuthSchema,
  MCPServerAuthType,
  MCPServerOAuth2ConfigSchema,
  MCPServerSchema,
  MCPServerTransport,
  type MCPServer as PbMCPServer,
  type MCPServerAuth as PbMCPServerAuth,
} from "@/gen/agents/v1/agent_pb";
import {
  MCPOAuthConnectionState,
  MCPServerService,
  type MCPOAuthConnectionStatus as PbMCPOAuthStatus,
} from "@/gen/agents/v1/agent_service_pb";
import {
  MCPServerStatus_State,
  type MCPServerStatus as PbMCPServerStatus,
  type MCPTool as PbMCPTool,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  MCPOAuthConnectionState as LegacyOAuthState,
  MCPOAuthConnectionStatus,
  MCPServer,
  MCPServerAuthType as LegacyAuthType,
  MCPServerState,
  MCPServerStatus,
  MCPServerTransport as LegacyTransport,
  MCPTool,
} from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(MCPServerService);

// Top-level proto enums keep their proto-style names as TS enum keys, so the
// generated `MCPServerTransport[2]` reverse-lookup already returns
// "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP". Nested enums (XYZ_State) strip the
// proto prefix, so those still need an explicit array.

function transportFromProto(t: MCPServerTransport): LegacyTransport {
  return (MCPServerTransport[t] as LegacyTransport) ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED";
}

function transportToProto(t: LegacyTransport | undefined): MCPServerTransport {
  if (!t) return MCPServerTransport.MCP_SERVER_TRANSPORT_UNSPECIFIED;
  const v = (MCPServerTransport as unknown as Record<string, number>)[t];
  return typeof v === "number" ? (v as MCPServerTransport) : MCPServerTransport.MCP_SERVER_TRANSPORT_UNSPECIFIED;
}

function authTypeFromProto(t: MCPServerAuthType): LegacyAuthType {
  return (MCPServerAuthType[t] as LegacyAuthType) ?? "MCP_SERVER_AUTH_TYPE_UNSPECIFIED";
}

function authTypeToProto(t: LegacyAuthType | undefined): MCPServerAuthType {
  if (!t) return MCPServerAuthType.MCP_SERVER_AUTH_TYPE_UNSPECIFIED;
  const v = (MCPServerAuthType as unknown as Record<string, number>)[t];
  return typeof v === "number" ? (v as MCPServerAuthType) : MCPServerAuthType.MCP_SERVER_AUTH_TYPE_UNSPECIFIED;
}

function oauthStateFromProto(s: MCPOAuthConnectionState): LegacyOAuthState {
  return (MCPOAuthConnectionState[s] as LegacyOAuthState) ?? "MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED";
}

const SERVER_STATE_NAMES: MCPServerState[] = [
  "STATE_UNSPECIFIED",
  "STATE_CONFIGURED",
  "STATE_CONNECTED",
  "STATE_DISCONNECTED",
  "STATE_ERROR",
];

function serverStateFromProto(s: MCPServerStatus_State): MCPServerState {
  return SERVER_STATE_NAMES[s] ?? "STATE_UNSPECIFIED";
}

function authFromProto(a: PbMCPServerAuth | undefined): MCPServer["auth"] {
  if (!a) return undefined;
  return {
    type: authTypeFromProto(a.type),
    oauth2: a.oauth2
      ? {
          client_id: a.oauth2.clientId,
          client_secret: a.oauth2.clientSecret,
          scopes: a.oauth2.scopes,
          authorization_url: a.oauth2.authorizationUrl,
          token_url: a.oauth2.tokenUrl,
          resource_metadata_url: a.oauth2.resourceMetadataUrl,
          authorization_server_url: a.oauth2.authorizationServerUrl,
          resource: a.oauth2.resource,
        }
      : undefined,
  };
}

function authToProto(a: MCPServer["auth"]): PbMCPServerAuth | undefined {
  if (!a) return undefined;
  return create(MCPServerAuthSchema, {
    type: authTypeToProto(a.type),
    oauth2: a.oauth2
      ? create(MCPServerOAuth2ConfigSchema, {
          clientId: a.oauth2.client_id ?? "",
          clientSecret: a.oauth2.client_secret ?? "",
          scopes: a.oauth2.scopes ?? [],
          authorizationUrl: a.oauth2.authorization_url ?? "",
          tokenUrl: a.oauth2.token_url ?? "",
          resourceMetadataUrl: a.oauth2.resource_metadata_url ?? "",
          authorizationServerUrl: a.oauth2.authorization_server_url ?? "",
          resource: a.oauth2.resource ?? "",
        })
      : undefined,
  });
}

function serverFromProto(s: PbMCPServer): MCPServer {
  return {
    id: s.id,
    name: s.name,
    transport: transportFromProto(s.transport),
    url: s.url,
    headers: s.headers,
    tool_filter: s.toolFilter,
    metadata: s.metadata,
    timeout_seconds: s.timeoutSeconds,
    auth: authFromProto(s.auth),
  };
}

function serverToProto(s: MCPServer): PbMCPServer {
  return create(MCPServerSchema, {
    id: s.id ?? "",
    name: s.name,
    transport: transportToProto(s.transport),
    url: s.url ?? "",
    headers: s.headers ?? {},
    toolFilter: s.tool_filter ?? [],
    metadata: s.metadata ?? {},
    timeoutSeconds: s.timeout_seconds ?? 0,
    auth: authToProto(s.auth),
  });
}

function oauthStatusFromProto(s: PbMCPOAuthStatus): MCPOAuthConnectionStatus {
  return {
    server_id: s.serverId,
    state: oauthStateFromProto(s.state),
    detail: s.detail,
    scopes: s.scopes,
    connected_at: tsToISO(s.connectedAt),
    expires_at: tsToISO(s.expiresAt),
    checked_at: tsToISO(s.checkedAt),
  };
}

function serverStatusFromProto(s: PbMCPServerStatus): MCPServerStatus {
  return {
    id: s.id,
    name: s.name,
    state: serverStateFromProto(s.state),
    tool_count: s.toolCount,
    detail: s.detail,
    checked_at: tsToISO(s.checkedAt),
  };
}

function toolFromProto(t: PbMCPTool): MCPTool {
  return {
    name: t.name,
    description: t.description,
    server_id: t.serverId,
    server_name: t.serverName,
    allowed: t.allowed,
  };
}

async function listMCPServers(): Promise<{ mcp_servers: MCPServer[] }> {
  const res = await client.listMCPServers({});
  return { mcp_servers: res.mcpServers.map(serverFromProto) };
}

async function getMCPServer(id: string): Promise<{ mcp_server: MCPServer }> {
  const res = await client.getMCPServer({ id });
  if (!res.mcpServer) throw new Error("not found");
  return { mcp_server: serverFromProto(res.mcpServer) };
}

async function createMCPServer(server: MCPServer): Promise<{ mcp_server: MCPServer }> {
  const res = await client.createMCPServer({ mcpServer: serverToProto(server) });
  if (!res.mcpServer) throw new Error("create returned nothing");
  return { mcp_server: serverFromProto(res.mcpServer) };
}

async function updateMCPServer(server: MCPServer): Promise<{ mcp_server: MCPServer }> {
  const res = await client.updateMCPServer({ mcpServer: serverToProto(server) });
  if (!res.mcpServer) throw new Error("update returned nothing");
  return { mcp_server: serverFromProto(res.mcpServer) };
}

async function deleteMCPServer(id: string): Promise<void> {
  await client.deleteMCPServer({ id });
}

async function getMCPServerOAuthStatus(serverId: string): Promise<{ status: MCPOAuthConnectionStatus }> {
  const res = await client.getMCPServerOAuthStatus({ serverId });
  if (!res.status) throw new Error("status not found");
  return { status: oauthStatusFromProto(res.status) };
}

async function startMCPServerOAuth(
  serverId: string,
  returnUrl?: string,
): Promise<{ authorization_url: string; flow_id: string }> {
  const res = await client.startMCPServerOAuth({ serverId, returnUrl: returnUrl ?? "" });
  return { authorization_url: res.authorizationUrl, flow_id: res.flowId };
}

async function disconnectMCPServerOAuth(serverId: string): Promise<{ status: MCPOAuthConnectionStatus }> {
  const res = await client.disconnectMCPServerOAuth({ serverId });
  if (!res.status) throw new Error("status not found");
  return { status: oauthStatusFromProto(res.status) };
}

async function getMCPServerStatus(id: string): Promise<{ status: MCPServerStatus }> {
  const res = await client.getMCPServerStatus({ id });
  if (!res.status) throw new Error("status not found");
  return { status: serverStatusFromProto(res.status) };
}

async function listMCPTools(
  serverId?: string,
): Promise<{ tools?: MCPTool[]; errors?: Record<string, string> }> {
  const res = await client.listMCPTools({ serverId: serverId ?? "" });
  return { tools: res.tools.map(toolFromProto), errors: res.errors };
}

export function useMCPServers() {
  return useQuery({ queryKey: ["mcp-servers"], queryFn: listMCPServers });
}

export function useMCPServer(id: string) {
  return useQuery({ queryKey: ["mcp-servers", id], queryFn: () => getMCPServer(id), enabled: !!id });
}

export function useCreateMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["mcp-servers"] }),
  });
}

export function useUpdateMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateMCPServer,
    onSuccess: (_data, server) => {
      qc.invalidateQueries({ queryKey: ["mcp-servers"] });
      qc.invalidateQueries({ queryKey: ["mcp-servers", server.id] });
    },
  });
}

export function useDeleteMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["mcp-servers"] }),
  });
}

export function useMCPServerOAuthStatus(serverId: string, enabled = true) {
  return useQuery({
    queryKey: ["mcp-servers", serverId, "oauth"],
    queryFn: () => getMCPServerOAuthStatus(serverId),
    enabled: enabled && !!serverId,
    refetchInterval: 30_000,
  });
}

export function useStartMCPServerOAuth() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ serverId, returnUrl }: { serverId: string; returnUrl?: string }) =>
      startMCPServerOAuth(serverId, returnUrl),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["mcp-servers", vars.serverId, "oauth"] });
    },
  });
}

export function useDisconnectMCPServerOAuth() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: disconnectMCPServerOAuth,
    onSuccess: (_data, serverId) => {
      qc.invalidateQueries({ queryKey: ["mcp-servers", serverId, "oauth"] });
      qc.invalidateQueries({ queryKey: ["mcp-tools"] });
    },
  });
}

export function useMCPServerStatus(id: string) {
  return useQuery({
    queryKey: ["mcp-servers", id, "status"],
    queryFn: () => getMCPServerStatus(id),
    enabled: !!id,
    refetchInterval: 30_000,
  });
}

export function useMCPTools(serverId?: string) {
  return useQuery({
    queryKey: ["mcp-tools", serverId ?? "all"],
    queryFn: () => listMCPTools(serverId),
    refetchInterval: 60_000,
  });
}
