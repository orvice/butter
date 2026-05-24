import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "./client";
import type { MCPServer } from "@/types/api";

const QUERY_KEY = ["global-mcp-servers"];

function listGlobalMCPServers() {
  return apiFetch<{ mcp_servers: MCPServer[] }>("/api/global-mcp-servers");
}

function createGlobalMCPServer(server: MCPServer) {
  return apiFetch<{ mcp_server: MCPServer }>("/api/admin/global-mcp-servers", {
    method: "POST",
    body: JSON.stringify(server),
  });
}

function updateGlobalMCPServer(server: MCPServer) {
  return apiFetch<{ mcp_server: MCPServer }>(`/api/admin/global-mcp-servers/${encodeURIComponent(server.id ?? "")}`, {
    method: "PUT",
    body: JSON.stringify(server),
  });
}

function deleteGlobalMCPServer(id: string) {
  return apiFetch<void>(`/api/admin/global-mcp-servers/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

function installGlobalMCPServer({ id, workspaceId }: { id: string; workspaceId?: string }) {
  return apiFetch<{ mcp_server: MCPServer }>(`/api/global-mcp-servers/${encodeURIComponent(id)}/install`, {
    method: "POST",
    body: JSON.stringify(workspaceId ? { workspace_id: workspaceId } : {}),
  });
}

export function useGlobalMCPServers() {
  return useQuery({ queryKey: QUERY_KEY, queryFn: listGlobalMCPServers });
}

export function useCreateGlobalMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createGlobalMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: QUERY_KEY }),
  });
}

export function useUpdateGlobalMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateGlobalMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: QUERY_KEY }),
  });
}

export function useDeleteGlobalMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteGlobalMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: QUERY_KEY }),
  });
}

export function useInstallGlobalMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: installGlobalMCPServer,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["mcp-servers"] });
      qc.invalidateQueries({ queryKey: ["mcp-tools"] });
    },
  });
}
