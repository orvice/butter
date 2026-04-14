import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { MCPServer } from "@/types/api";

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
