import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fromJson, toJson, type JsonValue } from "@bufbuild/protobuf";
import { MCPServerSchema } from "@/gen/agents/v1/agent_pb";
import { GlobalMCPServerService } from "@/gen/agents/v1/agent_service_pb";
import type { MCPServer } from "@/types/api";
import { makeClient } from "./transport";

const client = makeClient(GlobalMCPServerService);
const QUERY_KEY = ["global-mcp-servers"];

// MCPServer has a nested auth.oauth2 type and an MCPServerTransport enum;
// we already wrote the explicit toProto/fromProto mapping in api/mcp-servers.ts.
// Reusing it here would force a cross-file import and re-export. Since
// protojson accepts both snake_case and camelCase on input, and emits proto
// field names with useProtoFieldName=true, we round-trip through JSON
// instead — the same trick api/agents.ts uses for the deeply-nested Agent.
function toProto(server: MCPServer) {
  return fromJson(MCPServerSchema, server as unknown as JsonValue, { ignoreUnknownFields: true });
}

function fromProto(server: unknown): MCPServer {
  return toJson(MCPServerSchema, server as never, { useProtoFieldName: true }) as unknown as MCPServer;
}

async function listGlobalMCPServers(): Promise<{ mcp_servers: MCPServer[] }> {
  const res = await client.listGlobalMCPServers({});
  return { mcp_servers: res.mcpServers.map(fromProto) };
}

async function createGlobalMCPServer(server: MCPServer): Promise<{ mcp_server: MCPServer }> {
  const res = await client.createGlobalMCPServer({ mcpServer: toProto(server) });
  if (!res.mcpServer) throw new Error("create returned nothing");
  return { mcp_server: fromProto(res.mcpServer) };
}

async function updateGlobalMCPServer(server: MCPServer): Promise<{ mcp_server: MCPServer }> {
  const res = await client.updateGlobalMCPServer({ mcpServer: toProto(server) });
  if (!res.mcpServer) throw new Error("update returned nothing");
  return { mcp_server: fromProto(res.mcpServer) };
}

async function deleteGlobalMCPServer(id: string): Promise<void> {
  await client.deleteGlobalMCPServer({ id });
}

async function installGlobalMCPServer({
  id,
  workspaceId,
}: {
  id: string;
  workspaceId?: string;
}): Promise<{ mcp_server: MCPServer }> {
  const res = await client.installGlobalMCPServer({ id, workspaceId: workspaceId ?? "" });
  if (!res.mcpServer) throw new Error("install returned nothing");
  return { mcp_server: fromProto(res.mcpServer) };
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
