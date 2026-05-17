import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { Agent, AgentRuntimeStatus, Invocation } from "@/types/api";

const SVC = "agents.v1.AgentService";

interface ListAgentsParams {
  page_size?: number;
  page_token?: string;
}

interface ListAgentsResponse {
  agents: Agent[];
  next_page_token?: string;
  total?: number;
}

function listAgents(params: ListAgentsParams = {}) {
  return twirpFetch<ListAgentsParams, ListAgentsResponse>(SVC, "ListAgents", params);
}

function getAgent(name: string) {
  return twirpFetch<{ name: string }, { agent: Agent }>(SVC, "GetAgent", { name });
}

function createAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "CreateAgent", { agent });
}

function updateAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "UpdateAgent", { agent });
}

function deleteAgent(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteAgent", { name });
}

interface InvokeAgentParams {
  agent_name: string;
  input: string;
  app_name?: string;
  user_id?: string;
  session_id?: string;
  model_override?: string;
}

function invokeAgent(params: InvokeAgentParams) {
  return twirpFetch<InvokeAgentParams, { session_id: string; response: string }>(
    SVC,
    "InvokeAgent",
    params,
  );
}

function cancelAgentInvocation(invocationId: string) {
  return twirpFetch<{ invocation_id: string }, { cancelled: boolean }>(
    SVC,
    "CancelAgentInvocation",
    { invocation_id: invocationId },
  );
}

function reloadAgents() {
  return twirpFetch<object, { reloaded_at?: string }>(SVC, "ReloadAgents", {});
}

function getAgentRuntimeStatus(name: string) {
  return twirpFetch<{ name: string }, { status: AgentRuntimeStatus }>(
    SVC,
    "GetAgentRuntimeStatus",
    { name },
  );
}

function listAgentRuntimeStatuses(names?: string[]) {
  return twirpFetch<
    { names?: string[] },
    { statuses?: AgentRuntimeStatus[] }
  >(SVC, "ListAgentRuntimeStatuses", { names });
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

function listAgentInvocations(params: ListInvocationsParams) {
  return twirpFetch<ListInvocationsParams, ListInvocationsResponse>(
    SVC,
    "ListAgentInvocations",
    params,
  );
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
