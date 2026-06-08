import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  RemoteAgentProtocol,
  RemoteAgentSchema,
  type RemoteAgent as PbRemoteAgent,
} from "@/gen/agents/v1/agent_pb";
import { RemoteAgentService } from "@/gen/agents/v1/agent_service_pb";
import {
  RemoteAgentStatus_State,
  type RemoteAgentStatus as PbRemoteAgentStatus,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  RemoteAgent,
  RemoteAgentProtocol as LegacyProtocol,
  RemoteAgentState,
  RemoteAgentStatus,
} from "@/types/api";
import { bigintToNumber, tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(RemoteAgentService);

const PROTOCOL_NAMES: LegacyProtocol[] = [
  "REMOTE_AGENT_PROTOCOL_UNSPECIFIED",
  "REMOTE_AGENT_PROTOCOL_A2A",
  "REMOTE_AGENT_PROTOCOL_DAEMON",
];

const STATE_NAMES: RemoteAgentState[] = [
  "STATE_UNSPECIFIED",
  "STATE_CONFIGURED",
  "STATE_ACTIVE",
  "STATE_IDLE",
  "STATE_UNREACHABLE",
  "STATE_ERROR",
];

function protocolFromProto(p: RemoteAgentProtocol): LegacyProtocol {
  return PROTOCOL_NAMES[p] ?? "REMOTE_AGENT_PROTOCOL_UNSPECIFIED";
}

function protocolToProto(p: LegacyProtocol | undefined): RemoteAgentProtocol {
  const idx = PROTOCOL_NAMES.indexOf(p ?? "REMOTE_AGENT_PROTOCOL_UNSPECIFIED");
  return idx < 0 ? RemoteAgentProtocol.UNSPECIFIED : (idx as RemoteAgentProtocol);
}

function stateFromProto(s: RemoteAgentStatus_State): RemoteAgentState {
  return STATE_NAMES[s] ?? "STATE_UNSPECIFIED";
}

function fromProto(a: PbRemoteAgent): RemoteAgent {
  return {
    id: a.id,
    name: a.name,
    url: a.url,
    protocol: protocolFromProto(a.protocol),
    daemon_runtime_id: a.daemonRuntimeId,
    acp_runtime: a.acpRuntime,
  };
}

function toProto(a: RemoteAgent): PbRemoteAgent {
  return create(RemoteAgentSchema, {
    id: a.id,
    name: a.name,
    url: a.url,
    protocol: protocolToProto(a.protocol),
    daemonRuntimeId: a.daemon_runtime_id ?? "",
    acpRuntime: a.acp_runtime ?? "",
  });
}

function statusFromProto(s: PbRemoteAgentStatus): RemoteAgentStatus {
  return {
    id: s.id,
    protocol: protocolFromProto(s.protocol),
    state: stateFromProto(s.state),
    detail: s.detail,
    serving_daemon_runtime_id: s.servingDaemonRuntimeId,
    checked_at: tsToISO(s.checkedAt),
    latency_ms: bigintToNumber(s.latencyMs),
  };
}

async function listRemoteAgents(): Promise<{ remote_agents: RemoteAgent[] }> {
  const res = await client.listRemoteAgents({});
  return { remote_agents: res.remoteAgents.map(fromProto) };
}

async function getRemoteAgent(id: string): Promise<{ remote_agent: RemoteAgent }> {
  const res = await client.getRemoteAgent({ id });
  if (!res.remoteAgent) throw new Error("not found");
  return { remote_agent: fromProto(res.remoteAgent) };
}

async function createRemoteAgent(a: RemoteAgent): Promise<{ remote_agent: RemoteAgent }> {
  const res = await client.createRemoteAgent({ remoteAgent: toProto(a) });
  if (!res.remoteAgent) throw new Error("create returned nothing");
  return { remote_agent: fromProto(res.remoteAgent) };
}

async function updateRemoteAgent(a: RemoteAgent): Promise<{ remote_agent: RemoteAgent }> {
  const res = await client.updateRemoteAgent({ remoteAgent: toProto(a) });
  if (!res.remoteAgent) throw new Error("update returned nothing");
  return { remote_agent: fromProto(res.remoteAgent) };
}

async function deleteRemoteAgent(id: string): Promise<void> {
  await client.deleteRemoteAgent({ id });
}

async function getRemoteAgentStatus(id: string): Promise<{ status: RemoteAgentStatus }> {
  const res = await client.getRemoteAgentStatus({ id });
  if (!res.status) throw new Error("status not found");
  return { status: statusFromProto(res.status) };
}

export function useRemoteAgents() {
  return useQuery({ queryKey: ["remote-agents"], queryFn: listRemoteAgents });
}

export function useRemoteAgent(id: string) {
  return useQuery({ queryKey: ["remote-agents", id], queryFn: () => getRemoteAgent(id), enabled: !!id });
}

export function useCreateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}

export function useUpdateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateRemoteAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["remote-agents"] });
      qc.invalidateQueries({ queryKey: ["remote-agents", agent.id] });
    },
  });
}

export function useDeleteRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}

export function useRemoteAgentStatus(id: string) {
  return useQuery({
    queryKey: ["remote-agents", id, "status"],
    queryFn: () => getRemoteAgentStatus(id),
    enabled: !!id,
    refetchInterval: 30_000,
  });
}
