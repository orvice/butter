import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fromJson, toJson, type JsonValue } from "@bufbuild/protobuf";
import { AgentSchema } from "@/gen/agents/v1/agent_pb";
import {
  AgentRuntimeStatusSchema,
  AgentService,
  InvocationSchema,
} from "@/gen/agents/v1/agent_service_pb";
import type { Agent, AgentRuntimeStatus, Invocation } from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(AgentService);

// Agent / AgentConfig is deeply nested (sub_agents, mcp_servers, file_mounts,
// context_guard, ...). Rather than hand-rolling a 200-line toProto/fromProto,
// we leverage protojson: the proto-es runtime's fromJson accepts both
// snake_case and camelCase keys, and toJson with useProtoFieldName=true emits
// snake_case identical to the legacy Twirp wire format. So the legacy
// snake-cased Agent interface round-trips through the typed RPC call without
// extra mapping code.
function agentToProto(a: Agent) {
  return fromJson(AgentSchema, a as unknown as JsonValue, { ignoreUnknownFields: true });
}

function agentFromProto(a: unknown): Agent {
  return toJson(AgentSchema, a as never, { useProtoFieldName: true }) as unknown as Agent;
}

interface ListAgentsParams {
  page_size?: number;
  page_token?: string;
}

interface ListAgentsResponse {
  agents: Agent[];
  next_page_token?: string;
  total?: number;
}

async function listAgents(params: ListAgentsParams = {}): Promise<ListAgentsResponse> {
  const res = await client.listAgents({
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return {
    agents: res.agents.map(agentFromProto),
    next_page_token: res.nextPageToken,
    total: res.total,
  };
}

async function getAgent(name: string): Promise<{ agent: Agent }> {
  const res = await client.getAgent({ name });
  if (!res.agent) throw new Error("not found");
  return { agent: agentFromProto(res.agent) };
}

async function createAgent(agent: Agent): Promise<{ agent: Agent }> {
  const res = await client.createAgent({ agent: agentToProto(agent) });
  if (!res.agent) throw new Error("create returned nothing");
  return { agent: agentFromProto(res.agent) };
}

async function updateAgent(agent: Agent): Promise<{ agent: Agent }> {
  const res = await client.updateAgent({ agent: agentToProto(agent) });
  if (!res.agent) throw new Error("update returned nothing");
  return { agent: agentFromProto(res.agent) };
}

async function deleteAgent(name: string): Promise<void> {
  await client.deleteAgent({ name });
}

interface InvokeAgentParams {
  agent_name: string;
  input: string;
  app_name?: string;
  user_id?: string;
  session_id?: string;
  model_override?: string;
}

async function invokeAgent(params: InvokeAgentParams): Promise<{ session_id: string; response: string }> {
  const res = await client.invokeAgent({
    agentName: params.agent_name,
    input: params.input,
    appName: params.app_name ?? "",
    userId: params.user_id ?? "",
    sessionId: params.session_id ?? "",
    modelOverride: params.model_override ?? "",
  });
  return { session_id: res.sessionId, response: res.response };
}

export async function cancelAgentInvocation(invocationId: string): Promise<{ cancelled: boolean }> {
  const res = await client.cancelAgentInvocation({ invocationId });
  return { cancelled: res.cancelled };
}

async function reloadAgents(): Promise<{ reloaded_at?: string }> {
  const res = await client.reloadAgents({});
  return { reloaded_at: tsToISO(res.reloadedAt) };
}

function runtimeStatusFromProto(s: Parameters<typeof toJson<typeof AgentRuntimeStatusSchema>>[1]): AgentRuntimeStatus {
  return toJson(AgentRuntimeStatusSchema, s, { useProtoFieldName: true }) as unknown as AgentRuntimeStatus;
}

async function getAgentRuntimeStatus(name: string): Promise<{ status: AgentRuntimeStatus }> {
  const res = await client.getAgentRuntimeStatus({ name });
  if (!res.status) throw new Error("status not found");
  return { status: runtimeStatusFromProto(res.status) };
}

async function listAgentRuntimeStatuses(names?: string[]): Promise<{ statuses?: AgentRuntimeStatus[] }> {
  const res = await client.listAgentRuntimeStatuses({ names: names ?? [] });
  return { statuses: res.statuses.map(runtimeStatusFromProto) };
}

interface ListInvocationsParams {
  agent_name?: string;
  session_id?: string;
  page_size?: number;
  page_token?: string;
}

interface ListInvocationsResponse {
  invocations?: Invocation[];
  next_page_token?: string;
  total?: number;
}

async function listAgentInvocations(params: ListInvocationsParams): Promise<ListInvocationsResponse> {
  const res = await client.listAgentInvocations({
    agentName: params.agent_name ?? "",
    sessionId: params.session_id ?? "",
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  // Invocation also nests Timestamps; round-trip via toJson with proto names
  // so the legacy snake_case interface matches without manual mapping.
  return {
    invocations: res.invocations.map(
      (inv) => toJson(InvocationSchema, inv, { useProtoFieldName: true }) as unknown as Invocation,
    ),
    next_page_token: res.nextPageToken,
    total: res.total,
  };
}

export function useAgents(params: ListAgentsParams = {}) {
  return useQuery({
    queryKey: ["agents", params],
    queryFn: () => listAgents(params),
  });
}

export function useAgent(name: string) {
  return useQuery({ queryKey: ["agents", name], queryFn: () => getAgent(name), enabled: !!name });
}

export function useCreateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}

export function useUpdateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agents", agent.name] });
    },
  });
}

export function useDeleteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}

export function useInvokeAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: invokeAgent,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["agent-invocations"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useCancelAgentInvocation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: cancelAgentInvocation,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agent-invocations"] }),
  });
}

export function useReloadAgents() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: reloadAgents,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}

export function useAgentRuntimeStatus(name: string) {
  return useQuery({
    queryKey: ["agents", name, "runtime-status"],
    queryFn: () => getAgentRuntimeStatus(name),
    enabled: !!name,
    refetchInterval: 15_000,
  });
}

export function useAgentRuntimeStatuses(names?: string[]) {
  return useQuery({
    queryKey: ["agent-runtime-statuses", names],
    queryFn: () => listAgentRuntimeStatuses(names),
    refetchInterval: 15_000,
  });
}

export function useAgentInvocations(params: ListInvocationsParams = {}) {
  return useQuery({
    queryKey: ["agent-invocations", params],
    queryFn: () => listAgentInvocations(params),
  });
}
