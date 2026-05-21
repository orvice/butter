import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { MCPOAuthConnectionStatus, MCPServer, MCPServerStatus, MCPTool } from "@/types/api";

const SVC = "agents.v1.MCPServerService";

function listMCPServers() {
  return twirpFetch<object, { mcp_servers: MCPServer[] }>(SVC, "ListMCPServers", {});
}

function getMCPServer(id: string) {
  return twirpFetch<{ id: string }, { mcp_server: MCPServer }>(SVC, "GetMCPServer", { id });
}

function createMCPServer(mcp_server: MCPServer) {
  return twirpFetch<{ mcp_server: MCPServer }, { mcp_server: MCPServer }>(SVC, "CreateMCPServer", { mcp_server });
}

function updateMCPServer(mcp_server: MCPServer) {
  return twirpFetch<{ mcp_server: MCPServer }, { mcp_server: MCPServer }>(SVC, "UpdateMCPServer", { mcp_server });
}

function deleteMCPServer(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteMCPServer", { id });
}

function getMCPServerOAuthStatus(serverId: string) {
  return twirpFetch<{ server_id: string }, { status: MCPOAuthConnectionStatus }>(
    SVC,
    "GetMCPServerOAuthStatus",
    { server_id: serverId },
  );
}

function startMCPServerOAuth(serverId: string, returnUrl?: string) {
  return twirpFetch<
    { server_id: string; return_url?: string },
    { authorization_url: string; flow_id: string }
  >(SVC, "StartMCPServerOAuth", { server_id: serverId, return_url: returnUrl });
}

function disconnectMCPServerOAuth(serverId: string) {
  return twirpFetch<{ server_id: string }, { status: MCPOAuthConnectionStatus }>(
    SVC,
    "DisconnectMCPServerOAuth",
    { server_id: serverId },
  );
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

function getMCPServerStatus(id: string) {
  return twirpFetch<{ id: string }, { status: MCPServerStatus }>(
    SVC,
    "GetMCPServerStatus",
    { id },
  );
}

function listMCPTools(serverId?: string) {
  return twirpFetch<
    { server_id?: string },
    { tools?: MCPTool[]; errors?: Record<string, string> }
  >(SVC, "ListMCPTools", { server_id: serverId });
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
