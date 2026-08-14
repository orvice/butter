import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { create, fromJson, toJson, type JsonValue } from "@bufbuild/protobuf";
import { AgentSchema } from "@/gen/agents/v1/agent_pb";
import { AgentOperationStatus, type AgentOperation } from "@/gen/agents/v1/agent_operation_pb";
import {
  ContentFileActionSchema,
  ContentFileOperation,
  type ContentFileAction,
} from "@/gen/agents/v1/repobinding_pb";
import {
  AgentRuntimeStatusSchema,
  AgentService,
  InvocationSchema,
} from "@/gen/agents/v1/agent_service_pb";
import type { Agent, AgentRuntimeStatus, Invocation } from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(AgentService);

function stripUndefined(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(stripUndefined);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .filter(([, v]) => v !== undefined)
        .map(([k, v]) => [k, stripUndefined(v)]),
    );
  }
  return value;
}

// Agent / AgentConfig is deeply nested (sub_agents, mcp_servers, file_mounts,
// context_guard, ...). Rather than hand-rolling a 200-line toProto/fromProto,
// we leverage protojson: the proto-es runtime's fromJson accepts both
// snake_case and camelCase keys, and toJson with useProtoFieldName=true emits
// snake_case identical to the compatibility wire format. So the legacy
// snake-cased Agent interface round-trips through the typed RPC call without
// extra mapping code.
function agentToProto(a: Agent) {
  return fromJson(AgentSchema, stripUndefined(a) as JsonValue, { ignoreUnknownFields: true });
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

interface AgentRef {
  agent_id?: string;
  operation_id?: string;
}

// agentRefKey is the stable cache-key segment for an agent: its immutable
// agent_id.
export function agentRefKey(ref: AgentRef): string {
  return ref.agent_id ?? "";
}

async function getAgent(agentId: string): Promise<{ agent: Agent }> {
  const res = await client.getAgent({ agentId });
  if (!res.agent) throw new Error("not found");
  return { agent: agentFromProto(res.agent) };
}

export interface CreateAgentParams {
  agent: Agent;
  initial_content: {
    description?: string;
    prompt?: string;
    global_prompt?: string;
  };
  operation_id: string;
}

async function createAgent(params: CreateAgentParams): Promise<{ agent: Agent }> {
  const res = await client.createAgent({
    agent: agentToProto(params.agent),
    initialContent: {
      description: params.initial_content.description ?? "",
      prompt: params.initial_content.prompt ?? "",
      globalPrompt: params.initial_content.global_prompt ?? "",
    },
    operationId: params.operation_id,
  });
  if (!res.agent) throw new Error("create returned nothing");
  return { agent: agentFromProto(res.agent) };
}

export interface UpdateAgentParams {
  agent: Agent;
  previous_agent: Agent;
  repository_bound: boolean;
  base_commit_sha?: string;
  operation_id: string;
}

function contentAction(path: string, content: string): ContentFileAction {
  return create(ContentFileActionSchema, {
    path,
    operation: content === "" ? ContentFileOperation.DELETE : ContentFileOperation.PUT,
    content,
  });
}

function agentContentChanges(previous: Agent, next: Agent): ContentFileAction[] {
  const agentID = next.agent_id ?? previous.agent_id;
  if (!agentID) return [];

  const fields = [
    [`agents/${agentID}/description.md`, previous.description ?? "", next.description ?? ""],
    [`agents/${agentID}/prompt.md`, previous.config?.instruction ?? "", next.config?.instruction ?? ""],
    [
      `agents/${agentID}/global-prompt.md`,
      previous.config?.global_instruction ?? "",
      next.config?.global_instruction ?? "",
    ],
  ] as const;
  return fields.filter(([, before, after]) => before !== after).map(([path, , content]) => contentAction(path, content));
}

async function updateAgent(params: UpdateAgentParams): Promise<{ agent: Agent }> {
  if (!params.repository_bound) {
    const res = await client.updateAgent({ agent: agentToProto(params.agent) });
    if (!res.agent) throw new Error("update returned nothing");
    return { agent: agentFromProto(res.agent) };
  }

  const res = await client.updateAgentConfiguration({
    agentPatch: agentToProto(params.agent),
    contentChanges: agentContentChanges(params.previous_agent, params.agent),
    expectedAgentVersion: BigInt(params.previous_agent.version ?? 0),
    baseCommitSha: params.base_commit_sha ?? "",
    operationId: params.operation_id,
  });
  if (res.validationErrors.length > 0) {
    throw new Error(res.validationErrors.join("; "));
  }
  if (!res.agent) throw new Error("update returned nothing");
  return { agent: agentFromProto(res.agent) };
}

async function deleteAgent(ref: AgentRef): Promise<void> {
  await client.deleteAgent({
    agentId: ref.agent_id ?? "",
    operationId: ref.operation_id ?? "",
  });
}

async function restoreAgent(agentId: string): Promise<{ agent: Agent }> {
  const res = await client.restoreAgent({
    agentId,
    operationId: crypto.randomUUID(),
  });
  if (!res.agent) throw new Error("restore returned nothing");
  return { agent: agentFromProto(res.agent) };
}

export interface ListAgentOperationsParams {
  status?: AgentOperationStatus;
  page_size?: number;
  page_token?: string;
}

export interface ListAgentOperationsResponse {
  operations: AgentOperation[];
  next_page_token?: string;
}

async function listAgentOperations(
  params: ListAgentOperationsParams = {},
): Promise<ListAgentOperationsResponse> {
  const res = await client.listAgentOperations({
    status: params.status ?? AgentOperationStatus.UNSPECIFIED,
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return { operations: res.operations, next_page_token: res.nextPageToken };
}

async function getAgentOperation(operationId: string): Promise<AgentOperation> {
  const res = await client.getAgentOperation({ operationId });
  if (!res.operation) throw new Error("operation not found");
  return res.operation;
}

async function retryAgentOperation(operationId: string): Promise<AgentOperation> {
  const res = await client.retryAgentOperation({ operationId });
  if (!res.operation) throw new Error("retry returned no operation");
  return res.operation;
}

interface InvokeAgentParams {
  /** Immutable agent_id of the agent to invoke. */
  agent_id: string;
  input: string;
  app_name?: string;
  user_id?: string;
  session_id?: string;
  model_override?: string;
}

async function invokeAgent(params: InvokeAgentParams): Promise<{ session_id: string; response: string }> {
  const res = await client.invokeAgent({
    agentId: params.agent_id,
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

async function getAgentRuntimeStatus(ref: AgentRef): Promise<{ status: AgentRuntimeStatus }> {
  const res = await client.getAgentRuntimeStatus({
    agentId: ref.agent_id ?? "",
  });
  if (!res.status) throw new Error("status not found");
  return { status: runtimeStatusFromProto(res.status) };
}

interface ListRuntimeStatusesParams {
  /** Filter by immutable agent_ids (empty = all). */
  agent_ids?: string[];
}

async function listAgentRuntimeStatuses(
  params: ListRuntimeStatusesParams = {},
): Promise<{ statuses?: AgentRuntimeStatus[] }> {
  const res = await client.listAgentRuntimeStatuses({
    agentIds: params.agent_ids ?? [],
  });
  return { statuses: res.statuses.map(runtimeStatusFromProto) };
}

interface ListInvocationsParams {
  /** Legacy agent-name filter; fallback for agents without an agent_id. */
  agent_name?: string;
  /** Filter by immutable agent_id. Preferred over agent_name. */
  agent_id?: string;
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
    agentName: params.agent_id ? "" : (params.agent_name ?? ""),
    agentId: params.agent_id ?? "",
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

export function useAgents(params: ListAgentsParams = {}, options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: ["agents", params],
    queryFn: () => listAgents(params),
    enabled: options.enabled,
  });
}

// useAgent looks up an agent by its immutable agent_id.
export function useAgent(agentId: string) {
  return useQuery({ queryKey: ["agents", agentId], queryFn: () => getAgent(agentId), enabled: !!agentId });
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
    onSuccess: (_data, params) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agents", agentRefKey(params.agent)] });
      qc.invalidateQueries({ queryKey: ["agent-operations"] });
    },
  });
}

export function useDeleteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgent,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agent-operations"] });
    },
  });
}

export function useRestoreAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: restoreAgent,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agent-operations"] });
    },
  });
}

export function useAgentOperations(params: ListAgentOperationsParams = {}) {
  return useQuery({
    queryKey: ["agent-operations", params],
    queryFn: () => listAgentOperations(params),
    refetchInterval: (query) => {
      const operations = query.state.data?.operations ?? [];
      return operations.some(
        (op) => op.status === AgentOperationStatus.PENDING || op.status === AgentOperationStatus.RUNNING,
      )
        ? 5_000
        : 30_000;
    },
  });
}

export function useAgentOperation(operationId: string) {
  return useQuery({
    queryKey: ["agent-operations", operationId],
    queryFn: () => getAgentOperation(operationId),
    enabled: !!operationId,
  });
}

export function useRetryAgentOperation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: retryAgentOperation,
    onSuccess: (operation) => {
      qc.setQueryData(["agent-operations", operation.id], operation);
      qc.invalidateQueries({ queryKey: ["agent-operations"] });
      qc.invalidateQueries({ queryKey: ["agents"] });
    },
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

export function useAgentRuntimeStatus(ref: AgentRef) {
  return useQuery({
    queryKey: ["agents", agentRefKey(ref), "runtime-status"],
    queryFn: () => getAgentRuntimeStatus(ref),
    enabled: !!agentRefKey(ref),
    refetchInterval: 15_000,
  });
}

export function useAgentRuntimeStatuses(params: ListRuntimeStatusesParams = {}) {
  return useQuery({
    queryKey: ["agent-runtime-statuses", params],
    queryFn: () => listAgentRuntimeStatuses(params),
    refetchInterval: 15_000,
  });
}

export function useAgentInvocations(params: ListInvocationsParams = {}) {
  return useQuery({
    queryKey: ["agent-invocations", params],
    queryFn: () => listAgentInvocations(params),
  });
}
